package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/fasthttp/router"
	ba "github.com/fortis/backend/internal/modules/budget/application"
	bd "github.com/fortis/backend/internal/modules/budget/domain"
	bu "github.com/fortis/backend/internal/modules/budget/ui"
	aa "github.com/fortis/backend/internal/modules/defense_asset/application"
	ad "github.com/fortis/backend/internal/modules/defense_asset/domain"
	au "github.com/fortis/backend/internal/modules/defense_asset/ui"
	pa "github.com/fortis/backend/internal/modules/defense_project/application"
	pd "github.com/fortis/backend/internal/modules/defense_project/domain"
	pu "github.com/fortis/backend/internal/modules/defense_project/ui"
	demoUi "github.com/fortis/backend/internal/modules/demo_request/ui"
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
	"mime/multipart"
	"net/textproto"
	"strings"
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
	project.SetCostCalculationVersion(bd.CostCalculationV1)
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
func (r dr) CommitDocument(ctx context.Context, doc *ad.Document, actor, action string) error {
	return r.Save(ctx, doc)
}
func (r dr) DeleteDocument(ctx context.Context, id, actor string) error { return r.Delete(ctx, id) }
func (r dr) ListDocuments(ctx context.Context, id string, limit, offset int) ([]*ad.Document, int64, error) {
	items, err := r.FindByAssetID(ctx, id)
	total := int64(len(items))
	if offset >= len(items) {
		return []*ad.Document{}, total, err
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end], total, err
}

// Synthetic-only pipeline: real storage/scanner acceptance is exercised by FRC09.
type frc02DocumentBytes map[string][]byte

func (r frc02DocumentBytes) Ready() bool { return true }
func (r frc02DocumentBytes) Put(_ context.Context, id string, body []byte) (string, error) {
	r[id] = append([]byte(nil), body...)
	return id, nil
}
func (r frc02DocumentBytes) Read(_ context.Context, id string) ([]byte, error) { return r[id], nil }
func (r frc02DocumentBytes) Remove(_ context.Context, id string) error         { delete(r, id); return nil }

type frc02CleanScanner struct{}

func (frc02CleanScanner) Ready() bool                        { return true }
func (frc02CleanScanner) Scan(context.Context, string) error { return nil }

type frc02Multipart struct {
	body        []byte
	contentType string
}

func frc02Upload(asset, forgedOwner string) frc02Multipart {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	must(writer.WriteField("assetId", asset))
	if forgedOwner != "" {
		must(writer.WriteField("ownerId", forgedOwner))
	}
	part, err := writer.CreatePart(textproto.MIMEHeader{"Content-Disposition": {`form-data; name="file"; filename="synthetic.txt"`}, "Content-Type": {"text/plain"}})
	must(err)
	_, err = part.Write([]byte("Synthetic document fixture"))
	must(err)
	must(writer.Close())
	return frc02Multipart{body.Bytes(), writer.FormDataContentType()}
}

type br map[string]*bd.BudgetConfig

type costMemoryRepo map[string]*bd.CostProjection

func (r costMemoryRepo) FindCostProjection(_ context.Context, id string, version int, calculator string) (*bd.CostProjection, error) {
	for _, p := range r {
		i := p.Identity()
		if i.ProjectID() == id && i.ProjectVersion() == version && i.CalculationVersion() == calculator {
			return p, nil
		}
	}
	return nil, bd.ErrCostProjectionNotFound
}
func (r costMemoryRepo) SaveCostProjection(ctx context.Context, p *bd.CostProjection) (*bd.CostProjection, error) {
	i := p.Identity()
	if saved, err := r.FindCostProjection(ctx, i.ProjectID(), i.ProjectVersion(), i.CalculationVersion()); err == nil {
		return saved, nil
	}
	r[js([]any{i.ProjectID(), i.ProjectVersion(), i.CalculationVersion()})] = p
	return p, nil
}

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
	ds := aa.NewDocumentService(docs, as, aa.DocumentPipeline{Storage: frc02DocumentBytes{}, Scanner: frc02CleanScanner{}})
	budgets := br{}
	bs := ba.NewBudgetService(budgets, ps, costMemoryRepo{})
	rs := ra.NewReportService(ps, bs, nil, ru.RenderReport)
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
		initialBudget, budgetErr := bd.NewBudgetConfig(p.ProjectID(), bd.BudgetModeUnlimited, 0)
		must(budgetErr)
		must(budgets.Save(c, initialBudget))
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
	h := r.Handler
	return accessFixture{people, public.ID(), publicDoc.ID(), h, erp, pr, ar, docs, budgets, users}
}
func (f accessFixture) request(t *testing.T, actor person, method, path string, body any, want int) string {
	t.Helper()
	// This legacy access matrix always addresses a fixed revision for reports.
	if strings.HasPrefix(path, "/projects/report?id=") {
		path += "&projectVersion=1&hideCost=false"
	}
	var x fasthttp.RequestCtx
	x.Init2(nil, nil, false)
	x.Request.SetRequestURI("/api/v1" + path)
	x.Request.Header.SetMethod(method)
	x.Request.Header.SetContentType("application/json")
	if actor.token != "" {
		x.Request.Header.Set("Authorization", "Bearer "+actor.token)
	}
	if payload, ok := body.(frc02Multipart); ok {
		x.Request.Header.SetContentType(payload.contentType)
		x.Request.SetBody(payload.body)
	} else if body != nil {
		x.Request.SetBodyString(js(body))
	}
	f.h(&x)
	got := string(x.Response.Body())
	if x.Response.StatusCode() != want {
		t.Fatalf("%s %s: want %d got %d body=%s", method, path, want, x.Response.StatusCode(), got)
	}
	return got
}

