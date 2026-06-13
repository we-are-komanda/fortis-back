package ui

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/fortis/backend/internal/modules/defense_asset/application"
	"github.com/fortis/backend/internal/modules/defense_asset/domain"
)

// mockDefenseAssetService — мок сервиса для тестирования контроллера.
type mockDefenseAssetService struct {
	createFn  func(ctx context.Context, input application.CreateInput) (*domain.DefenseAsset, error)
	getByIDFn func(ctx context.Context, id string) (*domain.DefenseAsset, error)
	listFn    func(ctx context.Context, enterpriseID *string, isPublic *bool, category *domain.DefenseAssetCategory, limit, offset int) ([]*domain.DefenseAsset, int64, error)
	updateFn  func(ctx context.Context, input application.UpdateInput) (*domain.DefenseAsset, error)
	deleteFn  func(ctx context.Context, id string) error
}

func (m *mockDefenseAssetService) Create(ctx context.Context, input application.CreateInput) (*domain.DefenseAsset, error) {
	return m.createFn(ctx, input)
}

func (m *mockDefenseAssetService) GetByID(ctx context.Context, id string) (*domain.DefenseAsset, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockDefenseAssetService) List(ctx context.Context, enterpriseID *string, isPublic *bool, category *domain.DefenseAssetCategory, limit, offset int) ([]*domain.DefenseAsset, int64, error) {
	return m.listFn(ctx, enterpriseID, isPublic, category, limit, offset)
}

func (m *mockDefenseAssetService) Update(ctx context.Context, input application.UpdateInput) (*domain.DefenseAsset, error) {
	return m.updateFn(ctx, input)
}

func (m *mockDefenseAssetService) Delete(ctx context.Context, id string) error {
	return m.deleteFn(ctx, id)
}

// testAsset возвращает тестовый доменный объект.
func testAsset(id, name string) *domain.DefenseAsset {
	now := time.Now().UTC()
	eid := "550e8400-e29b-41d4-a716-446655440000"
	cat := domain.DefenseAssetCategoryRadar
	ct := domain.DefenseAssetCoverageCircle

	asset, _ := domain.NewDefenseAsset(
		id, name, "", "", cat, nil, nil, "RUB", "", nil, nil, nil, nil, "",
		nil, nil, ct, nil, nil,
		domain.DeploymentTypeStatic, domain.PlacementTypeMapObject,
		"", "", nil, nil, nil, nil, "", nil, nil,
		&eid, false, now, now,
	)
	return asset
}

