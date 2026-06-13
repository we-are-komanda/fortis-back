package ui

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/valyala/fasthttp"

	"github.com/fortis/backend/internal/modules/defense_asset/application"
	"github.com/fortis/backend/internal/modules/defense_asset/domain"
	"github.com/fortis/backend/pkg/handlers"
)

// DocumentServiceInterface — интерфейс сервиса для управления документами.
type DocumentServiceInterface interface {
	Create(ctx context.Context, input application.CreateDocumentInput) (*domain.Document, error)
	GetByID(ctx context.Context, id string) (*domain.Document, error)
	ListByAssetID(ctx context.Context, assetID string) ([]*domain.Document, error)
	Delete(ctx context.Context, id string) error
}

// DocumentController — контроллер для управления документами средства защиты.
type DocumentController struct {
	service DocumentServiceInterface
}

// NewDocumentController создаёт новый контроллер документов.
func NewDocumentController(service DocumentServiceInterface) *DocumentController {
	return &DocumentController{
		service: service,
	}
}

// swagger:route GET /api/v1/assets/documents/list api listAssetDocuments
// Получение списка документов средства защиты
//
// Возвращает список документов, прикреплённых к указанному средству защиты.
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: AssetDocumentListResponse
//	400: description: Bad Request — не указан assetId
//	500: description: Internal Server Error
func (c *DocumentController) List(ctx *fasthttp.RequestCtx) {
	assetID := string(ctx.QueryArgs().Peek("assetId"))
	if assetID == "" {
		handlers.ErrorHandler(ctx, "validation_error", "assetId query parameter is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	documents, err := c.service.ListByAssetID(ctx, assetID)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to list documents", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	items := make([]AssetDocumentDTO, len(documents))
	for i, doc := range documents {
		items[i] = documentToDTO(doc)
	}

	resp := AssetDocumentListResponse{
		Items:      items,
		TotalItems: len(items),
	}

	respJSON, err := json.Marshal(resp)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to marshal response", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetBody(respJSON)
}

// swagger:route GET /api/v1/assets/documents/get api getAssetDocument
// Получение метаданных документа
//
// Возвращает метаданные документа по его ID.
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: AssetDocumentDTO
//	400: description: Bad Request — не указан ID
//	404: description: Not Found — документ не найден
//	500: description: Internal Server Error
func (c *DocumentController) Get(ctx *fasthttp.RequestCtx) {
	id := string(ctx.QueryArgs().Peek("id"))
	if id == "" {
		handlers.ErrorHandler(ctx, "validation_error", "id query parameter is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	doc, err := c.service.GetByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDocumentNotFound):
			handlers.ErrorHandler(ctx, "not_found", "document not found", &handlers.ResponseBody{}, fasthttp.StatusNotFound)
		default:
			handlers.ErrorHandler(ctx, "internal_error", "failed to get document", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		}
		return
	}

	resp := documentToDTO(doc)
	respJSON, err := json.Marshal(resp)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to marshal response", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetBody(respJSON)
}

// swagger:route POST /api/v1/assets/documents api createAssetDocument
// Создание метаданных документа
//
// Создаёт запись о документе, прикреплённом к средству защиты.
//
// Consumes:
//   - application/json
//
// Produces:
//   - application/json
//
// Responses:
//
//	201: AssetDocumentDTO
//	400: description: Bad Request — неверные данные
//	500: description: Internal Server Error
func (c *DocumentController) Create(ctx *fasthttp.RequestCtx) {
	var req CreateDocumentRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		handlers.ErrorHandler(ctx, "parse_error", "invalid request body", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	if req.AssetID == "" {
		handlers.ErrorHandler(ctx, "validation_error", "assetId is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}
	if req.Name == "" {
		handlers.ErrorHandler(ctx, "validation_error", "name is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}
	if req.StorageKey == "" {
		handlers.ErrorHandler(ctx, "validation_error", "storageKey is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	input := mapCreateDocumentRequestToServiceInput(req)
	doc, err := c.service.Create(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDocumentInvalidName):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		case errors.Is(err, domain.ErrDocumentInvalidAssetID):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		case errors.Is(err, domain.ErrDocumentInvalidStorageKey):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		default:
			handlers.ErrorHandler(ctx, "internal_error", "failed to create document", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		}
		return
	}

	resp := documentToDTO(doc)
	respJSON, err := json.Marshal(resp)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to marshal response", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusCreated)
	ctx.SetBody(respJSON)
}

// swagger:route GET /api/v1/assets/documents/download api downloadAssetDocument
// Получение download URL документа
//
// Возвращает download URL для скачивания файла документа.
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: AssetDocumentDTO
//	400: description: Bad Request — не указан ID
//	404: description: Not Found — документ не найден
//	500: description: Internal Server Error
func (c *DocumentController) Download(ctx *fasthttp.RequestCtx) {
	id := string(ctx.QueryArgs().Peek("id"))
	if id == "" {
		handlers.ErrorHandler(ctx, "validation_error", "id query parameter is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	doc, err := c.service.GetByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDocumentNotFound):
			handlers.ErrorHandler(ctx, "not_found", "document not found", &handlers.ResponseBody{}, fasthttp.StatusNotFound)
		default:
			handlers.ErrorHandler(ctx, "internal_error", "failed to get document download url", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		}
		return
	}

	resp := documentToDTO(doc)
	respJSON, err := json.Marshal(resp)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to marshal response", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetBody(respJSON)
}

// swagger:route DELETE /api/v1/assets/documents/delete api deleteAssetDocument
// Удаление документа
//
// Удаляет метаданные документа по ID.
//
// Responses:
//
//	200: description: Документ успешно удалён
//	400: description: Bad Request — не указан ID документа
//	404: description: Not Found — документ не найден
//	500: description: Internal Server Error
func (c *DocumentController) Delete(ctx *fasthttp.RequestCtx) {
	id := string(ctx.QueryArgs().Peek("id"))
	if id == "" {
		handlers.ErrorHandler(ctx, "validation_error", "id query parameter is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	err := c.service.Delete(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDocumentNotFound):
			handlers.ErrorHandler(ctx, "not_found", "document not found", &handlers.ResponseBody{}, fasthttp.StatusNotFound)
		default:
			handlers.ErrorHandler(ctx, "internal_error", "failed to delete document", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		}
		return
	}

	ctx.SetBodyString(`{"status":"ok"}`)
}
