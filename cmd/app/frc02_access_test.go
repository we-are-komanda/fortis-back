package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/fasthttp/router"
	"github.com/fortis/backend/internal/auth"
	"github.com/fortis/backend/internal/middleware"
	ba "github.com/fortis/backend/internal/modules/budget/application"
	bd "github.com/fortis/backend/internal/modules/budget/domain"
	bu "github.com/fortis/backend/internal/modules/budget/ui"
	aa "github.com/fortis/backend/internal/modules/defense_asset/application"
	ad "github.com/fortis/backend/internal/modules/defense_asset/domain"
	au "github.com/fortis/backend/internal/modules/defense_asset/ui"
	pa "github.com/fortis/backend/internal/modules/defense_project/application"
	pd "github.com/fortis/backend/internal/modules/defense_project/domain"
	pu "github.com/fortis/backend/internal/modules/defense_project/ui"
	ea "github.com/fortis/backend/internal/modules/enterprise/application"
	ed "github.com/fortis/backend/internal/modules/enterprise/domain"
	eu "github.com/fortis/backend/internal/modules/enterprise/ui"
	platformUi "github.com/fortis/backend/internal/modules/platform/ui"
	ra "github.com/fortis/backend/internal/modules/report/application"
	ru "github.com/fortis/backend/internal/modules/report/ui"
	ua "github.com/fortis/backend/internal/modules/user/application"
	ud "github.com/fortis/backend/internal/modules/user/domain"
	uu "github.com/fortis/backend/internal/modules/user/ui"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/dig"
	"testing"
	"time"
)

// mockRepo — мок репозитория для тестирования сервиса.
type mockRepo struct {
	projects    map[string]*pd.DefenseProject
	memberships map[string]bool
	err         error
}

func newmockRepo() *mockRepo {
	return &mockRepo{
		projects: make(map[string]*pd.DefenseProject),
	}
}

func (m *mockRepo) Save(ctx context.Context, project *pd.DefenseProject) error {
	if m.err != nil {
		return m.err
	}
	if _, exists := m.projects[project.ProjectID()]; exists {
		project.SetVersion(project.Version() + 1)
	}
	m.projects[project.ProjectID()] = project
	return nil
}

func (m *mockRepo) FindByID(ctx context.Context, id string) (*pd.DefenseProject, error) {
	if m.err != nil {
		return nil, m.err
	}
	p, ok := m.projects[id]
	if !ok {
		return nil, pd.ErrProjectNotFound
	}
	return p, nil
}

func (m *mockRepo) FindAll(ctx context.Context, limit, offset int) ([]*pd.DefenseProject, int64, error) {
	projects := make([]*pd.DefenseProject, 0, len(m.projects))
	for _, p := range m.projects {
		projects = append(projects, p)
	}
	return projects, int64(len(projects)), nil
}

func (m *mockRepo) FindAllByEnterprise(ctx context.Context, enterpriseID string, limit, offset int) ([]*pd.DefenseProject, int64, error) {
	projects := make([]*pd.DefenseProject, 0)
	for _, p := range m.projects {
		if p.EnterpriseID() == enterpriseID {
			projects = append(projects, p)
		}
	}
	return projects, int64(len(projects)), nil
}

func (m *mockRepo) Delete(ctx context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.projects, id)
	return nil
}

// mockDefenseAssetRepo — мок репозитория для тестирования сервиса.
type mockDefenseAssetRepo struct {
	assets      map[string]*ad.DefenseAsset
	memberships map[string]bool
	saveErr     error
}

func newmockDefenseAssetRepo() *mockDefenseAssetRepo {
	return &mockDefenseAssetRepo{
		assets: make(map[string]*ad.DefenseAsset),
	}
}

func (m *mockDefenseAssetRepo) Save(ctx context.Context, asset *ad.DefenseAsset) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.assets[asset.ID()] = asset
	return nil
}

func (m *mockDefenseAssetRepo) FindByID(ctx context.Context, id string) (*ad.DefenseAsset, error) {
	asset, ok := m.assets[id]
	if !ok {
		return nil, ad.ErrDefenseAssetNotFound
	}
	return asset, nil
}

