package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fortis/backend/internal/audit"
	"github.com/fortis/backend/internal/middleware"
	pa "github.com/fortis/backend/internal/modules/defense_project/application"
	"github.com/fortis/backend/internal/modules/defense_project/document"
	pd "github.com/fortis/backend/internal/modules/defense_project/domain"
	pi "github.com/fortis/backend/internal/modules/defense_project/infrastructure"
	ea "github.com/fortis/backend/internal/modules/enterprise/application"
	ei "github.com/fortis/backend/internal/modules/enterprise/infrastructure"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func frc04Database(t *testing.T, legacy ...bool) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("FORTIS_FRC04_TEST_DSN")
	if dsn == "" {
		t.Skip("FORTIS_FRC04_TEST_DSN unset: PostgreSQL revision verification not run")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	admin, err := db.DB()
	require.NoError(t, err)
	schema := "frc04_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	require.NoError(t, db.Exec("CREATE SCHEMA "+schema).Error)
	t.Cleanup(func() { db.Exec("DROP SCHEMA " + schema + " CASCADE"); admin.Close() })
	scoped, err := gorm.Open(postgres.Open(dsn+" search_path="+schema+",public"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	conn, err := scoped.DB()
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })
	files, err := filepath.Glob("../../migrations/*.up.sql")
	require.NoError(t, err)
	for _, file := range files {
		if len(legacy) > 0 && legacy[0] && (strings.Contains(file, "project_revisions") || strings.Contains(file, "cost_projections")) {
			continue
		}
		body, err := os.ReadFile(file)
		require.NoError(t, err)
		require.NoError(t, scoped.Exec(string(body)).Error, file)
	}
	return scoped
}

func frc04ProjectService(db *gorm.DB) *pa.DefenseProjectService {
	return pa.NewDefenseProjectService(pi.NewDefenseProjectRepository(db), ea.NewEnterpriseService(ei.NewEnterpriseRepository(db)))
}

func TestFRC04PostgresCASAndRollback(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a := f.people[0]
	service := frc04ProjectService(db)
	created := frc04Request(t, f, a, "POST", "/projects", "", map[string]any{"name": "Atomic", "enterpriseId": a.e, "projectJson": frc04Document(a.e)}, 200)
	id := created["projectId"].(string)
	t.Run("two_writers_same_version", func(t *testing.T) {
		var wg sync.WaitGroup
		start := make(chan struct{})
		results := make(chan error, 2)
		for _, name := range []string{"Writer A", "Writer B"} {
			wg.Add(1)
			go func(name string) {
				defer wg.Done()
				<-start
				version := 1
				p, err := service.UpdateProject(context.Background(), a.u, id, name, nil, "", &version)
				if err == nil && p.Version() != 2 {
					err = errors.New("save did not return committed version 2")
				}
				results <- err
			}(name)
		}
		close(start)
		wg.Wait()
		close(results)
		success, conflict := 0, 0
		for err := range results {
			if err == nil {
				success++
			} else if errors.Is(err, pd.ErrVersionConflict) {
				conflict++
			} else {
				t.Errorf("unexpected writer error: %v", err)
			}
		}
		require.Equal(t, 1, success)
		require.Equal(t, 1, conflict)
		var count int64
		require.NoError(t, db.Table("project_revisions").Where("project_id = ?", id).Count(&count).Error)
		require.EqualValues(t, 2, count)
	})
	t.Run("revision_failure_rolls_back_current_budget_and_version", func(t *testing.T) {
		before, err := service.GetProject(context.Background(), a.u, id)
		require.NoError(t, err)
		require.NoError(t, db.Callback().Create().Before("gorm:create").Register("frc04_revision_failure", func(tx *gorm.DB) {
			if tx.Statement.Table == "project_revisions" {
				tx.AddError(errors.New("synthetic revision insertion failure"))
			}
		}))
		_, err = service.UpdateBudget(context.Background(), a.u, id, before.Version(), `{"budgetMode":"limited","budgetAmountMln":12.5}`)
		require.ErrorContains(t, err, "synthetic revision insertion failure")
		require.NoError(t, db.Callback().Create().Remove("frc04_revision_failure"))
		after, err := service.GetProject(context.Background(), a.u, id)
		require.NoError(t, err)
		require.Equal(t, before.Version(), after.Version())
		require.Equal(t, before.Document(), after.Document())
		var count int64
		require.NoError(t, db.Table("budget_configs").Where("project_id = ?", id).Count(&count).Error)
		require.Zero(t, count)
		require.NoError(t, db.Table("project_revisions").Where("project_id = ? AND version = 3", id).Count(&count).Error)
		require.Zero(t, count)
		saved, err := service.UpdateBudget(context.Background(), a.u, id, 2, `{"budgetMode":"limited","budgetAmountMln":12.5}`)
		require.NoError(t, err)
		require.Equal(t, 3, saved.Version())
		var snapshot map[string]any
		require.NoError(t, json.Unmarshal([]byte(saved.Snapshot()), &snapshot))
		require.Equal(t, 12.5, snapshot["budget"].(map[string]any)["budgetAmountMln"])
		v := 2
		old, err := service.GetRevision(context.Background(), a.u, id, &v)
		require.NoError(t, err)
		require.Contains(t, old.Snapshot(), `"budget": null`)
	})
}

