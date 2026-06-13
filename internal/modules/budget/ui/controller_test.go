package ui

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/valyala/fasthttp"

	"github.com/fortis/backend/internal/modules/budget/domain"
)

// ---- Mock Service ----

type mockBudgetService struct {
	getConfigFn    func(ctx context.Context, projectID string) (*domain.BudgetConfig, error)
	updateConfigFn func(ctx context.Context, projectID string, mode domain.BudgetMode, amountMln float64) error
	calcCostFn     func(ctx context.Context, projectID string) (*domain.CostCalculation, error)
	checkBudgetFn  func(ctx context.Context, projectID string, input domain.BudgetCheckInput) (*domain.BudgetCheckResult, error)
}

func (m *mockBudgetService) GetBudgetConfig(ctx context.Context, projectID string) (*domain.BudgetConfig, error) {
	return m.getConfigFn(ctx, projectID)
}

func (m *mockBudgetService) UpdateBudgetConfig(ctx context.Context, projectID string, mode domain.BudgetMode, amountMln float64) error {
	return m.updateConfigFn(ctx, projectID, mode, amountMln)
}

func (m *mockBudgetService) CalculateCost(ctx context.Context, projectID string) (*domain.CostCalculation, error) {
	return m.calcCostFn(ctx, projectID)
}

func (m *mockBudgetService) CheckBudget(ctx context.Context, projectID string, input domain.BudgetCheckInput) (*domain.BudgetCheckResult, error) {
	return m.checkBudgetFn(ctx, projectID, input)
}

// ---- Test helpers ----

func newTestCtx(method, path string, body []byte) *fasthttp.RequestCtx {
	ctx := &fasthttp.RequestCtx{}
	ctx.Init(&fasthttp.Request{}, nil, nil)

	// Set request method, URI and body
	var req fasthttp.Request
	req.SetRequestURI(path)
	req.Header.SetMethod(method)
	req.SetBody(body)
	ctx.Init(&req, nil, nil)

	return ctx
}

func parseResponse(t *testing.T, body []byte, target interface{}) {
	t.Helper()
	if err := json.Unmarshal(body, target); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
}

// ---- Tests ----

func TestBudgetController_GetBudgetConfig_Success(t *testing.T) {
	t.Parallel()

	config, _ := domain.NewBudgetConfig("proj-1", domain.BudgetModeLimited, 5000)

	ctrl := NewBudgetController(&mockBudgetService{
		getConfigFn: func(_ context.Context, projectID string) (*domain.BudgetConfig, error) {
			if projectID != "proj-1" {
				t.Errorf("unexpected projectID: %s", projectID)
			}
			return config, nil
		},
	})

	ctx := newTestCtx("GET", "/api/v1/projects/budget?id=proj-1", nil)
	ctrl.GetBudgetConfig(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("StatusCode = %d, want %d", ctx.Response.StatusCode(), fasthttp.StatusOK)
	}

	var resp BudgetConfigDTO
	parseResponse(t, ctx.Response.Body(), &resp)
	if resp.ProjectID != "proj-1" {
		t.Errorf("ProjectID = %s, want proj-1", resp.ProjectID)
	}
	if resp.BudgetMode != "limited" {
		t.Errorf("BudgetMode = %s, want limited", resp.BudgetMode)
	}
}

func TestBudgetController_GetBudgetConfig_MissingID(t *testing.T) {
	t.Parallel()

	ctrl := NewBudgetController(&mockBudgetService{})
	ctx := newTestCtx("GET", "/api/v1/projects/budget", nil)
	ctrl.GetBudgetConfig(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("StatusCode = %d, want %d", ctx.Response.StatusCode(), fasthttp.StatusBadRequest)
	}
}

