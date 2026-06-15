package ui

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/valyala/fasthttp"

	"github.com/fortis/backend/internal/modules/budget/domain"
	"github.com/fortis/backend/pkg/handlers"
)

// BudgetServiceInterface — интерфейс сервиса бюджета для контроллера.
type BudgetServiceInterface interface {
	GetBudgetConfig(ctx context.Context, projectID string) (*domain.BudgetConfig, error)
	UpdateBudgetConfig(ctx context.Context, projectID string, mode domain.BudgetMode, amountMln float64) error
	CalculateCost(ctx context.Context, projectID string) (*domain.CostCalculation, error)
	CheckBudget(ctx context.Context, projectID string, input domain.BudgetCheckInput) (*domain.BudgetCheckResult, error)
	CompareConfigs(ctx context.Context, projectID1, projectID2 string) (*domain.ConfigComparison, error)
}

// BudgetController — контроллер для API бюджета и стоимости.
type BudgetController struct {
	service BudgetServiceInterface
}

// NewBudgetController создаёт новый BudgetController.
func NewBudgetController(service BudgetServiceInterface) *BudgetController {
	return &BudgetController{
		service: service,
	}
}

// swagger:route GET /api/v1/projects/budget api getBudgetConfig
// Получить конфигурацию бюджета проекта
//
// Возвращает настройки бюджета для указанного проекта.
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: BudgetConfigResponse
//	400: description: Bad Request — не указан ID проекта
//	404: description: Not Found — бюджет не найден
//	500: description: Internal Server Error
func (c *BudgetController) GetBudgetConfig(ctx *fasthttp.RequestCtx) {
	projectID := string(ctx.QueryArgs().Peek("id"))
	if projectID == "" {
		handlers.ErrorHandler(ctx, "validation_error", "id query parameter is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	config, err := c.service.GetBudgetConfig(ctx, projectID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrBudgetConfigNotFound):
			handlers.ErrorHandler(ctx, "not_found", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusNotFound)
		default:
			handlers.ErrorHandler(ctx, "internal_error", "failed to get budget config", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		}
		return
	}

	resp := budgetConfigToDTO(config)
	respJSON, err := json.Marshal(resp)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to marshal response", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetBody(respJSON)
}

// swagger:route PUT /api/v1/projects/budget api updateBudgetConfig
// Обновить конфигурацию бюджета проекта
//
// Создаёт или обновляет настройки бюджета для указанного проекта.
//
// Consumes:
//   - application/json
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: description: Budget config updated successfully
//	400: description: Bad Request — неверные параметры
//	500: description: Internal Server Error
func (c *BudgetController) UpdateBudgetConfig(ctx *fasthttp.RequestCtx) {
	projectID := string(ctx.QueryArgs().Peek("id"))
	if projectID == "" {
		handlers.ErrorHandler(ctx, "validation_error", "id query parameter is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	var req UpdateBudgetRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		handlers.ErrorHandler(ctx, "parse_error", "invalid request body", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	if req.BudgetMode == "" {
		handlers.ErrorHandler(ctx, "validation_error", "budgetMode is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	if err := c.service.UpdateBudgetConfig(ctx, projectID, domain.BudgetMode(req.BudgetMode), req.BudgetAmountMln); err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidBudgetMode):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		case errors.Is(err, domain.ErrInvalidBudgetAmount):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		default:
			handlers.ErrorHandler(ctx, "internal_error", "failed to update budget config", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		}
		return
	}

	// Возвращаем обновлённую конфигурацию
	config, err := c.service.GetBudgetConfig(ctx, projectID)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to get updated budget config", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	resp := budgetConfigToDTO(config)
	respJSON, err := json.Marshal(resp)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to marshal response", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetBody(respJSON)
}

// swagger:route GET /api/v1/projects/cost api calculateCost
// Рассчитать стоимость конфигурации проекта
//
// Выполняет полный расчёт стоимости размещённых средств защиты с группировкой
// по эшелонам и типам.
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: CostCalculationResponse
//	400: description: Bad Request — не указан ID проекта
//	404: description: Not Found — проект не найден
//	500: description: Internal Server Error
func (c *BudgetController) CalculateCost(ctx *fasthttp.RequestCtx) {
	projectID := string(ctx.QueryArgs().Peek("id"))
	if projectID == "" {
		handlers.ErrorHandler(ctx, "validation_error", "id query parameter is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	calc, err := c.service.CalculateCost(ctx, projectID)
	if err != nil {
		handlers.ErrorHandler(ctx, "calculation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	resp := costCalculationToDTO(calc)
	respJSON, err := json.Marshal(resp)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to marshal response", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetBody(respJSON)
}

// swagger:route POST /api/v1/projects/budget/check api checkBudget
// Проверить добавление средства по бюджету
//
// Проверяет, помещается ли добавление указанного средства защиты
// в остаток бюджета проекта.
//
// Consumes:
//   - application/json
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: BudgetCheckResponse
//	400: description: Bad Request — неверные параметры
//	404: description: Not Found — проект или бюджет не найден
//	500: description: Internal Server Error
func (c *BudgetController) CheckBudget(ctx *fasthttp.RequestCtx) {
	projectID := string(ctx.QueryArgs().Peek("id"))
	if projectID == "" {
		handlers.ErrorHandler(ctx, "validation_error", "id query parameter is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	var req BudgetCheckRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		handlers.ErrorHandler(ctx, "parse_error", "invalid request body", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	if req.AssetID == "" {
		handlers.ErrorHandler(ctx, "validation_error", "assetId is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}
	if req.Quantity <= 0 {
		handlers.ErrorHandler(ctx, "validation_error", "quantity must be greater than 0", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	input := domain.NewBudgetCheckInput(req.AssetID, req.Quantity, req.EchelonID)
	result, err := c.service.CheckBudget(ctx, projectID, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrBudgetConfigNotFound):
			handlers.ErrorHandler(ctx, "not_found", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusNotFound)
		default:
			handlers.ErrorHandler(ctx, "internal_error", "failed to check budget", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		}
		return
	}

	resp := checkResultToDTO(result)
	respJSON, err := json.Marshal(resp)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to marshal response", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetBody(respJSON)
}

// swagger:route GET /api/v1/projects/compare api compareConfigs
// Сравнить две конфигурации проектов
//
// Возвращает структурный профиль и стоимость для каждой конфигурации,
// а также diff между ними.
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: ConfigComparisonResponse
//	400: description: Bad Request — не указаны ID проектов
//	404: description: Not Found — проект не найден
//	500: description: Internal Server Error
func (c *BudgetController) Compare(ctx *fasthttp.RequestCtx) {
	projectID1 := string(ctx.QueryArgs().Peek("id1"))
	projectID2 := string(ctx.QueryArgs().Peek("id2"))

	if projectID1 == "" || projectID2 == "" {
		handlers.ErrorHandler(ctx, "validation_error", "both id1 and id2 query parameters are required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	comp, err := c.service.CompareConfigs(ctx, projectID1, projectID2)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrBothIDsRequired):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		case errors.Is(err, domain.ErrComparisonFailed):
			handlers.ErrorHandler(ctx, "comparison_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		default:
			handlers.ErrorHandler(ctx, "internal_error", "failed to compare configs", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		}
		return
	}

	resp := comparisonToDTO(comp)
	respJSON, err := json.Marshal(resp)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to marshal response", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetBody(respJSON)
}
