package main

import (
	"context"
	"github.com/fasthttp/router"
	"github.com/fortis/backend/internal/middleware"
	bi "github.com/fortis/backend/internal/modules/budget/infrastructure"
	ai "github.com/fortis/backend/internal/modules/defense_asset/infrastructure"
	pi "github.com/fortis/backend/internal/modules/defense_project/infrastructure"
	demoUi "github.com/fortis/backend/internal/modules/demo_request/ui"
	ea "github.com/fortis/backend/internal/modules/enterprise/application"
	ed "github.com/fortis/backend/internal/modules/enterprise/domain"
	ei "github.com/fortis/backend/internal/modules/enterprise/infrastructure"
	eu "github.com/fortis/backend/internal/modules/enterprise/ui"
	platformUi "github.com/fortis/backend/internal/modules/platform/ui"
	ui "github.com/fortis/backend/internal/modules/user/infrastructure"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/dig"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	ba "github.com/fortis/backend/internal/modules/budget/application"
	bd "github.com/fortis/backend/internal/modules/budget/domain"
	bu "github.com/fortis/backend/internal/modules/budget/ui"
	aa "github.com/fortis/backend/internal/modules/defense_asset/application"
	ad "github.com/fortis/backend/internal/modules/defense_asset/domain"
	au "github.com/fortis/backend/internal/modules/defense_asset/ui"
	pa "github.com/fortis/backend/internal/modules/defense_project/application"
	pu "github.com/fortis/backend/internal/modules/defense_project/ui"

	ra "github.com/fortis/backend/internal/modules/report/application"
	ru "github.com/fortis/backend/internal/modules/report/ui"
	ua "github.com/fortis/backend/internal/modules/user/application"
	uu "github.com/fortis/backend/internal/modules/user/ui"
)

