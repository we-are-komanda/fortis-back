package domain

import (
	"fmt"
	"time"
)

// DefenseAssetCategory — категория средства защиты.
type DefenseAssetCategory string

const (
	DefenseAssetCategoryRadiotechnical DefenseAssetCategory = "radiotechnical"
	DefenseAssetCategoryRadar          DefenseAssetCategory = "radar"
	DefenseAssetCategoryEW             DefenseAssetCategory = "electronic-warfare"
	DefenseAssetCategoryMissile        DefenseAssetCategory = "missile"
	DefenseAssetCategoryAntiMissile    DefenseAssetCategory = "anti-missile"
	DefenseAssetCategoryShturmovaya    DefenseAssetCategory = "shturmovaya"
	DefenseAssetCategoryArtillery      DefenseAssetCategory = "artillery"
	DefenseAssetCategoryAircraft       DefenseAssetCategory = "aircraft"
	DefenseAssetCategoryHelicopter     DefenseAssetCategory = "helicopter"
	DefenseAssetCategoryUAV            DefenseAssetCategory = "uav"
	DefenseAssetCategoryShip           DefenseAssetCategory = "ship"
	DefenseAssetCategoryFortification  DefenseAssetCategory = "fortification"
	DefenseAssetCategoryInfrastructure DefenseAssetCategory = "infrastructure"
)

// ValidDefenseAssetCategories — список допустимых категорий.
var ValidDefenseAssetCategories = []DefenseAssetCategory{
	DefenseAssetCategoryRadiotechnical,
	DefenseAssetCategoryRadar,
	DefenseAssetCategoryEW,
	DefenseAssetCategoryMissile,
	DefenseAssetCategoryAntiMissile,
	DefenseAssetCategoryShturmovaya,
	DefenseAssetCategoryArtillery,
	DefenseAssetCategoryAircraft,
	DefenseAssetCategoryHelicopter,
	DefenseAssetCategoryUAV,
	DefenseAssetCategoryShip,
	DefenseAssetCategoryFortification,
	DefenseAssetCategoryInfrastructure,
}

// DefenseAssetRole — роль средства защиты.
type DefenseAssetRole string

const (
	DefenseAssetRoleDetection    DefenseAssetRole = "detection"
	DefenseAssetRoleDestruction  DefenseAssetRole = "destruction"
	DefenseAssetRoleEW           DefenseAssetRole = "ew"
	DefenseAssetRoleC2           DefenseAssetRole = "c2"
	DefenseAssetRoleCover        DefenseAssetRole = "cover"
	DefenseAssetRoleDeception    DefenseAssetRole = "deception"
	DefenseAssetRoleSupply       DefenseAssetRole = "supply"
	DefenseAssetRoleEngineering  DefenseAssetRole = "engineering"
	DefenseAssetRoleRecon        DefenseAssetRole = "recon"
	DefenseAssetRoleSpecial      DefenseAssetRole = "special"
)

// ValidDefenseAssetRoles — список допустимых ролей.
var ValidDefenseAssetRoles = []DefenseAssetRole{
	DefenseAssetRoleDetection,
	DefenseAssetRoleDestruction,
	DefenseAssetRoleEW,
	DefenseAssetRoleC2,
	DefenseAssetRoleCover,
	DefenseAssetRoleDeception,
	DefenseAssetRoleSupply,
	DefenseAssetRoleEngineering,
	DefenseAssetRoleRecon,
	DefenseAssetRoleSpecial,
}

// DefenseAssetCoverageType — тип покрытия средства защиты.
type DefenseAssetCoverageType string

const (
	DefenseAssetCoverageCircle  DefenseAssetCoverageType = "circle"
	DefenseAssetCoverageSector  DefenseAssetCoverageType = "sector"
	DefenseAssetCoverageLine    DefenseAssetCoverageType = "line"
	DefenseAssetCoveragePolygon DefenseAssetCoverageType = "polygon"
	DefenseAssetCoverageNone    DefenseAssetCoverageType = "none"
)

// ValidDefenseAssetCoverageTypes — список допустимых типов покрытия.
var ValidDefenseAssetCoverageTypes = []DefenseAssetCoverageType{
	DefenseAssetCoverageCircle,
	DefenseAssetCoverageSector,
	DefenseAssetCoverageLine,
	DefenseAssetCoveragePolygon,
	DefenseAssetCoverageNone,
}

