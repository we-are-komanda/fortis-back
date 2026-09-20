package main

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"testing"

	budgetApp "github.com/fortis/backend/internal/modules/budget/application"
	budgetDomain "github.com/fortis/backend/internal/modules/budget/domain"
	budgetInfra "github.com/fortis/backend/internal/modules/budget/infrastructure"
	projectInfra "github.com/fortis/backend/internal/modules/defense_project/infrastructure"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/stretchr/testify/require"
)

func TestFRC05PostgresLegacyBindingMigration(t *testing.T) {
	db := frc04Database(t, true)
	id := uuid.NewString()
	raw := `{"schemaVersion":1,"projectName":"Legacy cost","baseObject":{"id":"legacy","name":"Legacy","center":{"lat":0,"lng":0}},"layers":[],"assetLibrary":[],"placedObjects":[]}`
	require.NoError(t, db.Exec("INSERT INTO defense_projects (id,name,project_data,version) VALUES (?, ?, ?::jsonb, ?)", id, "Legacy", raw, 7).Error)
	for _, name := range []string{"1789900000_project_revisions.up.sql", "1789914500_cost_projections.up.sql"} {
		body, err := os.ReadFile("../../migrations/" + name)
		require.NoError(t, err)
		require.NoError(t, db.Exec(string(body)).Error)
	}
	var revisions []projectInfra.ProjectRevisionModel
	require.NoError(t, db.Find(&revisions).Error)
	require.Len(t, revisions, 1)
	require.Equal(t, 7, revisions[0].Version)
	require.Equal(t, budgetDomain.CostCalculationV1, revisions[0].CostCalculationVersion)
	var count int64
	require.NoError(t, db.Table("project_cost_projections").Count(&count).Error)
	require.Zero(t, count)
	down, err := os.ReadFile("../../migrations/1789914500_cost_projections.down.sql")
	require.NoError(t, err)
	require.NoError(t, db.Exec(string(down)).Error)
	var after projectInfra.ProjectRevisionModel
	require.NoError(t, db.Select("project_id", "version", "snapshot_json", "snapshot_digest", "created_by", "created_at").Where("project_id=?", id).Take(&after).Error)
	require.Equal(t, revisions[0].SnapshotDigest, after.SnapshotDigest)
	require.JSONEq(t, revisions[0].SnapshotJSON, after.SnapshotJSON)
	require.Equal(t, 7, after.Version)
}

func frc05Document(t *testing.T, enterpriseID string) string {
	t.Helper()
	var project map[string]any
	require.NoError(t, json.Unmarshal([]byte(frc04Document(enterpriseID)), &project))
	raw, err := os.ReadFile("../../internal/modules/budget/document/testdata/finance-v1.json")
	require.NoError(t, err)
	var fixture map[string]any
	require.NoError(t, json.Unmarshal(raw, &fixture))
	clone := func(source map[string]any) map[string]any {
		copy := map[string]any{}
		for k, v := range source {
			copy[k] = v
		}
		return copy
	}
	baseAsset := project["assetLibrary"].([]any)[0].(map[string]any)
	assets := []any{}
	for _, entry := range fixture["assets"].([]any) {
		asset := clone(baseAsset)
		for k, v := range entry.(map[string]any) {
			asset[k] = v
		}
		assets = append(assets, asset)
	}
	baseLayer := project["layers"].([]any)[0].(map[string]any)
	layers := []any{}
	for _, entry := range fixture["layers"].([]any) {
		layer := clone(baseLayer)
		for k, v := range entry.(map[string]any) {
			layer[k] = v
		}
		layers = append(layers, layer)
	}
	baseObject := project["placedObjects"].([]any)[0].(map[string]any)
	objects := []any{}
	for _, entry := range fixture["variantA"].(map[string]any)["objects"].([]any) {
		object := clone(baseObject)
		delete(object, "customPricePerUnitMln")
		for k, v := range entry.(map[string]any) {
			object[k] = v
		}
		objects = append(objects, object)
	}
	project["assetLibrary"], project["layers"], project["placedObjects"] = assets, layers, objects
	return js(project)
}