func TestFRC04PostgresFrontendFixture(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a := f.people[0]
	fixture, err := os.ReadFile("testdata/frc04-project.json")
	require.NoError(t, err)
	var original map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(fixture, &original))
	original["enterpriseId"], err = json.Marshal(a.e)
	require.NoError(t, err)
	raw, err := json.Marshal(original)
	require.NoError(t, err)
	created := frc04Request(t, f, a, "POST", "/projects", "frontend-fixture", map[string]any{"name": "Frontend fixture", "enterpriseId": a.e, "projectJson": string(raw)}, 200)
	id := created["projectId"].(string)
	exported := f.request(t, a, "GET", "/projects/export?id="+id+"&projectVersion=1", nil, 200)
	var loaded map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(exported), &loaded))
	for key, want := range original {
		switch key {
		case "projectId", "version", "enterpriseId", "name", "updatedAt":
			continue
		}
		require.JSONEq(t, string(want), string(loaded[key]), key)
	}
	// The same server-frozen payload, not the mutable global catalog, defines the digest.
	get := frc04Request(t, f, a, "GET", "/projects/get?id="+id+"&projectVersion=1", "", nil, 200)
	snapshotRaw, err := json.Marshal(get["snapshot"])
	require.NoError(t, err)
	canonical, err := document.Canonical(snapshotRaw)
	require.NoError(t, err)
	sum := sha256.Sum256(canonical)
	require.Equal(t, hex.EncodeToString(sum[:]), get["snapshotDigest"])
	require.NoError(t, db.Table("defense_assets").Where("id = ?", a.a).Update("asset_data", `{"name":"Changed global template"}`).Error)
	require.Equal(t, exported, f.request(t, a, "GET", "/projects/export?id="+id+"&projectVersion=1", nil, 200))
	require.Error(t, db.Table("project_revisions").Where("project_id = ?", id).Update("snapshot_json", `{}`).Error)
	if path := os.Getenv("FORTIS_FRC04_EXPORT_FILE"); path != "" {
		require.NoError(t, os.WriteFile(path, []byte(exported), 0600))
	}
}

