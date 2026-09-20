package ui

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"strconv"

	"github.com/valyala/fasthttp"

	"github.com/fortis/backend/internal/modules/defense_asset/application"
	"github.com/fortis/backend/internal/modules/defense_asset/domain"
	"github.com/fortis/backend/pkg/handlers"
)

// DocumentServiceInterface — интерфейс сервиса для управления документами.
type DocumentServiceInterface interface {
	Create(ctx context.Context, actorID string, input application.CreateDocumentInput) (*domain.Document, error)
	GetByID(ctx context.Context, actorID string, id string) (*domain.Document, error)
	ListByAssetID(ctx context.Context, actorID string, assetID string) ([]*domain.Document, error)
	Delete(ctx context.Context, actorID string, id string) error
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

	var documents []*domain.Document
	var total int64
	var err error
	if service, ok := c.service.(documentPipelineService); ok {
		limit, _ := strconv.Atoi(string(ctx.QueryArgs().Peek("limit")))
		offset, _ := strconv.Atoi(string(ctx.QueryArgs().Peek("offset")))
		documents, total, err = service.ListPage(ctx, handlers.ActorID(ctx), assetID, limit, offset)
	} else {
		documents, err = c.service.ListByAssetID(ctx, handlers.ActorID(ctx), assetID)
		total = int64(len(documents))
	}
	if err != nil {
		if handlers.AuthorizationError(ctx, err) {
			return
		}
		handlers.ErrorHandler(ctx, "internal_error", "failed to list documents", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	items := make([]AssetDocumentDTO, len(documents))
	for i, doc := range documents {
		items[i] = documentToDTO(doc)
	}

	resp := AssetDocumentListResponse{
		Items:      items,
		TotalItems: int(total),
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
//	200: AssetDocumentResponse
//	400: description: Bad Request — не указан ID
//	404: description: Not Found — документ не найден
//	500: description: Internal Server Error
func (c *DocumentController) Get(ctx *fasthttp.RequestCtx) {
	id := string(ctx.QueryArgs().Peek("id"))
	if id == "" {
		handlers.ErrorHandler(ctx, "validation_error", "id query parameter is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	doc, err := c.service.GetByID(ctx, handlers.ActorID(ctx), id)
	if err != nil {
		if handlers.AuthorizationError(ctx, err) {
			return
		}
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

// Create загружает файл по multipart-контракту UploadDocumentRequest.
func (c *DocumentController) Create(ctx *fasthttp.RequestCtx) {
	if len(ctx.PostBody()) > domain.MaxDocumentBytes+65536 || ctx.Request.Header.ContentLength() > domain.MaxDocumentBytes+65536 {
		documentError(ctx, domain.ErrDocumentTooLarge)
		return
	}
	kind, _, _ := mime.ParseMediaType(string(ctx.Request.Header.ContentType()))
	if kind != "multipart/form-data" {
		var req CreateDocumentRequest
		if json.Unmarshal(ctx.PostBody(), &req) != nil {
			documentError(ctx, domain.ErrDocumentUploadRequired)
			return
		}
		if req.AssetID == "" {
			documentError(ctx, domain.ErrDocumentInvalidAssetID)
			return
		}
		if req.Name == "" || req.StorageKey == "" {
			documentError(ctx, domain.ErrDocumentUploadRequired)
			return
		}
		// Keep the legacy parent permission check, while closing metadata URL writes.
		_, err := c.service.Create(ctx, handlers.ActorID(ctx), mapCreateDocumentRequestToServiceInput(req))
		if err == nil {
			err = domain.ErrDocumentUploadRequired
		}
		documentError(ctx, err)
		return
	}
	service, ok := c.service.(documentPipelineService)
	if !ok {
		documentError(ctx, domain.ErrDocumentUnavailable)
		return
	}
	form, err := ctx.MultipartForm()
	if err != nil {
		documentError(ctx, domain.ErrDocumentUploadRequired)
		return
	}
	files := form.File["file"]
	if len(files) != 1 || len(form.File) != 1 {
		documentError(ctx, domain.ErrDocumentUploadRequired)
		return
	}
	for key, values := range form.Value {
		if (key != "assetId" && key != "commercial") || len(values) != 1 {
			documentError(ctx, domain.ErrDocumentUploadRequired)
			return
		}
	}
	if len(form.Value["assetId"]) != 1 {
		documentError(ctx, domain.ErrDocumentInvalidAssetID)
		return
	}
	commercial := true
	if values := form.Value["commercial"]; len(values) > 0 {
		commercial, err = strconv.ParseBool(values[0])
		if err != nil {
			documentError(ctx, domain.ErrDocumentUploadRequired)
			return
		}
	}
	file := files[0]
	if file.Size > domain.MaxDocumentBytes {
		documentError(ctx, domain.ErrDocumentTooLarge)
		return
	}
	reader, err := file.Open()
	if err != nil {
		documentError(ctx, domain.ErrDocumentUnavailable)
		return
	}
	body, err := io.ReadAll(io.LimitReader(reader, domain.MaxDocumentBytes+1))
	reader.Close()
	if err != nil {
		documentError(ctx, domain.ErrDocumentUnavailable)
		return
	}
	doc, err := service.Upload(ctx, handlers.ActorID(ctx), application.UploadDocumentInput{AssetID: form.Value["assetId"][0], Name: file.Filename, MimeType: file.Header.Get("Content-Type"), Body: body, Commercial: commercial})
	if err != nil {
		documentError(ctx, err)
		return
	}
	raw, err := json.Marshal(documentToDTO(doc))
	if err != nil {
		documentError(ctx, err)
		return
	}
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusCreated)
	ctx.SetBody(raw)
}

// swagger:route GET /api/v1/assets/documents/download api downloadAssetDocument
// Скачивание приватного документа
//
// Повторно проверяет доступ к родителю и возвращает проверенные байты вложением.
//
// Produces:
//   - application/pdf
//   - image/png
//   - image/jpeg
//   - text/plain
//
// Responses:
//
//	200: AssetDocumentDownloadResponse
//	400: description: Bad Request — не указан ID
//	404: description: Not Found — документ не найден
//	500: description: Internal Server Error
func (c *DocumentController) Download(ctx *fasthttp.RequestCtx) {
	id := string(ctx.QueryArgs().Peek("id"))
	if id == "" {
		documentError(ctx, domain.ErrDocumentInvalidAssetID)
		return
	}
	service, ok := c.service.(documentPipelineService)
	if !ok {
		documentError(ctx, domain.ErrDocumentUnavailable)
		return
	}
	doc, body, err := service.Download(ctx, handlers.ActorID(ctx), id, string(ctx.QueryArgs().Peek("assetId")))
	if err != nil {
		documentError(ctx, err)
		return
	}
	ctx.SetContentType(doc.MimeType())
	ctx.Response.Header.Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": doc.Name()}))
	ctx.Response.Header.Set("X-Content-Type-Options", "nosniff")
	ctx.Response.Header.Set("Cache-Control", "private, no-store")
	ctx.SetBody(body)
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

	err := c.service.Delete(ctx, handlers.ActorID(ctx), id)
	if err != nil {
		if handlers.AuthorizationError(ctx, err) {
			return
		}
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

type documentPipelineService interface {
	Upload(context.Context, string, application.UploadDocumentInput) (*domain.Document, error)
	Download(context.Context, string, string, string) (*domain.Document, []byte, error)
	ListPage(context.Context, string, string, int, int) ([]*domain.Document, int64, error)
}

func documentError(ctx *fasthttp.RequestCtx, err error) {
	if handlers.AuthorizationError(ctx, err) {
		return
	}
	code, status, message := "internal_error", 500, "document operation failed"
	switch {
	case errors.Is(err, domain.ErrDocumentUnavailable):
		code, status, message = "document_unavailable", 503, err.Error()
	case errors.Is(err, domain.ErrDocumentNotReady), errors.Is(err, domain.ErrDocumentNotFound):
		code, status, message = "document_unavailable", 404, "document is unavailable"
	case errors.Is(err, domain.ErrDocumentTooLarge):
		code, status, message = "document_too_large", 413, err.Error()
	case errors.Is(err, domain.ErrDocumentMediaType):
		code, status, message = "unsupported_media_type", 415, err.Error()
	case errors.Is(err, domain.ErrDocumentRejected):
		code, status, message = "document_rejected", 422, err.Error()
	case errors.Is(err, domain.ErrDocumentUploadRequired):
		code, status, message = "document_upload_required", 400, err.Error()
	case errors.Is(err, domain.ErrDocumentInvalidName), errors.Is(err, domain.ErrDocumentInvalidAssetID):
		code, status, message = "validation_error", 400, err.Error()
	}
	handlers.ErrorHandler(ctx, code, message, &handlers.ResponseBody{}, status)
}