// DeploymentType — тип развёртывания средства защиты.
type DeploymentType string

const (
	DeploymentTypeStatic         DeploymentType = "static"
	DeploymentTypeMobile         DeploymentType = "mobile"
	DeploymentTypeInfrastructure DeploymentType = "infrastructure"
	DeploymentTypeSoftware       DeploymentType = "software"
	DeploymentTypeExternal       DeploymentType = "external"
)

// ValidDeploymentTypes — список допустимых типов развёртывания.
var ValidDeploymentTypes = []DeploymentType{
	DeploymentTypeStatic,
	DeploymentTypeMobile,
	DeploymentTypeInfrastructure,
	DeploymentTypeSoftware,
	DeploymentTypeExternal,
}

// PlacementType — тип размещения средства защиты.
type PlacementType string

const (
	PlacementTypeMapObject    PlacementType = "map-object"
	PlacementTypeZoneObject   PlacementType = "zone-object"
	PlacementTypeNonPhysical  PlacementType = "non-physical"
)

// ValidPlacementTypes — список допустимых типов размещения.
var ValidPlacementTypes = []PlacementType{
	PlacementTypeMapObject,
	PlacementTypeZoneObject,
	PlacementTypeNonPhysical,
}

// DefensePriority — приоритет средства защиты.
type DefensePriority string

const (
	DefensePriorityCritical DefensePriority = "critical"
	DefensePriorityHigh     DefensePriority = "high"
	DefensePriorityMedium   DefensePriority = "medium"
	DefensePriorityLow      DefensePriority = "low"
)

// ValidDefensePriorities — список допустимых приоритетов.
var ValidDefensePriorities = []DefensePriority{
	DefensePriorityCritical,
	DefensePriorityHigh,
	DefensePriorityMedium,
	DefensePriorityLow,
}

// DefenseAssetCompoundProfile — профиль составной установки (МОГ, ПВН, ГОР, КПП).
type DefenseAssetCompoundProfile struct {
	Kind           string  `json:"kind,omitempty"`
	PostType       string  `json:"postType,omitempty"`
	PersonnelCount string  `json:"personnelCount,omitempty"`
	Accountability string  `json:"accountability,omitempty"`
	Armament       string  `json:"armament,omitempty"`
	WeaponUnits    string  `json:"weaponUnits,omitempty"`
	SectorOrRange  string  `json:"sectorOrRange,omitempty"`
	Azimuth        float64 `json:"azimuth,omitempty"`
}

// WeaponSpecification — ТТХ для оружия/установок.
type WeaponSpecification struct {
	Caliber        *string `json:"caliber,omitempty"`
	AmmunitionType *string `json:"ammunitionType,omitempty"`
	OperationMode  *string `json:"operationMode,omitempty"`
	ModuleCount    *int    `json:"moduleCount,omitempty"`
	IsManual       *bool   `json:"isManual,omitempty"`
}

// DetectionSpecification — ТТХ для средств обнаружения.
type DetectionSpecification struct {
	FrequencyRange  *string  `json:"frequencyRange,omitempty"`
	DetectionMode   *string  `json:"detectionMode,omitempty"`
	RotationSpeed   *float64 `json:"rotationSpeed,omitempty"`
	FieldOfView     *float64 `json:"fieldOfView,omitempty"`
	HasThermalImager *bool   `json:"hasThermalImager,omitempty"`
}

// EWSpecification — ТТХ для средств РЭБ/спуферов.
type EWSpecification struct {
	FrequencyRange *string  `json:"frequencyRange,omitempty"`
	ActionRange    *float64 `json:"actionRange,omitempty"`
	Azimuth        *float64 `json:"azimuth,omitempty"`
}

// LayerType — тип слоя карты.
type LayerType string

