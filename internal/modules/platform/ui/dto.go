package ui

// PlatformDTO DTO платформы
// swagger:model PlatformDTO
type PlatformDTO struct {
	// Статус платформы
	// Example: ok
	Status string `json:"status"`
}

// SuccessGetResponse Успешный ответ
// swagger:response SuccessGetResponse
type SuccessGetResponse struct {
	// In: body
	Body PlatformDTO
}