func TestBudgetController_GetBudgetConfig_NotFound(t *testing.T) {
	t.Parallel()

	ctrl := NewBudgetController(&mockBudgetService{
		getConfigFn: func(_ context.Context, projectID string) (*domain.BudgetConfig, error) {
			return nil, domain.ErrBudgetConfigNotFound
		},
	})

	ctx := newTestCtx("GET", "/api/v1/projects/budget?id=nonexistent", nil)
	ctrl.GetBudgetConfig(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusNotFound {
		t.Errorf("StatusCode = %d, want %d", ctx.Response.StatusCode(), fasthttp.StatusNotFound)
	}
}

func TestBudgetController_UpdateBudgetConfig_Success(t *testing.T) {
	t.Parallel()

	config, _ := domain.NewBudgetConfig("proj-1", domain.BudgetModeLimited, 10000)

	ctrl := NewBudgetController(&mockBudgetService{
		updateConfigFn: func(_ context.Context, projectID string, mode domain.BudgetMode, amountMln float64) error {
			return nil
		},
		getConfigFn: func(_ context.Context, projectID string) (*domain.BudgetConfig, error) {
			return config, nil
		},
	})

	body, _ := json.Marshal(UpdateBudgetRequest{
		BudgetMode:      "limited",
		BudgetAmountMln: 10000,
	})

	ctx := newTestCtx("PUT", "/api/v1/projects/budget?id=proj-1", body)
	ctrl.UpdateBudgetConfig(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("StatusCode = %d, want %d", ctx.Response.StatusCode(), fasthttp.StatusOK)
	}

	var resp BudgetConfigDTO
	parseResponse(t, ctx.Response.Body(), &resp)
	if resp.BudgetAmountMln != 10000 {
		t.Errorf("BudgetAmountMln = %f, want 10000", resp.BudgetAmountMln)
	}
}

func TestBudgetController_UpdateBudgetConfig_MissingID(t *testing.T) {
	t.Parallel()

	ctrl := NewBudgetController(&mockBudgetService{})
	body, _ := json.Marshal(UpdateBudgetRequest{BudgetMode: "limited"})
	ctx := newTestCtx("PUT", "/api/v1/projects/budget", body)
	ctrl.UpdateBudgetConfig(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("StatusCode = %d, want %d", ctx.Response.StatusCode(), fasthttp.StatusBadRequest)
	}
}

func TestBudgetController_UpdateBudgetConfig_InvalidMode(t *testing.T) {
	t.Parallel()

	ctrl := NewBudgetController(&mockBudgetService{
		updateConfigFn: func(_ context.Context, projectID string, mode domain.BudgetMode, amountMln float64) error {
			return domain.ErrInvalidBudgetMode
		},
	})

	body, _ := json.Marshal(UpdateBudgetRequest{
		BudgetMode:      "invalid",
		BudgetAmountMln: 100,
	})

	ctx := newTestCtx("PUT", "/api/v1/projects/budget?id=proj-1", body)
	ctrl.UpdateBudgetConfig(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("StatusCode = %d, want %d", ctx.Response.StatusCode(), fasthttp.StatusBadRequest)
	}
}

func TestBudgetController_CalculateCost_Success(t *testing.T) {
	t.Parallel()

	calc := domain.NewCostCalculation(
		180,
		[]domain.EchelonEstimate{
			domain.NewEchelonEstimate("echelon-1", "Обнаружение", []domain.EstimateLine{}, 80),
		},
		[]domain.TypeEstimate{
			domain.NewTypeEstimate("radar", "radar", []domain.EstimateLine{}, 60),
		},
		[]domain.EstimateLine{
			domain.NewEstimateLine("obj-1", "asset-1", "Mobile Radar", "echelon-1", "Обнаружение", "radar", "radar", 3, 20, 60),
		},
	)

	ctrl := NewBudgetController(&mockBudgetService{
		calcCostFn: func(_ context.Context, projectID string) (*domain.CostCalculation, error) {
			return &calc, nil
		},
	})

	ctx := newTestCtx("GET", "/api/v1/projects/cost?id=proj-1", nil)
	ctrl.CalculateCost(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("StatusCode = %d, want %d", ctx.Response.StatusCode(), fasthttp.StatusOK)
	}

	var resp CostCalculationDTO
	parseResponse(t, ctx.Response.Body(), &resp)
	if resp.TotalMln != 180 {
		t.Errorf("TotalMln = %f, want 180", resp.TotalMln)
	}
	if len(resp.ByEchelon) != 1 {
		t.Errorf("ByEchelon length = %d, want 1", len(resp.ByEchelon))
	}
}

func TestBudgetController_CheckBudget_Success(t *testing.T) {
	t.Parallel()

	result := domain.NewBudgetCheckResult(true, 820, 20, domain.BudgetModeLimited)

	ctrl := NewBudgetController(&mockBudgetService{
		checkBudgetFn: func(_ context.Context, projectID string, input domain.BudgetCheckInput) (*domain.BudgetCheckResult, error) {
			return &result, nil
		},
	})

	body, _ := json.Marshal(BudgetCheckRequest{
		AssetID:  "asset-1",
		Quantity: 1,
		EchelonID: "echelon-1",
	})

	ctx := newTestCtx("POST", "/api/v1/projects/budget/check?id=proj-1", body)
	ctrl.CheckBudget(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("StatusCode = %d, want %d", ctx.Response.StatusCode(), fasthttp.StatusOK)
	}

	var resp BudgetCheckResponse
	parseResponse(t, ctx.Response.Body(), &resp)
	if !resp.Fits {
		t.Error("Fits = false, want true")
	}
	if resp.RemainingMln != 820 {
		t.Errorf("RemainingMln = %f, want 820", resp.RemainingMln)
	}
}

func TestBudgetController_CheckBudget_InvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body BudgetCheckRequest
	}{
		{
			name: "empty assetId",
			body: BudgetCheckRequest{AssetID: "", Quantity: 1, EchelonID: "echelon-1"},
		},
		{
			name: "zero quantity",
			body: BudgetCheckRequest{AssetID: "asset-1", Quantity: 0, EchelonID: "echelon-1"},
		},
		{
			name: "negative quantity",
			body: BudgetCheckRequest{AssetID: "asset-1", Quantity: -1, EchelonID: "echelon-1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewBudgetController(&mockBudgetService{})
			body, _ := json.Marshal(tt.body)
			ctx := newTestCtx("POST", "/api/v1/projects/budget/check?id=proj-1", body)
			ctrl.CheckBudget(ctx)

			if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
				t.Errorf("StatusCode = %d, want %d", ctx.Response.StatusCode(), fasthttp.StatusBadRequest)
			}
		})
	}
}