// DefenseAsset — агрегат средства защиты.
type DefenseAsset struct {
	id                    string
	name                  string
	shortName             string
	description           string
	category              DefenseAssetCategory
	roles                 []DefenseAssetRole
	pricePerUnitMln       *float64
	currency              string
	unitLabel             string
	compatibleLayerTypes  []LayerType
	recommendedLayerCodes []string
	compatibleLayerCodes  []string
	incompatibleLayerCodes []string
	protectionType        string
	minEffectiveDistance  *float64
	maxEffectiveDistance  *float64
	coverageType          DefenseAssetCoverageType
	coverageRadius        *float64
	coverageAngle         *float64
	deploymentType        DeploymentType
	placementType         PlacementType
	iconURL               string
	modelURL              string
	score                 *int
	priority              *DefensePriority
	compoundProfile       *DefenseAssetCompoundProfile
	weaponSpec            *WeaponSpecification
	detectionSpec         *DetectionSpecification
	ewSpec                *EWSpecification
	tags                  []string
	legacyItemID          string
	calculatorAssetID     *string
	mapCatalogGroupIDs    []string
	enterpriseID          *string
	isPublic              bool
	createdAt             time.Time
	updatedAt             time.Time
}

// NewDefenseAsset создаёт новое средство защиты с валидацией.
func NewDefenseAsset(
	id, name, shortName, description string,
	category DefenseAssetCategory,
	roles []DefenseAssetRole,
	pricePerUnitMln *float64,
	currency, unitLabel string,
	compatibleLayerTypes []LayerType,
	recommendedLayerCodes, compatibleLayerCodes, incompatibleLayerCodes []string,
	protectionType string,
	minEffectiveDistance, maxEffectiveDistance *float64,
	coverageType DefenseAssetCoverageType,
	coverageRadius, coverageAngle *float64,
	deploymentType DeploymentType,
	placementType PlacementType,
	iconURL, modelURL string,
	score *int,
	priority              *DefensePriority,
	compoundProfile       *DefenseAssetCompoundProfile,
	weaponSpec            *WeaponSpecification,
	detectionSpec         *DetectionSpecification,
	ewSpec                *EWSpecification,
	tags                  []string,
	legacyItemID string,
	calculatorAssetID *string,
	mapCatalogGroupIDs []string,
	enterpriseID *string,
	isPublic bool,
	createdAt, updatedAt time.Time,
) (*DefenseAsset, error) {
	if name == "" {
		return nil, ErrDefenseAssetInvalidName
	}
	if !isValidCategory(category) {
		return nil, fmt.Errorf("%w: %s", ErrDefenseAssetInvalidCategory, category)
	}
	if !isValidCoverageType(coverageType) {
		return nil, fmt.Errorf("%w: %s", ErrDefenseAssetInvalidCoverageType, coverageType)
	}

	if roles == nil {
		roles = []DefenseAssetRole{}
	}
	if compatibleLayerTypes == nil {
		compatibleLayerTypes = []LayerType{}
	}
	if recommendedLayerCodes == nil {
		recommendedLayerCodes = []string{}
	}
	if compatibleLayerCodes == nil {
		compatibleLayerCodes = []string{}
	}
	if incompatibleLayerCodes == nil {
		incompatibleLayerCodes = []string{}
	}
	if tags == nil {
		tags = []string{}
	}
	if mapCatalogGroupIDs == nil {
		mapCatalogGroupIDs = []string{}
	}

	if currency == "" {
		currency = "RUB"
	}
	if deploymentType == "" {
		deploymentType = DeploymentTypeStatic
	} else if !isValidDeploymentType(deploymentType) {
		return nil, fmt.Errorf("%w: %s", ErrDefenseAssetInvalidDeploymentType, deploymentType)
	}
	if placementType == "" {
		placementType = PlacementTypeMapObject
	} else if !isValidPlacementType(placementType) {
		return nil, fmt.Errorf("%w: %s", ErrDefenseAssetInvalidPlacementType, placementType)
	}

	if err := validateCategorySpecCompatibility(category, weaponSpec, detectionSpec, ewSpec); err != nil {
		return nil, err
	}

	return &DefenseAsset{
		id:                    id,
		name:                  name,
		shortName:             shortName,
		description:           description,
		category:              category,
		roles:                 roles,
		pricePerUnitMln:       pricePerUnitMln,
		currency:              currency,
		unitLabel:             unitLabel,
		compatibleLayerTypes:  compatibleLayerTypes,
		recommendedLayerCodes: recommendedLayerCodes,
		compatibleLayerCodes:  compatibleLayerCodes,
		incompatibleLayerCodes: incompatibleLayerCodes,
		protectionType:        protectionType,
		minEffectiveDistance:  minEffectiveDistance,
		maxEffectiveDistance:  maxEffectiveDistance,
		coverageType:          coverageType,
		coverageRadius:        coverageRadius,
		coverageAngle:         coverageAngle,
		deploymentType:        deploymentType,
		placementType:         placementType,
		iconURL:               iconURL,
		modelURL:              modelURL,
		score:                 score,
		priority:              priority,
		compoundProfile:       compoundProfile,
		weaponSpec:            weaponSpec,
		detectionSpec:         detectionSpec,
		ewSpec:                ewSpec,
		tags:                  tags,
		legacyItemID:          legacyItemID,
		calculatorAssetID:     calculatorAssetID,
		mapCatalogGroupIDs:    mapCatalogGroupIDs,
		enterpriseID:          enterpriseID,
		isPublic:              isPublic,
		createdAt:             createdAt,
		updatedAt:             updatedAt,
	}, nil
}

