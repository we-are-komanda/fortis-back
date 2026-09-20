package application

import (
	"context"
	"errors"

	"github.com/fortis/backend/internal/auth"

	"github.com/fortis/backend/internal/modules/defense_asset/domain"
)

// DocumentService — сервис для управления документами средства защиты.
type DocumentService struct {
	repo     domain.DocumentRepositoryInterface
	pipeline DocumentPipeline
	assets   *DefenseAssetService
}

// NewDocumentService создаёт новый сервис документов.
func NewDocumentService(repo domain.DocumentRepositoryInterface, assets *DefenseAssetService, pipeline ...DocumentPipeline) *DocumentService {
	var p DocumentPipeline
	if len(pipeline) > 0 {
		p = pipeline[0]
	}
	return &DocumentService{
		repo: repo, assets: assets, pipeline: p,
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
func (s *DocumentService) Create(ctx context.Context, userID string, input CreateDocumentInput) (*domain.Document, error) {
	if _, err := s.assets.GetForMutation(ctx, userID, input.AssetID); err != nil {
		return nil, err
	}
	return nil, domain.ErrDocumentUploadRequired
}

// GetByID возвращает документ по ID.
func (s *DocumentService) GetByID(ctx context.Context, userID, id string) (*domain.Document, error) {
	if err := auth.RequireIdentity(userID); err != nil {
		return nil, err
	}
	doc, err := s.repo.FindByID(ctx, id)
	if errors.Is(err, domain.ErrDocumentNotFound) {
		return nil, auth.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if _, err := s.assets.GetByID(ctx, userID, doc.AssetID()); err != nil {
		return nil, err
	}
	return doc, nil
}

func (s *DocumentService) ListByAssetID(ctx context.Context, userID, assetID string) ([]*domain.Document, error) {
	if _, err := s.assets.GetByID(ctx, userID, assetID); err != nil {
		return nil, err
	}
	return s.repo.FindByAssetID(ctx, assetID)
}

func (s *DocumentService) Delete(ctx context.Context, userID, id string) error {
	doc, err := s.GetByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if _, err := s.assets.GetForMutation(ctx, userID, doc.AssetID()); err != nil {
		return err
	}
	if repo, ok := s.repo.(domain.DocumentWriteRepository); ok {
		return repo.DeleteDocument(ctx, id, userID)
	}
	return domain.ErrDocumentUnavailable
}