func TestFRC04PostgresReplayScopesExpiryAndAccess(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a, b := f.people[0], f.people[1]
	payload := func(person person) map[string]any {
		return map[string]any{"name": "Scoped", "enterpriseId": person.e, "projectJson": frc04Document(person.e)}
	}
	created := frc04Request(t, f, a, "POST", "/projects", "scope-key", payload(a), 200)
	id := created["projectId"].(string)
	other := frc04Request(t, f, b, "POST", "/projects", "scope-key", payload(b), 200)
	require.NotEqual(t, id, other["projectId"])
	imported := frc04Request(t, f, a, "POST", "/projects/import", "scope-key", map[string]any{"projectJson": frc04Document(a.e)}, 200)
	require.NotEqual(t, id, imported["projectId"])
	require.Equal(t, imported, frc04Request(t, f, a, "POST", "/projects/import", "scope-key", map[string]any{"projectJson": frc04Document(a.e)}, 200))
	require.NoError(t, db.Table("project_idempotency").Where("actor_id = ? AND operation = ?", a.u, "create").Update("expires_at", time.Now().Add(-time.Second)).Error)
	fresh := frc04Request(t, f, a, "POST", "/projects", "scope-key", payload(a), 200)
	require.NotEqual(t, id, fresh["projectId"])
	f.request(t, a, "DELETE", "/projects/delete?id="+fresh["projectId"].(string), nil, 200)
	frc04Request(t, f, a, "POST", "/projects", "scope-key", payload(a), 404)
	frc04Request(t, f, a, "GET", "/projects/get?id="+fresh["projectId"].(string)+"&projectVersion=1", "", nil, 404)
	orphan := frc04Request(t, f, a, "POST", "/projects", "orphan-replay", payload(a), 200)
	require.NoError(t, db.Table("defense_projects").Where("id = ?", orphan["projectId"]).Update("enterprise_id", nil).Error)
	frc04Request(t, f, a, "POST", "/projects", "orphan-replay", payload(a), 404)
	require.NoError(t, ei.NewEnterpriseRepository(db).RemoveUserFromEnterprise(context.Background(), a.u, a.e))
	frc04Request(t, f, a, "GET", "/projects/get?id="+id+"&projectVersion=1", "", nil, 404)
	frc04Request(t, f, a, "GET", "/projects/export?id="+id+"&projectVersion=1", "", nil, 404)
	frc04Request(t, f, a, "POST", "/projects/import", "scope-key", map[string]any{"projectJson": frc04Document(a.e)}, 404)
}

func TestFRC04PostgresLegacyBackfill(t *testing.T) {
	db := frc04Database(t, true)
	id := uuid.NewString()
	raw := `{"schemaVersion":1,"projectName":"Legacy schema one","baseObject":{"id":"legacy","name":"Legacy object","center":{"lat":0,"lng":0}},"layers":[],"assetLibrary":[],"placedObjects":[]}`
	require.NoError(t, db.Exec("INSERT INTO defense_projects (id,name,project_data,version) VALUES (?, ?, ?::jsonb, ?)", id, "Legacy", raw, 7).Error)
	require.NoError(t, db.Exec("INSERT INTO budget_configs (project_id,config_data) VALUES (?,?::jsonb)", id, `{"budgetMode":"limited","budgetAmountMln":4}`).Error)
	migration, err := os.ReadFile("../../migrations/1789900000_project_revisions.up.sql")
	require.NoError(t, err)
	require.NoError(t, db.Exec(string(migration)).Error)
	var revisions []pi.ProjectRevisionModel
	require.NoError(t, db.Find(&revisions).Error)
	require.Len(t, revisions, 1)
	require.Equal(t, 7, revisions[0].Version)
	require.Nil(t, revisions[0].CreatedBy)
	canonical, err := document.Canonical([]byte(revisions[0].SnapshotJSON))
	require.NoError(t, err)
	sum := sha256.Sum256(canonical)
	require.Equal(t, hex.EncodeToString(sum[:]), revisions[0].SnapshotDigest)
	// The current repository also reads FRC-05's frozen calculator binding.
	costMigration, err := os.ReadFile("../../migrations/1789914500_cost_projections.up.sql")
	require.NoError(t, err)
	require.NoError(t, db.Exec(string(costMigration)).Error)
	repo := pi.NewDefenseProjectRepository(db).(pd.ProjectRevisionRepository)
	_, err = repo.FindRevision(context.Background(), id, 6)
	require.ErrorIs(t, err, pd.ErrRevisionNotFound)
	legacy, err := repo.FindRevision(context.Background(), id, 7)
	require.NoError(t, err)
	require.Contains(t, legacy.Snapshot(), `"budgetAmountMln": 4`)
	require.Equal(t, pd.DefenseProjectModeView, legacy.Mode())
	require.Equal(t, pd.DefenseProjectSourceCustom, legacy.Source())
	// Rollback removes only new revision/idempotency storage, leaving current data intact.
	costDown, err := os.ReadFile("../../migrations/1789914500_cost_projections.down.sql")
	require.NoError(t, err)
	require.NoError(t, db.Exec(string(costDown)).Error)
	down, err := os.ReadFile("../../migrations/1789900000_project_revisions.down.sql")
	require.NoError(t, err)
	require.NoError(t, db.Exec(string(down)).Error)
	var count int64
	require.NoError(t, db.Table("defense_projects").Where("id = ? AND version = 7", id).Count(&count).Error)
	require.EqualValues(t, 1, count)
}

