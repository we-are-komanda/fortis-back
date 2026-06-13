package ui

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/valyala/fasthttp"

	"github.com/fortis/backend/internal/modules/defense_asset/application"
	"github.com/fortis/backend/internal/modules/defense_asset/domain"
	"github.com/fortis/backend/pkg/handlers"
)

// DefenseAssetServiceInterface — интерфейс сервиса для управления средствами защиты.
type DefenseAssetServiceInterface interface {
	Create(ctx context.Context, input application.CreateInput) (*domain.DefenseAsset, error)
	GetByID(ctx context.Context, id string) (*domain.DefenseAsset, error)
	List(ctx context.Context, enterpriseID *string, isPublic *bool, category *domain.DefenseAssetCategory, limit, offset int) ([]*domain.DefenseAsset, int64, error)
	Update(ctx context.Context, input application.UpdateInput) (*domain.DefenseAsset, error)
	Delete(ctx context.Context, id string) error
}

// DefenseAssetController — контроллер для управления средствами защиты.
type DefenseAssetController struct {
	service DefenseAssetServiceInterface
}

// NewDefenseAssetController создаёт новый контроллер средств защиты.
func NewDefenseAssetController(service DefenseAssetServiceInterface) *DefenseAssetController {
	return &DefenseAssetController{
		service: service,
	}
}

