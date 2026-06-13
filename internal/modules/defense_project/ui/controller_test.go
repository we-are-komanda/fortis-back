package ui

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/fortis/backend/internal/modules/defense_project/domain"
)

// mockService — мок сервиса для тестирования контроллера.
type mockService struct {
	importProject *domain.DefenseProject
	importErr     error
	exportJSON    string
	exportErr     error
	projects      []*domain.DefenseProject
	totalItems    int64
	projectByID   *domain.DefenseProject
	crudErr       error
}

func (m *mockService) Import(ctx context.Context, rawJSON string) (*domain.DefenseProject, error) {
	return m.importProject, m.importErr
}

func (m *mockService) Export(ctx context.Context, projectID string) (string, error) {
	return m.exportJSON, m.exportErr
}

func (m *mockService) CreateFromJSON(ctx context.Context, name, enterpriseID, rawJSON string) (*domain.DefenseProject, error) {
	return m.importProject, m.importErr
}

func (m *mockService) ListProjects(ctx context.Context, enterpriseID string, limit, offset int) ([]*domain.DefenseProject, int64, error) {
	return m.projects, m.totalItems, m.crudErr
}

func (m *mockService) GetProject(ctx context.Context, id string) (*domain.DefenseProject, error) {
	return m.projectByID, m.crudErr
}

func (m *mockService) UpdateProject(ctx context.Context, id, name, enterpriseID string) (*domain.DefenseProject, error) {
	if m.crudErr != nil {
		return nil, m.crudErr
	}
	if m.projectByID == nil {
		return nil, domain.ErrProjectNotFound
	}
	return m.projectByID, nil
}

func (m *mockService) DeleteProject(ctx context.Context, id string) error {
	return m.crudErr
}

func validProject() *domain.DefenseProject {
	now := time.Now().UTC()
	project, _ := domain.NewDefenseProject(
		"550e8400-e29b-41d4-a716-446655440000",
		"Моя конфигурация",
		"",
		"Тестовый проект",
		domain.NewProtectedObject("obj-1", "Объект Альфа", domain.NewCoordinates(55.75, 37.62)),
		[]domain.EditableDefenseLayer{},
		[]domain.DefenseAsset{},
		[]domain.PlacedDefenseObject{},
		nil, nil, nil,
		domain.DefenseProjectModeView,
		domain.DefenseProjectSourceCustom,
		nil,
		now,
	)
	return project
}

func TestDefenseProjectController_Import_Success(t *testing.T) {
	svc := &mockService{
		importProject: validProject(),
	}
	ctrl := NewDefenseProjectController(svc)

	reqBody := `{"projectJson": "{\"schemaVersion\":1,\"projectName\":\"Test\"}"}`
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetBody([]byte(reqBody))

	ctrl.Import(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected status 200, got %d", ctx.Response.StatusCode())
	}

	var respBody struct {
		ProjectID   string `json:"projectId"`
		ProjectName string `json:"projectName"`
		Version     int    `json:"version"`
		UpdatedAt   string `json:"updatedAt"`
	}
	if err := json.Unmarshal(ctx.Response.Body(), &respBody); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if respBody.ProjectID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected projectId, got %q", respBody.ProjectID)
	}
	if respBody.ProjectName != "Тестовый проект" {
		t.Errorf("expected projectName 'Тестовый проект', got %q", respBody.ProjectName)
	}
	if respBody.Version != 1 {
		t.Errorf("expected version 1, got %d", respBody.Version)
	}
}

func TestDefenseProjectController_Import_EmptyBody(t *testing.T) {
	ctrl := NewDefenseProjectController(&mockService{})

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetBody([]byte(`{}`))

	ctrl.Import(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status 400, got %d", ctx.Response.StatusCode())
	}
}

func TestDefenseProjectController_Import_BadJSON(t *testing.T) {
	ctrl := NewDefenseProjectController(&mockService{})

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetBody([]byte(`not json`))

	ctrl.Import(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status 400, got %d", ctx.Response.StatusCode())
	}
}

func TestDefenseProjectController_Import_InvalidSchemaVersion(t *testing.T) {
	svc := &mockService{
		importErr: domain.ErrInvalidSchemaVersion,
	}
	ctrl := NewDefenseProjectController(svc)

	reqBody := `{"projectJson": "{\"schemaVersion\":999}"}`
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetBody([]byte(reqBody))

	ctrl.Import(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status 400, got %d", ctx.Response.StatusCode())
	}
}

func TestDefenseProjectController_Import_InvalidProjectData(t *testing.T) {
	svc := &mockService{
		importErr: domain.ErrInvalidProjectData,
	}
	ctrl := NewDefenseProjectController(svc)

	reqBody := `{"projectJson": "{}"}`
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetBody([]byte(reqBody))

	ctrl.Import(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status 400, got %d", ctx.Response.StatusCode())
	}
}