func (m *mockDefenseAssetRepo) FindAll(ctx context.Context, filter ad.DefenseAssetFilter) ([]*ad.DefenseAsset, int64, error) {
	var result []*ad.DefenseAsset
	for _, a := range m.assets {
		if !a.IsPublic() && (a.EnterpriseID() == nil || !m.memberships[filter.UserID+*a.EnterpriseID()]) {
			continue
		}
		if filter.EnterpriseID != nil && (a.EnterpriseID() == nil || *a.EnterpriseID() != *filter.EnterpriseID) {
			continue
		}
		if filter.IsPublic != nil && a.IsPublic() != *filter.IsPublic {
			continue
		}
		if filter.Category != nil && a.Category() != *filter.Category {
			continue
		}
		result = append(result, a)
	}
	total := int64(len(result))

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	if offset >= len(result) {
		return []*ad.DefenseAsset{}, total, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}

	return result[offset:end], total, nil
}

func (m *mockDefenseAssetRepo) Update(ctx context.Context, asset *ad.DefenseAsset) error {
	if _, ok := m.assets[asset.ID()]; !ok {
		return ad.ErrDefenseAssetNotFound
	}
	m.assets[asset.ID()] = asset
	return nil
}

func (m *mockDefenseAssetRepo) Delete(ctx context.Context, id string) error {
	if _, ok := m.assets[id]; !ok {
		return ad.ErrDefenseAssetNotFound
	}
	delete(m.assets, id)
	return nil
}

type er struct {
	es        map[string]*ed.Enterprise
	links     map[string]bool
	accessErr error
}

func (r *er) Save(c context.Context, e *ed.Enterprise) error { r.es[e.ID()] = e; return nil }
func (r *er) FindByID(c context.Context, id string) (*ed.Enterprise, error) {
	if e := r.es[id]; e != nil {
		return e, nil
	}
	return nil, ed.ErrEnterpriseNotFound
}
func (r *er) FindAll(c context.Context, l, o int) ([]*ed.Enterprise, int64, error) {
	var x []*ed.Enterprise
	for _, e := range r.es {
		x = append(x, e)
	}
	return x, int64(len(x)), nil
}
func (r *er) FindAllByUserID(c context.Context, u string, l, o int) ([]*ed.Enterprise, int64, error) {
	var x []*ed.Enterprise
	for _, e := range r.es {
		if r.links[u+e.ID()] {
			x = append(x, e)
		}
	}
	return x, int64(len(x)), nil
}
func (r *er) CheckUserEnterpriseAccess(c context.Context, u, e string) (bool, error) {
	return r.links[u+e], r.accessErr
}
func (r *er) AddUserToEnterprise(c context.Context, u, e string) error {
	r.links[u+e] = true
	return nil
}
func (r *er) RemoveUserFromEnterprise(c context.Context, u, e string) error {
	delete(r.links, u+e)
	return nil
}
func (r *er) Delete(c context.Context, id string) error { delete(r.es, id); return nil }

type dr map[string]*ad.Document

func (r dr) Save(c context.Context, d *ad.Document) error { r[d.ID()] = d; return nil }
func (r dr) FindByID(c context.Context, id string) (*ad.Document, error) {
	if d := r[id]; d != nil {
		return d, nil
	}
	return nil, ad.ErrDocumentNotFound
}
func (r dr) FindByAssetID(c context.Context, id string) ([]*ad.Document, error) {
	var x []*ad.Document
	for _, d := range r {
		if d.AssetID() == id {
			x = append(x, d)
		}
	}
	return x, nil
}
func (r dr) Delete(c context.Context, id string) error { delete(r, id); return nil }

type br map[string]*bd.BudgetConfig

func (r br) FindByProjectID(c context.Context, id string) (*bd.BudgetConfig, error) {
	if b := r[id]; b != nil {
		return b, nil
	}
	return nil, bd.ErrBudgetConfigNotFound
}
func (r br) Save(c context.Context, b *bd.BudgetConfig) error { r[b.ProjectID()] = b; return nil }
func (r br) Delete(c context.Context, id string) error        { delete(r, id); return nil }

type ur map[string]*ud.User

