package domain

import "errors"

var (
	// ErrDocumentNotFound возвращается когда документ не найден.
	ErrDocumentNotFound = errors.New("document not found")

	// ErrDocumentInvalidName возвращается при пустом названии документа.
	ErrDocumentInvalidName = errors.New("invalid document name: must not be empty")

	// ErrDocumentInvalidAssetID возвращается при пустом asset_id.
	ErrDocumentInvalidAssetID = errors.New("invalid document asset id: must not be empty")

	// ErrDocumentInvalidSizeBytes возвращается при отрицательном размере файла.
	ErrDocumentInvalidSizeBytes = errors.New("invalid document size bytes: must not be negative")

	// ErrDocumentInvalidStorageKey возвращается при пустом storage_key.
	ErrDocumentInvalidStorageKey = errors.New("invalid document storage key: must not be empty")
)