func TestFRC04PostgresFreezesActualTemplateAndBudget(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a := f.people[0]
	raw := strings.ReplaceAll(frc04Document(a.e), `"asset"`, `"`+a.a+`"`)
	created := frc04Request(t, f, a, "POST", "/projects", "", map[string]any{"name": "Frozen library", "enterpriseId": a.e, "projectJson": raw}, 200)
	id := created["projectId"].(string)
	before := f.request(t, a, "GET", "/projects/export?id="+id+"&projectVersion=1", nil, 200)
	f.request(t, a, "PUT", "/assets/update?id="+a.a, map[string]any{"name": "Changed same-ID template", "pricePerUnitMln": 99}, 200)
	require.Equal(t, before, f.request(t, a, "GET", "/projects/export?id="+id+"&projectVersion=1", nil, 200))
	service := frc04ProjectService(db)
	first, err := service.UpdateBudget(context.Background(), a.u, id, 1, `{"budgetMode":"limited","budgetAmountMln":10}`)
	require.NoError(t, err)
	_, err = service.UpdateBudget(context.Background(), a.u, id, 2, `{"budgetMode":"limited","budgetAmountMln":20}`)
	require.NoError(t, err)
	version := 2
	historical, err := service.GetRevision(context.Background(), a.u, id, &version)
	require.NoError(t, err)
	require.JSONEq(t, first.Snapshot(), historical.Snapshot())
	require.Equal(t, first.SnapshotDigest(), historical.SnapshotDigest())
}

func TestFRC04PostgresLegacyGeometryDefaults(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a := f.people[0]
	for _, kind := range []string{"circle", "ring"} {
		t.Run(kind, func(t *testing.T) {
			raw := js(map[string]any{"schemaVersion": 1, "projectName": "Compatible geometry", "baseObject": map[string]any{"id": "base", "name": "Base", "center": map[string]any{"lat": 0, "lng": 0}}, "layers": []any{map[string]any{"id": "legacy-layer", "geometryType": kind, "geometry": map[string]any{"type": kind, "extension": "preserved"}}}, "assetLibrary": []any{}, "placedObjects": []any{}})
			require.NotPanics(t, func() {
				created := frc04Request(t, f, a, "POST", "/projects", "", map[string]any{"name": "Legacy geometry", "enterpriseId": a.e, "projectJson": raw}, 200)
				f.request(t, a, "GET", "/projects/export?id="+created["projectId"].(string), nil, 200)
			})
		})
	}
}

func TestFRC04PostgresConcurrentIdempotency(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a := f.people[0]
	service := frc04ProjectService(db)
	start := make(chan struct{})
	results := make(chan *pd.DefenseProject, 2)
	failures := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			p, err := service.CreateIdempotent(context.Background(), a.u, "Concurrent", a.e, frc04Document(a.e), "double-click")
			results <- p
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
	var first *pd.DefenseProject
	for p := range results {
		if first == nil {
			first = p
		} else {
			require.Equal(t, first.ProjectID(), p.ProjectID())
			require.Equal(t, first.SnapshotDigest(), p.SnapshotDigest())
		}
	}
	var count int64
	require.NoError(t, db.Table("defense_projects").Where("name = ?", "Concurrent").Count(&count).Error)
	require.EqualValues(t, 1, count)
	require.NoError(t, db.Table("project_revisions").Where("project_id = ?", first.ProjectID()).Count(&count).Error)
	require.EqualValues(t, 1, count)
}

