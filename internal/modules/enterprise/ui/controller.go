package ui

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/valyala/fasthttp"

	"github.com/fortis/backend/internal/modules/enterprise/domain"
	"github.com/fortis/backend/pkg/handlers"
)

// EnterpriseServiceInterface — интерфейс сервиса для управления предприятиями.
type EnterpriseServiceInterface interface {
	Create(ctx context.Context, name, address string, status domain.EnterpriseStatus, latitude, longitude float64) (*domain.Enterprise, error)
	Get(ctx context.Context, id string) (*domain.Enterprise, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Enterprise, int64, error)
	Update(ctx context.Context, id, name, address string, status domain.EnterpriseStatus, latitude, longitude float64) (*domain.Enterprise, error)
	Delete(ctx context.Context, id string) error
}

// EnterpriseController — контроллер для управления предприятиями.
type EnterpriseController struct {
	service EnterpriseServiceInterface
}

// NewEnterpriseController создаёт новый контроллер предприятий.
func NewEnterpriseController(service EnterpriseServiceInterface) *EnterpriseController {
	return &EnterpriseController{
		service: service,
	}
}

// swagger:route POST /api/v1/enterprises api createEnterprise
// Создание нового предприятия
//
// Создаёт предприятие с указанными названием, адресом, статусом и координатами.
//
// Consumes:
//   - application/json
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: EnterpriseResponse
//	400: description: Bad Request — неверные данные
//	500: description: Internal Server Error
func (c *EnterpriseController) Create(ctx *fasthttp.RequestCtx) {
	var req CreateEnterpriseRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		handlers.ErrorHandler(ctx, "parse_error", "invalid request body", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	if req.Name == "" {
		handlers.ErrorHandler(ctx, "validation_error", "name is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	enterprise, err := c.service.Create(
		ctx,
		req.Name,
		req.Address,
		domain.EnterpriseStatus(req.Status),
		req.Latitude,
		req.Longitude,
	)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidEnterpriseName):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		case errors.Is(err, domain.ErrInvalidEnterpriseStatus):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		case errors.Is(err, domain.ErrInvalidCoordinates):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		default:
			handlers.ErrorHandler(ctx, "internal_error", "failed to create enterprise", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		}
		return
	}

	resp := enterpriseToResponse(enterprise)
	respJSON, err := json.Marshal(resp)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to marshal response", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetBody(respJSON)
}

// swagger:route GET /api/v1/enterprises api listEnterprises
// Получение списка предприятий или одного предприятия по ID
//
// Если передан query-параметр id, возвращается одно предприятие.
// Иначе возвращается список предприятий с пагинацией (limit, offset).
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: EnterpriseListResponse
//	400: description: Bad Request — неверные параметры
//	404: description: Not Found — предприятие не найдено
//	500: description: Internal Server Error
func (c *EnterpriseController) GetOrList(ctx *fasthttp.RequestCtx) {
	id := string(ctx.QueryArgs().Peek("id"))
	if id != "" {
		c.getByID(ctx, id)
		return
	}

	c.list(ctx)
}

func (c *EnterpriseController) getByID(ctx *fasthttp.RequestCtx, id string) {
	enterprise, err := c.service.Get(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrEnterpriseNotFound):
			handlers.ErrorHandler(ctx, "not_found", "enterprise not found", &handlers.ResponseBody{}, fasthttp.StatusNotFound)
		default:
			handlers.ErrorHandler(ctx, "internal_error", "failed to get enterprise", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		}
		return
	}

	resp := enterpriseToResponse(enterprise)
	respJSON, err := json.Marshal(resp)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to marshal response", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetBody(respJSON)
}

func (c *EnterpriseController) list(ctx *fasthttp.RequestCtx) {
	limit := 20
	offset := 0

	if limitStr := string(ctx.QueryArgs().Peek("limit")); limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
			limit = v
		}
	}
	if offsetStr := string(ctx.QueryArgs().Peek("offset")); offsetStr != "" {
		if v, err := strconv.Atoi(offsetStr); err == nil && v >= 0 {
			offset = v
		}
	}

	enterprises, total, err := c.service.List(ctx, limit, offset)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to list enterprises", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	items := make([]EnterpriseResponse, len(enterprises))
	for i, e := range enterprises {
		items[i] = enterpriseToResponse(e)
	}

	resp := EnterpriseListResponse{
		Items:      items,
		TotalItems: total,
	}

	respJSON, err := json.Marshal(resp)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to marshal response", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetBody(respJSON)
}

// swagger:route PUT /api/v1/enterprises api updateEnterprise
// Обновление предприятия
//
// Обновляет поля существующего предприятия по ID.
//
// Consumes:
//   - application/json
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: EnterpriseResponse
//	400: description: Bad Request — неверные параметры
//	404: description: Not Found — предприятие не найдено
//	500: description: Internal Server Error
func (c *EnterpriseController) Update(ctx *fasthttp.RequestCtx) {
	id := string(ctx.QueryArgs().Peek("id"))
	if id == "" {
		handlers.ErrorHandler(ctx, "validation_error", "id query parameter is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	var req UpdateEnterpriseRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		handlers.ErrorHandler(ctx, "parse_error", "invalid request body", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	enterprise, err := c.service.Update(
		ctx,
		id,
		req.Name,
		req.Address,
		domain.EnterpriseStatus(req.Status),
		req.Latitude,
		req.Longitude,
	)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrEnterpriseNotFound):
			handlers.ErrorHandler(ctx, "not_found", "enterprise not found", &handlers.ResponseBody{}, fasthttp.StatusNotFound)
		case errors.Is(err, domain.ErrInvalidEnterpriseName):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		case errors.Is(err, domain.ErrInvalidEnterpriseStatus):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		case errors.Is(err, domain.ErrInvalidCoordinates):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		default:
			handlers.ErrorHandler(ctx, "internal_error", "failed to update enterprise", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		}
		return
	}

	resp := enterpriseToResponse(enterprise)
	respJSON, err := json.Marshal(resp)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to marshal response", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetBody(respJSON)
}

// swagger:route DELETE /api/v1/enterprises api deleteEnterprise
// Удаление предприятия
//
// Удаляет предприятие по ID.
//
// Responses:
//
//	200: description: Предприятие успешно удалено
//	400: description: Bad Request — не указан ID предприятия
//	404: description: Not Found — предприятие не найдено
//	500: description: Internal Server Error
func (c *EnterpriseController) Delete(ctx *fasthttp.RequestCtx) {
	id := string(ctx.QueryArgs().Peek("id"))
	if id == "" {
		handlers.ErrorHandler(ctx, "validation_error", "id query parameter is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	err := c.service.Delete(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrEnterpriseNotFound):
			handlers.ErrorHandler(ctx, "not_found", "enterprise not found", &handlers.ResponseBody{}, fasthttp.StatusNotFound)
		default:
			handlers.ErrorHandler(ctx, "internal_error", "failed to delete enterprise", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		}
		return
	}

	ctx.SetBodyString(`{"status":"ok"}`)
}
