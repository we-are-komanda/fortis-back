package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/fortis/backend/internal/auth"
	"time"

	"github.com/google/uuid"

	"github.com/fortis/backend/internal/modules/defense_asset/domain"
)

// DocumentService — сервис для управления документами средства защиты.
type DocumentService struct {
	repo   domain.DocumentRepositoryInterface
	assets *DefenseAssetService
}

// NewDocumentService создаёт новый сервис документов.
func NewDocumentService(repo domain.DocumentRepositoryInterface, assets *DefenseAssetService) *DocumentService {
	return &DocumentService{
		repo: repo, assets: assets,
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
	input.OwnerID = &userID
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
	return s.repo.Delete(ctx, id)
}
