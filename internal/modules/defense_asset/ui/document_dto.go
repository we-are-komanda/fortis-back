package ui

// AssetDocumentDTO DTO ответа с метаданными документа.
// swagger:model AssetDocumentDTO
type AssetDocumentDTO struct {
	// ID документа
	// Example: 550e8400-e29b-41d4-a716-446655440000
	ID string `json:"id"`
	// ID средства защиты
	// Example: 550e8400-e29b-41d4-a716-446655440001
	AssetID string `json:"assetId"`
	// Оригинальное имя файла
	// Example: схема_рлс.pdf
	Name string `json:"name"`
	// MIME-тип файла
	// Example: application/pdf
	MimeType string `json:"mimeType"`
	// Размер файла в байтах
	// Example: 1048576
	SizeBytes int64 `json:"sizeBytes"`
	// URL для скачивания
	// Example: https://storage.example.com/documents/file.pdf
	DownloadURL string `json:"downloadUrl,omitempty"`
	// ID пользователя, загрузившего документ
	OwnerID *string `json:"ownerId,omitempty"`
	// Дата создания (RFC3339)
	// Example: 2026-06-12T14:00:00Z
	CreatedAt string `json:"createdAt"`
	// Дата обновления (RFC3339)
	// Example: 2026-06-12T14:00:00Z
	UpdatedAt string `json:"updatedAt"`
}

// CreateDocumentRequest DTO запроса на создание документа.
// swagger:parameters CreateDocumentRequest
type CreateDocumentRequest struct {
	// ID средства защиты
	// Required: true
	// Example: 550e8400-e29b-41d4-a716-446655440001
	AssetID string `json:"assetId"`
	// Оригинальное имя файла
	// Required: true
	// Example: схема_рлс.pdf
	Name string `json:"name"`
	// MIME-тип файла
	// Example: application/pdf
	MimeType string `json:"mimeType"`
	// Размер файла в байтах
	// Example: 1048576
	SizeBytes int64 `json:"sizeBytes"`
	// Ключ в файловом хранилище
	// Required: true
	// Example: assets/abc-123/schema.pdf
	StorageKey string `json:"storageKey"`
	// URL для скачивания (presigned URL или прямая ссылка)
	DownloadURL string `json:"downloadUrl,omitempty"`
	// ID пользователя, загрузившего документ
	OwnerID *string `json:"ownerId,omitempty"`
}

// AssetDocumentListResponse DTO ответа со списком документов.
// swagger:response AssetDocumentListResponse
type AssetDocumentListResponse struct {
	// Список документов
	Items []AssetDocumentDTO `json:"items"`
	// Общее количество документов
	TotalItems int `json:"totalItems"`
}

// DocumentListQuery DTO query-параметров для списка документов.
// swagger:parameters DocumentListQuery
type DocumentListQuery struct {
	// ID средства защиты
	// Required: true
	// In: query
	AssetID string `json:"assetId"`
}

// DocumentGetQuery DTO query-параметров для получения/удаления документа.
// swagger:parameters DocumentGetQuery
type DocumentGetQuery struct {
	// ID документа
	// Required: true
	// In: query
	ID string `json:"id"`
}
