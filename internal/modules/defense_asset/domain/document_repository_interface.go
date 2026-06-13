package domain

import "context"

// DocumentRepositoryInterface — интерфейс репозитория для Document.
type DocumentRepositoryInterface interface {
	// Save создаёт новый документ.
	Save(ctx context.Context, document *Document) error

	// FindByID возвращает документ по ID.
	FindByID(ctx context.Context, id string) (*Document, error)

	// FindByAssetID возвращает список документов по asset_id, отсортированный по created_at.
	FindByAssetID(ctx context.Context, assetID string) ([]*Document, error)

	// Delete удаляет документ по ID.
	Delete(ctx context.Context, id string) error
}
