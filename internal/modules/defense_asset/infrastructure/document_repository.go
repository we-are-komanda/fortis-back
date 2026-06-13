package infrastructure

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/fortis/backend/internal/modules/defense_asset/domain"
	"github.com/fortis/backend/internal/rdbms"
)

// DocumentRepository — реализация репозитория Document.
type DocumentRepository struct {
	executor rdbms.Executor
}

// NewDocumentRepository создаёт новый репозиторий документов.
func NewDocumentRepository(executor rdbms.Executor) domain.DocumentRepositoryInterface {
	return &DocumentRepository{
		executor: executor,
	}
}

// Save создаёт новый документ.
func (r *DocumentRepository) Save(ctx context.Context, document *domain.Document) error {
	model := toDocumentModel(document)
	return r.executor.WithContext(ctx).Create(model).Error
}

// FindByID ищет документ по ID.
func (r *DocumentRepository) FindByID(ctx context.Context, id string) (*domain.Document, error) {
	var model DefenseAssetDocumentModel
	result := r.executor.WithContext(ctx).Where("id = ?", id).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrDocumentNotFound
		}
		return nil, result.Error
	}

	return model.toDomain()
}

// FindByAssetID возвращает список документов по asset_id, отсортированный по created_at.
func (r *DocumentRepository) FindByAssetID(ctx context.Context, assetID string) ([]*domain.Document, error) {
	var models []DefenseAssetDocumentModel
	result := r.executor.WithContext(ctx).
		Where("asset_id = ?", assetID).
		Order("created_at ASC").
		Find(&models)
	if result.Error != nil {
		return nil, result.Error
	}

	documents := make([]*domain.Document, len(models))
	for i, m := range models {
		doc, err := m.toDomain()
		if err != nil {
			return nil, err
		}
		documents[i] = doc
	}

	return documents, nil
}

// Delete удаляет документ по ID.
func (r *DocumentRepository) Delete(ctx context.Context, id string) error {
	result := r.executor.WithContext(ctx).Where("id = ?", id).Delete(&DefenseAssetDocumentModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrDocumentNotFound
	}
	return nil
}

// toDocumentModel преобразует доменный Document в GORM-модель.
func toDocumentModel(doc *domain.Document) *DefenseAssetDocumentModel {
	id, _ := uuid.Parse(doc.ID())
	assetID, _ := uuid.Parse(doc.AssetID())

	var ownerID *uuid.UUID
	if doc.OwnerID() != nil {
		uid, err := uuid.Parse(*doc.OwnerID())
		if err == nil {
			ownerID = &uid
		}
	}

	return &DefenseAssetDocumentModel{
		ID:          id,
		AssetID:     assetID,
		Name:        doc.Name(),
		MimeType:    doc.MimeType(),
		SizeBytes:   doc.SizeBytes(),
		StorageKey:  doc.StorageKey(),
		DownloadURL: doc.DownloadURL(),
		OwnerID:     ownerID,
		CreatedAt:   doc.CreatedAt(),
		UpdatedAt:   doc.UpdatedAt(),
	}
}
