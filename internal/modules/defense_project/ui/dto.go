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