func (r ur) Save(c context.Context, u *ud.User) error { r[u.ID()] = u; return nil }
func (r ur) FindByID(c context.Context, id string) (*ud.User, error) {
	if u := r[id]; u != nil {
		return u, nil
	}
	return nil, ud.ErrUserNotFound
}
func (r ur) FindByEmail(c context.Context, email string) (*ud.User, error) {
	for _, u := range r {
		if u.Email() == email {
			return u, nil
		}
	}
	return nil, ud.ErrUserNotFound
}
func (r ur) Delete(c context.Context, id string) error { delete(r, id); return nil }
func must(err error) {
	if err != nil {
		panic(err)
	}
}
func js(v any) string { b, e := json.Marshal(v); must(e); return string(b) }

type person struct{ u, e, p, a, d, token string }
type accessFixture struct {
	people                      []person
	publicAsset, publicDocument string
	h                           fasthttp.RequestHandler
	enterprises                 *er
	projects                    *mockRepo
	assets                      *mockDefenseAssetRepo
	documents                   dr
	budgets                     br
	users                       ur
}

func newAccessFixture(t *testing.T) accessFixture {
	c := context.Background()
	secret := "frc02-local-fixture-only"
	users := ur{}
	us := ua.NewUserService(users, secret, 1)
	erp := &er{es: map[string]*ed.Enterprise{}, links: map[string]bool{}}
	es := ea.NewEnterpriseService(erp)
	pr := newmockRepo()
	pr.memberships = erp.links
	ps := pa.NewDefenseProjectService(pr, es)
	ar := newmockDefenseAssetRepo()
	ar.memberships = erp.links
	as := aa.NewDefenseAssetService(ar, es)
	docs := dr{}
	ds := aa.NewDocumentService(docs, as)
	budgets := br{}
	bs := ba.NewBudgetService(budgets, ps)
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
		d, e := ds.Create(c, u.ID(), aa.CreateDocumentInput{AssetID: a.ID(), Name: "Document " + name, StorageKey: "fixture/" + name, DownloadURL: "https://storage.example.test/" + name})
		must(e)
		must(bs.UpdateBudgetConfig(c, u.ID(), p.ProjectID(), bd.BudgetModeUnlimited, 0))
		people = append(people, person{u.ID(), ent.ID(), p.ProjectID(), a.ID(), d.ID(), t})
	}
	public, e := as.Create(c, people[0].u, aa.CreateInput{Name: "Public reference", Category: ad.DefenseAssetCategoryRadar, CoverageType: ad.DefenseAssetCoverageCircle, EnterpriseID: &people[0].e})
	must(e)
	publicDoc, e := ds.Create(c, people[0].u, aa.CreateDocumentInput{AssetID: public.ID(), Name: "Public document", StorageKey: "fixture/public", DownloadURL: "https://storage.example.test/public"})
	must(e)
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
		func() *eu.EnterpriseController { return ec }, func() *pu.DefenseProjectController { return pc }, func() *au.DefenseAssetController { return ac }, func() *au.DocumentController { return dc }, func() *bu.BudgetController { return bc }, func() *ru.ReportController { return rc }, func() *uu.UserController { return uc }, func() *platformUi.ExampleController { return &platformUi.ExampleController{} },
	} {
		require.NoError(t, app.container.Provide(provider))
	}
	r := router.New()
	require.NoError(t, app.registerHandlers(r))
	h := middleware.NewAuthRequired(secret, []string{"^/api/v1/auth/register$", "^/api/v1/auth/login$", "^/api/v1/token_validate$"}, us).Process(r.Handler)
	return accessFixture{people, public.ID(), publicDoc.ID(), h, erp, pr, ar, docs, budgets, users}
}
func (f accessFixture) request(t *testing.T, actor person, method, path string, body any, want int) string {
	t.Helper()
	var x fasthttp.RequestCtx
	x.Init2(nil, nil, false)
	x.Request.SetRequestURI("/api/v1" + path)
	x.Request.Header.SetMethod(method)
	x.Request.Header.SetContentType("application/json")
	if actor.token != "" {
		x.Request.Header.Set("Authorization", "Bearer "+actor.token)
	}
	if body != nil {
		x.Request.SetBodyString(js(body))
	}
	f.h(&x)
	got := string(x.Response.Body())
	if x.Response.StatusCode() != want {
		t.Fatalf("%s %s: want %d got %d body=%s", method, path, want, x.Response.StatusCode(), got)
	}
	return got
}

