package ui

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/fortis/backend/internal/modules/enterprise/domain"
)

// mockService — мок сервиса для тестирования контроллера.
type mockService struct {
	createEnterprise *domain.Enterprise
	createErr        error
	getEnterprise    *domain.Enterprise
	getErr           error
	listEnterprises  []*domain.Enterprise
	listTotal        int64
	listErr          error
	updateEnterprise *domain.Enterprise
	updateErr        error
	deleteErr        error
}

func (m *mockService) Create(ctx context.Context, name, address string, status domain.EnterpriseStatus, latitude, longitude float64) (*domain.Enterprise, error) {
	return m.createEnterprise, m.createErr
}

func (m *mockService) Get(ctx context.Context, id string) (*domain.Enterprise, error) {
	return m.getEnterprise, m.getErr
}

func (m *mockService) List(ctx context.Context, limit, offset int) ([]*domain.Enterprise, int64, error) {
	return m.listEnterprises, m.listTotal, m.listErr
}

func (m *mockService) Update(ctx context.Context, id, name, address string, status domain.EnterpriseStatus, latitude, longitude float64) (*domain.Enterprise, error) {
	return m.updateEnterprise, m.updateErr
}

func (m *mockService) Delete(ctx context.Context, id string) error {
	return m.deleteErr
}

func (m *mockService) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Enterprise, int64, error) {
	return m.listEnterprises, m.listTotal, m.listErr
}

func (m *mockService) CheckUserAccess(ctx context.Context, userID, enterpriseID string) error {
	return nil
}

func (m *mockService) AddUser(ctx context.Context, userID, enterpriseID string) error {
	return nil
}

func (m *mockService) RemoveUser(ctx context.Context, userID, enterpriseID string) error {
	return nil
}

func validEnterprise() *domain.Enterprise {
	now := time.Now().UTC()
	e, _ := domain.NewEnterprise(
		"550e8400-e29b-41d4-a716-446655440000",
		"Тестовое предприятие",
		"г. Москва, ул. Ленина, д. 1",
		domain.EnterpriseStatusActive,
		55.75,
		37.62,
		now,
		now,
	)
	return e
}

