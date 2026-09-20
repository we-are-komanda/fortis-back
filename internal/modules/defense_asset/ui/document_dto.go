package ui

// AssetDocumentDTO DTO ответа с метаданными документа.
// swagger:model AssetDocumentDTO
type AssetDocumentDTO struct {
	Revision   string  `json:"revision"`
	Checksum   *string `json:"checksum"`
	Status     string  `json:"status"`
	Commercial bool    `json:"commercial"`
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
	// Относительный URL авторизованного скачивания; отсутствует для недоступного файла.
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
// swagger:model AssetDocumentListResponse
type AssetDocumentListResponse struct {
	// Список документов
	Items []AssetDocumentDTO `json:"items"`
	// Общее количество документов
	TotalItems int `json:"totalItems"`
}

// DocumentListQuery DTO query-параметров для списка документов.
// swagger:parameters listAssetDocuments
type DocumentListQuery struct {
	// ID средства защиты
	// Required: true
	// In: query
	AssetID string `json:"assetId"`
	// In: query
	// Maximum: 100
	Limit int `json:"limit"`
	// In: query
	// Minimum: 0
	Offset int `json:"offset"`
}

// DocumentGetQuery DTO query-параметров для получения/удаления документа.
// swagger:parameters getAssetDocument deleteAssetDocument
type DocumentGetQuery struct {
	// ID документа
	// Required: true
	// In: query
	ID string `json:"id"`
}

// UploadDocumentRequest — multipart-контракт: один file, assetId и optional commercial.
// ponytail: v0.36.6 drops required on swagger:file fields; keep the operation here until its field parser supports both.
// swagger:operation POST /api/v1/assets/documents api createAssetDocument
//
// # Загрузка приватного документа
//
// Загружает один файл до 10 MiB в карантин и публикует после сканирования.
// ---
// consumes: [multipart/form-data]
// produces: [application/json]
// parameters:
//   - name: assetId
//     in: formData
//     type: string
//     required: true
//   - name: file
//     in: formData
//     type: file
//     required: true
//   - name: commercial
//     in: formData
//     type: boolean
//     default: true
//
// responses:
//
//	'201':
//	  $ref: '#/responses/AssetDocumentResponse'
//	'400':
//	  description: Неверные поля загрузки
//	'403':
//	  description: Изменение общего каталога запрещено
//	'404':
//	  description: Карточка не найдена
//	'413':
//	  description: Файл превышает 10 MiB
//	'415':
//	  description: Формат или содержимое файла не поддерживается
//	'422':
//	  description: Файл отклонён проверкой
//	'503':
//	  description: Хранилище или сканер недоступны
type UploadDocumentRequest struct {
	AssetID    string `json:"assetId"`
	File       string `json:"file"`
	Commercial *bool  `json:"commercial,omitempty"`
}

// DownloadDocumentQuery описывает авторизованное скачивание по постоянному ID.
// swagger:parameters downloadAssetDocument
type DownloadDocumentQuery struct {
	// In: query
	// Required: true
	ID string `json:"id"`
	// Необязательная дополнительная проверка родительской карточки.
	// In: query
	AssetID string `json:"assetId,omitempty"`
}

// AssetDocumentResponse описывает тело метаданных документа.
// swagger:response AssetDocumentResponse
type AssetDocumentResponse struct {
	// In: body
	Body AssetDocumentDTO
}

// AssetDocumentListResponseBody описывает тело списка документов.
// swagger:response AssetDocumentListResponse
type AssetDocumentListResponseBody struct {
	// In: body
	Body AssetDocumentListResponse
}

// AssetDocumentDownloadResponse описывает байты проверенного файла.
// swagger:response AssetDocumentDownloadResponse
type AssetDocumentDownloadResponse struct {
	// In: body
	// swagger:file
	Body []byte
}