// Getters.
func (a *DefenseAsset) ID() string                           { return a.id }
func (a *DefenseAsset) Name() string                         { return a.name }
func (a *DefenseAsset) ShortName() string                    { return a.shortName }
func (a *DefenseAsset) Description() string                  { return a.description }
func (a *DefenseAsset) Category() DefenseAssetCategory       { return a.category }
func (a *DefenseAsset) Roles() []DefenseAssetRole            { return a.roles }
func (a *DefenseAsset) PricePerUnitMln() *float64            { return a.pricePerUnitMln }
func (a *DefenseAsset) Currency() string                     { return a.currency }
func (a *DefenseAsset) UnitLabel() string                    { return a.unitLabel }
func (a *DefenseAsset) CompatibleLayerTypes() []LayerType    { return a.compatibleLayerTypes }
func (a *DefenseAsset) RecommendedLayerCodes() []string      { return a.recommendedLayerCodes }
func (a *DefenseAsset) CompatibleLayerCodes() []string       { return a.compatibleLayerCodes }
func (a *DefenseAsset) IncompatibleLayerCodes() []string     { return a.incompatibleLayerCodes }
func (a *DefenseAsset) ProtectionType() string               { return a.protectionType }
func (a *DefenseAsset) MinEffectiveDistance() *float64       { return a.minEffectiveDistance }
func (a *DefenseAsset) MaxEffectiveDistance() *float64       { return a.maxEffectiveDistance }
func (a *DefenseAsset) CoverageType() DefenseAssetCoverageType { return a.coverageType }
func (a *DefenseAsset) CoverageRadius() *float64             { return a.coverageRadius }
func (a *DefenseAsset) CoverageAngle() *float64              { return a.coverageAngle }
func (a *DefenseAsset) DeploymentType() DeploymentType       { return a.deploymentType }
func (a *DefenseAsset) PlacementType() PlacementType         { return a.placementType }
func (a *DefenseAsset) IconURL() string                      { return a.iconURL }
func (a *DefenseAsset) ModelURL() string                     { return a.modelURL }
func (a *DefenseAsset) Score() *int                          { return a.score }
func (a *DefenseAsset) Priority() *DefensePriority           { return a.priority }
func (a *DefenseAsset) CompoundProfile() *DefenseAssetCompoundProfile { return a.compoundProfile }
func (a *DefenseAsset) WeaponSpec() *WeaponSpecification            { return a.weaponSpec }
func (a *DefenseAsset) DetectionSpec() *DetectionSpecification      { return a.detectionSpec }
func (a *DefenseAsset) EWSpec() *EWSpecification                   { return a.ewSpec }
func (a *DefenseAsset) Tags() []string                       { return a.tags }
func (a *DefenseAsset) LegacyItemID() string                 { return a.legacyItemID }
func (a *DefenseAsset) CalculatorAssetID() *string           { return a.calculatorAssetID }
func (a *DefenseAsset) MapCatalogGroupIDs() []string         { return a.mapCatalogGroupIDs }
func (a *DefenseAsset) EnterpriseID() *string                { return a.enterpriseID }
func (a *DefenseAsset) IsPublic() bool                       { return a.isPublic }
func (a *DefenseAsset) CreatedAt() time.Time                 { return a.createdAt }
func (a *DefenseAsset) UpdatedAt() time.Time                 { return a.updatedAt }

