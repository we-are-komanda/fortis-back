package ui

import (
	"time"

	"github.com/fortis/backend/internal/modules/defense_asset/application"
	"github.com/fortis/backend/internal/modules/defense_asset/domain"
)

// defenseAssetToDTO преобразует доменный DefenseAsset в DTO ответа.
func defenseAssetToDTO(a *domain.DefenseAsset) DefenseAssetDTO {
	roles := make([]string, len(a.Roles()))
	for i, r := range a.Roles() {
		roles[i] = string(r)
	}

	layerTypes := make([]string, len(a.CompatibleLayerTypes()))
	for i, lt := range a.CompatibleLayerTypes() {
		layerTypes[i] = string(lt)
	}

	var priority *string
	if a.Priority() != nil {
		p := string(*a.Priority())
		priority = &p
	}

	var compoundProfile *CompoundProfileDTO
	if a.CompoundProfile() != nil {
		cp := a.CompoundProfile()
		compoundProfile = &CompoundProfileDTO{
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

	var weaponSpec *WeaponSpecDTO
	if a.WeaponSpec() != nil {
		ws := a.WeaponSpec()
		weaponSpec = &WeaponSpecDTO{
			Caliber:        ws.Caliber,
			AmmunitionType: ws.AmmunitionType,
			OperationMode:  ws.OperationMode,
			ModuleCount:    ws.ModuleCount,
			IsManual:       ws.IsManual,
		}
	}

	var detectionSpec *DetectionSpecDTO
	if a.DetectionSpec() != nil {
		ds := a.DetectionSpec()
		detectionSpec = &DetectionSpecDTO{
			FrequencyRange:   ds.FrequencyRange,
			DetectionMode:    ds.DetectionMode,
			RotationSpeed:    ds.RotationSpeed,
			FieldOfView:      ds.FieldOfView,
			HasThermalImager: ds.HasThermalImager,
		}
	}

	var ewSpec *EWSpecDTO
	if a.EWSpec() != nil {
		es := a.EWSpec()
		ewSpec = &EWSpecDTO{
			FrequencyRange: es.FrequencyRange,
			ActionRange:    es.ActionRange,
			Azimuth:        es.Azimuth,
		}
	}

	return DefenseAssetDTO{
		ID:                    a.ID(),
		Name:                  a.Name(),
		ShortName:             a.ShortName(),
		Description:           a.Description(),
		Category:              string(a.Category()),
		Roles:                 roles,
		PricePerUnitMln:       a.PricePerUnitMln(),
		Currency:              a.Currency(),
		UnitLabel:             a.UnitLabel(),
		CompatibleLayerTypes:  layerTypes,
		RecommendedLayerCodes: a.RecommendedLayerCodes(),
		CompatibleLayerCodes:  a.CompatibleLayerCodes(),
		IncompatibleLayerCodes: a.IncompatibleLayerCodes(),
		ProtectionType:        a.ProtectionType(),
		MinEffectiveDistance:  a.MinEffectiveDistance(),
		MaxEffectiveDistance:  a.MaxEffectiveDistance(),
		CoverageType:          string(a.CoverageType()),
		CoverageRadius:        a.CoverageRadius(),
		CoverageAngle:         a.CoverageAngle(),
		DeploymentType:        string(a.DeploymentType()),
		PlacementType:         string(a.PlacementType()),
		IconURL:               a.IconURL(),
		ModelURL:              a.ModelURL(),
		Score:                 a.Score(),
		Priority:              priority,
		CompoundProfile:       compoundProfile,
		WeaponSpec:            weaponSpec,
		DetectionSpec:         detectionSpec,
		EWSpec:                ewSpec,
		Tags:                  a.Tags(),
		LegacyItemID:          a.LegacyItemID(),
		CalculatorAssetID:     a.CalculatorAssetID(),
		MapCatalogGroupIDs:    a.MapCatalogGroupIDs(),
		EnterpriseID:          a.EnterpriseID(),
		IsPublic:              a.IsPublic(),
		CreatedAt:             a.CreatedAt().UTC().Format(time.RFC3339),
		UpdatedAt:             a.UpdatedAt().UTC().Format(time.RFC3339),
	}
}

// mapCreateRequestToDomain преобразует DTO создания в доменную модель.
func mapCreateRequestToDomain(req CreateDefenseAssetRequest) application.CreateInput {
	roles := make([]domain.DefenseAssetRole, len(req.Roles))
	for i, r := range req.Roles {
		roles[i] = domain.DefenseAssetRole(r)
	}

	layerTypes := make([]domain.LayerType, len(req.CompatibleLayerTypes))
	for i, lt := range req.CompatibleLayerTypes {
		layerTypes[i] = domain.LayerType(lt)
	}

	var priority *domain.DefensePriority
	if req.Priority != nil {
		p := domain.DefensePriority(*req.Priority)
		priority = &p
	}

	var compoundProfile *domain.DefenseAssetCompoundProfile
	if req.CompoundProfile != nil {
		compoundProfile = &domain.DefenseAssetCompoundProfile{
			Kind:           req.CompoundProfile.Kind,
			PostType:       req.CompoundProfile.PostType,
			PersonnelCount: req.CompoundProfile.PersonnelCount,
			Accountability: req.CompoundProfile.Accountability,
			Armament:       req.CompoundProfile.Armament,
			WeaponUnits:    req.CompoundProfile.WeaponUnits,
			SectorOrRange:  req.CompoundProfile.SectorOrRange,
			Azimuth:        req.CompoundProfile.Azimuth,
		}
	}

	var weaponSpec *domain.WeaponSpecification
	if req.WeaponSpec != nil {
		weaponSpec = &domain.WeaponSpecification{
			Caliber:        req.WeaponSpec.Caliber,
			AmmunitionType: req.WeaponSpec.AmmunitionType,
			OperationMode:  req.WeaponSpec.OperationMode,
			ModuleCount:    req.WeaponSpec.ModuleCount,
			IsManual:       req.WeaponSpec.IsManual,
		}
	}

	var detectionSpec *domain.DetectionSpecification
	if req.DetectionSpec != nil {
		detectionSpec = &domain.DetectionSpecification{
			FrequencyRange:   req.DetectionSpec.FrequencyRange,
			DetectionMode:    req.DetectionSpec.DetectionMode,
			RotationSpeed:    req.DetectionSpec.RotationSpeed,
			FieldOfView:      req.DetectionSpec.FieldOfView,
			HasThermalImager: req.DetectionSpec.HasThermalImager,
		}
	}

	var ewSpec *domain.EWSpecification
	if req.EWSpec != nil {
		ewSpec = &domain.EWSpecification{
			FrequencyRange: req.EWSpec.FrequencyRange,
			ActionRange:    req.EWSpec.ActionRange,
			Azimuth:        req.EWSpec.Azimuth,
		}
	}

	return application.CreateInput{
		Name:                  req.Name,
		ShortName:             req.ShortName,
		Description:           req.Description,
		Category:              domain.DefenseAssetCategory(req.Category),
		Roles:                 roles,
		PricePerUnitMln:       req.PricePerUnitMln,
		Currency:              req.Currency,
		UnitLabel:             req.UnitLabel,
		CompatibleLayerTypes:  layerTypes,
		RecommendedLayerCodes: req.RecommendedLayerCodes,
		CompatibleLayerCodes:  req.CompatibleLayerCodes,
		IncompatibleLayerCodes: req.IncompatibleLayerCodes,
		ProtectionType:        req.ProtectionType,
		MinEffectiveDistance:  req.MinEffectiveDistance,
		MaxEffectiveDistance:  req.MaxEffectiveDistance,
		CoverageType:          domain.DefenseAssetCoverageType(req.CoverageType),
		CoverageRadius:        req.CoverageRadius,
		CoverageAngle:         req.CoverageAngle,
		DeploymentType:        domain.DeploymentType(req.DeploymentType),
		PlacementType:         domain.PlacementType(req.PlacementType),
		IconURL:               req.IconURL,
		ModelURL:              req.ModelURL,
		Score:                 req.Score,
		Priority:              priority,
		CompoundProfile:       compoundProfile,
		WeaponSpec:            weaponSpec,
		DetectionSpec:         detectionSpec,
		EWSpec:                ewSpec,
		Tags:                  req.Tags,
		LegacyItemID:          req.LegacyItemID,
		CalculatorAssetID:     req.CalculatorAssetID,
		MapCatalogGroupIDs:    req.MapCatalogGroupIDs,
		EnterpriseID:          req.EnterpriseID,
		IsPublic:              req.IsPublic,
	}
}

// mapSpecDTOtoDomain преобразует DTO спецификации в доменные типы.
func mapWeaponSpecDTO(s *WeaponSpecDTO) *domain.WeaponSpecification {
	if s == nil {
		return nil
	}
	return &domain.WeaponSpecification{
		Caliber:        s.Caliber,
		AmmunitionType: s.AmmunitionType,
		OperationMode:  s.OperationMode,
		ModuleCount:    s.ModuleCount,
		IsManual:       s.IsManual,
	}
}

func mapDetectionSpecDTO(s *DetectionSpecDTO) *domain.DetectionSpecification {
	if s == nil {
		return nil
	}
	return &domain.DetectionSpecification{
		FrequencyRange:   s.FrequencyRange,
		DetectionMode:    s.DetectionMode,
		RotationSpeed:    s.RotationSpeed,
		FieldOfView:      s.FieldOfView,
		HasThermalImager: s.HasThermalImager,
	}
}

func mapEWSpecDTO(s *EWSpecDTO) *domain.EWSpecification {
	if s == nil {
		return nil
	}
	return &domain.EWSpecification{
		FrequencyRange: s.FrequencyRange,
		ActionRange:    s.ActionRange,
		Azimuth:        s.Azimuth,
	}
}

// mapUpdateRequestToServiceInput преобразует DTO обновления в входные данные сервиса.
func mapUpdateRequestToServiceInput(id string, req UpdateDefenseAssetRequest) application.UpdateInput {
	var name *string
	if req.Name != nil {
		name = req.Name
	}
	var shortName *string
	if req.ShortName != nil {
		shortName = req.ShortName
	}
	var description *string
	if req.Description != nil {
		description = req.Description
	}
	var category *domain.DefenseAssetCategory
	if req.Category != nil {
		c := domain.DefenseAssetCategory(*req.Category)
		category = &c
	}
	var coverageType *domain.DefenseAssetCoverageType
	if req.CoverageType != nil {
		ct := domain.DefenseAssetCoverageType(*req.CoverageType)
		coverageType = &ct
	}
	var deploymentType *domain.DeploymentType
	if req.DeploymentType != nil {
		dt := domain.DeploymentType(*req.DeploymentType)
		deploymentType = &dt
	}
	var placementType *domain.PlacementType
	if req.PlacementType != nil {
		pt := domain.PlacementType(*req.PlacementType)
		placementType = &pt
	}
	var iconURL *string
	if req.IconURL != nil {
		iconURL = req.IconURL
	}
	var modelURL *string
	if req.ModelURL != nil {
		modelURL = req.ModelURL
	}
	var legacyItemID *string
	if req.LegacyItemID != nil {
		legacyItemID = req.LegacyItemID
	}
	var currency *string
	if req.Currency != nil {
		currency = req.Currency
	}
	var unitLabel *string
	if req.UnitLabel != nil {
		unitLabel = req.UnitLabel
	}
	var protectionType *string
	if req.ProtectionType != nil {
		protectionType = req.ProtectionType
	}

	var priority *domain.DefensePriority
	if req.Priority != nil {
		p := domain.DefensePriority(*req.Priority)
		priority = &p
	}

	var compoundProfile *domain.DefenseAssetCompoundProfile
	if req.CompoundProfile != nil {
		compoundProfile = &domain.DefenseAssetCompoundProfile{
			Kind:           req.CompoundProfile.Kind,
			PostType:       req.CompoundProfile.PostType,
			PersonnelCount: req.CompoundProfile.PersonnelCount,
			Accountability: req.CompoundProfile.Accountability,
			Armament:       req.CompoundProfile.Armament,
			WeaponUnits:    req.CompoundProfile.WeaponUnits,
			SectorOrRange:  req.CompoundProfile.SectorOrRange,
			Azimuth:        req.CompoundProfile.Azimuth,
		}
	}

	weaponSpec := mapWeaponSpecDTO(req.WeaponSpec)
	detectionSpec := mapDetectionSpecDTO(req.DetectionSpec)
	ewSpec := mapEWSpecDTO(req.EWSpec)

	roles := make([]domain.DefenseAssetRole, len(req.Roles))
	for i, r := range req.Roles {
		roles[i] = domain.DefenseAssetRole(r)
	}

	layerTypes := make([]domain.LayerType, len(req.CompatibleLayerTypes))
	for i, lt := range req.CompatibleLayerTypes {
		layerTypes[i] = domain.LayerType(lt)
	}

	return application.UpdateInput{
		ID:                    id,
		Name:                  name,
		ShortName:             shortName,
		Description:           description,
		Category:              category,
		Roles:                 roles,
		PricePerUnitMln:       req.PricePerUnitMln,
		Currency:              currency,
		UnitLabel:             unitLabel,
		CompatibleLayerTypes:  layerTypes,
		RecommendedLayerCodes: req.RecommendedLayerCodes,
		CompatibleLayerCodes:  req.CompatibleLayerCodes,
		IncompatibleLayerCodes: req.IncompatibleLayerCodes,
		ProtectionType:        protectionType,
		MinEffectiveDistance:  req.MinEffectiveDistance,
		MaxEffectiveDistance:  req.MaxEffectiveDistance,
		CoverageType:          coverageType,
		CoverageRadius:        req.CoverageRadius,
		CoverageAngle:         req.CoverageAngle,
		DeploymentType:        deploymentType,
		PlacementType:         placementType,
		IconURL:               iconURL,
		ModelURL:              modelURL,
		Score:                 req.Score,
		Priority:              priority,
		CompoundProfile:       compoundProfile,
		WeaponSpec:            weaponSpec,
		DetectionSpec:         detectionSpec,
		EWSpec:                ewSpec,
		Tags:                  req.Tags,
		LegacyItemID:          legacyItemID,
		CalculatorAssetID:     req.CalculatorAssetID,
		MapCatalogGroupIDs:    req.MapCatalogGroupIDs,
		IsPublic:              req.IsPublic,
	}
}