// swagger:route GET /api/v1/assets api listDefenseAssets
// Получение списка средств защиты
//
// Возвращает список средств защиты с пагинацией и фильтрацией по предприятию, признаку общего доступа и категории.
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: DefenseAssetListResponse
//	500: description: Internal Server Error
func (c *DefenseAssetController) List(ctx *fasthttp.RequestCtx) {
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

	enterpriseIDStr := string(ctx.QueryArgs().Peek("enterpriseId"))
	var enterpriseID *string
	if enterpriseIDStr != "" {
		enterpriseID = &enterpriseIDStr
	}

	isPublicStr := string(ctx.QueryArgs().Peek("isPublic"))
	var isPublic *bool
	if isPublicStr == "true" {
		v := true
		isPublic = &v
	} else if isPublicStr == "false" {
		v := false
		isPublic = &v
	}

	categoryStr := string(ctx.QueryArgs().Peek("category"))
	var category *domain.DefenseAssetCategory
	if categoryStr != "" {
		cat := domain.DefenseAssetCategory(categoryStr)
		category = &cat
	}

	assets, total, err := c.service.List(ctx, enterpriseID, isPublic, category, limit, offset)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to list defense assets", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	items := make([]DefenseAssetDTO, len(assets))
	for i, a := range assets {
		items[i] = defenseAssetToDTO(a)
	}

	resp := DefenseAssetListResponse{
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

// swagger:route GET /api/v1/assets/get api getDefenseAsset
// Получение средства защиты по ID
//
// Возвращает одно средство защиты по его ID.
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: DefenseAssetDTO
//	400: description: Bad Request — не указан ID
//	404: description: Not Found — средство защиты не найдено
//	500: description: Internal Server Error
func (c *DefenseAssetController) Get(ctx *fasthttp.RequestCtx) {
	id := string(ctx.QueryArgs().Peek("id"))
	if id == "" {
		handlers.ErrorHandler(ctx, "validation_error", "id query parameter is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	asset, err := c.service.GetByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDefenseAssetNotFound):
			handlers.ErrorHandler(ctx, "not_found", "defense asset not found", &handlers.ResponseBody{}, fasthttp.StatusNotFound)
		default:
			handlers.ErrorHandler(ctx, "internal_error", "failed to get defense asset", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		}
		return
	}

	resp := defenseAssetToDTO(asset)
	respJSON, err := json.Marshal(resp)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to marshal response", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetBody(respJSON)
}

// swagger:route POST /api/v1/assets api createDefenseAsset
// Создание нового средства защиты
//
// Создаёт новое средство защиты с указанными параметрами.
//
// Consumes:
//   - application/json
//
// Produces:
//   - application/json
//
// Responses:
//
//	201: DefenseAssetDTO
//	400: description: Bad Request — неверные данные
//	500: description: Internal Server Error
func (c *DefenseAssetController) Create(ctx *fasthttp.RequestCtx) {
	var req CreateDefenseAssetRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		handlers.ErrorHandler(ctx, "parse_error", "invalid request body", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	if req.Name == "" {
		handlers.ErrorHandler(ctx, "validation_error", "name is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}
	if req.Category == "" {
		handlers.ErrorHandler(ctx, "validation_error", "category is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}
	if req.CoverageType == "" {
		handlers.ErrorHandler(ctx, "validation_error", "coverageType is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	input := mapCreateRequestToDomain(req)
	asset, err := c.service.Create(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDefenseAssetInvalidName):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		case errors.Is(err, domain.ErrDefenseAssetInvalidCategory):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		case errors.Is(err, domain.ErrDefenseAssetInvalidCoverageType):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		default:
			handlers.ErrorHandler(ctx, "internal_error", "failed to create defense asset", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		}
		return
	}

	resp := defenseAssetToDTO(asset)
	respJSON, err := json.Marshal(resp)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to marshal response", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusCreated)
	ctx.SetBody(respJSON)
}

// swagger:route PUT /api/v1/assets/update api updateDefenseAsset
// Обновление средства защиты
//
// Обновляет поля существующего средства защиты по ID.
//
// Consumes:
//   - application/json
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: DefenseAssetDTO
//	400: description: Bad Request — неверные параметры
//	404: description: Not Found — средство защиты не найдено
//	500: description: Internal Server Error
func (c *DefenseAssetController) Update(ctx *fasthttp.RequestCtx) {
	id := string(ctx.QueryArgs().Peek("id"))
	if id == "" {
		handlers.ErrorHandler(ctx, "validation_error", "id query parameter is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	var req UpdateDefenseAssetRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		handlers.ErrorHandler(ctx, "parse_error", "invalid request body", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	input := mapUpdateRequestToServiceInput(id, req)
	asset, err := c.service.Update(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDefenseAssetNotFound):
			handlers.ErrorHandler(ctx, "not_found", "defense asset not found", &handlers.ResponseBody{}, fasthttp.StatusNotFound)
		case errors.Is(err, domain.ErrDefenseAssetInvalidName):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		case errors.Is(err, domain.ErrDefenseAssetInvalidCategory):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		case errors.Is(err, domain.ErrDefenseAssetInvalidCoverageType):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		default:
			handlers.ErrorHandler(ctx, "internal_error", "failed to update defense asset", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		}
		return
	}

	resp := defenseAssetToDTO(asset)
	respJSON, err := json.Marshal(resp)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to marshal response", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetBody(respJSON)
}

// swagger:route DELETE /api/v1/assets/delete api deleteDefenseAsset
// Удаление средства защиты
//
// Удаляет средство защиты по ID.
//
// Responses:
//
//	200: description: Средство защиты успешно удалено
//	400: description: Bad Request — не указан ID средства защиты
//	404: description: Not Found — средство защиты не найдено
//	500: description: Internal Server Error
func (c *DefenseAssetController) Delete(ctx *fasthttp.RequestCtx) {
	id := string(ctx.QueryArgs().Peek("id"))
	if id == "" {
		handlers.ErrorHandler(ctx, "validation_error", "id query parameter is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	err := c.service.Delete(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDefenseAssetNotFound):
			handlers.ErrorHandler(ctx, "not_found", "defense asset not found", &handlers.ResponseBody{}, fasthttp.StatusNotFound)
		default:
			handlers.ErrorHandler(ctx, "internal_error", "failed to delete defense asset", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		}
		return
	}

	ctx.SetBodyString(`{"status":"ok"}`)
}
