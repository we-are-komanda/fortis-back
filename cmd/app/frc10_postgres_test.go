package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	ba "github.com/fortis/backend/internal/modules/budget/application"
	bi "github.com/fortis/backend/internal/modules/budget/infrastructure"
	pa "github.com/fortis/backend/internal/modules/defense_project/application"
	pi "github.com/fortis/backend/internal/modules/defense_project/infrastructure"
	ea "github.com/fortis/backend/internal/modules/enterprise/application"
	ei "github.com/fortis/backend/internal/modules/enterprise/infrastructure"
	ra "github.com/fortis/backend/internal/modules/report/application"
	rd "github.com/fortis/backend/internal/modules/report/domain"
	ri "github.com/fortis/backend/internal/modules/report/infrastructure"
	ru "github.com/fortis/backend/internal/modules/report/ui"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"gorm.io/gorm"
)

func frc10Query(id string, version int, hidden bool, view string) string {
	return fmt.Sprintf("/projects/report?id=%s&projectVersion=%d&hideCost=%t&mapView=%s", id, version, hidden, view)
}
func frc10Count(t *testing.T, db *gorm.DB, table string) int64 {
	t.Helper()
	var n int64
	require.NoError(t, db.Table(table).Count(&n).Error)
	return n
}
func TestFRC10PostgresImmutableReport(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a := f.people[0]
	id := frc04Request(t, f, a, "POST", "/projects", "", map[string]any{"name": "Issued frozen variant", "enterpriseId": a.e, "projectJson": frc05Document(t, a.e)}, 200)["projectId"].(string)
	before := frc10Count(t, db, "audit_events")
	query := frc10Query(id, 1, false, "all")
	got := frc04Request(t, f, a, "GET", query, "", nil, 200)
	reportID := got["reportId"].(string)
	require.NotEmpty(t, reportID)
	require.Equal(t, "report-v1", got["reportSchemaVersion"])
	require.Equal(t, "Issued frozen variant", got["title"])
	require.Nil(t, got["enterprise"].(map[string]any)["name"])
	cost := got["cost"].(map[string]any)
	budget := got["budget"].(map[string]any)
	require.Equal(t, "47500000", cost["totalMinor"])
	require.Equal(t, got["identity"], cost["identity"])
	require.Equal(t, got["identity"], budget["identity"])
	require.Len(t, got["objects"], 2)
	require.Contains(t, js(got["warnings"]), "preliminary_report")
	require.Equal(t, before+1, frc10Count(t, db, "audit_events"))
	require.Equal(t, got, frc04Request(t, f, a, "GET", query, "", nil, 200))
	require.Equal(t, got, frc04Request(t, f, a, "GET", "/projects/report?reportId="+reportID, "", nil, 200))
	require.Equal(t, before+1, frc10Count(t, db, "audit_events"))
	frc04Request(t, f, a, "PUT", "/projects/update?id="+id, "", map[string]any{"version": 1, "name": "Later variant"}, 200)
	require.NoError(t, db.Table("defense_assets").Where("id=?", a.a).Update("asset_data", `{"name":"Later catalogue"}`).Error)
	require.Equal(t, got, frc04Request(t, f, a, "GET", query, "", nil, 200))
	require.Equal(t, got, frc04Request(t, f, a, "GET", "/projects/report?reportId="+reportID, "", nil, 200))
	newReport := frc04Request(t, f, a, "GET", frc10Query(id, 2, false, "all"), "", nil, 200)
	require.NotEqual(t, reportID, newReport["reportId"])
	require.Equal(t, "Later variant", newReport["title"])
	require.Error(t, db.Table("issued_reports").Where("id=?", reportID).Update("payload_json", `{}`).Error)
	frc04Request(t, f, a, "DELETE", "/projects/delete?id="+id, "", nil, 200)
	frc04Request(t, f, a, "GET", "/projects/report?reportId="+reportID, "", nil, 404)
	require.Zero(t, frc10Count(t, db, "issued_reports"))
}
func TestFRC10PostgresUnpricedSkipsFinanceAndAllowlist(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a := f.people[0]
	var doc map[string]any
	require.NoError(t, json.Unmarshal([]byte(frc07Document(t, a.e, nil)), &doc))
	asset := doc["assetLibrary"].([]any)[0].(map[string]any)
	asset["iconUrl"] = "PRIVATE_FINANCIAL_SENTINEL"
	asset["documents"] = []any{"PRIVATE_FINANCIAL_SENTINEL"}
	asset["provenance"] = map[string]any{"sourceUrl": "PRIVATE_FINANCIAL_SENTINEL", "sourceLabel": "SECRET FINANCIAL SOURCE", "quality": "demo"}
	asset["weaponSpec"] = map[string]any{"caliber": "Frozen caliber", "price": "PRIVATE_FINANCIAL_SENTINEL"}
	layer := doc["layers"].([]any)[0].(map[string]any)
	layer["isVisible"] = false
	layer["geometry"] = map[string]any{"type": "polygon", "coordinates": []any{map[string]any{"lat": 0, "lng": 0}, map[string]any{"lat": 1, "lng": 0}, map[string]any{"lat": 0, "lng": 1}}, "isClosed": true}
	layer["geometryType"] = "polygon"
	// Test-only insertion of a future binding: unpriced must never invoke its calculator.
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register("frc10_future", func(tx *gorm.DB) {
		if tx.Statement.Table == "project_revisions" {
			tx.Statement.SetColumn("CostCalculationVersion", "future-unavailable")
		}
	}))
	id := frc04Request(t, f, a, "POST", "/projects", "", map[string]any{"name": "Hidden", "enterpriseId": a.e, "projectJson": js(doc)}, 200)["projectId"].(string)
	require.NoError(t, db.Callback().Create().Remove("frc10_future"))
	before := frc10Count(t, db, "project_cost_projections")
	hidden := frc04Request(t, f, a, "GET", frc10Query(id, 1, true, "all"), "", nil, 200)
	for _, key := range []string{"cost", "budget", "sources", "structuralProfile"} {
		require.NotContains(t, hidden, key)
	}
	for _, text := range []string{"PRIVATE_FINANCIAL_SENTINEL", "SECRET FINANCIAL SOURCE", "price_unknown", "sourceUrl", "unitPrice", "documentId", "pricePerUnit", "coveredObjCount"} {
		require.NotContains(t, js(hidden), text)
	}
	require.Contains(t, js(hidden), "Frozen caliber")
	require.Contains(t, js(hidden), "coordinates")
	require.Contains(t, js(hidden), "isClosed")
	require.Equal(t, before, frc10Count(t, db, "project_cost_projections"))
	selective := frc04Request(t, f, a, "GET", frc10Query(id, 1, true, "saved"), "", nil, 200)
	require.NotEqual(t, hidden["reportId"], selective["reportId"])
	require.Equal(t, hidden["objects"], selective["objects"])
	require.Empty(t, selective["mapVisibility"].(map[string]any)["objectIds"])
	failure := frc04Request(t, f, a, "GET", frc10Query(id, 1, false, "all"), "", nil, 503)
	require.Equal(t, "unsupported_calculation_version", failure["error"].(map[string]any)["code"])
	require.Equal(t, int64(2), frc10Count(t, db, "issued_reports"))
}
func TestFRC10PostgresAccessAndIssuanceRollback(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a, b := f.people[0], f.people[1]
	query := frc10Query(a.p, 1, true, "all")
	before := frc10Count(t, db, "audit_events")
	for _, table := range []string{"issued_reports", "audit_events"} {
		name := "frc10_fail_" + table
		require.NoError(t, db.Callback().Create().Before("gorm:create").Register(name, func(tx *gorm.DB) {
			if tx.Statement.Table == table {
				tx.AddError(errors.New("synthetic report failure"))
			}
		}))
		frc04Request(t, f, a, "GET", query, "", nil, 500)
		require.NoError(t, db.Callback().Create().Remove(name))
		require.Zero(t, frc10Count(t, db, "issued_reports"))
		require.Equal(t, before, frc10Count(t, db, "audit_events"))
	}
	issued := frc04Request(t, f, a, "GET", query, "", nil, 200)
	reportID := issued["reportId"].(string)
	for _, path := range []string{query, "/projects/report?reportId=" + reportID} {
		shared := frc04Request(t, f, b, "GET", path, "", nil, 200)
		require.Equal(t, reportID, shared["reportId"])
	}
	for _, path := range []string{query + "&reportId=" + reportID, query + "&hideCost=false", "/projects/report?reportId=" + reportID + "&mapView=all", "/projects/report?id=" + a.p + "&projectVersion=1&hideCost=1", "/projects/report?id=" + a.p + "&hideCost=true"} {
		frc04Request(t, f, a, "GET", path, "", nil, 400)
	}
	frc04Request(t, f, a, "GET", frc10Query(a.p, 999, true, "all"), "", nil, 404)
	require.NoError(t, ei.NewEnterpriseRepository(db).RemoveUserFromEnterprise(context.Background(), a.u, a.e))
	for _, path := range []string{query, "/projects/report?reportId=" + reportID} {
		frc04Request(t, f, a, "GET", path, "", nil, 200)
	}
	require.Equal(t, int64(1), frc10Count(t, db, "issued_reports"))
}
func TestFRC10PostgresRevokeBetweenRenderAndIssue(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a := f.people[0]
	ctx := context.Background()
	enterprises := ei.NewEnterpriseRepository(db)
	projects := pa.NewDefenseProjectService(pi.NewDefenseProjectRepository(db), ea.NewEnterpriseService(enterprises))
	budgets := bi.NewBudgetConfigRepository(db)
	render := func(d *rd.ReportDraft) (string, error) {
		payload, err := ru.RenderReport(d)
		require.NoError(t, err)
		require.NoError(t, enterprises.RemoveUserFromEnterprise(ctx, a.u, a.e))
		return payload, nil
	}
	service := ra.NewReportService(projects, ba.NewBudgetService(budgets, projects, budgets), ri.NewReportRepository(db), render)
	key, err := rd.NewReportKey(a.p, 1, true, "all")
	require.NoError(t, err)
	before := frc10Count(t, db, "audit_events")
	_, err = service.GetReport(ctx, a.u, ra.GetReportQuery{Key: key})
	require.NoError(t, err)
	require.EqualValues(t, 1, frc10Count(t, db, "issued_reports"))
	require.Equal(t, before+1, frc10Count(t, db, "audit_events"))
}
func TestFRC10PostgresConcurrentIssuance(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a := f.people[0]
	query := frc10Query(a.p, 1, false, "all")
	before := frc10Count(t, db, "audit_events")
	type response struct {
		status int
		body   string
	}
	out := make(chan response, 8)
	start := make(chan struct{})
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			var x fasthttp.RequestCtx
			x.Init2(nil, nil, false)
			x.Request.SetRequestURI("/api/v1" + query)
			x.Request.Header.SetMethod("GET")
			x.Request.Header.Set("Authorization", "Bearer "+a.token)
			f.h(&x)
			out <- response{x.Response.StatusCode(), string(x.Response.Body())}
		}()
	}
	close(start)
	group.Wait()
	close(out)
	var winner map[string]any
	for r := range out {
		require.Equal(t, 200, r.status, r.body)
		var got map[string]any
		require.NoError(t, json.Unmarshal([]byte(r.body), &got))
		if winner == nil {
			winner = got
		} else {
			require.Equal(t, winner, got)
		}
	}
	require.Equal(t, int64(1), frc10Count(t, db, "issued_reports"))
	require.Equal(t, before+1, frc10Count(t, db, "audit_events"))
	require.Equal(t, winner, frc04Request(t, f, a, "GET", "/projects/report?reportId="+winner["reportId"].(string), "", nil, 200))
	// Stored payload reuse must not read the project revision or financial cache.
	require.NoError(t, db.Callback().Query().Before("gorm:query").Register("frc10_reopen_no_calculation", func(tx *gorm.DB) {
		if strings.Contains(tx.Statement.Table, "cost_projection") || tx.Statement.Table == "project_revisions" {
			tx.AddError(errors.New("reopen recalculated"))
		}
	}))
	require.Equal(t, winner, frc04Request(t, f, a, "GET", "/projects/report?reportId="+winner["reportId"].(string), "", nil, 200))
	require.NoError(t, db.Callback().Query().Remove("frc10_reopen_no_calculation"))
}

func TestFRC10PostgresMembershipDoesNotRestrictReports(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a := f.people[0]
	issued := frc04Request(t, f, person{}, "GET", frc10Query(a.p, 1, true, "all"), "", nil, 200)
	require.NoError(t, ei.NewEnterpriseRepository(db).RemoveUserFromEnterprise(context.Background(), a.u, a.e))
	frc04Request(t, f, person{}, "GET", "/projects/report?reportId="+issued["reportId"].(string), "", nil, 200)
	require.Equal(t, int64(1), frc10Count(t, db, "issued_reports"))
}