func TestFRC05PostgresConcurrentCacheAndAccess(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a := f.people[0]
	p, err := frc04ProjectService(db).CreateFromJSON(context.Background(), a.u, "Concurrent cost", a.e, frc05Document(t, a.e))
	require.NoError(t, err)
	repo := budgetInfra.NewBudgetConfigRepository(db)
	service := budgetApp.NewBudgetService(repo, frc04ProjectService(db), repo)
	start := make(chan struct{})
	failures := make(chan error, 8)
	results := make(chan *budgetDomain.CostProjection, 8)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			version := 1
			got, err := service.ProjectCost(context.Background(), a.u, p.ProjectID(), &version)
			results <- got
			failures <- err
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	close(failures)
	for err := range failures {
		require.NoError(t, err)
	}
	for result := range results {
		require.Equal(t, "47500000", *result.TotalMinor())
		require.Equal(t, p.SnapshotDigest(), result.Identity().SnapshotDigest())
	}
	var count int64
	require.NoError(t, db.Table("project_cost_projections").Where("project_id = ?", p.ProjectID()).Count(&count).Error)
	require.EqualValues(t, 1, count)
	require.NoError(t, db.Exec("DELETE FROM user_enterprises WHERE user_id = ? AND enterprise_id = ?", a.u, a.e).Error)
	frc04Request(t, f, a, "GET", "/projects/cost?projectId="+p.ProjectID()+"&projectVersion=1", "", nil, 404)
}

func TestFRC05PostgresFrozenAlgorithmBinding(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a := f.people[0]
	service := frc04ProjectService(db)
	old, err := service.CreateFromJSON(context.Background(), a.u, "Unread v1", a.e, frc05Document(t, a.e))
	require.NoError(t, err)
	// Simulate a later deployment producing a new binding. Old unread revisions
	// still dispatch v1; an unavailable binding must not select the current code.
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register("frc05_future_binding", func(tx *gorm.DB) {
		if tx.Statement.Table == "project_revisions" {
			tx.Statement.SetColumn("CostCalculationVersion", "cost-rub-v2")
		}
	}))
	future, err := service.CreateFromJSON(context.Background(), a.u, "Future binding", a.e, frc05Document(t, a.e))
	require.NoError(t, err)
	require.NoError(t, db.Callback().Create().Remove("frc05_future_binding"))
	projection := frc04Request(t, f, a, "GET", "/projects/cost?projectId="+old.ProjectID()+"&projectVersion=1", "", nil, 200)
	require.Equal(t, "cost-rub-v1", projection["identity"].(map[string]any)["calculationVersion"])
	failure := frc04Request(t, f, a, "GET", "/projects/cost?projectId="+future.ProjectID()+"&projectVersion=1", "", nil, 503)
	require.Equal(t, "unsupported_calculation_version", failure["error"].(map[string]any)["code"])
	var count int64
	require.NoError(t, db.Table("project_cost_projections").Where("project_id = ?", future.ProjectID()).Count(&count).Error)
	require.Zero(t, count)
	require.Error(t, db.Exec("UPDATE project_revisions SET cost_calculation_version='cost-rub-v2' WHERE project_id=?", old.ProjectID()).Error)
	require.NoError(t, service.DeleteProject(context.Background(), a.u, old.ProjectID()))
	frc04Request(t, f, a, "GET", "/projects/cost?projectId="+old.ProjectID()+"&projectVersion=1", "", nil, 404)
	require.NoError(t, db.Table("project_cost_projections").Where("project_id = ?", old.ProjectID()).Count(&count).Error)
	require.Zero(t, count)
}

func TestFRC05PostgresMoneyValidationAndNumericStorage(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a := f.people[0]
	for _, tc := range []struct {
		name     string
		price    any
		quantity int
		status   int
		total    any
	}{
		{"unknown", nil, 2, 200, nil}, {"invalid", "-1", 2, 400, nil}, {"large", "1000000000000000", 1000000, 200, "1000000000000000000000"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var project map[string]any
			require.NoError(t, json.Unmarshal([]byte(frc05Document(t, a.e)), &project))
			project["assetLibrary"].([]any)[0].(map[string]any)["unitPriceMinor"] = tc.price
			object := project["placedObjects"].([]any)[0].(map[string]any)
			object["quantity"] = tc.quantity
			project["placedObjects"] = []any{object}
			created := frc04Request(t, f, a, "POST", "/projects", "", map[string]any{"name": tc.name, "enterpriseId": a.e, "projectJson": js(project)}, 200)
			id := created["projectId"].(string)
			response := frc04Request(t, f, a, "GET", "/projects/cost?projectId="+id+"&projectVersion=1", "", nil, tc.status)
			if tc.status == 400 {
				detail := response["error"].(map[string]any)
				require.Equal(t, "invalid_price", detail["code"])
				require.Equal(t, "assetLibrary.asset-a.unitPriceMinor", detail["details"].(map[string]any)["field"])
				return
			}
			require.Equal(t, tc.total, response["totalMinor"])
			var stored budgetInfra.CostProjectionModel
			require.NoError(t, db.Where("project_id=?", id).First(&stored).Error)
			if tc.total == nil {
				require.Nil(t, stored.TotalMinor)
				require.Equal(t, "0", stored.KnownSubtotalMinor)
				require.Equal(t, false, response["isComplete"])
			} else {
				require.Equal(t, tc.total, *stored.TotalMinor)
			}
		})
	}
}