func TestFRC04PostgresConcurrentExpiredIdempotency(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a := f.people[0]
	service := frc04ProjectService(db)
	old, err := service.CreateIdempotent(context.Background(), a.u, "Expired concurrent", a.e, frc04Document(a.e), "expired-concurrent")
	require.NoError(t, err)
	require.NoError(t, db.Table("project_idempotency").Where("actor_id = ?", a.u).Update("expires_at", time.Now().Add(-time.Hour)).Error)
	// Force the former DELETE/INSERT replacement to overlap waiting requests.
	require.NoError(t, db.Exec(`CREATE FUNCTION slow_receipt_delete() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN PERFORM pg_sleep(0.3); RETURN OLD; END; $$; CREATE TRIGGER slow_receipt_delete BEFORE DELETE ON project_idempotency FOR EACH ROW EXECUTE FUNCTION slow_receipt_delete();`).Error)
	start := make(chan struct{})
	results := make(chan *pd.DefenseProject, 8)
	failures := make(chan error, 8)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			p, err := service.CreateIdempotent(context.Background(), a.u, "Expired concurrent", a.e, frc04Document(a.e), "expired-concurrent")
			results <- p
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
	var first *pd.DefenseProject
	for p := range results {
		require.NotEqual(t, old.ProjectID(), p.ProjectID())
		if first == nil {
			first = p
		} else {
			require.Equal(t, first.ProjectID(), p.ProjectID())
			require.Equal(t, first.SnapshotDigest(), p.SnapshotDigest())
		}
	}
	var count int64
	require.NoError(t, db.Table("defense_projects").Where("name = ?", "Expired concurrent").Count(&count).Error)
	require.EqualValues(t, 2, count)
	require.NoError(t, db.Table("project_revisions").Where("project_id = ?", first.ProjectID()).Count(&count).Error)
	require.EqualValues(t, 1, count)
	require.NoError(t, db.Table("audit_events").Where("entity_id = ?", first.ProjectID()).Count(&count).Error)
	require.EqualValues(t, 1, count)
}

type frc04RevokingAccess struct {
	db   *gorm.DB
	base *ea.EnterpriseService
}

func (access frc04RevokingAccess) CheckUserAccess(ctx context.Context, userID, enterpriseID string) error {
	if err := access.base.CheckUserAccess(ctx, userID, enterpriseID); err != nil {
		return err
	}
	return ei.NewEnterpriseRepository(access.db).RemoveUserFromEnterprise(ctx, userID, enterpriseID)
}

func TestFRC04PostgresMembershipRecheckedInsideWrite(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a := f.people[0]
	access := frc04RevokingAccess{db: db, base: ea.NewEnterpriseService(ei.NewEnterpriseRepository(db))}
	service := pa.NewDefenseProjectService(pi.NewDefenseProjectRepository(db), access)
	_, err := service.CreateFromJSON(context.Background(), a.u, "Revoked during request", a.e, frc04Document(a.e))
	require.Error(t, err)
	var count int64
	require.NoError(t, db.Table("defense_projects").Where("name = ?", "Revoked during request").Count(&count).Error)
	require.Zero(t, count)
}

func TestFRC04PostgresDeletingIdentityDoesNotRewriteRevisions(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a := f.people[0]
	var before pi.ProjectRevisionModel
	require.NoError(t, db.Where("project_id = ? AND version = 1", a.p).First(&before).Error)
	require.NoError(t, db.Exec("DELETE FROM users WHERE id = ?", a.u).Error)
	var after pi.ProjectRevisionModel
	require.NoError(t, db.Where("project_id = ? AND version = 1", a.p).First(&after).Error)
	require.Equal(t, before, after)
	f.request(t, a, "GET", "/projects/get?id="+a.p+"&projectVersion=1", nil, 401)
}