func TestDefenseProjectController_Import_VersionConflict(t *testing.T) {
	svc := &mockService{
		importErr: domain.ErrVersionConflict,
	}
	ctrl := NewDefenseProjectController(svc)

	reqBody := `{"projectJson": "{\"schemaVersion\":1,\"projectName\":\"Test\"}"}`
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetBody([]byte(reqBody))

	ctrl.Import(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusConflict {
		t.Errorf("expected status 409, got %d", ctx.Response.StatusCode())
	}
}

func TestDefenseProjectController_Export_Success(t *testing.T) {
	svc := &mockService{
		exportJSON: `{"schemaVersion":1,"projectId":"test-id","projectName":"Test"}`,
	}
	ctrl := NewDefenseProjectController(svc)

	ctx := &fasthttp.RequestCtx{}
	ctx.QueryArgs().Set("id", "test-id")

	ctrl.Export(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected status 200, got %d", ctx.Response.StatusCode())
	}

	body := string(ctx.Response.Body())
	if body != svc.exportJSON {
		t.Errorf("expected body %q, got %q", svc.exportJSON, body)
	}

	contentType := string(ctx.Response.Header.ContentType())
	if contentType != "application/json" {
		t.Errorf("expected content-type application/json, got %q", contentType)
	}
}

func TestDefenseProjectController_Export_MissingID(t *testing.T) {
	ctrl := NewDefenseProjectController(&mockService{})

	ctx := &fasthttp.RequestCtx{}
	// No id param

	ctrl.Export(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status 400, got %d", ctx.Response.StatusCode())
	}
}

func TestDefenseProjectController_Export_NotFound(t *testing.T) {
	svc := &mockService{
		exportErr: domain.ErrProjectNotFound,
	}
	ctrl := NewDefenseProjectController(svc)

	ctx := &fasthttp.RequestCtx{}
	ctx.QueryArgs().Set("id", "nonexistent")

	ctrl.Export(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusNotFound {
		t.Errorf("expected status 404, got %d", ctx.Response.StatusCode())
	}
}

func TestDefenseProjectController_Export_InternalError(t *testing.T) {
	svc := &mockService{
		exportErr: errors.New("unexpected error"),
	}
	ctrl := NewDefenseProjectController(svc)

	ctx := &fasthttp.RequestCtx{}
	ctx.QueryArgs().Set("id", "some-id")

	ctrl.Export(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", ctx.Response.StatusCode())
	}
}

// ---- CRUD tests ----

func TestDefenseProjectController_Create_Success(t *testing.T) {
	svc := &mockService{
		importProject: validProject(),
	}
	ctrl := NewDefenseProjectController(svc)

	reqBody := `{"name":"Моя конфигурация","projectJson":"{\"schemaVersion\":1,\"projectName\":\"Test\"}"}`
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetBody([]byte(reqBody))

	ctrl.Create(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected status 200, got %d", ctx.Response.StatusCode())
	}

	var resp ProjectResponse
	if err := json.Unmarshal(ctx.Response.Body(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.ProjectID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected projectId, got %q", resp.ProjectID)
	}
	if resp.Name != "Моя конфигурация" {
		t.Errorf("expected name 'Моя конфигурация', got %q", resp.Name)
	}
	if resp.Version != 1 {
		t.Errorf("expected version 1, got %d", resp.Version)
	}
}

func TestDefenseProjectController_Create_VersionConflict(t *testing.T) {
	svc := &mockService{
		importErr: domain.ErrVersionConflict,
	}
	ctrl := NewDefenseProjectController(svc)

	reqBody := `{"name":"Моя конфигурация","projectJson":"{\"schemaVersion\":1,\"projectName\":\"Test\"}"}`
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetBody([]byte(reqBody))

	ctrl.Create(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusConflict {
		t.Errorf("expected status 409, got %d", ctx.Response.StatusCode())
	}
}

func TestDefenseProjectController_Create_EmptyName(t *testing.T) {
	ctrl := NewDefenseProjectController(&mockService{})

	reqBody := `{"name":"","projectJson":"{}"}`
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetBody([]byte(reqBody))

	ctrl.Create(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status 400, got %d", ctx.Response.StatusCode())
	}
}

func TestDefenseProjectController_List_Success(t *testing.T) {
	svc := &mockService{
		projects:   []*domain.DefenseProject{validProject()},
		totalItems: 1,
	}
	ctrl := NewDefenseProjectController(svc)

	ctx := &fasthttp.RequestCtx{}

	ctrl.List(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected status 200, got %d", ctx.Response.StatusCode())
	}

	var resp ProjectListResponse
	if err := json.Unmarshal(ctx.Response.Body(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.TotalItems != 1 {
		t.Errorf("expected totalItems 1, got %d", resp.TotalItems)
	}
	if len(resp.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(resp.Items))
	}
	if resp.Items[0].Name != "Моя конфигурация" {
		t.Errorf("expected name 'Моя конфигурация', got %q", resp.Items[0].Name)
	}
	if resp.Items[0].Version != 1 {
		t.Errorf("expected version 1, got %d", resp.Items[0].Version)
	}
}

func TestDefenseProjectController_Get_Success(t *testing.T) {
	svc := &mockService{
		projectByID: validProject(),
	}
	ctrl := NewDefenseProjectController(svc)

	ctx := &fasthttp.RequestCtx{}
	ctx.QueryArgs().Set("id", "550e8400-e29b-41d4-a716-446655440000")

	ctrl.Get(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected status 200, got %d", ctx.Response.StatusCode())
	}

	var resp ProjectResponse
	if err := json.Unmarshal(ctx.Response.Body(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.ProjectID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected projectId, got %q", resp.ProjectID)
	}
	if resp.Version != 1 {
		t.Errorf("expected version 1, got %d", resp.Version)
	}
}

func TestDefenseProjectController_Get_MissingID(t *testing.T) {
	ctrl := NewDefenseProjectController(&mockService{})

	ctx := &fasthttp.RequestCtx{}

	ctrl.Get(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status 400, got %d", ctx.Response.StatusCode())
	}
}

func TestDefenseProjectController_Get_NotFound(t *testing.T) {
	svc := &mockService{
		crudErr: domain.ErrProjectNotFound,
	}
	ctrl := NewDefenseProjectController(svc)

	ctx := &fasthttp.RequestCtx{}
	ctx.QueryArgs().Set("id", "nonexistent")

	ctrl.Get(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusNotFound {
		t.Errorf("expected status 404, got %d", ctx.Response.StatusCode())
	}
}

func TestDefenseProjectController_Update_Success(t *testing.T) {
	svc := &mockService{
		projectByID: validProject(),
	}
	ctrl := NewDefenseProjectController(svc)

	reqBody := `{"name":"Новое имя"}`
	ctx := &fasthttp.RequestCtx{}
	ctx.QueryArgs().Set("id", "550e8400-e29b-41d4-a716-446655440000")
	ctx.Request.SetBody([]byte(reqBody))

	ctrl.Update(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected status 200, got %d", ctx.Response.StatusCode())
	}

	var resp ProjectResponse
	if err := json.Unmarshal(ctx.Response.Body(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.ProjectID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected projectId, got %q", resp.ProjectID)
	}
	if resp.Version != 1 {
		t.Errorf("expected version 1, got %d", resp.Version)
	}
}

func TestDefenseProjectController_Update_VersionConflict(t *testing.T) {
	svc := &mockService{
		crudErr: domain.ErrVersionConflict,
	}
	ctrl := NewDefenseProjectController(svc)

	reqBody := `{"name":"Новое имя"}`
	ctx := &fasthttp.RequestCtx{}
	ctx.QueryArgs().Set("id", "550e8400-e29b-41d4-a716-446655440000")
	ctx.Request.SetBody([]byte(reqBody))

	ctrl.Update(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusConflict {
		t.Errorf("expected status 409, got %d", ctx.Response.StatusCode())
	}
}

func TestDefenseProjectController_Delete_Success(t *testing.T) {
	svc := &mockService{}
	ctrl := NewDefenseProjectController(svc)

	ctx := &fasthttp.RequestCtx{}
	ctx.QueryArgs().Set("id", "550e8400-e29b-41d4-a716-446655440000")

	ctrl.Delete(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected status 200, got %d", ctx.Response.StatusCode())
	}
}

func TestDefenseProjectController_Delete_MissingID(t *testing.T) {
	ctrl := NewDefenseProjectController(&mockService{})

	ctx := &fasthttp.RequestCtx{}

	ctrl.Delete(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status 400, got %d", ctx.Response.StatusCode())
	}
}

func TestDefenseProjectController_Delete_NotFound(t *testing.T) {
	svc := &mockService{
		crudErr: domain.ErrProjectNotFound,
	}
	ctrl := NewDefenseProjectController(svc)

	ctx := &fasthttp.RequestCtx{}
	ctx.QueryArgs().Set("id", "nonexistent")

	ctrl.Delete(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusNotFound {
		t.Errorf("expected status 404, got %d", ctx.Response.StatusCode())
	}
}
