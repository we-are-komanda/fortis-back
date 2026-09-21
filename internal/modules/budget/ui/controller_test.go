package ui

import (
	"context"
	"encoding/json"
	"errors"
	budgetApp "github.com/fortis/backend/internal/modules/budget/application"
	"testing"

	"github.com/valyala/fasthttp"

	"github.com/fortis/backend/internal/modules/budget/domain"
	defenseDomain "github.com/fortis/backend/internal/modules/defense_project/domain"
)

// ---- Mock Service ----

type mockBudgetService struct {
	calcCostFn       func(ctx context.Context, projectID string) (*domain.CostCalculation, error)
	projectCostFn    func(ctx context.Context, projectID string, version *int) (*domain.CostProjection, error)
	compareConfigsFn func(ctx context.Context, projectID1, projectID2 string) (*domain.ConfigComparison, error)
}

func (m *mockBudgetService) CalculateCost(ctx context.Context, actorID string, projectID string) (*domain.CostCalculation, error) {
	return m.calcCostFn(ctx, projectID)
}
func (m *mockBudgetService) ProjectCost(ctx context.Context, actorID, projectID string, version *int) (*domain.CostProjection, error) {
	return m.projectCostFn(ctx, projectID, version)
}

func (m *mockBudgetService) CompareConfigs(ctx context.Context, actorID string, projectID1, projectID2 string) (*domain.ConfigComparison, error) {
	return m.compareConfigsFn(ctx, projectID1, projectID2)
}

func (m *mockBudgetService) ProjectBudget(context.Context, string, string, *int) (*domain.BudgetProjection, error) {
	return nil, errors.New("unexpected budget call")
}
func (m *mockBudgetService) UpdateBudget(context.Context, string, string, budgetApp.UpdateBudgetCommand) (*defenseDomain.DefenseProject, *domain.BudgetProjection, error) {
	return nil, nil, errors.New("unexpected budget call")
}
func (m *mockBudgetService) CheckCandidate(context.Context, string, string, budgetApp.CheckBudgetCommand) (*domain.CandidateBudgetCheck, error) {
	return nil, errors.New("unexpected budget call")
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

func TestBudgetController_CalculateCost_Success(t *testing.T) {
	t.Parallel()

	price, _ := domain.ParseUnitMinor("2000000000")
	line, _ := domain.NewFinancialLine("obj-1", "asset-1", "custom", "synthetic", "Module", 3, price, "template", "")
	projection := domain.NewCostProjection(domain.NewCalculationIdentity("proj-1", 7, domain.CostCalculationV1, "digest", nil), []domain.FinancialLine{line}, map[string]string{"custom": "Custom"}, nil)
	ctrl := NewBudgetController(&mockBudgetService{projectCostFn: func(_ context.Context, id string, version *int) (*domain.CostProjection, error) {
		if id != "proj-1" || version == nil || *version != 7 {
			t.Fatalf("wrong revision request: %s %v", id, version)
		}
		return projection, nil
	}})
	ctx := newTestCtx("GET", "/api/v1/projects/cost?projectId=proj-1&projectVersion=7", nil)
	ctrl.CalculateCost(ctx)
	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Fatalf("status=%d body=%s", ctx.Response.StatusCode(), ctx.Response.Body())
	}
	var response CostProjectionDTO
	parseResponse(t, ctx.Response.Body(), &response)
	if response.TotalMinor == nil || *response.TotalMinor != "6000000000" || response.Identity.ProjectVersion != 7 || len(response.ByLayer) != 1 {
		t.Fatalf("unexpected projection: %+v", response)
	}
}

func TestBudgetController_Compare_Success(t *testing.T) {
	t.Parallel()

	calcA := domain.NewCostCalculation(100, nil, nil, nil)
	calcB := domain.NewCostCalculation(200, nil, nil, nil)

	profA := domain.NewStructuralProfile(5, 10, 2, 3, 1, 4, 100, nil)
	profB := domain.NewStructuralProfile(8, 16, 3, 4, 2, 6, 200, nil)

	snapA := domain.NewConfigSnapshot("proj-1", "Config A", profA, calcA)
	snapB := domain.NewConfigSnapshot("proj-2", "Config B", profB, calcB)
	diff := domain.NewConfigDiff(3, 6, 1, 1, 1, 2, 100, nil)
	comp := domain.NewConfigComparison(snapA, snapB, diff)

	ctrl := NewBudgetController(&mockBudgetService{
		compareConfigsFn: func(_ context.Context, id1, id2 string) (*domain.ConfigComparison, error) {
			if id1 != "proj-1" || id2 != "proj-2" {
				t.Errorf("unexpected ids: %s, %s", id1, id2)
			}
			return &comp, nil
		},
	})

	ctx := newTestCtx("GET", "/api/v1/projects/compare?id1=proj-1&id2=proj-2", nil)
	ctrl.Compare(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("StatusCode = %d, want %d", ctx.Response.StatusCode(), fasthttp.StatusOK)
	}

	var resp ConfigComparisonResponse
	parseResponse(t, ctx.Response.Body(), &resp)

	if resp.ProjectA.ProjectID != "proj-1" {
		t.Errorf("ProjectA.ProjectID = %s, want proj-1", resp.ProjectA.ProjectID)
	}
	if resp.ProjectB.ProjectID != "proj-2" {
		t.Errorf("ProjectB.ProjectID = %s, want proj-2", resp.ProjectB.ProjectID)
	}
	if resp.Diff.ObjectCountDelta != 3 {
		t.Errorf("Diff.ObjectCountDelta = %d, want 3", resp.Diff.ObjectCountDelta)
	}
	if resp.Diff.CostDeltaMln != 100 {
		t.Errorf("Diff.CostDeltaMln = %f, want 100", resp.Diff.CostDeltaMln)
	}
}

func TestBudgetController_Compare_MissingIDs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
	}{
		{
			name: "both ids missing",
			path: "/api/v1/projects/compare",
		},
		{
			name: "id1 missing",
			path: "/api/v1/projects/compare?id2=proj-2",
		},
		{
			name: "id2 missing",
			path: "/api/v1/projects/compare?id1=proj-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewBudgetController(&mockBudgetService{})
			ctx := newTestCtx("GET", tt.path, nil)
			ctrl.Compare(ctx)

			if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
				t.Errorf("StatusCode = %d, want %d", ctx.Response.StatusCode(), fasthttp.StatusBadRequest)
			}
		})
	}
}

func TestBudgetController_Compare_ProjectNotFound(t *testing.T) {
	t.Parallel()

	ctrl := NewBudgetController(&mockBudgetService{
		compareConfigsFn: func(_ context.Context, id1, id2 string) (*domain.ConfigComparison, error) {
			return nil, defenseDomain.ErrProjectNotFound
		},
	})

	ctx := newTestCtx("GET", "/api/v1/projects/compare?id1=nonexistent&id2=proj-2", nil)
	ctrl.Compare(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", ctx.Response.StatusCode(), fasthttp.StatusInternalServerError)
	}
}
