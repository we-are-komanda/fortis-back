package ui

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/valyala/fasthttp"

	"github.com/fortis/backend/internal/modules/report/domain"
	"github.com/fortis/backend/pkg/handlers"
)

// ReportServiceInterface — интерфейс сервиса отчёта для контроллера.
type ReportServiceInterface interface {
	GetReport(ctx context.Context, projectID string, hideCost bool) (*domain.ReportPayload, error)
}

// ReportController — контроллер для API отчёта GIS MVP.
type ReportController struct {
	service ReportServiceInterface
}

// NewReportController создаёт новый ReportController.
func NewReportController(service ReportServiceInterface) *ReportController {
	return &ReportController{
		service: service,
	}
}

// swagger:route GET /api/v1/projects/report api getReport
// Получить отчёт GIS MVP для проекта
//
// Возвращает единый payload для построения отчёта: карту, объекты, слои,
// типы, стоимость, ТТХ, режим скрытия стоимости.
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: ReportResponse
//	400: description: Bad Request — не указан ID проекта
//	404: description: Not Found — проект не найден
//	500: description: Internal Server Error
func (c *ReportController) Get(ctx *fasthttp.RequestCtx) {
	projectID := string(ctx.QueryArgs().Peek("id"))
	if projectID == "" {
		handlers.ErrorHandler(ctx, "validation_error", "id query parameter is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	hideCostStr := string(ctx.QueryArgs().Peek("hideCost"))
	hideCost, _ := strconv.ParseBool(hideCostStr)

	payload, err := c.service.GetReport(ctx, projectID, hideCost)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrProjectNotFound):
			handlers.ErrorHandler(ctx, "not_found", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusNotFound)
		case errors.Is(err, domain.ErrInvalidProjectID):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		default:
			handlers.ErrorHandler(ctx, "internal_error", "failed to get report", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		}
		return
	}

	resp := domainToReportResponse(payload)
	respJSON, err := json.Marshal(resp)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to marshal response", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetBody(respJSON)
}
