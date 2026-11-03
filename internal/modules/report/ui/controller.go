package ui

import (
	"context"
	"errors"
	budget "github.com/fortis/backend/internal/modules/budget/domain"
	project "github.com/fortis/backend/internal/modules/defense_project/domain"
	"github.com/fortis/backend/internal/modules/report/application"
	"github.com/fortis/backend/internal/modules/report/domain"
	"github.com/fortis/backend/pkg/handlers"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"strconv"
)

type ReportServiceInterface interface {
	GetReport(context.Context, string, application.GetReportQuery) (*domain.IssuedReport, error)
}
type ReportController struct{ service ReportServiceInterface }

func NewReportController(service ReportServiceInterface) *ReportController {
	return &ReportController{service}
}

// swagger:route GET /api/v1/projects/report api getReport
// Выпустить отчёт фиксированной версии или открыть неизменяемый ранее выпущенный отчёт.
// Стоимость скрывается отдельным безопасным DTO; доступ проверяется при каждом открытии.
// Produces:
//   - application/json
//
// Responses:
//
//	200: PricedReportResponse
//	400: description: Некорректные параметры выпуска или открытия
//	404: description: Отчёт или версия недоступны
//	503: description: Сохранённая версия алгоритма недоступна
func (c *ReportController) Get(ctx *fasthttp.RequestCtx) {
	bad := func(code, message string) { handlers.ErrorHandler(ctx, code, message, &handlers.ResponseBody{}, 400) }
	args := ctx.QueryArgs()
	seen := map[string]bool{}
	invalid := false
	args.VisitAll(func(k, v []byte) {
		key := string(k)
		if seen[key] {
			invalid = true
		}
		seen[key] = true
		switch key {
		case "reportId", "id", "projectVersion", "hideCost", "mapView":
		default:
			invalid = true
		}
	})
	if invalid {
		bad("validation_error", "invalid or duplicate report query parameter")
		return
	}
	q := application.GetReportQuery{}
	if args.Has("reportId") {
		if len(seen) != 1 {
			bad("validation_error", "reportId cannot be combined with issuance parameters")
			return
		}
		id, err := uuid.Parse(string(args.Peek("reportId")))
		if err != nil {
			bad("validation_error", "invalid reportId")
			return
		}
		q.ReportID = id.String()
	} else {
		id, err := uuid.Parse(string(args.Peek("id")))
		if err != nil {
			bad("validation_error", "valid id is required")
			return
		}
		if !args.Has("projectVersion") {
			bad("version_required", "projectVersion is required")
			return
		}
		version, err := strconv.Atoi(string(args.Peek("projectVersion")))
		if err != nil || version <= 0 {
			bad("validation_error", "projectVersion must be positive")
			return
		}
		hidden := string(args.Peek("hideCost"))
		if hidden != "true" && hidden != "false" {
			bad("validation_error", "hideCost must be explicitly true or false")
			return
		}
		view := "all"
		if args.Has("mapView") {
			view = string(args.Peek("mapView"))
		}
		key, err := domain.NewReportKey(id.String(), version, hidden == "true", view)
		if err != nil {
			bad("validation_error", err.Error())
			return
		}
		q.Key = key
	}
	report, err := c.service.GetReport(ctx, handlers.ActorID(ctx), q)
	if err != nil {
		if handlers.AuthorizationError(ctx, err) {
			return
		}
		switch {
		case errors.Is(err, domain.ErrReportNotFound), errors.Is(err, project.ErrProjectNotFound), errors.Is(err, project.ErrRevisionNotFound):
			handlers.ErrorHandler(ctx, "not_found", "report not found", &handlers.ResponseBody{}, 404)
		case errors.Is(err, budget.ErrUnsupportedCalculationVersion):
			handlers.ErrorHandler(ctx, "unsupported_calculation_version", "saved calculation version is unavailable", &handlers.ResponseBody{}, 503)
		case errors.Is(err, domain.ErrInvalidSnapshot):
			handlers.ErrorHandler(ctx, "invalid_report_snapshot", "saved geometry cannot be rendered safely", &handlers.ResponseBody{}, 422)
		default:
			handlers.ErrorHandler(ctx, "internal_error", "failed to get report", &handlers.ResponseBody{}, 500)
		}
		return
	}
	ctx.SetContentType("application/json")
	ctx.SetBodyString(report.Payload())
}
