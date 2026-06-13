package infrastructure

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/fortis/backend/internal/modules/defense_asset/domain"
)

// compoundProfileDTO — промежуточная структура для сериализации/десериализации
// DefenseAssetCompoundProfile в JSONB с обработкой обратной совместимости.
type compoundProfileDTO struct {
	Kind           string  `json:"kind,omitempty"`
	PostType       string  `json:"postType,omitempty"`
	PersonnelCount string  `json:"personnelCount,omitempty"`
	Accountability string  `json:"accountability,omitempty"`
	Armament       string  `json:"armament,omitempty"`
	WeaponUnits    string  `json:"weaponUnits,omitempty"`
	SectorOrRange  string  `json:"sectorOrRange,omitempty"`
	Azimuth        float64 `json:"azimuth,omitempty"`
}

// UnmarshalJSON для compoundProfileDTO с поддержкой старого формата (int → string).
func (c *compoundProfileDTO) UnmarshalJSON(data []byte) error {
	type alias compoundProfileDTO
	aux := &alias{}
	if err := json.Unmarshal(data, aux); err == nil {
		*c = compoundProfileDTO(*aux)
		return nil
	}

	// Пробуем с преобразованием числа в строку для personnelCount
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if v, ok := raw["personnelCount"]; ok {
		if val, ok := v.(float64); ok {
			raw["personnelCount"] = fmt.Sprintf("%.0f", val)
		}
	}
	fixed, err := json.Marshal(raw)
	if err != nil {
		return err
	}
	return json.Unmarshal(fixed, (*alias)(c))
}

// DefenseAssetModel — GORM-модель для хранения DefenseAsset.
type DefenseAssetModel struct {
	ID           uuid.UUID  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	EnterpriseID *uuid.UUID `gorm:"type:uuid"`
	AssetData    string     `gorm:"type:jsonb;not null"`
	IsPublic     bool       `gorm:"not null;default:false"`
	CreatedAt    time.Time  `gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime"`
}

func (DefenseAssetModel) TableName() string {
	return "defense_assets"
}

