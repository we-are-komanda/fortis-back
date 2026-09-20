package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/fortis/backend/internal/auth"
	enterpriseApp "github.com/fortis/backend/internal/modules/enterprise/application"
	"time"

	"github.com/google/uuid"

	"github.com/fortis/backend/internal/modules/defense_asset/domain"
)

// DefenseAssetService — сервис для управления средствами защиты.
type DefenseAssetService struct {
	repo      domain.DefenseAssetRepositoryInterface
	documents domain.DocumentRepositoryInterface
	access    enterpriseApp.AccessChecker
}

// NewDefenseAssetService создаёт новый сервис средств защиты.
func NewDefenseAssetService(repo domain.DefenseAssetRepositoryInterface, access enterpriseApp.AccessChecker, documents ...domain.DocumentRepositoryInterface) *DefenseAssetService {
	var docs domain.DocumentRepositoryInterface
	if len(documents) > 0 {
		docs = documents[0]
	}
	return &DefenseAssetService{
		repo: repo, access: access, documents: docs,
	}
}

// CreateInput — входные данные для создания средства защиты.
type CreateInput struct {
	CatalogMetadataInput
	Name                   string
	ShortName              string
	Description            string
	Category               domain.DefenseAssetCategory
	Roles                  []domain.DefenseAssetRole
	PricePerUnitMln        *float64
	Currency               string
	UnitLabel              string
	CompatibleLayerTypes   []domain.LayerType
	RecommendedLayerCodes  []string
	CompatibleLayerCodes   []string
	IncompatibleLayerCodes []string
	ProtectionType         string
	MinEffectiveDistance   *float64
	MaxEffectiveDistance   *float64
	CoverageType           domain.DefenseAssetCoverageType
	CoverageRadius         *float64
	CoverageAngle          *float64
	DeploymentType         domain.DeploymentType
	PlacementType          domain.PlacementType
	IconURL                string
	ModelURL               string
	Score                  *int
	Priority               *domain.DefensePriority
	CompoundProfile        *domain.DefenseAssetCompoundProfile
	WeaponSpec             *domain.WeaponSpecification
	DetectionSpec          *domain.DetectionSpecification
	EWSpec                 *domain.EWSpecification
	Tags                   []string
	LegacyItemID           string
	CalculatorAssetID      *string
	MapCatalogGroupIDs     []string
	EnterpriseID           *string
	IsPublic               bool
}

// Create создаёт новое средство защиты.
func (s *DefenseAssetService) Create(ctx context.Context, userID string, input CreateInput) (*domain.DefenseAsset, error) {
	if err := auth.RequireIdentity(userID); err != nil {
		return nil, err
	}
	if input.IsPublic {
		return nil, auth.ErrForbidden
	}
	enterpriseID := ""
	if input.EnterpriseID != nil {
		enterpriseID = *input.EnterpriseID
	}
	if err := s.access.CheckUserAccess(ctx, userID, enterpriseID); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	id := uuid.New().String()

	asset, err := domain.NewDefenseAsset(
		id,
		input.Name,
		input.ShortName,
		input.Description,
		input.Category,
		input.Roles,
		input.PricePerUnitMln,
		input.Currency,
		input.UnitLabel,
		input.CompatibleLayerTypes,
		input.RecommendedLayerCodes,
		input.CompatibleLayerCodes,
		input.IncompatibleLayerCodes,
		input.ProtectionType,
		input.MinEffectiveDistance,
		input.MaxEffectiveDistance,
		input.CoverageType,
		input.CoverageRadius,
		input.CoverageAngle,
		input.DeploymentType,
		input.PlacementType,
		input.IconURL,
		input.ModelURL,
		input.Score,
		input.Priority,
		input.CompoundProfile,
		input.WeaponSpec,
		input.DetectionSpec,
		input.EWSpec,
		input.Tags,
		input.LegacyItemID,
		input.CalculatorAssetID,
		input.MapCatalogGroupIDs,
		input.EnterpriseID,
		input.IsPublic,
		now,
		now,
	)
	if err != nil {
		return nil, err
	}

	if err := s.applyCatalogMetadata(ctx, userID, asset, input.CatalogMetadataInput); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, asset); err != nil {
		return nil, fmt.Errorf("save defense asset: %w", err)
	}

	return asset, nil
}