// Explicit opt-in: runs existing migrations inside a fresh transaction/schema only.
// FRC02_TEST_DSN must point to a disposable local PostgreSQL database.
func TestFRC02Postgres(t *testing.T) {
	tx := frc02TestDatabase(t)
	f := newPostgresFixture(t, tx)
	for i, a := range f.people {
		b := f.people[1-i]
		t.Run([]string{"A", "B"}[i], func(t *testing.T) {
			body := f.request(t, a, "GET", "/enterprises?limit=1", nil, 200)
			require.Contains(t, body, a.e)
			require.NotContains(t, body, b.e)
			require.Contains(t, body, `"totalItems":1`)
			f.request(t, a, "GET", "/enterprises?id="+b.e, nil, 404)
			f.request(t, a, "POST", "/enterprises/members", map[string]string{"userId": a.u, "enterpriseId": b.e}, 403)
			body = f.request(t, a, "GET", "/projects?limit=1", nil, 200)
			require.Contains(t, body, a.p)
			require.NotContains(t, body, b.p)
			require.Contains(t, body, `"totalItems":1`)
			for _, p := range []string{"/projects/get", "/projects/export", "/projects/budget", "/projects/cost", "/projects/report"} {
				f.request(t, a, "GET", p+"?id="+a.p, nil, 200)
				f.request(t, a, "GET", p+"?id="+b.p, nil, 404)
			}
			f.request(t, a, "PUT", "/projects/update?id="+b.p, map[string]string{"name": "stolen"}, 404)
			f.request(t, a, "PUT", "/projects/update?id="+a.p, map[string]any{"name": "allowed", "version": 1}, 200)
			f.request(t, a, "PUT", "/projects/update?id="+a.p, map[string]string{"enterpriseId": b.e}, 400)
			for _, pair := range [][2]string{{a.p, b.p}, {b.p, a.p}, {b.p, b.p}} {
				f.request(t, a, "GET", "/projects/compare?id1="+pair[0]+"&id2="+pair[1], nil, 404)
			}
			body = f.request(t, a, "GET", "/assets?limit=1", nil, 200)
			require.NotContains(t, body, b.a)
			require.Contains(t, body, `"totalItems":2`)
			f.request(t, a, "GET", "/assets/get?id="+a.a, nil, 200)
			f.request(t, a, "GET", "/assets/get?id="+b.a, nil, 404)
			f.request(t, a, "GET", "/assets/documents/get?id="+a.d, nil, 200)
			f.request(t, a, "GET", "/assets/documents/download?id="+a.d, nil, 404)
			body = f.request(t, a, "GET", "/assets/documents/download?id="+b.d, nil, 404)
			require.NotContains(t, body, "storage")
			f.request(t, a, "GET", "/assets/documents/download?id="+f.publicDocument, nil, 404)
			f.request(t, a, "DELETE", "/assets/documents/delete?id="+f.publicDocument, nil, 403)
		})
	}
	a, b := f.people[0], f.people[1]
	// Stored ownership is authoritative even if JSON still embeds an enterprise.
	require.NoError(t, tx.Exec("UPDATE defense_projects SET enterprise_id = NULL WHERE id = ?", b.p).Error)
	require.NoError(t, tx.Exec("UPDATE defense_assets SET enterprise_id = NULL WHERE id = ?", b.a).Error)
	f.request(t, b, "GET", "/projects/get?id="+b.p, nil, 404)
	f.request(t, b, "GET", "/assets/get?id="+b.a, nil, 404)
	f.request(t, b, "GET", "/assets/documents/download?id="+b.d, nil, 404)
	require.NoError(t, tx.Exec("UPDATE defense_assets SET enterprise_id = NULL WHERE id = ?", f.publicAsset).Error)
	f.request(t, a, "GET", "/assets/documents/download?id="+f.publicDocument, nil, 404)
	require.NoError(t, ei.NewEnterpriseRepository(tx).RemoveUserFromEnterprise(context.Background(), a.u, a.e))
	for _, path := range []string{"/enterprises?id=" + a.e, "/projects/get?id=" + a.p, "/projects/budget?id=" + a.p, "/projects/cost?id=" + a.p, "/projects/report?id=" + a.p, "/assets/get?id=" + a.a, "/assets/documents/download?id=" + a.d, "/projects/compare?id1=" + a.p + "&id2=" + a.p} {
		f.request(t, a, "GET", path, nil, 404)
	}
	f.request(t, a, "PUT", "/projects/update?id="+a.p, map[string]string{"name": "revoked"}, 404)
}
func newPostgresFixture(t *testing.T, db *gorm.DB, pipeline ...aa.DocumentPipeline) accessFixture {
	c := context.Background()
	secret := "frc02-local-fixture-only"
	users := ui.NewUserRepository(db)
	us := ua.NewUserService(users, secret, 1)
	erp := ei.NewEnterpriseRepository(db)
	es := ea.NewEnterpriseService(erp)
	pr := pi.NewDefenseProjectRepository(db)
	ps := pa.NewDefenseProjectService(pr, es)
	ar := ai.NewDefenseAssetRepository(db)
	docs := ai.NewDocumentRepository(db)
	as := aa.NewDefenseAssetService(ar, es, docs)
	ds := aa.NewDocumentService(docs, as, pipeline...)
	budgets := bi.NewBudgetConfigRepository(db)
	bs := ba.NewBudgetService(budgets, ps, budgets)
	rs := ra.NewReportService(ps, bs, as)
	people := []person{}
	for _, name := range []string{"A", "B"} {
		u, t, e := us.Register(c, name+"@example.test", "fixture-password-only", name)
		must(e)
		ent, e := ed.NewEnterprise(uuid.NewString(), "Enterprise "+name, "", ed.EnterpriseStatus("active"), 0, 0, time.Now(), time.Now())
		must(e)
		must(erp.Save(c, ent))
		must(erp.AddUserToEnterprise(c, u.ID(), ent.ID()))
		raw := js(map[string]any{"schemaVersion": 1, "projectName": "Project " + name, "baseObject": map[string]any{"id": ent.ID(), "name": "Facility " + name, "center": map[string]int{"lat": 0, "lng": 0}}, "layers": []any{}, "assetLibrary": []any{}, "placedObjects": []any{}})
		p, e := ps.CreateFromJSON(c, u.ID(), "Config "+name, ent.ID(), raw)
		must(e)
		eid := ent.ID()
		a, e := as.Create(c, u.ID(), aa.CreateInput{Name: "Private " + name, Category: ad.DefenseAssetCategoryRadar, CoverageType: ad.DefenseAssetCoverageCircle, EnterpriseID: &eid})
		must(e)
		ownerID := u.ID()
		d, e := ad.NewDocument(uuid.NewString(), a.ID(), "Document "+name, "text/plain", "fixture/"+name, "https://storage.example.test/"+name, 0, &ownerID, time.Now(), time.Now())
		must(e)
		must(docs.Save(c, d))
		must(bs.UpdateBudgetConfig(c, u.ID(), p.ProjectID(), bd.BudgetModeUnlimited, 0))
		people = append(people, person{u.ID(), ent.ID(), p.ProjectID(), a.ID(), d.ID(), t})
	}
	public, e := as.Create(c, people[0].u, aa.CreateInput{Name: "Public reference", Category: ad.DefenseAssetCategoryRadar, CoverageType: ad.DefenseAssetCoverageCircle, EnterpriseID: &people[0].e})
	must(e)
	publicDoc, e := ad.NewDocument(uuid.NewString(), public.ID(), "Public document", "text/plain", "fixture/public", "https://storage.example.test/public", 0, &people[0].u, time.Now(), time.Now())
	must(e)
	must(docs.Save(c, publicDoc))
	public.SetIsPublic(true)
	must(ar.Update(c, public))
	ec := eu.NewEnterpriseController(es)
	pc := pu.NewDefenseProjectController(ps)
	ac := au.NewDefenseAssetController(as)
	dc := au.NewDocumentController(ds)
	bc := bu.NewBudgetController(bs)
	rc := ru.NewReportController(rs)
	uc := uu.NewUserController(us)

	app := &Application{container: dig.New()}
	for _, provider := range []any{
		func() *demoUi.Controller { return demoUi.NewController(nil, demoUi.TransportConfig{}) },
		func() *eu.EnterpriseController { return ec }, func() *pu.DefenseProjectController { return pc }, func() *au.DefenseAssetController { return ac }, func() *au.DocumentController { return dc }, func() *bu.BudgetController { return bc }, func() *ru.ReportController { return rc }, func() *uu.UserController { return uc }, func() *platformUi.ExampleController { return &platformUi.ExampleController{} },
	} {
		require.NoError(t, app.container.Provide(provider))
	}
	r := router.New()
	require.NoError(t, app.registerHandlers(r))
	h := middleware.NewAuthRequired(secret, publicAuthPaths, us).Process(r.Handler)
	return accessFixture{people: people, publicAsset: public.ID(), publicDocument: publicDoc.ID(), h: h}
}