// DefenseAssetDocumentModel — GORM-модель для хранения документов средства защиты.
type DefenseAssetDocumentModel struct {
	ID          uuid.UUID  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	AssetID     uuid.UUID  `gorm:"type:uuid;not null;index"`
	Name        string     `gorm:"not null"`
	MimeType    string     `gorm:"not null;default:'application/octet-stream'"`
	SizeBytes   int64      `gorm:"not null;default:0"`
	StorageKey  string     `gorm:"not null"`
	DownloadURL string     `gorm:""`
	OwnerID     *uuid.UUID `gorm:"type:uuid"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
}

func (DefenseAssetDocumentModel) TableName() string {
	return "defense_asset_documents"
}

// toDomain преобразует GORM-модель документа в доменный Document.
func (m *DefenseAssetDocumentModel) toDomain() (*domain.Document, error) {
	var ownerID *string
	if m.OwnerID != nil {
		s := m.OwnerID.String()
		ownerID = &s
	}

	doc, err := domain.NewDocument(
		m.ID.String(),
		m.AssetID.String(),
		m.Name,
		m.MimeType,
		m.StorageKey,
		m.DownloadURL,
		m.SizeBytes,
		ownerID,
		m.CreatedAt,
		m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

// AssetDataDTO — промежуточная структура для сериализации/десериализации JSONB.
// Повторяет поля DefenseAsset для JSONB-хранения.
type AssetDataDTO struct {
	ID                    string                       `json:"id"`
	Name                  string                       `json:"name"`
	ShortName             string                       `json:"shortName,omitempty"`
	Description           string                       `json:"description,omitempty"`
	Category              string                       `json:"category"`
	Roles                 []string                     `json:"roles,omitempty"`
	PricePerUnitMln       *float64                     `json:"pricePerUnitMln,omitempty"`
	Currency              string                       `json:"currency,omitempty"`
	UnitLabel             string                       `json:"unitLabel,omitempty"`
	CompatibleLayerTypes  []string                     `json:"compatibleLayerTypes,omitempty"`
	RecommendedLayerCodes []string                     `json:"recommendedLayerCodes,omitempty"`
	CompatibleLayerCodes  []string                     `json:"compatibleLayerCodes,omitempty"`
	IncompatibleLayerCodes []string                    `json:"incompatibleLayerCodes,omitempty"`
	ProtectionType        string                       `json:"protectionType,omitempty"`
	MinEffectiveDistance  *float64                     `json:"minEffectiveDistance,omitempty"`
	MaxEffectiveDistance  *float64                     `json:"maxEffectiveDistance,omitempty"`
	CoverageType          string                       `json:"coverageType"`
	CoverageRadius        *float64                     `json:"coverageRadius,omitempty"`
	CoverageAngle         *float64                     `json:"coverageAngle,omitempty"`
	DeploymentType        string                       `json:"deploymentType,omitempty"`
	PlacementType         string                       `json:"placementType,omitempty"`
	IconURL               string                       `json:"iconUrl,omitempty"`
	ModelURL              string                       `json:"modelUrl,omitempty"`
	Score                 *int                         `json:"score,omitempty"`
	Priority              *string                      `json:"priority,omitempty"`
	CompoundProfile       *compoundProfileDTO              `json:"compoundProfile,omitempty"`
	WeaponSpec            *domain.WeaponSpecification      `json:"weaponSpec,omitempty"`
	DetectionSpec         *domain.DetectionSpecification    `json:"detectionSpec,omitempty"`
	EWSpec                *domain.EWSpecification          `json:"ewSpec,omitempty"`
	Tags                  []string                         `json:"tags,omitempty"`
	LegacyItemID          string                       `json:"legacyItemId,omitempty"`
	CalculatorAssetID     *string                      `json:"calculatorAssetId,omitempty"`
	MapCatalogGroupIDs    []string                     `json:"mapCatalogGroupIds,omitempty"`
}

// ToDomain преобразует GORM-модель в доменный агрегат.
func (m *DefenseAssetModel) ToDomain() (*domain.DefenseAsset, error) {
	var dto AssetDataDTO
	if err := json.Unmarshal([]byte(m.AssetData), &dto); err != nil {
		return nil, fmt.Errorf("unmarshal asset_data: %w", err)
	}

	// Преобразование строк обратно в VO
	category := domain.DefenseAssetCategory(dto.Category)
	coverageType := domain.DefenseAssetCoverageType(dto.CoverageType)

	roles := make([]domain.DefenseAssetRole, len(dto.Roles))
	for i, r := range dto.Roles {
		roles[i] = domain.DefenseAssetRole(r)
	}

	layerTypes := make([]domain.LayerType, len(dto.CompatibleLayerTypes))
	for i, lt := range dto.CompatibleLayerTypes {
		layerTypes[i] = domain.LayerType(lt)
	}

	var priority *domain.DefensePriority
	if dto.Priority != nil {
		p := domain.DefensePriority(*dto.Priority)
		priority = &p
	}

	enterpriseID := m.EnterpriseID
	var eid *string
	if enterpriseID != nil {
		s := enterpriseID.String()
		eid = &s
	}

	// Конвертация compoundProfileDTO → domain.DefenseAssetCompoundProfile
	var compoundProfile *domain.DefenseAssetCompoundProfile
	if dto.CompoundProfile != nil {
		compoundProfile = &domain.DefenseAssetCompoundProfile{
			Kind:           dto.CompoundProfile.Kind,
			PostType:       dto.CompoundProfile.PostType,
			PersonnelCount: dto.CompoundProfile.PersonnelCount,
			Accountability: dto.CompoundProfile.Accountability,
			Armament:       dto.CompoundProfile.Armament,
			WeaponUnits:    dto.CompoundProfile.WeaponUnits,
			SectorOrRange:  dto.CompoundProfile.SectorOrRange,
			Azimuth:        dto.CompoundProfile.Azimuth,
		}
	}

	asset, err := domain.NewDefenseAsset(
		dto.ID,
		dto.Name,
		dto.ShortName,
		dto.Description,
		category,
		roles,
		dto.PricePerUnitMln,
		dto.Currency,
		dto.UnitLabel,
		layerTypes,
		dto.RecommendedLayerCodes,
		dto.CompatibleLayerCodes,
		dto.IncompatibleLayerCodes,
		dto.ProtectionType,
		dto.MinEffectiveDistance,
		dto.MaxEffectiveDistance,
		coverageType,
		dto.CoverageRadius,
		dto.CoverageAngle,
		domain.DeploymentType(dto.DeploymentType),
		domain.PlacementType(dto.PlacementType),
		dto.IconURL,
		dto.ModelURL,
		dto.Score,
		priority,
		compoundProfile,
		dto.WeaponSpec,
		dto.DetectionSpec,
		dto.EWSpec,
		dto.Tags,
		dto.LegacyItemID,
		dto.CalculatorAssetID,
		dto.MapCatalogGroupIDs,
		eid,
		m.IsPublic,
		m.CreatedAt,
		m.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid asset data in database: %w", err)
	}

	return asset, nil
}

// ToModel преобразует доменный агрегат в GORM-модель.
func ToModel(asset *domain.DefenseAsset) (*DefenseAssetModel, error) {
	var eid *uuid.UUID
	if asset.EnterpriseID() != nil {
		uid, err := uuid.Parse(*asset.EnterpriseID())
		if err != nil {
			return nil, fmt.Errorf("parse enterprise_id: %w", err)
		}
		eid = &uid
	}

	roles := make([]string, len(asset.Roles()))
	for i, r := range asset.Roles() {
		roles[i] = string(r)
	}

	layerTypes := make([]string, len(asset.CompatibleLayerTypes()))
	for i, lt := range asset.CompatibleLayerTypes() {
		layerTypes[i] = string(lt)
	}

	var priority *string
	if asset.Priority() != nil {
		p := string(*asset.Priority())
		priority = &p
	}

	// Конвертация domain.DefenseAssetCompoundProfile → compoundProfileDTO
	var compoundProfile *compoundProfileDTO
	if cp := asset.CompoundProfile(); cp != nil {
		compoundProfile = &compoundProfileDTO{
			Kind:           cp.Kind,
			PostType:       cp.PostType,
			PersonnelCount: cp.PersonnelCount,
			Accountability: cp.Accountability,
			Armament:       cp.Armament,
			WeaponUnits:    cp.WeaponUnits,
			SectorOrRange:  cp.SectorOrRange,
			Azimuth:        cp.Azimuth,
		}
	}

	cat := string(asset.Category())
	dto := AssetDataDTO{
		ID:                    asset.ID(),
		Name:                  asset.Name(),
		ShortName:             asset.ShortName(),
		Description:           asset.Description(),
		Category:              cat,
		Roles:                 roles,
		PricePerUnitMln:       asset.PricePerUnitMln(),
		Currency:              asset.Currency(),
		UnitLabel:             asset.UnitLabel(),
		CompatibleLayerTypes:  layerTypes,
		RecommendedLayerCodes: asset.RecommendedLayerCodes(),
		CompatibleLayerCodes:  asset.CompatibleLayerCodes(),
		IncompatibleLayerCodes: asset.IncompatibleLayerCodes(),
		ProtectionType:        asset.ProtectionType(),
		MinEffectiveDistance:  asset.MinEffectiveDistance(),
		MaxEffectiveDistance:  asset.MaxEffectiveDistance(),
		CoverageType:          string(asset.CoverageType()),
		CoverageRadius:        asset.CoverageRadius(),
		CoverageAngle:         asset.CoverageAngle(),
		DeploymentType:        string(asset.DeploymentType()),
		PlacementType:         string(asset.PlacementType()),
		IconURL:               asset.IconURL(),
		ModelURL:              asset.ModelURL(),
		Score:                 asset.Score(),
		Priority:              priority,
		CompoundProfile:       compoundProfile,
		WeaponSpec:            asset.WeaponSpec(),
		DetectionSpec:         asset.DetectionSpec(),
		EWSpec:                asset.EWSpec(),
		Tags:                  asset.Tags(),
		LegacyItemID:          asset.LegacyItemID(),
		CalculatorAssetID:     asset.CalculatorAssetID(),
		MapCatalogGroupIDs:    asset.MapCatalogGroupIDs(),
	}

	assetData, err := json.Marshal(dto)
	if err != nil {
		return nil, fmt.Errorf("marshal asset_data: %w", err)
	}

	id, err := uuid.Parse(asset.ID())
	if err != nil {
		return nil, fmt.Errorf("parse asset id: %w", err)
	}

	return &DefenseAssetModel{
		ID:           id,
		EnterpriseID: eid,
		AssetData:    string(assetData),
		IsPublic:     asset.IsPublic(),
		CreatedAt:    asset.CreatedAt(),
		UpdatedAt:    asset.UpdatedAt(),
	}, nil
}
