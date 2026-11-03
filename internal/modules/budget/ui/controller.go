package ui

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"strconv"

	"github.com/valyala/fasthttp"

	"github.com/fortis/backend/internal/audit"
	budgetApp "github.com/fortis/backend/internal/modules/budget/application"
	"github.com/fortis/backend/internal/modules/budget/domain"
	projectDomain "github.com/fortis/backend/internal/modules/defense_project/domain"
	projectUi "github.com/fortis/backend/internal/modules/defense_project/ui"
	"github.com/fortis/backend/pkg/handlers"
)

// BudgetServiceInterface — интерфейс сервиса бюджета для контроллера.
type BudgetServiceInterface interface {
	ProjectBudget(ctx context.Context, actorID string, projectID string, version *int) (*domain.BudgetProjection, error)
	UpdateBudget(ctx context.Context, actorID string, projectID string, command budgetApp.UpdateBudgetCommand) (*projectDomain.DefenseProject, *domain.BudgetProjection, error)
	CalculateCost(ctx context.Context, actorID string, projectID string) (*domain.CostCalculation, error)
	ProjectCost(ctx context.Context, actorID string, projectID string, version *int) (*domain.CostProjection, error)
	CheckCandidate(ctx context.Context, actorID string, projectID string, command budgetApp.CheckBudgetCommand) (*domain.CandidateBudgetCheck, error)
	CompareConfigs(ctx context.Context, actorID string, a, b projectDomain.ProjectRef) (*domain.RevisionComparison, error)
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
// Возвращает точные настройки и стоимость сохранённой ревизии. Без версии читается текущая.
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: BudgetProjectionResponse
//	400: description: Bad Request — не указан ID проекта
//	404: description: Not Found — бюджет не найден
//	500: description: Internal Server Error
func (c *BudgetController) GetBudgetConfig(ctx *fasthttp.RequestCtx) {
	id, version, ok := budgetQuery(ctx, true)
	if !ok {
		return
	}
	projection, err := c.service.ProjectBudget(ctx, handlers.ActorID(ctx), id, version)
	if budgetResponseError(ctx, err) {
		return
	}
	raw, _ := json.Marshal(budgetProjectionToDTO(projection))
	ctx.SetBody(raw)
}

// swagger:route PUT /api/v1/projects/budget api updateBudgetConfig
// Обновить конфигурацию бюджета проекта
//
// CAS по общей версии проекта; записывает полную ревизию и возвращает её вместе с бюджетом.
// Снижение лимита ниже стоимости допускается; дальнейшие превышающие лимит изменения не могут увеличивать стоимость.
//
// Consumes:
//   - application/json
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: BudgetUpdateResponse
//	409: description: Conflict — версия изменилась
//	400: description: Bad Request — неверные параметры
//	500: description: Internal Server Error
func (c *BudgetController) UpdateBudgetConfig(ctx *fasthttp.RequestCtx) {
	id, _, ok := budgetQuery(ctx, false)
	if !ok {
		return
	}
	var req UpdateBudgetRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		costError(ctx, 400, "validation_error", "Invalid budget request", nil)
		return
	}
	// Canonical limit is required explicitly, including null for unlimited.
	if len(req.BudgetLimitMinor) == 0 {
		costError(ctx, 400, "validation_error", "budgetLimitMinor is required", nil)
		return
	}
	var limit *string
	if err := json.Unmarshal(req.BudgetLimitMinor, &limit); err != nil {
		costError(ctx, 400, "validation_error", "budgetLimitMinor must be a string or null", nil)
		return
	}
	project, projection, err := c.service.UpdateBudget(ctx, handlers.ActorID(ctx), id, budgetApp.UpdateBudgetCommand{Version: req.Version, Mode: domain.BudgetMode(req.BudgetMode), LimitMinor: limit})
	if budgetResponseError(ctx, err) {
		return
	}
	raw, _ := json.Marshal(BudgetUpdateResponse{ProjectResponse: projectUi.ProjectResponseFromDomain(project), Budget: budgetProjectionToDTO(projection)})
	ctx.SetBody(raw)
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
// Проверить полный кандидат конфигурации по бюджету
//
// Проверка без записи или резервирования. 200 содержит allowed/blocked и точную причину;
// сохранение повторяет проверку в транзакции и при запрете возвращает 422.
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
	id, _, ok := budgetQuery(ctx, false)
	if !ok {
		return
	}
	var req BudgetCheckRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		costError(ctx, 400, "validation_error", "Invalid candidate request", nil)
		return
	}
	result, err := c.service.CheckCandidate(ctx, handlers.ActorID(ctx), id, budgetApp.CheckBudgetCommand{Version: req.Version, ProjectJSON: req.ProjectJSON})
	if budgetResponseError(ctx, err) {
		return
	}
	status := "blocked"
	if result.Allowed() {
		status = "allowed"
	}
	identity := result.Projection().Identity()
	raw, _ := json.Marshal(BudgetCheckResponse{BasedOn: ProjectRefDTO{ProjectID: identity.ProjectID(), ProjectVersion: identity.ProjectVersion()}, Status: status, ReasonCode: result.ReasonCode(), ProjectedSpentMinor: result.Projection().SpentMinor(), ProjectedRemainingMinor: result.Projection().RemainingMinor()})
	ctx.SetBody(raw)
}

