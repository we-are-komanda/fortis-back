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
	compareConfigsFn func(ctx context.Context, a, b defenseDomain.ProjectRef) (*domain.RevisionComparison, error)
}

func (m *mockBudgetService) CalculateCost(ctx context.Context, actorID string, projectID string) (*domain.CostCalculation, error) {
	return m.calcCostFn(ctx, projectID)
}
func (m *mockBudgetService) ProjectCost(ctx context.Context, actorID, projectID string, version *int) (*domain.CostProjection, error) {
	return m.projectCostFn(ctx, projectID, version)
}

func (m *mockBudgetService) CompareConfigs(ctx context.Context, actorID string, a, b defenseDomain.ProjectRef) (*domain.RevisionComparison, error) {
	return m.compareConfigsFn(ctx, a, b)
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

func TestBudgetController_CompareRequiresPinnedVersions(t *testing.T) {
	const a = "00000000-0000-4000-8000-000000000001"
	const b = "00000000-0000-4000-8000-000000000002"
	for _, query := range []string{"", "id1=" + a + "&id2=" + b, "projectIdA=" + a + "&projectIdB=" + b + "&projectVersionA=0&projectVersionB=1", "projectIdA=" + a + "&projectIdB=" + b + "&projectVersionA=1&projectVersionB=1.2", "projectIdA=" + a + "&id1=" + b + "&projectVersionA=1&projectIdB=" + b + "&projectVersionB=2"} {
		ctx := newTestCtx("GET", "/api/v1/projects/compare?"+query, nil)
		NewBudgetController(&mockBudgetService{}).Compare(ctx)
		if ctx.Response.StatusCode() != 400 {
			t.Fatalf("query %s: status %d", query, ctx.Response.StatusCode())
		}
	}
}