// Setters for repository and service usage.
func (a *DefenseAsset) SetID(id string)             { a.id = id }
func (a *DefenseAsset) SetCreatedAt(t time.Time)    { a.createdAt = t }
func (a *DefenseAsset) SetUpdatedAt(t time.Time)    { a.updatedAt = t }
func (a *DefenseAsset) SetEnterpriseID(id *string)   { a.enterpriseID = id }
func (a *DefenseAsset) SetIsPublic(v bool)          { a.isPublic = v }
func (a *DefenseAsset) SetRoles(v []DefenseAssetRole) { a.roles = v; a.updatedAt = time.Now().UTC() }
func (a *DefenseAsset) SetShortName(v string)        { a.shortName = v; a.updatedAt = time.Now().UTC() }
func (a *DefenseAsset) SetPricePerUnitMln(v *float64) { a.pricePerUnitMln = v; a.updatedAt = time.Now().UTC() }
func (a *DefenseAsset) SetCurrency(v string)         { a.currency = v; a.updatedAt = time.Now().UTC() }
func (a *DefenseAsset) SetUnitLabel(v string)        { a.unitLabel = v; a.updatedAt = time.Now().UTC() }
func (a *DefenseAsset) SetCompatibleLayerTypes(v []LayerType) { a.compatibleLayerTypes = v; a.updatedAt = time.Now().UTC() }
func (a *DefenseAsset) SetRecommendedLayerCodes(v []string)   { a.recommendedLayerCodes = v; a.updatedAt = time.Now().UTC() }
func (a *DefenseAsset) SetCompatibleLayerCodes(v []string)    { a.compatibleLayerCodes = v; a.updatedAt = time.Now().UTC() }
func (a *DefenseAsset) SetIncompatibleLayerCodes(v []string)  { a.incompatibleLayerCodes = v; a.updatedAt = time.Now().UTC() }
func (a *DefenseAsset) SetProtectionType(v string)  { a.protectionType = v; a.updatedAt = time.Now().UTC() }
func (a *DefenseAsset) SetMinEffectiveDistance(v *float64) { a.minEffectiveDistance = v; a.updatedAt = time.Now().UTC() }
func (a *DefenseAsset) SetMaxEffectiveDistance(v *float64) { a.maxEffectiveDistance = v; a.updatedAt = time.Now().UTC() }
func (a *DefenseAsset) SetCoverageRadius(v *float64) { a.coverageRadius = v; a.updatedAt = time.Now().UTC() }
func (a *DefenseAsset) SetCoverageAngle(v *float64)  { a.coverageAngle = v; a.updatedAt = time.Now().UTC() }
func (a *DefenseAsset) SetDeploymentType(v DeploymentType)     { a.deploymentType = v; a.updatedAt = time.Now().UTC() }
func (a *DefenseAsset) SetPlacementType(v PlacementType)       { a.placementType = v; a.updatedAt = time.Now().UTC() }
func (a *DefenseAsset) SetIconURL(v string)         { a.iconURL = v; a.updatedAt = time.Now().UTC() }
func (a *DefenseAsset) SetModelURL(v string)        { a.modelURL = v; a.updatedAt = time.Now().UTC() }
func (a *DefenseAsset) SetScore(v *int)             { a.score = v; a.updatedAt = time.Now().UTC() }
func (a *DefenseAsset) SetPriority(v *DefensePriority)    { a.priority = v; a.updatedAt = time.Now().UTC() }
func (a *DefenseAsset) SetCompoundProfile(v *DefenseAssetCompoundProfile) { a.compoundProfile = v; a.updatedAt = time.Now().UTC() }
func (a *DefenseAsset) SetWeaponSpec(v *WeaponSpecification)              { a.weaponSpec = v; a.updatedAt = time.Now().UTC() }
func (a *DefenseAsset) SetDetectionSpec(v *DetectionSpecification)        { a.detectionSpec = v; a.updatedAt = time.Now().UTC() }
func (a *DefenseAsset) SetEWSpec(v *EWSpecification)                      { a.ewSpec = v; a.updatedAt = time.Now().UTC() }
func (a *DefenseAsset) SetTags(v []string)          { a.tags = v; a.updatedAt = time.Now().UTC() }
func (a *DefenseAsset) SetLegacyItemID(v string)    { a.legacyItemID = v; a.updatedAt = time.Now().UTC() }
func (a *DefenseAsset) SetCalculatorAssetID(v *string)   { a.calculatorAssetID = v; a.updatedAt = time.Now().UTC() }
func (a *DefenseAsset) SetMapCatalogGroupIDs(v []string) { a.mapCatalogGroupIDs = v; a.updatedAt = time.Now().UTC() }