// Removing service authorization makes these direct-ID and mutation assertions fail.
func TestFRC02TenantIsolation(t *testing.T) {
	f := newAccessFixture(t)
	for i, own := range f.people {
		other := f.people[1-i]
		t.Run([]string{"A", "B"}[i], func(t *testing.T) {
			t.Run("A_enterprises", func(t *testing.T) {
				body := f.request(t, own, "GET", "/enterprises", nil, 200)
				require.Contains(t, body, own.e)
				require.NotContains(t, body, other.e)
				f.request(t, own, "GET", "/enterprises?id="+own.e, nil, 200)
				f.request(t, own, "GET", "/enterprises?id="+other.e, nil, 404)
				f.request(t, own, "PUT", "/enterprises?id="+other.e, map[string]any{"name": "stolen"}, 404)
				f.request(t, own, "DELETE", "/enterprises?id="+own.e, nil, 403)
				f.request(t, own, "DELETE", "/enterprises?id="+other.e, nil, 404)
				f.request(t, own, "POST", "/enterprises/members", map[string]string{"userId": own.u, "enterpriseId": other.e}, 403)
				f.request(t, own, "DELETE", "/enterprises/members?userId="+other.u+"&enterpriseId="+other.e, nil, 403)
				require.False(t, f.enterprises.links[own.u+other.e])
				require.True(t, f.enterprises.links[other.u+other.e])
			})
			t.Run("B_C_projects", func(t *testing.T) {
				f.request(t, own, "GET", "/projects/get?id="+own.p, nil, 200)
				for _, path := range []string{"/projects/get", "/projects/export"} {
					f.request(t, own, "GET", path+"?id="+other.p, nil, 404)
				}
				f.request(t, own, "PUT", "/projects/update?id="+other.p, map[string]any{"name": "stolen", "version": -99}, 404)
				f.request(t, own, "DELETE", "/projects/delete?id="+other.p, nil, 404)
				f.request(t, own, "PUT", "/projects/update?id="+own.p, map[string]string{"enterpriseId": other.e}, 400)
				require.Equal(t, own.e, f.projects.projects[own.p].EnterpriseID())
				raw := map[string]any{"schemaVersion": 1, "projectName": "injected", "name": "injected", "enterpriseId": other.e, "baseObject": map[string]any{"id": other.e, "name": "foreign", "center": map[string]int{"lat": 0, "lng": 0}}}
				before := len(f.projects.projects)
				f.request(t, own, "POST", "/projects", map[string]string{"name": "injected", "enterpriseId": other.e, "projectJson": js(raw)}, 404)
				f.request(t, own, "POST", "/projects/import", map[string]string{"projectJson": js(raw)}, 404)
				f.request(t, own, "PUT", "/projects/update?id="+own.p, map[string]string{"projectJson": js(raw)}, 400)
				require.Equal(t, before, len(f.projects.projects))
				require.Equal(t, own.e, f.projects.projects[own.p].EnterpriseID())
				body := f.request(t, own, "GET", "/projects?limit=1&offset=0", nil, 200)
				require.Contains(t, body, own.p)
				require.NotContains(t, body, other.p)
				require.Contains(t, body, `"totalItems":1`)
				f.request(t, own, "GET", "/projects?enterpriseId="+other.e, nil, 404)
			})
			t.Run("D_E_financial_compare", func(t *testing.T) {
				for _, path := range []string{"/projects/budget", "/projects/cost", "/projects/report"} {
					f.request(t, own, "GET", path+"?id="+own.p, nil, 200)
					f.request(t, own, "GET", path+"?id="+other.p, nil, 404)
				}
				f.request(t, own, "PUT", "/projects/budget?id="+other.p, map[string]any{"budgetMode": "limited", "budgetAmountMln": 999}, 404)
				f.request(t, own, "POST", "/projects/budget/check?id="+other.p, map[string]any{"assetId": other.a, "quantity": 1}, 404)
				require.Equal(t, float64(0), f.budgets[other.p].BudgetAmountMln())
				for _, x := range []struct {
					a, b string
					want int
				}{{own.p, own.p, 200}, {own.p, other.p, 404}, {other.p, own.p, 404}, {other.p, other.p, 404}} {
					f.request(t, own, "GET", "/projects/compare?id1="+x.a+"&id2="+x.b, nil, x.want)
				}
			})
			t.Run("F_G_private_assets_documents", func(t *testing.T) {
				body := f.request(t, own, "GET", "/assets", nil, 200)
				require.Contains(t, body, own.a)
				require.Contains(t, body, f.publicAsset)
				require.NotContains(t, body, other.a)
				f.request(t, own, "GET", "/assets/get?id="+own.a, nil, 200)
				f.request(t, own, "GET", "/assets/get?id="+other.a, nil, 404)
				f.request(t, own, "PUT", "/assets/update?id="+other.a, map[string]any{"name": "stolen"}, 404)
				f.request(t, own, "DELETE", "/assets/delete?id="+other.a, nil, 404)
				f.request(t, own, "POST", "/assets", map[string]any{"name": "injected", "category": "radar", "coverageType": "circle", "enterpriseId": other.e}, 404)
				f.request(t, own, "GET", "/assets/documents/list?assetId="+other.a, nil, 404)
				for _, path := range []string{"/assets/documents/get", "/assets/documents/download"} {
					body = f.request(t, own, "GET", path+"?id="+other.d, nil, 404)
					require.NotContains(t, body, "storage.example.test")
					require.NotContains(t, body, "fixture/")
				}
				f.request(t, own, "DELETE", "/assets/documents/delete?id="+other.d, nil, 404)
				f.request(t, own, "POST", "/assets/documents", map[string]string{"assetId": other.a, "name": "injected", "storageKey": "fixture/injected", "ownerId": other.u}, 404)
				require.NotNil(t, f.documents[other.d])
				require.NotNil(t, f.assets.assets[other.a])
			})
			t.Run("J_public_catalog", func(t *testing.T) {
				f.request(t, own, "GET", "/assets/get?id="+f.publicAsset, nil, 200)
				f.request(t, own, "GET", "/assets/documents/list?assetId="+f.publicAsset, nil, 200)
				for _, path := range []string{"/assets/documents/get", "/assets/documents/download"} {
					f.request(t, own, "GET", path+"?id="+f.publicDocument, nil, 200)
				}
				f.request(t, own, "POST", "/assets", map[string]any{"name": "global", "category": "radar", "coverageType": "circle", "isPublic": true}, 403)
				f.request(t, own, "PUT", "/assets/update?id="+f.publicAsset, map[string]string{"name": "stolen"}, 403)
				f.request(t, own, "DELETE", "/assets/delete?id="+f.publicAsset, nil, 403)
				f.request(t, own, "POST", "/assets/documents", map[string]string{"assetId": f.publicAsset, "name": "global", "storageKey": "fixture/global"}, 403)
				f.request(t, own, "DELETE", "/assets/documents/delete?id="+f.publicDocument, nil, 403)
			})
		})
	}
}
func TestFRC02RevocationSameJWT(t *testing.T) {
	f := newAccessFixture(t)
	a := f.people[0]
	f.request(t, a, "GET", "/projects/get?id="+a.p, nil, 200)
	require.NoError(t, f.enterprises.RemoveUserFromEnterprise(context.Background(), a.u, a.e))
	for _, path := range []string{"/enterprises?id=" + a.e, "/projects/get?id=" + a.p, "/projects/export?id=" + a.p, "/projects/budget?id=" + a.p, "/projects/cost?id=" + a.p, "/projects/report?id=" + a.p, "/projects/compare?id1=" + a.p + "&id2=" + a.p, "/assets/get?id=" + a.a, "/assets/documents/download?id=" + a.d} {
		f.request(t, a, "GET", path, nil, 404)
	}
	f.request(t, a, "PUT", "/projects/update?id="+a.p, map[string]string{"name": "revoked"}, 404)
	f.request(t, a, "DELETE", "/projects/delete?id="+a.p, nil, 404)
	f.request(t, a, "PUT", "/projects/budget?id="+a.p, map[string]any{"budgetMode": "limited", "budgetAmountMln": 5}, 404)
	require.NotNil(t, f.projects.projects[a.p])
}
func TestFRC02OrphansAndMissing(t *testing.T) {
	f := newAccessFixture(t)
	a, b := f.people[0], f.people[1]
	f.projects.projects[b.p].SetEnterpriseID("")
	f.assets.assets[b.a].SetEnterpriseID(nil)
	for _, path := range []string{"/projects/get", "/projects/export", "/projects/cost", "/projects/report", "/projects/budget"} {
		missing := f.request(t, a, "GET", path+"?id=99999999-9999-4999-8999-999999999999", nil, 404)
		foreign := f.request(t, a, "GET", path+"?id="+b.p, nil, 404)
		require.JSONEq(t, missing, foreign)
	}
	for _, path := range []string{"/assets/get?id=" + b.a, "/assets/documents/get?id=" + b.d, "/assets/documents/download?id=" + b.d, "/assets/documents/list?assetId=" + b.a} {
		f.request(t, a, "GET", path, nil, 404)
	}
	delete(f.assets.assets, b.a)
	f.request(t, a, "GET", "/assets/documents/download?id="+b.d, nil, 404)
	f.request(t, person{}, "GET", "/assets/get?id="+f.publicAsset, nil, 401)
}