func TestCreateAsset_Success(t *testing.T) {
	svc := &mockDefenseAssetService{
		createFn: func(ctx context.Context, input application.CreateInput) (*domain.DefenseAsset, error) {
			return testAsset("test-id", input.Name), nil
		},
	}

	controller := NewDefenseAssetController(svc)

	body := `{"name":"Test Asset","category":"radar","coverageType":"circle"}`
	req := fasthttp.AcquireRequest()
	req.SetBody([]byte(body))
	req.Header.SetContentType("application/json")

	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.Create(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusCreated {
		t.Errorf("expected status %d, got %d", fasthttp.StatusCreated, ctx.Response.StatusCode())
	}

	var resp DefenseAssetDTO
	if err := json.Unmarshal(ctx.Response.Body(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Name != "Test Asset" {
		t.Errorf("expected name 'Test Asset', got %q", resp.Name)
	}
}

func TestCreateAsset_MissingName(t *testing.T) {
	controller := NewDefenseAssetController(&mockDefenseAssetService{})

	body := `{"category":"radar","coverageType":"circle"}`
	req := fasthttp.AcquireRequest()
	req.SetBody([]byte(body))
	req.Header.SetContentType("application/json")

	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.Create(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status %d, got %d", fasthttp.StatusBadRequest, ctx.Response.StatusCode())
	}
}

func TestCreateAsset_InvalidJSON(t *testing.T) {
	controller := NewDefenseAssetController(&mockDefenseAssetService{})

	body := `{invalid json}`
	req := fasthttp.AcquireRequest()
	req.SetBody([]byte(body))
	req.Header.SetContentType("application/json")

	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.Create(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status %d, got %d", fasthttp.StatusBadRequest, ctx.Response.StatusCode())
	}
}

func TestGetAsset_Success(t *testing.T) {
	svc := &mockDefenseAssetService{
		getByIDFn: func(ctx context.Context, id string) (*domain.DefenseAsset, error) {
			return testAsset(id, "Test Asset"), nil
		},
	}

	controller := NewDefenseAssetController(svc)

	req := fasthttp.AcquireRequest()
	req.URI().SetQueryString("id=test-id")

	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.Get(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected status %d, got %d", fasthttp.StatusOK, ctx.Response.StatusCode())
	}

	var resp DefenseAssetDTO
	if err := json.Unmarshal(ctx.Response.Body(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Name != "Test Asset" {
		t.Errorf("expected name 'Test Asset', got %q", resp.Name)
	}
}

func TestGetAsset_NotFound(t *testing.T) {
	svc := &mockDefenseAssetService{
		getByIDFn: func(ctx context.Context, id string) (*domain.DefenseAsset, error) {
			return nil, domain.ErrDefenseAssetNotFound
		},
	}

	controller := NewDefenseAssetController(svc)

	req := fasthttp.AcquireRequest()
	req.URI().SetQueryString("id=non-existent")

	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.Get(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusNotFound {
		t.Errorf("expected status %d, got %d", fasthttp.StatusNotFound, ctx.Response.StatusCode())
	}
}

func TestGetAsset_MissingID(t *testing.T) {
	controller := NewDefenseAssetController(&mockDefenseAssetService{})

	req := fasthttp.AcquireRequest()
	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.Get(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status %d, got %d", fasthttp.StatusBadRequest, ctx.Response.StatusCode())
	}
}

func TestListAsset_Success(t *testing.T) {
	svc := &mockDefenseAssetService{
		listFn: func(ctx context.Context, enterpriseID *string, isPublic *bool, category *domain.DefenseAssetCategory, limit, offset int) ([]*domain.DefenseAsset, int64, error) {
			assets := []*domain.DefenseAsset{
				testAsset("id-1", "Asset 1"),
				testAsset("id-2", "Asset 2"),
			}
			return assets, 2, nil
		},
	}

	controller := NewDefenseAssetController(svc)

	req := fasthttp.AcquireRequest()
	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.List(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected status %d, got %d", fasthttp.StatusOK, ctx.Response.StatusCode())
	}

	var resp DefenseAssetListResponse
	if err := json.Unmarshal(ctx.Response.Body(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.TotalItems != 2 {
		t.Errorf("expected totalItems 2, got %d", resp.TotalItems)
	}
	if len(resp.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(resp.Items))
	}
}

func TestDeleteAsset_Success(t *testing.T) {
	svc := &mockDefenseAssetService{
		deleteFn: func(ctx context.Context, id string) error {
			return nil
		},
	}

	controller := NewDefenseAssetController(svc)

	req := fasthttp.AcquireRequest()
	req.URI().SetQueryString("id=test-id")

	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.Delete(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected status %d, got %d", fasthttp.StatusOK, ctx.Response.StatusCode())
	}
}

func TestDeleteAsset_NotFound(t *testing.T) {
	svc := &mockDefenseAssetService{
		deleteFn: func(ctx context.Context, id string) error {
			return domain.ErrDefenseAssetNotFound
		},
	}

	controller := NewDefenseAssetController(svc)

	req := fasthttp.AcquireRequest()
	req.URI().SetQueryString("id=non-existent")

	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.Delete(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusNotFound {
		t.Errorf("expected status %d, got %d", fasthttp.StatusNotFound, ctx.Response.StatusCode())
	}
}

func TestDeleteAsset_MissingID(t *testing.T) {
	controller := NewDefenseAssetController(&mockDefenseAssetService{})

	req := fasthttp.AcquireRequest()
	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.Delete(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status %d, got %d", fasthttp.StatusBadRequest, ctx.Response.StatusCode())
	}
}

func TestUpdateAsset_Success(t *testing.T) {
	svc := &mockDefenseAssetService{
		updateFn: func(ctx context.Context, input application.UpdateInput) (*domain.DefenseAsset, error) {
			name := "Updated Asset"
			if input.Name != nil {
				name = *input.Name
			}
			return testAsset(input.ID, name), nil
		},
	}

	controller := NewDefenseAssetController(svc)

	body := `{"name":"Updated Asset"}`
	req := fasthttp.AcquireRequest()
	req.SetBody([]byte(body))
	req.Header.SetContentType("application/json")
	req.URI().SetQueryString("id=test-id")

	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.Update(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected status %d, got %d", fasthttp.StatusOK, ctx.Response.StatusCode())
	}

	var resp DefenseAssetDTO
	if err := json.Unmarshal(ctx.Response.Body(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Name != "Updated Asset" {
		t.Errorf("expected name 'Updated Asset', got %q", resp.Name)
	}
}

func TestUpdateAsset_MissingID(t *testing.T) {
	controller := NewDefenseAssetController(&mockDefenseAssetService{})

	body := `{"name":"Updated Asset"}`
	req := fasthttp.AcquireRequest()
	req.SetBody([]byte(body))
	req.Header.SetContentType("application/json")

	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.Update(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status %d, got %d", fasthttp.StatusBadRequest, ctx.Response.StatusCode())
	}
}

func TestUpdateAsset_NotFound(t *testing.T) {
	svc := &mockDefenseAssetService{
		updateFn: func(ctx context.Context, input application.UpdateInput) (*domain.DefenseAsset, error) {
			return nil, domain.ErrDefenseAssetNotFound
		},
	}

	controller := NewDefenseAssetController(svc)

	body := `{"name":"Updated Asset"}`
	req := fasthttp.AcquireRequest()
	req.SetBody([]byte(body))
	req.Header.SetContentType("application/json")
	req.URI().SetQueryString("id=non-existent")

	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.Update(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusNotFound {
		t.Errorf("expected status %d, got %d", fasthttp.StatusNotFound, ctx.Response.StatusCode())
	}
}
