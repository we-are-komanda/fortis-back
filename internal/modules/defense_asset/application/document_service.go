package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/fortis/backend/internal/modules/defense_asset/domain"
)

// DocumentService — сервис для управления документами средства защиты.
type DocumentService struct {
	repo domain.DocumentRepositoryInterface
}

// NewDocumentService создаёт новый сервис документов.
func NewDocumentService(repo domain.DocumentRepositoryInterface) *DocumentService {
	return &DocumentService{
		repo: repo,
	}
}

// CreateDocumentInput — входные данные для создания документа.
type CreateDocumentInput struct {
	AssetID     string
	Name        string
	MimeType    string
	SizeBytes   int64
	StorageKey  string
	DownloadURL string
	OwnerID     *string
}

// Create создаёт новый документ.
func (s *DocumentService) Create(ctx context.Context, input CreateDocumentInput) (*domain.Document, error) {
	now := time.Now().UTC()
	id := uuid.New().String()

	doc, err := domain.NewDocument(
		id,
		input.AssetID,
		input.Name,
		input.MimeType,
		input.StorageKey,
		input.DownloadURL,
		input.SizeBytes,
		input.OwnerID,
		now,
		now,
	)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, doc); err != nil {
		return nil, fmt.Errorf("save document: %w", err)
	}

	return doc, nil
}

// GetByID возвращает документ по ID.
func (s *DocumentService) GetByID(ctx context.Context, id string) (*domain.Document, error) {
	return s.repo.FindByID(ctx, id)
}

// ListByAssetID возвращает список документов средства защиты.
func (s *DocumentService) ListByAssetID(ctx context.Context, assetID string) ([]*domain.Document, error) {
	return s.repo.FindByAssetID(ctx, assetID)
}

// Delete удаляет документ по ID.
func (s *DocumentService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
