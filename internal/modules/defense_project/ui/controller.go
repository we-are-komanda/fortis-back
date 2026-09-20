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
	Import(ctx context.Context, actorID string, rawJSON string) (*domain.DefenseProject, error)
	Export(ctx context.Context, actorID string, projectID string) (string, error)
	CreateFromJSON(ctx context.Context, actorID string, name, enterpriseID, rawJSON string) (*domain.DefenseProject, error)
	ListProjects(ctx context.Context, actorID string, enterpriseID string, limit, offset int) ([]*domain.DefenseProject, int64, error)
	GetProject(ctx context.Context, actorID string, id string) (*domain.DefenseProject, error)
	UpdateProject(ctx context.Context, actorID string, id, name string, enterpriseID *string, projectJSON string, version *int) (*domain.DefenseProject, error)
	DeleteProject(ctx context.Context, actorID string, id string) error
}

type projectRevisionService interface {
	CreateIdempotent(context.Context, string, string, string, string, string) (*domain.DefenseProject, error)
	ImportIdempotent(context.Context, string, string, string) (*domain.DefenseProject, error)
	GetRevision(context.Context, string, string, *int) (*domain.DefenseProject, error)
	ExportRevision(context.Context, string, string, *int) (string, error)
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

	var project *domain.DefenseProject
	var err error
	if key := string(ctx.Request.Header.Peek("Idempotency-Key")); key != "" {
		if service, ok := c.service.(projectRevisionService); ok {
			project, err = service.ImportIdempotent(ctx, handlers.ActorID(ctx), req.ProjectJSON, key)
		} else {
			err = domain.ErrInvalidProjectData
		}
	} else {
		project, err = c.service.Import(ctx, handlers.ActorID(ctx), req.ProjectJSON)
	}
	if err != nil {
		if projectRevisionError(ctx, err) {
			return
		}
		if handlers.AuthorizationError(ctx, err) {
			return
		}
		switch {
		case errors.Is(err, domain.ErrInvalidSchemaVersion):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		case errors.Is(err, domain.ErrInvalidProjectData), errors.Is(err, domain.ErrProjectOwnershipImmutable):
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

	version, ok := requestedProjectVersion(ctx)
	if !ok {
		return
	}
	var jsonStr string
	var err error
	if service, ok := c.service.(projectRevisionService); ok {
		jsonStr, err = service.ExportRevision(ctx, handlers.ActorID(ctx), projectID, version)
	} else if version != nil {
		err = domain.ErrRevisionNotFound
	} else {
		jsonStr, err = c.service.Export(ctx, handlers.ActorID(ctx), projectID)
	}
	if err != nil {
		if projectRevisionError(ctx, err) {
			return
		}
		if handlers.AuthorizationError(ctx, err) {
			return
		}
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

	var project *domain.DefenseProject
	var err error
	if key := string(ctx.Request.Header.Peek("Idempotency-Key")); key != "" {
		if service, ok := c.service.(projectRevisionService); ok {
			project, err = service.CreateIdempotent(ctx, handlers.ActorID(ctx), req.Name, req.EnterpriseID, req.ProjectJSON, key)
		} else {
			err = domain.ErrInvalidProjectData
		}
	} else {
		project, err = c.service.CreateFromJSON(ctx, handlers.ActorID(ctx), req.Name, req.EnterpriseID, req.ProjectJSON)
	}
	if err != nil {
		if projectRevisionError(ctx, err) {
			return
		}
		if handlers.AuthorizationError(ctx, err) {
			return
		}
		switch {
		case errors.Is(err, domain.ErrInvalidSchemaVersion):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		case errors.Is(err, domain.ErrInvalidProjectData), errors.Is(err, domain.ErrProjectOwnershipImmutable):
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

	projects, total, err := c.service.ListProjects(ctx, handlers.ActorID(ctx), enterpriseID, limit, offset)
	if err != nil {
		if handlers.AuthorizationError(ctx, err) {
			return
		}
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

	version, ok := requestedProjectVersion(ctx)
	if !ok {
		return
	}
	var project *domain.DefenseProject
	var err error
	if service, ok := c.service.(projectRevisionService); ok {
		project, err = service.GetRevision(ctx, handlers.ActorID(ctx), projectID, version)
	} else if version != nil {
		err = domain.ErrRevisionNotFound
	} else {
		project, err = c.service.GetProject(ctx, handlers.ActorID(ctx), projectID)
	}
	if err != nil {
		if projectRevisionError(ctx, err) {
			return
		}
		if handlers.AuthorizationError(ctx, err) {
			return
		}
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
// Обновляет имя и содержимое с обязательной expected version. Enterprise ID неизменяем.
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

	project, err := c.service.UpdateProject(ctx, handlers.ActorID(ctx), projectID, req.Name, req.EnterpriseID, req.ProjectJSON, req.Version)
	if err != nil {
		if handlers.AuthorizationError(ctx, err) {
			return
		}
		switch {
		case errors.Is(err, domain.ErrVersionRequired):
			handlers.ErrorHandler(ctx, "version_required", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		case errors.Is(err, domain.ErrProjectNotFound):
			handlers.ErrorHandler(ctx, "not_found", "project not found", &handlers.ResponseBody{}, fasthttp.StatusNotFound)
		case errors.Is(err, domain.ErrInvalidSchemaVersion):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		case errors.Is(err, domain.ErrInvalidProjectData), errors.Is(err, domain.ErrProjectOwnershipImmutable):
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

	if err := c.service.DeleteProject(ctx, handlers.ActorID(ctx), projectID); err != nil {
		if handlers.AuthorizationError(ctx, err) {
			return
		}
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

func requestedProjectVersion(ctx *fasthttp.RequestCtx) (*int, bool) {
	if !ctx.QueryArgs().Has("projectVersion") {
		return nil, true
	}
	version, err := strconv.Atoi(string(ctx.QueryArgs().Peek("projectVersion")))
	if err != nil || version <= 0 {
		handlers.ErrorHandler(ctx, "validation_error", "projectVersion must be a positive integer", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return nil, false
	}
	return &version, true
}

func projectRevisionError(ctx *fasthttp.RequestCtx, err error) bool {
	switch {
	case errors.Is(err, domain.ErrRevisionNotFound):
		handlers.ErrorHandler(ctx, "revision_not_found", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusNotFound)
	case errors.Is(err, domain.ErrIdempotencyConflict):
		handlers.ErrorHandler(ctx, "idempotency_conflict", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusConflict)
	case errors.Is(err, domain.ErrInvalidIdempotencyKey):
		handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
	default:
		return false
	}
	return true
}
