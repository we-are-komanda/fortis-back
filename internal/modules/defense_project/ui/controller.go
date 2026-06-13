package ui

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/valyala/fasthttp"

	"github.com/fortis/backend/internal/modules/defense_project/domain"
	"github.com/fortis/backend/pkg/handlers"
)

// DefenseProjectServiceInterface — интерфейс сервиса для операций с проектами защиты.
type DefenseProjectServiceInterface interface {
	Import(ctx context.Context, rawJSON string) (*domain.DefenseProject, error)
	Export(ctx context.Context, projectID string) (string, error)
	CreateFromJSON(ctx context.Context, name, enterpriseID, rawJSON string) (*domain.DefenseProject, error)
	ListProjects(ctx context.Context, enterpriseID string, limit, offset int) ([]*domain.DefenseProject, int64, error)
	GetProject(ctx context.Context, id string) (*domain.DefenseProject, error)
	UpdateProject(ctx context.Context, id, name, enterpriseID, projectJSON string) (*domain.DefenseProject, error)
	DeleteProject(ctx context.Context, id string) error
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
		case errors.Is(err, domain.ErrVersionConflict):
			handlers.ErrorHandler(ctx, "version_conflict", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusConflict)
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

// swagger:route POST /api/v1/projects api createProject
// Создание новой конфигурации проекта защиты
//
// Принимает имя конфигурации и полный JSON проекта защиты в теле запроса.
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: ProjectResponse
//	400: description: Bad Request — ошибка валидации
//	500: description: Internal Server Error
func (c *DefenseProjectController) Create(ctx *fasthttp.RequestCtx) {
	var req CreateProjectRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		handlers.ErrorHandler(ctx, "parse_error", "invalid request body", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	if req.Name == "" {
		handlers.ErrorHandler(ctx, "validation_error", "name is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	if req.ProjectJSON == "" {
		handlers.ErrorHandler(ctx, "validation_error", "projectJson is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	project, err := c.service.CreateFromJSON(ctx, req.Name, req.EnterpriseID, req.ProjectJSON)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidSchemaVersion):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		case errors.Is(err, domain.ErrInvalidProjectData):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		case errors.Is(err, domain.ErrInvalidConfigName):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		case errors.Is(err, domain.ErrVersionConflict):
			handlers.ErrorHandler(ctx, "version_conflict", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusConflict)
		default:
			handlers.ErrorHandler(ctx, "internal_error", "failed to create project", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		}
		return
	}

	resp := domainToProjectResponse(project)
	respJSON, err := json.Marshal(resp)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to marshal response", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetBody(respJSON)
}

// swagger:route GET /api/v1/projects api listProjects
// Список конфигураций проектов защиты
//
// Возвращает список проектов с пагинацией.
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: ProjectListResponse
//	500: description: Internal Server Error
func (c *DefenseProjectController) List(ctx *fasthttp.RequestCtx) {
	limit := 50
	if v := string(ctx.QueryArgs().Peek("limit")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	offset := 0
	if v := string(ctx.QueryArgs().Peek("offset")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	enterpriseID := string(ctx.QueryArgs().Peek("enterpriseId"))

	projects, total, err := c.service.ListProjects(ctx, enterpriseID, limit, offset)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to list projects", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	resp := domainToProjectListResponse(projects, total)
	respJSON, err := json.Marshal(resp)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to marshal response", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetBody(respJSON)
}

// swagger:route GET /api/v1/projects/get api getProject
// Получение конфигурации проекта защиты по ID
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: ProjectResponse
//	400: description: Bad Request — не указан ID
//	404: description: Not Found — проект не найден
//	500: description: Internal Server Error
func (c *DefenseProjectController) Get(ctx *fasthttp.RequestCtx) {
	projectID := string(ctx.QueryArgs().Peek("id"))
	if projectID == "" {
		handlers.ErrorHandler(ctx, "validation_error", "id query parameter is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	project, err := c.service.GetProject(ctx, projectID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrProjectNotFound):
			handlers.ErrorHandler(ctx, "not_found", "project not found", &handlers.ResponseBody{}, fasthttp.StatusNotFound)
		default:
			handlers.ErrorHandler(ctx, "internal_error", "failed to get project", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		}
		return
	}

	resp := domainToProjectResponse(project)
	respJSON, err := json.Marshal(resp)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to marshal response", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetBody(respJSON)
}

// swagger:route PUT /api/v1/projects/update api updateProject
// Обновление конфигурации проекта защиты
//
// Обновляет имя и enterprise ID проекта.
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: ProjectResponse
//	400: description: Bad Request — не указан ID
//	404: description: Not Found — проект не найден
//	500: description: Internal Server Error
func (c *DefenseProjectController) Update(ctx *fasthttp.RequestCtx) {
	projectID := string(ctx.QueryArgs().Peek("id"))
	if projectID == "" {
		handlers.ErrorHandler(ctx, "validation_error", "id query parameter is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	var req UpdateProjectRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		handlers.ErrorHandler(ctx, "parse_error", "invalid request body", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	project, err := c.service.UpdateProject(ctx, projectID, req.Name, req.EnterpriseID, req.ProjectJSON)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrProjectNotFound):
			handlers.ErrorHandler(ctx, "not_found", "project not found", &handlers.ResponseBody{}, fasthttp.StatusNotFound)
		case errors.Is(err, domain.ErrInvalidSchemaVersion):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		case errors.Is(err, domain.ErrInvalidProjectData):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		case errors.Is(err, domain.ErrVersionConflict):
			handlers.ErrorHandler(ctx, "version_conflict", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusConflict)
		default:
			handlers.ErrorHandler(ctx, "internal_error", "failed to update project", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		}
		return
	}

	resp := domainToProjectResponse(project)
	respJSON, err := json.Marshal(resp)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to marshal response", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetBody(respJSON)
}

// swagger:route DELETE /api/v1/projects/delete api deleteProject
// Удаление конфигурации проекта защиты
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: description: OK — проект удалён
//	400: description: Bad Request — не указан ID
//	404: description: Not Found — проект не найден
//	500: description: Internal Server Error
func (c *DefenseProjectController) Delete(ctx *fasthttp.RequestCtx) {
	projectID := string(ctx.QueryArgs().Peek("id"))
	if projectID == "" {
		handlers.ErrorHandler(ctx, "validation_error", "id query parameter is required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	if err := c.service.DeleteProject(ctx, projectID); err != nil {
		switch {
		case errors.Is(err, domain.ErrProjectNotFound):
			handlers.ErrorHandler(ctx, "not_found", "project not found", &handlers.ResponseBody{}, fasthttp.StatusNotFound)
		default:
			handlers.ErrorHandler(ctx, "internal_error", "failed to delete project", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		}
		return
	}

	ctx.SetBodyString(`{"status":"ok"}`)
}


