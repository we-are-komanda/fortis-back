package ui

// ImportRequest DTO запроса на импорт проекта.
// swagger:parameters ImportRequest
type ImportRequest struct {
	// JSON проекта защиты
	// Required: true
	// In: body
	ProjectJSON string `json:"projectJson"`
}

// ImportResponse DTO ответа после успешного импорта.
// swagger:response ImportResponse
type ImportResponse struct {
	// In: body
	Body struct {
		// ID сохранённого проекта
		// Example: 550e8400-e29b-41d4-a716-446655440000
		ProjectID string `json:"projectId"`
		// Название проекта
		// Example: Защита объекта Альфа
		ProjectName string `json:"projectName"`
		// Дата обновления проекта
		// Example: 2026-06-12T14:00:00.000Z
		UpdatedAt string `json:"updatedAt"`
	} `json:"body"`
}

// ExportQuery DTO query-параметров для экспорта проекта.
// swagger:parameters ExportQuery
type ExportQuery struct {
	// ID проекта для экспорта
	// Required: true
	// In: query
	ID string `json:"id"`
}

// ---- CRUD DTOs ----

// CreateProjectRequest DTO запроса на создание конфигурации.
// swagger:parameters CreateProjectRequest
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

// ProjectResponse DTO ответа с данными проекта.
// swagger:response ProjectResponse
type ProjectResponse struct {
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
	// Дата обновления
	// Example: 2026-06-12T14:00:00.000Z
	UpdatedAt string `json:"updatedAt"`
}

// ProjectListResponse DTO ответа со списком проектов.
// swagger:response ProjectListResponse
type ProjectListResponse struct {
	Items      []ProjectResponse `json:"items"`
	TotalItems int64             `json:"totalItems"`
}

// ListProjectsQuery DTO query-параметров для списка проектов.
// swagger:parameters ListProjectsQuery
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
// swagger:parameters GetProjectQuery
type GetProjectQuery struct {
	// ID проекта
	// Required: true
	// In: query
	ID string `json:"id"`
}

// UpdateProjectRequest DTO запроса на обновление проекта.
// swagger:parameters UpdateProjectRequest
type UpdateProjectRequest struct {
	// ID проекта
	// Required: true
	// In: query
	ID string `json:"id"`
	// Новое имя конфигурации
	// In: body
	Name string `json:"name"`
	// Новый enterprise ID
	// In: body
	EnterpriseID string `json:"enterpriseId"`
}

// DeleteProjectQuery DTO query-параметров для удаления проекта.
// swagger:parameters DeleteProjectQuery
type DeleteProjectQuery struct {
	// ID проекта
	// Required: true
	// In: query
	ID string `json:"id"`
}