func (m *mockRepo) FindAllByUserID(ctx context.Context, userID string, limit, offset int) ([]*pd.DefenseProject, int64, error) {
	var items []*pd.DefenseProject
	for _, p := range m.projects {
		if m.memberships[userID+p.EnterpriseID()] {
			items = append(items, p)
		}
	}
	total := int64(len(items))
	if offset >= len(items) {
		return []*pd.DefenseProject{}, total, nil
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end], total, nil
}

func TestFRC02IdentityAndAccessFailure(t *testing.T) {
	f := newAccessFixture(t)
	a := f.people[0]
	f.enterprises.accessErr = errors.New("synthetic membership store unavailable")
	for _, path := range []string{"/enterprises?id=" + a.e, "/projects/get?id=" + a.p, "/projects/budget?id=" + a.p, "/projects/cost?id=" + a.p, "/projects/report?id=" + a.p, "/assets/get?id=" + a.a, "/assets/documents/download?id=" + a.d} {
		body := f.request(t, a, "GET", path, nil, 500)
		require.NotContains(t, body, "synthetic")
	}
	f.enterprises.accessErr = nil
	_, err := pa.NewDefenseProjectService(f.projects, ea.NewEnterpriseService(f.enterprises)).GetProject(context.Background(), "", a.p)
	require.ErrorIs(t, err, auth.ErrIdentityRequired)
	require.NoError(t, f.users.Delete(context.Background(), a.u))
	f.request(t, a, "GET", "/projects/get?id="+a.p, nil, 401)
}