// UpdateName обновляет название средства защиты.
func (a *DefenseAsset) UpdateName(name string) error {
	if name == "" {
		return ErrDefenseAssetInvalidName
	}
	a.name = name
	a.updatedAt = time.Now().UTC()
	return nil
}

// UpdateDescription обновляет описание средства защиты.
func (a *DefenseAsset) UpdateDescription(desc string) {
	a.description = desc
	a.updatedAt = time.Now().UTC()
}

// UpdateCategory обновляет категорию средства защиты.
func (a *DefenseAsset) UpdateCategory(cat DefenseAssetCategory) error {
	if !isValidCategory(cat) {
		return fmt.Errorf("%w: %s", ErrDefenseAssetInvalidCategory, cat)
	}
	a.category = cat
	a.updatedAt = time.Now().UTC()
	return nil
}

// UpdateCoverageType обновляет тип покрытия.
func (a *DefenseAsset) UpdateCoverageType(ct DefenseAssetCoverageType) error {
	if !isValidCoverageType(ct) {
		return fmt.Errorf("%w: %s", ErrDefenseAssetInvalidCoverageType, ct)
	}
	a.coverageType = ct
	a.updatedAt = time.Now().UTC()
	return nil
}

// validateCategorySpecCompatibility проверяет, что спецификация соответствует категории.
func validateCategorySpecCompatibility(
	category DefenseAssetCategory,
	weaponSpec *WeaponSpecification,
	detectionSpec *DetectionSpecification,
	ewSpec *EWSpecification,
) error {
	// Определяем категории, совместимые с каждым типом спецификации
	weaponCategories := map[DefenseAssetCategory]bool{
		DefenseAssetCategoryArtillery:   true,
		DefenseAssetCategoryMissile:     true,
		DefenseAssetCategoryAntiMissile: true,
		DefenseAssetCategoryShturmovaya: true,
		DefenseAssetCategoryAircraft:    true,
		DefenseAssetCategoryHelicopter:  true,
		DefenseAssetCategoryUAV:         true,
		DefenseAssetCategoryShip:        true,
	}
	detectionCategories := map[DefenseAssetCategory]bool{
		DefenseAssetCategoryRadiotechnical: true,
		DefenseAssetCategoryRadar:          true,
	}
	ewCategories := map[DefenseAssetCategory]bool{
		DefenseAssetCategoryEW: true,
	}

	if weaponSpec != nil && !weaponCategories[category] {
		return fmt.Errorf("%w: weapon specification is not compatible with category %s", ErrDefenseAssetInvalidSpecification, category)
	}
	if detectionSpec != nil && !detectionCategories[category] {
		return fmt.Errorf("%w: detection specification is not compatible with category %s", ErrDefenseAssetInvalidSpecification, category)
	}
	if ewSpec != nil && !ewCategories[category] {
		return fmt.Errorf("%w: ew specification is not compatible with category %s", ErrDefenseAssetInvalidSpecification, category)
	}
	return nil
}

func isValidCategory(c DefenseAssetCategory) bool {
	for _, v := range ValidDefenseAssetCategories {
		if c == v {
			return true
		}
	}
	return false
}

func isValidCoverageType(ct DefenseAssetCoverageType) bool {
	for _, v := range ValidDefenseAssetCoverageTypes {
		if ct == v {
			return true
		}
	}
	return false
}

func isValidDeploymentType(d DeploymentType) bool {
	for _, v := range ValidDeploymentTypes {
		if d == v {
			return true
		}
	}
	return false
}

func isValidPlacementType(p PlacementType) bool {
	for _, v := range ValidPlacementTypes {
		if p == v {
			return true
		}
	}
	return false
}