func TestFRC05PostgresFrozenProjection(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a := f.people[0]
	created := frc04Request(t, f, a, "POST", "/projects", "", map[string]any{"name": "Golden", "enterpriseId": a.e, "projectJson": frc05Document(t, a.e)}, 200)
	id := created["projectId"].(string)
	projection := frc04Request(t, f, a, "GET", "/projects/cost?projectId="+id+"&projectVersion=1", "", nil, 200)
	if path := os.Getenv("FORTIS_FRC05_PROJECTION_FILE"); path != "" {
		require.NoError(t, os.WriteFile(path, []byte(js(projection)), 0600))
	}
	if path := os.Getenv("FORTIS_FRC05_PROJECT_FILE"); path != "" {
		require.NoError(t, os.WriteFile(path, []byte(js(created["snapshot"].(map[string]any)["project"])), 0600))
	}
	require.Equal(t, "47500000", projection["totalMinor"])
	identity := projection["identity"].(map[string]any)
	require.Equal(t, id, identity["projectId"])
	require.EqualValues(t, 1, identity["projectVersion"])
	require.Equal(t, "cost-rub-v1", identity["calculationVersion"])
	require.Equal(t, created["snapshotDigest"], identity["snapshotDigest"])
	var count int64
	require.NoError(t, db.Table("project_cost_projections").Where("project_id = ?", id).Count(&count).Error)
	require.EqualValues(t, 1, count)
	frc04Request(t, f, a, "PUT", "/projects/update?id="+id, "", map[string]any{"name": "Changed", "version": 1}, 200)
	again := frc04Request(t, f, a, "GET", "/projects/cost?projectId="+id+"&projectVersion=1", "", nil, 200)
	require.JSONEq(t, js(projection), js(again))
	frc04Request(t, f, a, "GET", "/projects/cost?projectId="+id+"&projectVersion=99", "", nil, 404)
	frc04Request(t, f, f.people[1], "GET", "/projects/cost?projectId="+id+"&projectVersion=1", "", nil, 404)
	require.Error(t, db.Exec("UPDATE project_cost_projections SET total_minor=0 WHERE project_id=?", id).Error)
}

func TestFRC05PostgresFrozenSourceProvenance(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a := f.people[0]
	var project map[string]any
	require.NoError(t, json.Unmarshal([]byte(frc05Document(t, a.e)), &project))
	provenance := map[string]any{
		"sourceLabel": "Synthetic price document", "sourceDocumentId": uuid.NewString(),
		"sourceUrl": nil, "sourceDate": "2026-09-20", "recordedAt": "2026-09-20T00:00:00Z",
		"recordedBy": a.u, "quality": "confirmed", "revision": "catalog-v3",
		"sourceDocumentRevision": "document-v2", "sourceDocumentChecksum": "sha256:synthetic",
	}
	project["assetLibrary"].([]any)[0].(map[string]any)["fieldProvenance"] = map[string]any{"unitPriceMinor": provenance}
	created := frc04Request(t, f, a, "POST", "/projects", "", map[string]any{"name": "Source identity", "enterpriseId": a.e, "projectJson": js(project)}, 200)
	id := created["projectId"].(string)
	for i := 0; i < 2; i++ {
		projection := frc04Request(t, f, a, "GET", "/projects/cost?projectId="+id+"&projectVersion=1", "", nil, 200)
		require.JSONEq(t, js(provenance), js(projection["lines"].([]any)[0].(map[string]any)["provenance"]))
	}
	var stored budgetInfra.CostProjectionModel
	require.NoError(t, db.Where("project_id=?", id).First(&stored).Error)
	require.Contains(t, stored.ProjectionJSON, "document-v2")
	require.Contains(t, stored.ProjectionJSON, "sha256:synthetic")
}
