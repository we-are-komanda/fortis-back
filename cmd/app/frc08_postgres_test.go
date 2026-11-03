package main

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"testing"
)

func frc08Query(a string, av int, b string, bv int) string {
	return fmt.Sprintf("/projects/compare?projectIdA=%s&projectVersionA=%d&projectIdB=%s&projectVersionB=%d", a, av, b, bv)
}
func TestFRC08PostgresPinnedComparison(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a := f.people[0]
	create := func(name, raw string) string {
		return frc04Request(t, f, a, "POST", "/projects", "", map[string]any{"name": name, "enterpriseId": a.e, "projectJson": raw}, 200)["projectId"].(string)
	}
	idA := create("Frozen A", frc05Document(t, a.e))
	var b map[string]any
	require.NoError(t, json.Unmarshal([]byte(frc05Document(t, a.e)), &b))
	objects := b["placedObjects"].([]any)
	objects[0].(map[string]any)["quantity"] = 3
	objects[1].(map[string]any)["quantity"] = 2
	idB := create("Frozen B", js(b))
	// Reach real persisted A7/B3 without fabricating historical rows.
	frc04Request(t, f, a, "PUT", "/projects/budget?projectId="+idA, "", map[string]any{"version": 1, "budgetMode": "limited", "budgetLimitMinor": "50000000"}, 200)
	for v := 2; v < 7; v++ {
		frc04Request(t, f, a, "PUT", "/projects/update?id="+idA, "", map[string]any{"version": v, "name": "Frozen A"}, 200)
	}
	for v := 1; v < 3; v++ {
		frc04Request(t, f, a, "PUT", "/projects/update?id="+idB, "", map[string]any{"version": v, "name": "Frozen B"}, 200)
	}
	query := frc08Query(idA, 7, idB, 3)
	reads := map[string]int{}
	require.NoError(t, db.Callback().Query().Before("gorm:query").Register("frc08_read_count", func(tx *gorm.DB) { reads[tx.Statement.Table]++ }))
	got := frc04Request(t, f, a, "GET", query, "", nil, 200)
	require.NoError(t, db.Callback().Query().Remove("frc08_read_count"))
	require.Equal(t, 2, reads["defense_projects"])
	require.Equal(t, 2, reads["project_revisions"])
	require.Zero(t, reads["defense_assets"])
	require.Zero(t, reads["budget_configs"])
	require.Equal(t, "7500000", got["differenceMinor"])
	require.Equal(t, "15.79", got["differencePercent"])
	sideA, sideB := got["a"].(map[string]any), got["b"].(map[string]any)
	require.Equal(t, "Frozen A", sideA["name"])
	require.Equal(t, "Frozen B", sideB["name"])
	require.NotEmpty(t, sideA["updatedAt"])
	require.Equal(t, "47500000", sideA["cost"].(map[string]any)["totalMinor"])
	require.Equal(t, "55000000", sideB["cost"].(map[string]any)["totalMinor"])
	require.Equal(t, "2500000", sideA["budget"].(map[string]any)["remainingMinor"])
	for _, side := range []map[string]any{sideA, sideB} {
		require.Equal(t, side["identity"], side["cost"].(map[string]any)["identity"])
		require.Equal(t, side["identity"], side["budget"].(map[string]any)["identity"])
	}
	require.NotContains(t, js(got), "coveredObjCount")
	require.NotContains(t, js(got), "costDeltaMln")
	frc04Request(t, f, a, "PUT", "/projects/update?id="+idA, "", map[string]any{"version": 7, "name": "Later A"}, 200)
	require.NoError(t, db.Table("defense_assets").Where("id = ?", a.a).Update("asset_data", `{"name":"Later catalogue"}`).Error)
	require.Equal(t, got, frc04Request(t, f, a, "GET", query, "", nil, 200))
	frc04Request(t, f, a, "GET", frc08Query(idA, 7, idA, 8), "", nil, 200)
	frc04Request(t, f, a, "GET", frc08Query(idA, 7, idA, 7), "", nil, 400)
}

