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
}

func (m *mockService) Import(ctx context.Context, rawJSON string) (*domain.DefenseProject, error) {
	return m.importProject, m.importErr
}

func (m *mockService) Export(ctx context.Context, projectID string) (string, error) {
	return m.exportJSON, m.exportErr
}

func validProject() *domain.DefenseProject {
	now := time.Now().UTC()
	project, _ := domain.NewDefenseProject(
		"550e8400-e29b-41d4-a716-446655440000",
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
