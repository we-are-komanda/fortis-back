package ui

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/valyala/fasthttp"

	"github.com/fortis/backend/internal/audit"
	"github.com/fortis/backend/internal/modules/budget/domain"
	projectDomain "github.com/fortis/backend/internal/modules/defense_project/domain"
	"github.com/fortis/backend/pkg/handlers"
)

// BudgetServiceInterface — интерфейс сервиса бюджета для контроллера.
type BudgetServiceInterface interface {
	GetBudgetConfig(ctx context.Context, actorID string, projectID string) (*domain.BudgetConfig, error)
	UpdateBudgetConfig(ctx context.Context, actorID string, projectID string, mode domain.BudgetMode, amountMln float64) error
	CalculateCost(ctx context.Context, actorID string, projectID string) (*domain.CostCalculation, error)
	ProjectCost(ctx context.Context, actorID string, projectID string, version *int) (*domain.CostProjection, error)
	CheckBudget(ctx context.Context, actorID string, projectID string, input domain.BudgetCheckInput) (*domain.BudgetCheckResult, error)
	CompareConfigs(ctx context.Context, actorID string, projectID1, projectID2 string) (*domain.ConfigComparison, error)
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

	config, err := c.service.GetBudgetConfig(ctx, handlers.ActorID(ctx), projectID)
	if err != nil {
		if handlers.AuthorizationError(ctx, err) {
			return
		}
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

	if err := c.service.UpdateBudgetConfig(ctx, handlers.ActorID(ctx), projectID, domain.BudgetMode(req.BudgetMode), req.BudgetAmountMln); err != nil {
		if handlers.AuthorizationError(ctx, err) {
			return
		}
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
	config, err := c.service.GetBudgetConfig(ctx, handlers.ActorID(ctx), projectID)
	if err != nil {
		if handlers.AuthorizationError(ctx, err) {
			return
		}
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
//	200: CostProjectionResponse
//	400: description: Bad Request — не указан ID проекта
//	404: description: Not Found — проект не найден
//	500: description: Internal Server Error
func (c *BudgetController) CalculateCost(ctx *fasthttp.RequestCtx) {
	projectID := string(ctx.QueryArgs().Peek("projectId"))
	alias := string(ctx.QueryArgs().Peek("id"))
	if projectID != "" && alias != "" && projectID != alias {
		costError(ctx, 400, "validation_error", "projectId and id disagree", nil)
		return
	}
	if projectID == "" {
		projectID = alias
	}
	if projectID == "" {
		handlers.ErrorHandler(ctx, "validation_error", "id query parameter is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	var version *int
	if raw := string(ctx.QueryArgs().Peek("projectVersion")); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			costError(ctx, 400, "version_required", "projectVersion must be positive", nil)
			return
		}
		version = &n
	}
	calc, err := c.service.ProjectCost(ctx, handlers.ActorID(ctx), projectID, version)
	if err != nil {
		if handlers.AuthorizationError(ctx, err) {
			return
		}
		var invalid *domain.CostValidationError
		switch {
		case errors.As(err, &invalid):
			costError(ctx, 400, invalid.Code, invalid.Message, map[string]any{"field": invalid.Field, "objectIds": invalid.ObjectIDs})
		case errors.Is(err, projectDomain.ErrRevisionNotFound):
			costError(ctx, 404, "revision_not_found", "Project revision not found", nil)
		case errors.Is(err, domain.ErrUnsupportedCalculationVersion):
			costError(ctx, 503, "unsupported_calculation_version", "Frozen calculation version is unavailable", nil)
		default:
			costError(ctx, 500, "calculation_error", "Failed to calculate project cost", nil)
		}
		return
	}

	resp := costProjectionToDTO(calc)
	respJSON, err := json.Marshal(resp)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to marshal response", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetBody(respJSON)
}

func costError(ctx *fasthttp.RequestCtx, status int, code, message string, details map[string]any) {
	requestID := audit.RequestID(ctx)
	ctx.Response.Header.Set("X-Request-ID", requestID)
	errorBody := map[string]any{"code": code, "message": message, "requestId": requestID, "retryable": status >= 500}
	if details != nil {
		errorBody["details"] = details
	}
	body, _ := json.Marshal(map[string]any{"error": errorBody})
	ctx.SetStatusCode(status)
	ctx.SetBody(body)
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
	result, err := c.service.CheckBudget(ctx, handlers.ActorID(ctx), projectID, input)
	if err != nil {
		if handlers.AuthorizationError(ctx, err) {
			return
		}
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

	comp, err := c.service.CompareConfigs(ctx, handlers.ActorID(ctx), projectID1, projectID2)
	if err != nil {
		if handlers.AuthorizationError(ctx, err) {
			return
		}
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
