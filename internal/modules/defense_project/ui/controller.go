package ui

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/valyala/fasthttp"

	"github.com/fortis/backend/internal/modules/defense_project/domain"
	"github.com/fortis/backend/pkg/handlers"
)

// DefenseProjectServiceInterface — интерфейс сервиса для импорта/экспорта.
type DefenseProjectServiceInterface interface {
	Import(ctx context.Context, rawJSON string) (*domain.DefenseProject, error)
	Export(ctx context.Context, projectID string) (string, error)
}

// DefenseProjectController — контроллер для импорта/экспорта проектов защиты.
type DefenseProjectController struct {
	service DefenseProjectServiceInterface
}

func NewDefenseProjectController(service DefenseProjectServiceInterface) *DefenseProjectController {
	return &DefenseProjectController{
		service: service,
	}
}

// swagger:route POST /api/v1/projects/import api importProject
// Импорт проекта защиты из JSON
//
// Принимает JSON проекта защиты в теле запроса, валидирует,
// сохраняет в БД и возвращает ID сохранённого проекта.
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: ImportResponse
//	400: description: Bad Request — неверный JSON или ошибка валидации
//	500: description: Internal Server Error
func (c *DefenseProjectController) Import(ctx *fasthttp.RequestCtx) {
	var req ImportRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		handlers.ErrorHandler(ctx, "parse_error", "invalid request body", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	if req.ProjectJSON == "" {
		handlers.ErrorHandler(ctx, "validation_error", "projectJson is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	project, err := c.service.Import(ctx, req.ProjectJSON)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidSchemaVersion):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		case errors.Is(err, domain.ErrInvalidProjectData):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		default:
			handlers.ErrorHandler(ctx, "internal_error", "failed to import project", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		}
		return
	}

	resp := domainToImportResponse(project)
	respJSON, err := json.Marshal(resp.Body)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to marshal response", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetBody(respJSON)
}

// swagger:route GET /api/v1/projects/export api exportProject
// Экспорт проекта защиты в JSON
//
// Возвращает полный JSON проекта защиты по ID.
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: description: JSON проекта защиты
//	400: description: Bad Request — не указан ID проекта
//	404: description: Not Found — проект не найден
//	500: description: Internal Server Error
func (c *DefenseProjectController) Export(ctx *fasthttp.RequestCtx) {
	projectID := string(ctx.QueryArgs().Peek("id"))
	if projectID == "" {
		handlers.ErrorHandler(ctx, "validation_error", "id query parameter is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	jsonStr, err := c.service.Export(ctx, projectID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrProjectNotFound):
			handlers.ErrorHandler(ctx, "not_found", "project not found", &handlers.ResponseBody{}, fasthttp.StatusNotFound)
		default:
			handlers.ErrorHandler(ctx, "internal_error", "failed to export project", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		}
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetBodyString(jsonStr)
}


