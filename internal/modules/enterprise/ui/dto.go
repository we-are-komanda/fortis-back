package ui

// CreateEnterpriseRequest DTO запроса на создание предприятия.
// swagger:parameters CreateEnterpriseRequest
type CreateEnterpriseRequest struct {
	// Название предприятия
	// Required: true
	// Example: Нефтебаза "Альфа"
	Name string `json:"name"`
	// Адрес предприятия
	// Example: г. Москва, ул. Ленина, д. 1
	Address string `json:"address"`
	// Статус предприятия (active, configuring, offline)
	// Enum: ["active","configuring","offline"]
	// Example: active
	Status string `json:"status"`
	// Широта центра предприятия
	// Required: true
	// Example: 55.75
	Latitude float64 `json:"latitude"`
	// Долгота центра предприятия
	// Required: true
	// Example: 37.62
	Longitude float64 `json:"longitude"`
}

// UpdateEnterpriseRequest DTO запроса на обновление предприятия.
// swagger:parameters UpdateEnterpriseRequest
type UpdateEnterpriseRequest struct {
	// Название предприятия
	// Example: Нефтебаза "Бета"
	Name string `json:"name"`
	// Адрес предприятия
	// Example: г. Москва, ул. Тверская, д. 10
	Address string `json:"address"`
	// Статус предприятия (active, configuring, offline)
	// Enum: ["active","configuring","offline"]
	// Example: active
	Status string `json:"status"`
	// Широта центра предприятия
	// Example: 55.76
	Latitude float64 `json:"latitude"`
	// Долгота центра предприятия
	// Example: 37.63
	Longitude float64 `json:"longitude"`
}

// EnterpriseResponse DTO ответа с данными предприятия.
// swagger:response EnterpriseResponse
type EnterpriseResponse struct {
	// ID предприятия
	// Example: 550e8400-e29b-41d4-a716-446655440000
	ID string `json:"id"`
	// Название предприятия
	// Example: Нефтебаза "Альфа"
	Name string `json:"name"`
	// Адрес предприятия
	// Example: г. Москва, ул. Ленина, д. 1
	Address string `json:"address"`
	// Статус предприятия
	// Example: active
	Status string `json:"status"`
	// Широта центра
	// Example: 55.75
	Latitude float64 `json:"latitude"`
	// Долгота центра
	// Example: 37.62
	Longitude float64 `json:"longitude"`
	// Дата создания (RFC3339)
	// Example: 2026-06-12T14:00:00Z
	CreatedAt string `json:"createdAt"`
	// Дата обновления (RFC3339)
	// Example: 2026-06-12T14:00:00Z
	UpdatedAt string `json:"updatedAt"`
}

// EnterpriseListResponse DTO ответа со списком предприятий.
// swagger:response EnterpriseListResponse
type EnterpriseListResponse struct {
	Items      []EnterpriseResponse `json:"items"`
	TotalItems int64                `json:"totalItems"`
}

// EnterpriseQuery DTO query-параметров для получения по ID.
// swagger:parameters EnterpriseQuery
type EnterpriseQuery struct {
	// ID предприятия
	// Required: true
	// In: query
	ID string `json:"id"`
}

// ListQuery DTO query-параметров пагинации.
// swagger:parameters ListQuery
type ListQuery struct {
	// Количество записей на странице (по умолчанию 20, макс 100)
	// In: query
	// Example: 20
	Limit int `json:"limit"`
	// Смещение от начала списка
	// In: query
	// Example: 0
	Offset int `json:"offset"`
}