func TestFRC04PostgresAuditAtomicity(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	f.h = middleware.NewHttpResponse().Process(f.h)
	a := f.people[0]
	body := map[string]any{"name": "Audited", "enterpriseId": a.e, "projectJson": frc04Document(a.e)}
	var ctx fasthttp.RequestCtx
	ctx.Init2(nil, nil, false)
	ctx.Request.SetRequestURI("/api/v1/projects")
	ctx.Request.Header.SetMethod("POST")
	ctx.Request.Header.SetContentType("application/json")
	ctx.Request.Header.Set("Authorization", "Bearer "+a.token)
	ctx.Request.Header.Set("Idempotency-Key", "audited-create")
	ctx.Request.Header.Set("X-Request-ID", "untrusted-client-value")
	ctx.Request.SetBodyString(js(body))
	f.h(&ctx)
	require.Equal(t, 200, ctx.Response.StatusCode(), string(ctx.Response.Body()))
	var created map[string]any
	require.NoError(t, json.Unmarshal(ctx.Response.Body(), &created))
	id := created["projectId"].(string)
	var firstEvent audit.Event
	require.NoError(t, db.Where("entity_id = ?", id).First(&firstEvent).Error)
	require.Equal(t, string(ctx.Response.Header.Peek("X-Request-ID")), firstEvent.RequestID)
	require.NotEqual(t, "untrusted-client-value", firstEvent.RequestID)
	_, err := uuid.Parse(firstEvent.RequestID)
	require.NoError(t, err)
	var count int64
	require.NoError(t, db.Table("audit_events").Where("entity_id = ? AND action = ?", id, "project.create").Count(&count).Error)
	require.EqualValues(t, 1, count)
	frc04Request(t, f, a, "POST", "/projects", "audited-create", body, 200)
	require.NoError(t, db.Table("audit_events").Where("entity_id = ?", id).Count(&count).Error)
	require.EqualValues(t, 1, count)
	service := frc04ProjectService(db)
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register("frc04_audit_failure", func(tx *gorm.DB) {
		if tx.Statement.Table == "audit_events" {
			tx.AddError(errors.New("synthetic audit append failure"))
		}
	}))
	_, err = service.UpdateBudget(context.Background(), a.u, id, 1, `{"budgetMode":"limited","budgetAmountMln":123}`)
	require.ErrorContains(t, err, "synthetic audit append failure")
	err = service.DeleteProject(context.Background(), a.u, id)
	require.ErrorContains(t, err, "synthetic audit append failure")
	require.NoError(t, db.Callback().Create().Remove("frc04_audit_failure"))
	current, err := service.GetProject(context.Background(), a.u, id)
	require.NoError(t, err)
	require.Equal(t, 1, current.Version())
	require.NoError(t, db.Table("budget_configs").Where("project_id = ?", id).Count(&count).Error)
	require.Zero(t, count)
	require.NoError(t, db.Table("project_revisions").Where("project_id = ?", id).Count(&count).Error)
	require.EqualValues(t, 1, count)
	version := 1
	_, err = service.UpdateProject(context.Background(), a.u, id, "Audit update", nil, "", &version)
	require.NoError(t, err)
	_, err = service.UpdateBudget(context.Background(), a.u, id, 2, `{"budgetMode":"limited","budgetAmountMln":123}`)
	require.NoError(t, err)
	require.NoError(t, service.DeleteProject(context.Background(), a.u, id))
	var events []map[string]any
	require.NoError(t, db.Table("audit_events").Where("entity_id = ?", id).Order("occurred_at").Find(&events).Error)
	require.Len(t, events, 4)
	for _, event := range events {
		require.Equal(t, a.u, event["actor_id"])
		require.NotEmpty(t, event["request_id"])
		require.NotContains(t, js(event), "projectJson")
		require.NotContains(t, js(event), "budgetAmountMln")
	}
	require.Equal(t, "project.create", events[0]["action"])
	require.Equal(t, "project.update", events[1]["action"])
	require.Equal(t, "budget.update", events[2]["action"])
	require.Equal(t, "budget", events[2]["entity_type"])
	require.Equal(t, "project.delete", events[3]["action"])
	require.Nil(t, events[0]["previous_version"])
	require.EqualValues(t, 1, events[0]["new_version"])
	require.EqualValues(t, 1, events[1]["previous_version"])
	require.EqualValues(t, 2, events[1]["new_version"])
	require.EqualValues(t, 2, events[2]["previous_version"])
	require.EqualValues(t, 3, events[2]["new_version"])
	require.EqualValues(t, 3, events[3]["previous_version"])
	require.Nil(t, events[3]["new_version"])
	require.NoError(t, db.Table("project_revisions").Where("project_id = ?", id).Count(&count).Error)
	require.Zero(t, count)
	require.Error(t, db.Exec("UPDATE audit_events SET action = 'tamper' WHERE entity_id = ?", id).Error)
	require.Error(t, db.Exec("DELETE FROM audit_events WHERE entity_id = ?", id).Error)
	imported, err := service.ImportIdempotent(context.Background(), a.u, frc04Document(a.e), "audited-import")
	require.NoError(t, err)
	var importedEvent audit.Event
	require.NoError(t, db.Where("entity_id = ?", imported.ProjectID()).First(&importedEvent).Error)
	require.Equal(t, "project.import", importedEvent.Action)
	require.Equal(t, a.u, importedEvent.ActorID)
}