func TestFRC02MemberCRUDAndPagination(t *testing.T) {
	f := newAccessFixture(t)
	a := f.people[0]
	f.request(t, a, "POST", "/enterprises", map[string]string{"name": "ordinary-provisioning"}, 403)
	f.request(t, a, "PUT", "/enterprises?id="+a.e, map[string]string{"name": "Member metadata"}, 200)
	raw := map[string]any{"schemaVersion": 1, "projectName": "Member project", "enterpriseId": a.e, "baseObject": map[string]any{"id": a.e, "name": "Member facility", "center": map[string]int{"lat": 0, "lng": 0}}, "layers": []any{}, "assetLibrary": []any{}, "placedObjects": []any{}}
	var created struct {
		ProjectID string `json:"projectId"`
	}
	require.NoError(t, json.Unmarshal([]byte(f.request(t, a, "POST", "/projects", map[string]string{"name": "New configuration", "enterpriseId": a.e, "projectJson": js(raw)}, 200)), &created))
	require.NotEmpty(t, created.ProjectID)
	f.request(t, a, "PUT", "/projects/update?id="+created.ProjectID, map[string]any{"name": "Updated", "version": 1}, 200)
	f.request(t, a, "PUT", "/projects/update?id="+created.ProjectID, map[string]any{"name": "Stale", "version": 1}, 409)
	f.request(t, a, "PUT", "/projects/update?id="+created.ProjectID, map[string]string{"enterpriseId": ""}, 400)
	exported := f.request(t, a, "GET", "/projects/export?id="+created.ProjectID, nil, 200)
	var imported struct {
		ProjectID string `json:"projectId"`
	}
	require.NoError(t, json.Unmarshal([]byte(f.request(t, a, "POST", "/projects/import", map[string]string{"projectJson": exported}, 200)), &imported))
	require.NotEmpty(t, imported.ProjectID)
	body := f.request(t, a, "GET", "/projects?limit=1&offset=1", nil, 200)
	var page struct {
		Items []any `json:"items"`
		Total int   `json:"totalItems"`
	}
	require.NoError(t, json.Unmarshal([]byte(body), &page))
	require.Len(t, page.Items, 1)
	require.Equal(t, 3, page.Total)
	f.request(t, a, "DELETE", "/projects/delete?id="+created.ProjectID, nil, 200)
	f.request(t, a, "GET", "/projects/get?id="+created.ProjectID, nil, 404)
	f.request(t, a, "PUT", "/assets/update?id="+a.a, map[string]string{"name": "Updated private"}, 200)
	f.request(t, a, "PUT", "/assets/update?id="+a.a, map[string]bool{"isPublic": true}, 403)
	var doc struct {
		ID      string `json:"id"`
		OwnerID string `json:"ownerId"`
	}
	require.NoError(t, json.Unmarshal([]byte(f.request(t, a, "POST", "/assets/documents", map[string]string{"assetId": a.a, "name": "Member document", "storageKey": "fixture/member", "ownerId": f.people[1].u}, 201)), &doc))
	require.Equal(t, a.u, doc.OwnerID)
	require.NotEmpty(t, doc.ID)
	f.request(t, a, "GET", "/assets/documents/get?id="+doc.ID, nil, 200)
	f.request(t, a, "DELETE", "/assets/documents/delete?id="+doc.ID, nil, 200)
	f.request(t, a, "GET", "/assets/documents/get?id="+doc.ID, nil, 404)
	f.request(t, a, "DELETE", "/assets/delete?id="+a.a, nil, 200)
	f.request(t, a, "GET", "/assets/get?id="+a.a, nil, 404)
}