// Shared access deliberately ignores historical accounts and memberships.
func TestSharedWorkspaceAnonymousAccess(t *testing.T) {
	f := newAccessFixture(t)
	for _, actor := range []person{{}, {token: "invalid"}, f.people[0]} {
		for _, p := range f.people {
			for _, path := range []string{"/enterprises?id=" + p.e, "/projects/get?id=" + p.p, "/projects/export?id=" + p.p, "/projects/budget?id=" + p.p, "/projects/cost?id=" + p.p, "/assets/get?id=" + p.a, "/assets/documents/get?id=" + p.d} {
				f.request(t, actor, "GET", path, nil, 200)
			}
			require.Contains(t, f.request(t, actor, "GET", "/projects", nil, 200), p.p)
			require.Contains(t, f.request(t, actor, "GET", "/assets", nil, 200), p.a)
		}
	}
	for _, path := range []string{"/auth/me", "/auth/login", "/auth/register", "/enterprises/members"} {
		f.request(t, person{}, "GET", path, nil, 404)
	}
	f.request(t, person{}, "GET", "/projects/get?id="+uuid.NewString(), nil, 404)
	f.request(t, person{}, "PUT", "/assets/update?id="+f.publicAsset, map[string]string{"name": "Shared catalog"}, 200)
}

func (m *mockRepo) FindAllByUserID(ctx context.Context, userID string, limit, offset int) ([]*pd.DefenseProject, int64, error) {
	var items []*pd.DefenseProject
	for _, p := range m.projects {
		items = append(items, p)
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

func TestFRC02MemberCRUDAndPagination(t *testing.T) {
	f := newAccessFixture(t)
	a := f.people[0]
	a.token = ""
	f.request(t, a, "POST", "/enterprises", map[string]string{"name": "ordinary-provisioning", "status": "active"}, 200)
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
	require.Equal(t, 4, page.Total)
	f.request(t, a, "DELETE", "/projects/delete?id="+created.ProjectID, nil, 200)
	f.request(t, a, "GET", "/projects/get?id="+created.ProjectID, nil, 404)
	f.request(t, a, "PUT", "/assets/update?id="+a.a, map[string]string{"name": "Updated private"}, 200)
	f.request(t, a, "PUT", "/assets/update?id="+a.a, map[string]bool{"isPublic": true}, 200)
	var doc struct {
		ID      string `json:"id"`
		OwnerID string `json:"ownerId"`
	}
	f.request(t, a, "POST", "/assets/documents", map[string]string{"assetId": a.a, "name": "Forged metadata", "storageKey": "fixture/member", "ownerId": f.people[1].u}, 400)
	f.request(t, a, "POST", "/assets/documents", frc02Upload(a.a, f.people[1].u), 400)
	require.NoError(t, json.Unmarshal([]byte(f.request(t, a, "POST", "/assets/documents", frc02Upload(a.a, ""), 201)), &doc))
	require.Equal(t, "00000000-0000-0000-0000-000000000000", doc.OwnerID)
	require.NotEmpty(t, doc.ID)
	f.request(t, a, "GET", "/assets/documents/get?id="+doc.ID, nil, 200)
	f.request(t, a, "DELETE", "/assets/documents/delete?id="+doc.ID, nil, 200)
	f.request(t, a, "GET", "/assets/documents/get?id="+doc.ID, nil, 404)
	f.request(t, a, "DELETE", "/assets/delete?id="+a.a, nil, 200)
	f.request(t, a, "GET", "/assets/get?id="+a.a, nil, 404)
}

func TestFRC02ProductionDependencyGraph(t *testing.T) {
	// Resolve the actual DI graph without opening a configured database or applying migrations.
	app := NewApplication(dig.New(dig.DryRun(true)))
	app.registerCoreDependencies()
	app.provideDependencies()
	_, err := app.getWepApp()
	require.NoError(t, err)
}