func frc04Request(t *testing.T, f accessFixture, actor person, method, path, key string, body any, want int) map[string]any {
	t.Helper()
	var ctx fasthttp.RequestCtx
	ctx.Init2(nil, nil, false)
	ctx.Request.SetRequestURI("/api/v1" + path)
	ctx.Request.Header.SetMethod(method)
	ctx.Request.Header.SetContentType("application/json")
	ctx.Request.Header.Set("Authorization", "Bearer "+actor.token)
	if key != "" {
		ctx.Request.Header.Set("Idempotency-Key", key)
	}
	if body != nil {
		ctx.Request.SetBodyString(js(body))
	}
	f.h(&ctx)
	require.Equal(t, want, ctx.Response.StatusCode(), string(ctx.Response.Body()))
	var result map[string]any
	require.NoError(t, json.Unmarshal(ctx.Response.Body(), &result))
	return result
}

func TestFRC04PostgresRevisionAndIdempotency(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a := f.people[0]
	body := map[string]any{"name": "Lossless", "enterpriseId": a.e, "projectJson": frc04Document(a.e)}
	created := frc04Request(t, f, a, "POST", "/projects", "same-request", body, 200)
	id := created["projectId"].(string)
	t.Run("immutable_revision", func(t *testing.T) {
		var count int64
		require.NoError(t, db.Table("project_revisions").Where("project_id = ? AND version = 1", id).Count(&count).Error)
		require.EqualValues(t, 1, count)
		get := frc04Request(t, f, a, "GET", "/projects/get?id="+id+"&projectVersion=1", "", nil, 200)
		require.NotEmpty(t, get["snapshotDigest"])
		snapshot, ok := get["snapshot"].(map[string]any)
		require.True(t, ok, "missing revision envelope")
		var original map[string]any
		require.NoError(t, json.Unmarshal([]byte(frc04Document(a.e)), &original))
		project := snapshot["project"].(map[string]any)
		for _, key := range []string{"baseObject", "layers", "assetLibrary", "placedObjects", "customFields"} {
			require.Equal(t, original[key], project[key], key)
		}
		frc04Request(t, f, a, "PUT", "/projects/update?id="+id, "", map[string]any{"version": 1, "name": "Version 2"}, 200)
		historical := frc04Request(t, f, a, "GET", "/projects/get?id="+id+"&projectVersion=1", "", nil, 200)
		require.Equal(t, get, historical)
		missing := frc04Request(t, f, a, "GET", "/projects/get?id="+id+"&projectVersion=999", "", nil, 404)
		require.Contains(t, js(missing), "revision_not_found")
	})
	t.Run("idempotent_replay", func(t *testing.T) {
		replay := frc04Request(t, f, a, "POST", "/projects", "same-request", body, 200)
		require.Equal(t, created, replay)
		changed := map[string]any{"name": "Changed", "enterpriseId": a.e, "projectJson": frc04Document(a.e)}
		conflict := frc04Request(t, f, a, "POST", "/projects", "same-request", changed, 409)
		require.Contains(t, js(conflict), "idempotency_conflict")
	})
}