func TestFRC08PostgresUnknownZeroAndAccess(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a, b := f.people[0], f.people[1]
	create := func(name string, price any) string {
		return frc04Request(t, f, a, "POST", "/projects", "", map[string]any{"name": name, "enterpriseId": a.e, "projectJson": frc07Document(t, a.e, price)}, 200)["projectId"].(string)
	}
	zero, known, unknown := create("Zero", "0"), create("Known", "1"), create("Unknown", nil)
	got := frc04Request(t, f, a, "GET", frc08Query(zero, 1, known, 1), "", nil, 200)
	require.Equal(t, "1", got["differenceMinor"])
	require.Nil(t, got["differencePercent"])
	incomplete := frc04Request(t, f, a, "GET", frc08Query(known, 1, unknown, 1), "", nil, 200)
	require.Nil(t, incomplete["differenceMinor"])
	require.Nil(t, incomplete["differencePercent"])
	require.Contains(t, js(incomplete["warnings"]), "incomplete_cost")
	require.Nil(t, incomplete["b"].(map[string]any)["budget"].(map[string]any)["spentMinor"])
	count := func() int64 {
		var n int64
		require.NoError(t, db.Table("project_cost_projections").Count(&n).Error)
		return n
	}
	fresh := create("No partial cache", "2")
	before := count()
	for _, pair := range [][2]string{{fresh, b.p}, {b.p, fresh}} {
		failure := frc04Request(t, f, a, "GET", frc08Query(pair[0], 1, pair[1], 1), "", nil, 400)
		require.NotContains(t, js(failure), "totalMinor")
		require.NotContains(t, js(failure), "\"a\":")
		require.Equal(t, before, count())
	}
	require.NoError(t, db.Table("user_enterprises").Create(map[string]any{"user_id": a.u, "enterprise_id": b.e}).Error)
	incompatible := frc04Request(t, f, a, "GET", frc08Query(fresh, 1, b.p, 1), "", nil, 400)
	require.Equal(t, "incompatible_projects", incompatible["error"].(map[string]any)["code"])
	require.Equal(t, before, count())
	frc04Request(t, f, a, "GET", frc08Query(known, 999, zero, 1), "", nil, 404)
	alias := fmt.Sprintf("/projects/compare?id1=%s&id2=%s&projectVersionA=1&projectVersionB=1", zero, known)
	require.Equal(t, got, frc04Request(t, f, a, "GET", alias, "", nil, 200))
	for _, query := range []string{"/projects/compare?id1=" + zero + "&id2=" + known, frc08Query(zero, 0, known, 1), frc08Query(zero, 1, known, 1) + "&id1=" + known, frc08Query("invalid", 1, known, 1)} {
		frc04Request(t, f, a, "GET", query, "", nil, 400)
	}
	require.NoError(t, db.Exec("DELETE FROM user_enterprises WHERE user_id=? AND enterprise_id=?", a.u, a.e).Error)
	frc04Request(t, f, a, "GET", frc08Query(zero, 1, known, 1), "", nil, 200)
}

func TestFRC08PostgresIdentityAndProvenance(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a := f.people[0]
	document := func(layer, asset, label string) string {
		var value map[string]any
		require.NoError(t, json.Unmarshal([]byte(frc07Document(t, a.e, "1")), &value))
		l := value["layers"].([]any)[0].(map[string]any)
		l["id"], l["name"] = layer, "Identical display name"
		av := value["assetLibrary"].([]any)[0].(map[string]any)
		av["id"], av["name"] = asset, "Identical asset name"
		av["provenance"] = map[string]any{"sourceLabel": label, "sourceDate": "2026-09-20", "quality": "estimated", "revision": "synthetic-v1", "recordedAt": "2026-09-20T00:00:00Z", "recordedBy": "fixture"}
		obj := value["placedObjects"].([]any)[0].(map[string]any)
		obj["layerId"], obj["assetId"] = layer, asset
		return js(value)
	}
	create := func(name, raw string) string {
		return frc04Request(t, f, a, "POST", "/projects", "", map[string]any{"name": name, "enterpriseId": a.e, "projectJson": raw}, 200)["projectId"].(string)
	}
	idA := create("A", document("layer-A", "asset-A", "Source A"))
	idB := create("B", document("layer-B", "asset-B", "Source B"))
	separate := frc04Request(t, f, a, "GET", frc08Query(idA, 1, idB, 1), "", nil, 200)
	require.Contains(t, js(separate["warnings"]), "unmatched_layer_ids")
	require.Contains(t, js(separate["warnings"]), "unmatched_asset_ids")
	require.Contains(t, js(separate["a"]), "layer-A")
	require.NotContains(t, js(separate["a"]), "layer-B")
	idC := create("C", document("layer-A", "asset-A", "Source B"))
	provenance := frc04Request(t, f, a, "GET", frc08Query(idA, 1, idC, 1), "", nil, 200)
	require.Contains(t, js(provenance["warnings"]), "price_provenance_differs")
	require.Contains(t, js(provenance["a"]), "Source A")
	require.Contains(t, js(provenance["b"]), "Source B")
	require.Equal(t, provenance, frc04Request(t, f, a, "GET", frc08Query(idA, 1, idC, 1), "", nil, 200))
	// A future binding is unavailable, even when another side is readable.
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register("frc08_binding", func(tx *gorm.DB) {
		if tx.Statement.Table == "project_revisions" {
			tx.Statement.SetColumn("CostCalculationVersion", "cost-rub-unavailable")
		}
	}))
	future := create("Future", document("layer-A", "asset-A", "Source A"))
	require.NoError(t, db.Callback().Create().Remove("frc08_binding"))
	response := frc04Request(t, f, a, "GET", frc08Query(idA, 1, future, 1), "", nil, 503)
	require.Equal(t, "unsupported_calculation_version", response["error"].(map[string]any)["code"])
	require.NotContains(t, js(response), "totalMinor")
}
