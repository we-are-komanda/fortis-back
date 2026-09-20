package ui

import "encoding/json"

// ImportRequest DTO запроса на импорт проекта.
type ImportRequest struct {
	// JSON проекта защиты
	// Required: true
	// In: body
	ProjectJSON string `json:"projectJson"`
}

// ImportProjectParams описывает тело и ключ идемпотентности импорта.
// swagger:parameters importProject
type ImportProjectParams struct {
	// Ключ повторного запроса; область — текущий пользователь и операция, срок — 24 часа.
	// In: header
	IdempotencyKey string `json:"Idempotency-Key"`
	// In: body
	// Required: true
	Body ImportRequest
}

// ImportResponse DTO ответа после успешного импорта.
// swagger:response ImportResponse
type ImportResponse struct {
	// In: body
	Body struct {
		ProjectVersion int             `json:"projectVersion"`
		SnapshotDigest string          `json:"snapshotDigest,omitempty"`
		Snapshot       json.RawMessage `json:"snapshot,omitempty"`
		// ID сохранённого проекта
		// Example: 550e8400-e29b-41d4-a716-446655440000
		ProjectID string `json:"projectId"`
		// Название проекта
		// Example: Защита объекта Альфа
		ProjectName string `json:"projectName"`
		// Версия проекта для optimistic locking
		// Example: 1
		Version int `json:"version"`
		// Дата обновления проекта
		// Example: 2026-06-12T14:00:00.000Z
		UpdatedAt string `json:"updatedAt"`
	} `json:"body"`
}

// ExportQuery DTO query-параметров для экспорта проекта.
// swagger:parameters exportProject
type ExportQuery struct {
	// Версия сохранённого снимка; без параметра возвращается текущая.
	// In: query
	// Minimum: 1
	ProjectVersion *int `json:"projectVersion,omitempty"`
	// ID проекта для экспорта
	// Required: true
	// In: query
	ID string `json:"id"`
}

// ---- CRUD DTOs ----

// CreateProjectRequest DTO запроса на создание конфигурации.
type CreateProjectRequest struct {
	// Имя конфигурации
	// Required: true
	// Example: Моя конфигурация
	// In: body
	Name string `json:"name"`
	// Enterprise ID для мультитенантности
	// Example: 550e8400-e29b-41d4-a716-446655440000
	// In: body
	EnterpriseID string `json:"enterpriseId,omitempty"`
	// JSON проекта защиты (полный DefenseProject)
	// Required: true
	// In: body
	ProjectJSON string `json:"projectJson"`
}

// CreateProjectParams описывает тело и ключ идемпотентности создания.
// swagger:parameters createProject
type CreateProjectParams struct {
	// Ключ повторного запроса; область — текущий пользователь и операция, срок — 24 часа.
	// In: header
	IdempotencyKey string `json:"Idempotency-Key"`
	// In: body
	// Required: true
	Body CreateProjectRequest
}

// ProjectResponse DTO ответа с данными проекта.
// swagger:model ProjectResponse
type ProjectResponse struct {
	ProjectVersion int             `json:"projectVersion"`
	SnapshotDigest string          `json:"snapshotDigest,omitempty"`
	Snapshot       json.RawMessage `json:"snapshot,omitempty"`
	// ID проекта
	// Example: 550e8400-e29b-41d4-a716-446655440000
	ProjectID string `json:"projectId"`
	// Имя конфигурации
	// Example: Моя конфигурация
	Name string `json:"name"`
	// Enterprise ID для мультитенантности
	// Example: 550e8400-e29b-41d4-a716-446655440000
	EnterpriseID string `json:"enterpriseId,omitempty"`
	// Название проекта
	// Example: Защита объекта Альфа
	ProjectName string `json:"projectName"`
	// Версия проекта для optimistic locking
	// Example: 1
	Version int `json:"version"`
	// Дата обновления
	// Example: 2026-06-12T14:00:00.000Z
	UpdatedAt string `json:"updatedAt"`
}

// ProjectListResponse DTO ответа со списком проектов.
// swagger:model ProjectListResponse
type ProjectListResponse struct {
	Items      []ProjectResponse `json:"items"`
	TotalItems int64             `json:"totalItems"`
}

// ListProjectsQuery DTO query-параметров для списка проектов.
// swagger:parameters listProjects
type ListProjectsQuery struct {
	// Лимит записей
	// In: query
	// Required: false
	Limit int `json:"limit"`
	// Смещение
	// In: query
	// Required: false
	Offset int `json:"offset"`
	// Фильтр по enterprise ID
	// In: query
	// Required: false
	EnterpriseID string `json:"enterpriseId"`
}

// GetProjectQuery DTO query-параметров для получения проекта.
// swagger:parameters getProject
type GetProjectQuery struct {
	// Версия сохранённого снимка; без параметра возвращается текущая.
	// In: query
	// Minimum: 1
	ProjectVersion *int `json:"projectVersion,omitempty"`
	// ID проекта
	// Required: true
	// In: query
	ID string `json:"id"`
}

// UpdateProjectRequest DTO запроса на обновление проекта.
type UpdateProjectRequest struct {
	// Новое имя конфигурации
	// In: body
	Name string `json:"name"`
	// Новый enterprise ID
	// In: body
	EnterpriseID *string `json:"enterpriseId"`
	// JSON проекта защиты для перезаписи содержимого карты (опционально)
	// In: body
	ProjectJSON string `json:"projectJson,omitempty"`
	// Обязательная положительная ожидаемая версия проекта.
	// Required: true
	// Minimum: 1
	// In: body
	Version *int `json:"version,omitempty"`
}

// UpdateProjectParams описывает query ID и тело обновления с ожидаемой версией.
// swagger:parameters updateProject
type UpdateProjectParams struct {
	// ID проекта
	// Required: true
	// In: query
	ID string `json:"id"`
	// In: body
	// Required: true
	Body UpdateProjectRequest
}

// DeleteProjectQuery DTO query-параметров для удаления проекта.
// swagger:parameters deleteProject
type DeleteProjectQuery struct {
	// ID проекта
	// Required: true
	// In: query
	ID string `json:"id"`
}

// swagger:response ProjectResponse
type ProjectResponseBody struct {
	// In: body
	Body ProjectResponse
}

// swagger:response ProjectListResponse
type ProjectListResponseBody struct {
	// In: body
	Body ProjectListResponse
}