func budgetQuery(ctx *fasthttp.RequestCtx, versioned bool) (string, *int, bool) {
	id, alias := string(ctx.QueryArgs().Peek("projectId")), string(ctx.QueryArgs().Peek("id"))
	if id != "" && alias != "" && id != alias {
		costError(ctx, 400, "validation_error", "projectId and id disagree", nil)
		return "", nil, false
	}
	if id == "" {
		id = alias
	}
	if id == "" {
		costError(ctx, 400, "validation_error", "projectId is required", nil)
		return "", nil, false
	}
	var version *int
	if versioned && ctx.QueryArgs().Has("projectVersion") {
		v, err := strconv.Atoi(string(ctx.QueryArgs().Peek("projectVersion")))
		if err != nil || v <= 0 {
			costError(ctx, 400, "version_required", "projectVersion must be positive", nil)
			return "", nil, false
		}
		version = &v
	}
	return id, version, true
}

func budgetResponseError(ctx *fasthttp.RequestCtx, err error) bool {
	if err == nil {
		return false
	}
	if handlers.AuthorizationError(ctx, err) || handlers.FinancialError(ctx, err) {
		return true
	}
	switch {
	case errors.Is(err, projectDomain.ErrVersionConflict):
		costError(ctx, 409, "version_conflict", "Project version has changed", nil)
	case errors.Is(err, projectDomain.ErrVersionRequired):
		costError(ctx, 400, "version_required", "Positive version is required", nil)
	case errors.Is(err, projectDomain.ErrRevisionNotFound):
		costError(ctx, 404, "revision_not_found", "Project revision not found", nil)
	case errors.Is(err, projectDomain.ErrProjectNotFound):
		costError(ctx, 404, "not_found", "Project not found", nil)
	case errors.Is(err, projectDomain.ErrInvalidProjectData), errors.Is(err, projectDomain.ErrProjectOwnershipImmutable):
		costError(ctx, 400, "validation_error", "Invalid candidate project", nil)
	case errors.Is(err, domain.ErrUnsupportedCalculationVersion):
		costError(ctx, 503, "unsupported_calculation_version", "Frozen calculation version is unavailable", nil)
	default:
		costError(ctx, 500, "budget_error", "Budget operation failed", nil)
	}
	return true
}

// swagger:route GET /api/v1/projects/compare api compareConfigs
// Сравнить две конфигурации проектов
//
// Сравнивает две явно зафиксированные ревизии одного предприятия.
// Разница B−A — точные копейки; процент null при A=0 или неполных итогах.
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
	read := func(side, alias string) (projectDomain.ProjectRef, bool) {
		id, old := string(ctx.QueryArgs().Peek("projectId"+side)), string(ctx.QueryArgs().Peek(alias))
		if id != "" && old != "" && id != old {
			costError(ctx, 400, "validation_error", "Project ID aliases disagree", nil)
			return projectDomain.ProjectRef{}, false
		}
		if id == "" {
			id = old
		}
		parsed, err := uuid.Parse(id)
		if err != nil {
			costError(ctx, 400, "validation_error", "Both valid project IDs are required", nil)
			return projectDomain.ProjectRef{}, false
		}
		version, err := strconv.Atoi(string(ctx.QueryArgs().Peek("projectVersion" + side)))
		if err != nil || version <= 0 {
			costError(ctx, 400, "version_required", "Both positive project versions are required", nil)
			return projectDomain.ProjectRef{}, false
		}
		return projectDomain.ProjectRef{ProjectID: parsed.String(), ProjectVersion: version}, true
	}
	a, ok := read("A", "id1")
	if !ok {
		return
	}
	b, ok := read("B", "id2")
	if !ok {
		return
	}
	result, err := c.service.CompareConfigs(ctx, handlers.ActorID(ctx), a, b)
	if errors.Is(err, domain.ErrIncompatibleProjects) {
		costError(ctx, 400, "incompatible_projects", "Choose distinct revisions from the same enterprise", nil)
		return
	}
	if budgetResponseError(ctx, err) {
		return
	}
	raw, _ := json.Marshal(comparisonToDTO(result))
	ctx.SetBody(raw)
}
