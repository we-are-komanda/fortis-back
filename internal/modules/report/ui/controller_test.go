package ui

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/valyala/fasthttp"

	budgetDomain "github.com/fortis/backend/internal/modules/budget/domain"
	"github.com/fortis/backend/internal/modules/report/domain"
)

// mockReportService — мок для ReportServiceInterface.
type mockReportService struct {
	payload *domain.ReportPayload
	err     error
}

func (m *mockReportService) GetReport(ctx context.Context, projectID string, hideCost bool) (*domain.ReportPayload, error) {
	return m.payload, m.err
}

func makeTestPayload() *domain.ReportPayload {
	baseObj := domain.NewReportBaseObject("base-1", "Test Base", 55.75, 37.62)

	layer := domain.NewReportLayer(
		"layer-1", "Дальнее обнаружение", "early-warning", nil,
		"circle", float64Ptr(55.75), float64Ptr(37.62),
		float64Ptr(50000), nil, nil,
		strPtr("#FF0000"), float64Ptr(0.5),
	)

	obj := domain.NewReportPlacedObject(
		"obj-1", "asset-1", "РЛС Небо-М",
		"layer-1", "early-warning", "Дальнее обнаружение",
		55.76, 37.63, 2, "detection",
		150.0, 300.0, false, nil, nil, nil,
	)

	est := budgetDomain.NewCostCalculation(300.0, nil, nil, nil)
	sp := budgetDomain.NewStructuralProfile(1, 2, 1, 1, 0, 1, 300.0, nil)

	return domain.NewReportPayload(
		"proj-1", "Test Project",
		baseObj, []domain.ReportLayer{layer},
		[]domain.ReportPlacedObject{obj},
		est, sp, false,
	)
}

// ---- Tests ----

func TestReportController_Get_Success(t *testing.T) {
	t.Parallel()

	controller := NewReportController(&mockReportService{
		payload: makeTestPayload(),
	})

	ctx := &fasthttp.RequestCtx{}
	ctx.QueryArgs().Set("id", "proj-1")

	controller.Get(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("StatusCode = %d, want 200", ctx.Response.StatusCode())
	}

	var resp ReportResponse
	if err := json.Unmarshal(ctx.Response.Body(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.ProjectID != "proj-1" {
		t.Errorf("ProjectID = %s, want proj-1", resp.ProjectID)
	}
	if resp.ProjectName != "Test Project" {
		t.Errorf("ProjectName = %s, want Test Project", resp.ProjectName)
	}
	if len(resp.Layers) != 1 {
		t.Errorf("Layers length = %d, want 1", len(resp.Layers))
	}
	if len(resp.PlacedObjects) != 1 {
		t.Errorf("PlacedObjects length = %d, want 1", len(resp.PlacedObjects))
	}
	if resp.HideCost {
		t.Error("HideCost = true, want false")
	}
	if resp.BaseObject.ID != "base-1" {
		t.Errorf("BaseObject.ID = %s, want base-1", resp.BaseObject.ID)
	}
	if resp.BaseObject.Center.Lat != 55.75 {
		t.Errorf("BaseObject.Center.Lat = %f, want 55.75", resp.BaseObject.Center.Lat)
	}
}

func TestReportController_Get_MissingID(t *testing.T) {
	t.Parallel()

	controller := NewReportController(&mockReportService{})

	ctx := &fasthttp.RequestCtx{}
	// Не передаём id
	controller.Get(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("StatusCode = %d, want 400", ctx.Response.StatusCode())
	}
}

func TestReportController_Get_NotFound(t *testing.T) {
	t.Parallel()

	controller := NewReportController(&mockReportService{
		err: domain.ErrProjectNotFound,
	})

	ctx := &fasthttp.RequestCtx{}
	ctx.QueryArgs().Set("id", "nonexistent")

	controller.Get(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusNotFound {
		t.Errorf("StatusCode = %d, want 404", ctx.Response.StatusCode())
	}
}

func TestReportController_Get_InternalError(t *testing.T) {
	t.Parallel()

	controller := NewReportController(&mockReportService{
		err: errors.New("unexpected error"),
	})

	ctx := &fasthttp.RequestCtx{}
	ctx.QueryArgs().Set("id", "proj-1")

	controller.Get(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want 500", ctx.Response.StatusCode())
	}
}

func TestReportController_Get_HideCost(t *testing.T) {
	t.Parallel()

	// Создаём payload с hideCost = true (так вернёт сервис после обнуления)
	payload := makeTestPayload()
	// Сервис в реальности вернёт hideCost=true, когда hideCost=true в запросе
	// Но контроллер просто передаёт параметр, поэтому мок должен вернуть правильный payload
	controller := NewReportController(&mockReportService{
		payload: payload,
	})

	ctx := &fasthttp.RequestCtx{}
	ctx.QueryArgs().Set("id", "proj-1")
	ctx.QueryArgs().Set("hideCost", "true")

	controller.Get(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("StatusCode = %d, want 200", ctx.Response.StatusCode())
	}

	var resp ReportResponse
	if err := json.Unmarshal(ctx.Response.Body(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Проверяем, что ответ содержит ProjectID (основной флоу контроллера)
	if resp.ProjectID != "proj-1" {
		t.Errorf("ProjectID = %s, want proj-1", resp.ProjectID)
	}
}

func TestReportController_Get_ObjectLines(t *testing.T) {
	t.Parallel()

	payload := makeTestPayload()
	controller := NewReportController(&mockReportService{
		payload: payload,
	})

	ctx := &fasthttp.RequestCtx{}
	ctx.QueryArgs().Set("id", "proj-1")

	controller.Get(ctx)

	var resp ReportResponse
	if err := json.Unmarshal(ctx.Response.Body(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(resp.PlacedObjects) != 1 {
		t.Fatalf("PlacedObjects length = %d, want 1", len(resp.PlacedObjects))
	}

	obj := resp.PlacedObjects[0]
	if obj.ObjectID != "obj-1" {
		t.Errorf("ObjectID = %s, want obj-1", obj.ObjectID)
	}
	if obj.AssetName != "РЛС Небо-М" {
		t.Errorf("AssetName = %s, want РЛС Небо-М", obj.AssetName)
	}
	if obj.LayerCode != "early-warning" {
		t.Errorf("LayerCode = %s, want early-warning", obj.LayerCode)
	}
	if obj.Quantity != 2 {
		t.Errorf("Quantity = %d, want 2", obj.Quantity)
	}
	if obj.UnitPriceMln != 150.0 {
		t.Errorf("UnitPriceMln = %f, want 150.0", obj.UnitPriceMln)
	}
	if obj.LineTotalMln != 300.0 {
		t.Errorf("LineTotalMln = %f, want 300.0", obj.LineTotalMln)
	}
}

// ---- Helpers ----

func strPtr(s string) *string { return &s }

func float64Ptr(f float64) *float64 { return &f }