func frc02TestDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("FRC02_TEST_DSN")
	if dsn == "" {
		t.Skip("FRC02_TEST_DSN unset: PostgreSQL isolation not verified")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { sqlDB.Close() })
	tx := db.Begin()
	require.NoError(t, tx.Error)
	t.Cleanup(func() { tx.Rollback() })
	schema := "frc02_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	require.NoError(t, tx.Exec("CREATE SCHEMA "+schema).Error)
	require.NoError(t, tx.Exec("SET LOCAL search_path TO "+schema+", public").Error)
	files, err := filepath.Glob("../../migrations/*.up.sql")
	require.NoError(t, err)
	require.NotEmpty(t, files)
	for _, file := range files {
		sql, err := os.ReadFile(file)
		require.NoError(t, err)
		require.NoError(t, tx.Exec(string(sql)).Error, file)
	}
	return tx
}

// Explicit test-only loopback server for browser tests; never compiled into the application.
func TestFRC02BrowserServer(t *testing.T) {
	if os.Getenv("FRC02_BROWSER_SERVER") != "true" {
		t.Skip("browser fixture disabled")
	}
	// HTTP requests need independent connections; a shared outer transaction is not concurrent-safe.
	t.Setenv("FORTIS_FRC04_TEST_DSN", os.Getenv("FRC02_TEST_DSN"))
	tx := frc04Database(t)
	storage := ai.NewLocalDocumentStorage(ai.DocumentStorageConfig{RootDir: filepath.Join(t.TempDir(), "private"), Enabled: true})
	// This scanner accepts synthetic fixtures only; it is not production scanner evidence.
	f := newPostgresFixture(t, tx, aa.DocumentPipeline{Storage: storage, Scanner: frc02CleanScanner{}})
	listener, err := net.Listen("tcp", "127.0.0.1:8092")
	require.NoError(t, err)
	stopped := make(chan struct{})
	server := &fasthttp.Server{MaxRequestBodySize: 11 * 1024 * 1024, Handler: func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/__fixture":
			ctx.SetContentType("application/json")
			ctx.SetBodyString(js(map[string]any{"a": map[string]string{"userId": f.people[0].u, "enterpriseId": f.people[0].e, "projectId": f.people[0].p}, "b": map[string]string{"userId": f.people[1].u, "enterpriseId": f.people[1].e, "projectId": f.people[1].p}}))
		case "/__revoke":
			if !ctx.IsPost() {
				ctx.SetStatusCode(405)
				return
			}
			err := ei.NewEnterpriseRepository(tx).RemoveUserFromEnterprise(context.Background(), f.people[0].u, f.people[0].e)
			if err != nil {
				ctx.SetStatusCode(500)
			}
		case "/__stop":
			if !ctx.IsPost() {
				ctx.SetStatusCode(405)
				return
			}
			select {
			case <-stopped:
			default:
				close(stopped)
			}
		default:
			f.h(ctx)
		}
	}}
	t.Cleanup(func() { server.Shutdown(); listener.Close() })
	go func() { _ = server.Serve(listener) }()
	t.Log("FRC02 browser fixture ready on 127.0.0.1:8092")
	select {
	case <-stopped:
	case <-time.After(30 * time.Minute):
		t.Fatal("browser fixture timed out")
	}
}