func TestFRC02ForeignMissingAndInvalidIdentity(t *testing.T) {
	f := newAccessFixture(t)
	a, b := f.people[0], f.people[1]
	for _, resource := range []struct{ path, id string }{{"/enterprises", b.e}, {"/projects/get", b.p}, {"/projects/export", b.p}, {"/projects/budget", b.p}, {"/projects/cost", b.p}, {"/projects/report", b.p}, {"/assets/get", b.a}, {"/assets/documents/get", b.d}, {"/assets/documents/download", b.d}} {
		missing := f.request(t, a, "GET", resource.path+"?id="+uuid.NewString(), nil, 404)
		foreign := f.request(t, a, "GET", resource.path+"?id="+resource.id, nil, 404)
		require.JSONEq(t, missing, foreign)
	}
	expired, err := auth.GenerateToken(a.u, "A@example.test", "frc02-local-fixture-only", -1)
	require.NoError(t, err)
	empty, err := auth.GenerateToken("", "", "frc02-local-fixture-only", 1)
	require.NoError(t, err)
	for _, token := range []string{"", "invalid", expired, empty} {
		f.request(t, person{token: token}, "GET", "/projects/get?id="+a.p, nil, 401)
	}
	f.assets.assets[f.publicAsset].SetEnterpriseID(nil)
	f.request(t, a, "GET", "/assets/get?id="+f.publicAsset, nil, 200)
	f.request(t, b, "GET", "/assets/documents/download?id="+f.publicDocument, nil, 200)
}

func TestFRC02ProductionDependencyGraph(t *testing.T) {
	// Resolve the actual DI graph without opening a configured database or applying migrations.
	app := NewApplication(dig.New(dig.DryRun(true)))
	app.registerCoreDependencies()
	app.provideDependencies()
	_, err := app.getWepApp()
	require.NoError(t, err)
}