func TestCreate_Success(t *testing.T) {
	svc := &mockService{
		createEnterprise: validEnterprise(),
	}
	ctrl := NewEnterpriseController(svc)

	reqBody := `{"name":"Тестовое предприятие","address":"г. Москва, ул. Ленина, д. 1","status":"active","latitude":55.75,"longitude":37.62}`
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetBody([]byte(reqBody))

	ctrl.Create(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected status 200, got %d", ctx.Response.StatusCode())
	}

	var resp EnterpriseResponse
	if err := json.Unmarshal(ctx.Response.Body(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.ID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected ID, got %q", resp.ID)
	}
	if resp.Name != "Тестовое предприятие" {
		t.Errorf("expected name 'Тестовое предприятие', got %q", resp.Name)
	}
}

func TestCreate_EmptyBody(t *testing.T) {
	ctrl := NewEnterpriseController(&mockService{})

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetBody([]byte(`{}`))

	ctrl.Create(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status 400, got %d", ctx.Response.StatusCode())
	}
}

func TestCreate_BadJSON(t *testing.T) {
	ctrl := NewEnterpriseController(&mockService{})

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetBody([]byte(`not json`))

	ctrl.Create(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status 400, got %d", ctx.Response.StatusCode())
	}
}

func TestCreate_EmptyName(t *testing.T) {
	ctrl := NewEnterpriseController(&mockService{})

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetBody([]byte(`{"name":""}`))

	ctrl.Create(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status 400, got %d", ctx.Response.StatusCode())
	}
}

func TestCreate_InvalidStatus(t *testing.T) {
	svc := &mockService{
		createErr: domain.ErrInvalidEnterpriseStatus,
	}
	ctrl := NewEnterpriseController(svc)

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetBody([]byte(`{"name":"Test","status":"invalid","latitude":55.75,"longitude":37.62}`))

	ctrl.Create(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status 400, got %d", ctx.Response.StatusCode())
	}
}

func TestCreate_InvalidCoordinates(t *testing.T) {
	svc := &mockService{
		createErr: domain.ErrInvalidCoordinates,
	}
	ctrl := NewEnterpriseController(svc)

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetBody([]byte(`{"name":"Test","latitude":100,"longitude":200}`))

	ctrl.Create(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status 400, got %d", ctx.Response.StatusCode())
	}
}

func TestGetByID_Success(t *testing.T) {
	svc := &mockService{
		getEnterprise: validEnterprise(),
	}
	ctrl := NewEnterpriseController(svc)

	ctx := &fasthttp.RequestCtx{}
	ctx.QueryArgs().Set("id", "550e8400-e29b-41d4-a716-446655440000")

	ctrl.GetOrList(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected status 200, got %d", ctx.Response.StatusCode())
	}

	var resp EnterpriseResponse
	if err := json.Unmarshal(ctx.Response.Body(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.ID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected ID, got %q", resp.ID)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	svc := &mockService{
		getErr: domain.ErrEnterpriseNotFound,
	}
	ctrl := NewEnterpriseController(svc)

	ctx := &fasthttp.RequestCtx{}
	ctx.QueryArgs().Set("id", "nonexistent")

	ctrl.GetOrList(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusNotFound {
		t.Errorf("expected status 404, got %d", ctx.Response.StatusCode())
	}
}

func TestList_Success(t *testing.T) {
	svc := &mockService{
		listEnterprises: []*domain.Enterprise{validEnterprise()},
		listTotal:       1,
	}
	ctrl := NewEnterpriseController(svc)

	ctx := &fasthttp.RequestCtx{}

	ctrl.GetOrList(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected status 200, got %d", ctx.Response.StatusCode())
	}

	var resp EnterpriseListResponse
	if err := json.Unmarshal(ctx.Response.Body(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(resp.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(resp.Items))
	}
	if resp.TotalItems != 1 {
		t.Errorf("expected totalItems 1, got %d", resp.TotalItems)
	}
}

func TestList_Empty(t *testing.T) {
	svc := &mockService{
		listEnterprises: []*domain.Enterprise{},
		listTotal:       0,
	}
	ctrl := NewEnterpriseController(svc)

	ctx := &fasthttp.RequestCtx{}

	ctrl.GetOrList(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected status 200, got %d", ctx.Response.StatusCode())
	}

	var resp EnterpriseListResponse
	if err := json.Unmarshal(ctx.Response.Body(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(resp.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(resp.Items))
	}
}

func TestUpdate_Success(t *testing.T) {
	svc := &mockService{
		updateEnterprise: validEnterprise(),
	}
	ctrl := NewEnterpriseController(svc)

	reqBody := `{"name":"Обновлённое предприятие","address":"Новый адрес","status":"configuring","latitude":56.0,"longitude":38.0}`
	ctx := &fasthttp.RequestCtx{}
	ctx.QueryArgs().Set("id", "550e8400-e29b-41d4-a716-446655440000")
	ctx.Request.SetBody([]byte(reqBody))

	ctrl.Update(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected status 200, got %d", ctx.Response.StatusCode())
	}

	var resp EnterpriseResponse
	if err := json.Unmarshal(ctx.Response.Body(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.ID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected ID, got %q", resp.ID)
	}
}

func TestUpdate_MissingID(t *testing.T) {
	ctrl := NewEnterpriseController(&mockService{})

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetBody([]byte(`{"name":"Test"}`))

	ctrl.Update(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status 400, got %d", ctx.Response.StatusCode())
	}
}

func TestUpdate_NotFound(t *testing.T) {
	svc := &mockService{
		updateErr: domain.ErrEnterpriseNotFound,
	}
	ctrl := NewEnterpriseController(svc)

	ctx := &fasthttp.RequestCtx{}
	ctx.QueryArgs().Set("id", "nonexistent")
	ctx.Request.SetBody([]byte(`{"name":"Test"}`))

	ctrl.Update(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusNotFound {
		t.Errorf("expected status 404, got %d", ctx.Response.StatusCode())
	}
}

func TestDelete_Success(t *testing.T) {
	svc := &mockService{}
	ctrl := NewEnterpriseController(svc)

	ctx := &fasthttp.RequestCtx{}
	ctx.QueryArgs().Set("id", "test-id")

	ctrl.Delete(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected status 200, got %d", ctx.Response.StatusCode())
	}
}

func TestDelete_MissingID(t *testing.T) {
	ctrl := NewEnterpriseController(&mockService{})

	ctx := &fasthttp.RequestCtx{}

	ctrl.Delete(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status 400, got %d", ctx.Response.StatusCode())
	}
}

func TestDelete_NotFound(t *testing.T) {
	svc := &mockService{
		deleteErr: domain.ErrEnterpriseNotFound,
	}
	ctrl := NewEnterpriseController(svc)

	ctx := &fasthttp.RequestCtx{}
	ctx.QueryArgs().Set("id", "nonexistent")

	ctrl.Delete(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusNotFound {
		t.Errorf("expected status 404, got %d", ctx.Response.StatusCode())
	}
}

func TestGetByID_InternalError(t *testing.T) {
	svc := &mockService{
		getErr: errors.New("unexpected error"),
	}
	ctrl := NewEnterpriseController(svc)

	ctx := &fasthttp.RequestCtx{}
	ctx.QueryArgs().Set("id", "test-id")

	ctrl.GetOrList(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", ctx.Response.StatusCode())
	}
}

func TestCreate_InternalError(t *testing.T) {
	svc := &mockService{
		createErr: errors.New("unexpected error"),
	}
	ctrl := NewEnterpriseController(svc)

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetBody([]byte(`{"name":"Test","latitude":55.75,"longitude":37.62}`))

	ctrl.Create(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", ctx.Response.StatusCode())
	}
}

func TestList_InternalError(t *testing.T) {
	svc := &mockService{
		listErr: errors.New("unexpected error"),
	}
	ctrl := NewEnterpriseController(svc)

	ctx := &fasthttp.RequestCtx{}

	ctrl.GetOrList(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", ctx.Response.StatusCode())
	}
}

func TestUpdate_BadJSON(t *testing.T) {
	ctrl := NewEnterpriseController(&mockService{})

	ctx := &fasthttp.RequestCtx{}
	ctx.QueryArgs().Set("id", "test-id")
	ctx.Request.SetBody([]byte(`not json`))

	ctrl.Update(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status 400, got %d", ctx.Response.StatusCode())
	}
}

func TestDelete_InternalError(t *testing.T) {
	svc := &mockService{
		deleteErr: errors.New("unexpected error"),
	}
	ctrl := NewEnterpriseController(svc)

	ctx := &fasthttp.RequestCtx{}
	ctx.QueryArgs().Set("id", "test-id")

	ctrl.Delete(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", ctx.Response.StatusCode())
	}
}

func TestList_WithPagination(t *testing.T) {
	svc := &mockService{
		listEnterprises: []*domain.Enterprise{validEnterprise()},
		listTotal:       1,
	}
	ctrl := NewEnterpriseController(svc)

	ctx := &fasthttp.RequestCtx{}
	ctx.QueryArgs().Set("limit", "10")
	ctx.QueryArgs().Set("offset", "0")

	ctrl.GetOrList(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected status 200, got %d", ctx.Response.StatusCode())
	}
}