// GetByID возвращает средство защиты по ID.
func (s *DefenseAssetService) GetByID(ctx context.Context, userID, id string) (*domain.DefenseAsset, error) {
	if err := auth.RequireIdentity(userID); err != nil {
		return nil, err
	}
	asset, err := s.repo.FindByID(ctx, id)
	if errors.Is(err, domain.ErrDefenseAssetNotFound) {
		return nil, auth.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if !asset.IsPublic() {
		enterpriseID := ""
		if asset.EnterpriseID() != nil {
			enterpriseID = *asset.EnterpriseID()
		}
		if err := s.access.CheckUserAccess(ctx, userID, enterpriseID); err != nil {
			return nil, err
		}
	}
	return asset, nil
}

// GetForMutation applies the same parent policy to assets and their documents.
func (s *DefenseAssetService) GetForMutation(ctx context.Context, userID, id string) (*domain.DefenseAsset, error) {
	asset, err := s.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if asset.IsPublic() {
		return nil, auth.ErrForbidden
	}
	return asset, nil
}

// List возвращает список средств защиты с пагинацией и фильтрацией.
func (s *DefenseAssetService) List(
	ctx context.Context, userID string,
	enterpriseID *string,
	isPublic *bool,
	category *domain.DefenseAssetCategory,
	limit, offset int,
) ([]*domain.DefenseAsset, int64, error) {
	if err := auth.RequireIdentity(userID); err != nil {
		return nil, 0, err
	}
	if enterpriseID != nil {
		if err := s.access.CheckUserAccess(ctx, userID, *enterpriseID); err != nil {
			return nil, 0, err
		}
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	filter := domain.DefenseAssetFilter{
		UserID:       userID,
		EnterpriseID: enterpriseID,
		IsPublic:     isPublic,
		Category:     category,
		Limit:        limit,
		Offset:       offset,
	}

	return s.repo.FindAll(ctx, filter)
}

// UpdateInput — входные данные для обновления средства защиты.
type UpdateInput struct {
	CatalogMetadataInput
	ID                     string
	Name                   *string
	ShortName              *string
	Description            *string
	Category               *domain.DefenseAssetCategory
	Roles                  []domain.DefenseAssetRole
	PricePerUnitMln        *float64
	Currency               *string
	UnitLabel              *string
	CompatibleLayerTypes   []domain.LayerType
	RecommendedLayerCodes  []string
	CompatibleLayerCodes   []string
	IncompatibleLayerCodes []string
	ProtectionType         *string
	MinEffectiveDistance   *float64
	MaxEffectiveDistance   *float64
	CoverageType           *domain.DefenseAssetCoverageType
	CoverageRadius         *float64
	CoverageAngle          *float64
	DeploymentType         *domain.DeploymentType
	PlacementType          *domain.PlacementType
	IconURL                *string
	ModelURL               *string
	Score                  *int
	Priority               *domain.DefensePriority
	CompoundProfile        *domain.DefenseAssetCompoundProfile
	WeaponSpec             *domain.WeaponSpecification
	DetectionSpec          *domain.DetectionSpecification
	EWSpec                 *domain.EWSpecification
	Tags                   []string
	LegacyItemID           *string
	CalculatorAssetID      *string
	MapCatalogGroupIDs     []string
	IsPublic               *bool
}

// Update обновляет существующее средство защиты.
func (s *DefenseAssetService) Update(ctx context.Context, userID string, input UpdateInput) (*domain.DefenseAsset, error) {
	asset, err := s.GetForMutation(ctx, userID, input.ID)
	if err != nil {
		return nil, err
	}
	// Setters replace fields; keep the loaded value unchanged if validation fails.
	next := *asset
	asset = &next

	if input.IsPublic != nil && *input.IsPublic {
		return nil, auth.ErrForbidden
	}

	if input.Name != nil {
		if err := asset.UpdateName(*input.Name); err != nil {
			return nil, err
		}
	}
	if input.Description != nil {
		asset.UpdateDescription(*input.Description)
	}
	if input.Category != nil {
		if err := asset.UpdateCategory(*input.Category); err != nil {
			return nil, err
		}
	}
	if input.CoverageType != nil {
		if err := asset.UpdateCoverageType(*input.CoverageType); err != nil {
			return nil, err
		}
	}
	if input.Roles != nil {
		asset.SetRoles(input.Roles)
	}
	if input.ShortName != nil {
		asset.SetShortName(*input.ShortName)
	}
	if input.PricePerUnitMln != nil {
		asset.SetPricePerUnitMln(input.PricePerUnitMln)
	}
	if input.Currency != nil {
		asset.SetCurrency(*input.Currency)
	}
	if input.UnitLabel != nil {
		asset.SetUnitLabel(*input.UnitLabel)
	}
	if input.CompatibleLayerTypes != nil {
		asset.SetCompatibleLayerTypes(input.CompatibleLayerTypes)
	}
	if input.RecommendedLayerCodes != nil {
		asset.SetRecommendedLayerCodes(input.RecommendedLayerCodes)
	}
	if input.CompatibleLayerCodes != nil {
		asset.SetCompatibleLayerCodes(input.CompatibleLayerCodes)
	}
	if input.IncompatibleLayerCodes != nil {
		asset.SetIncompatibleLayerCodes(input.IncompatibleLayerCodes)
	}
	if input.ProtectionType != nil {
		asset.SetProtectionType(*input.ProtectionType)
	}
	if input.MinEffectiveDistance != nil {
		asset.SetMinEffectiveDistance(input.MinEffectiveDistance)
	}
	if input.MaxEffectiveDistance != nil {
		asset.SetMaxEffectiveDistance(input.MaxEffectiveDistance)
	}
	if input.CoverageRadius != nil {
		asset.SetCoverageRadius(input.CoverageRadius)
	}
	if input.CoverageAngle != nil {
		asset.SetCoverageAngle(input.CoverageAngle)
	}
	if input.DeploymentType != nil {
		asset.SetDeploymentType(*input.DeploymentType)
	}
	if input.PlacementType != nil {
		asset.SetPlacementType(*input.PlacementType)
	}
	if input.IconURL != nil {
		asset.SetIconURL(*input.IconURL)
	}
	if input.ModelURL != nil {
		asset.SetModelURL(*input.ModelURL)
	}
	if input.Score != nil {
		asset.SetScore(input.Score)
	}
	if input.Priority != nil {
		asset.SetPriority(input.Priority)
	}
	if input.CompoundProfile != nil {
		asset.SetCompoundProfile(input.CompoundProfile)
	}
	if input.WeaponSpec != nil {
		asset.SetWeaponSpec(input.WeaponSpec)
	}
	if input.DetectionSpec != nil {
		asset.SetDetectionSpec(input.DetectionSpec)
	}
	if input.EWSpec != nil {
		asset.SetEWSpec(input.EWSpec)
	}
	if input.Tags != nil {
		asset.SetTags(input.Tags)
	}
	if input.LegacyItemID != nil {
		asset.SetLegacyItemID(*input.LegacyItemID)
	}
	if input.CalculatorAssetID != nil {
		asset.SetCalculatorAssetID(input.CalculatorAssetID)
	}
	if input.MapCatalogGroupIDs != nil {
		asset.SetMapCatalogGroupIDs(input.MapCatalogGroupIDs)
	}
	if input.IsPublic != nil {
		asset.SetIsPublic(*input.IsPublic)
	}

	asset.SetUpdatedAt(time.Now().UTC())

	if err := s.applyCatalogMetadata(ctx, userID, asset, input.CatalogMetadataInput); err != nil {
		return nil, err
	}
	if err := asset.Validate(); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, asset); err != nil {
		return nil, fmt.Errorf("update defense asset: %w", err)
	}

	return asset, nil
}

// Delete удаляет средство защиты по ID.
func (s *DefenseAssetService) Delete(ctx context.Context, userID string, id string) error {
	if _, err := s.GetForMutation(ctx, userID, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}
