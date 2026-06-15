package application

import (
	"context"
	"errors"
	"fmt"

	budgetApp "github.com/fortis/backend/internal/modules/budget/application"
	budgetDomain "github.com/fortis/backend/internal/modules/budget/domain"
	defenseAssetDomain "github.com/fortis/backend/internal/modules/defense_asset/domain"
	defenseDomain "github.com/fortis/backend/internal/modules/defense_project/domain"
	reportDomain "github.com/fortis/backend/internal/modules/report/domain"
)

// BudgetServiceInterface — интерфейс сервиса бюджета для отчёта.
type BudgetServiceInterface interface {
	CalculateCost(ctx context.Context, projectID string) (*budgetDomain.CostCalculation, error)
}

// ReportService — сервис для построения отчёта GIS MVP.
type ReportService struct {
	projectRepo      defenseDomain.DefenseProjectRepositoryInterface
	budgetSvc        BudgetServiceInterface
	defenseAssetRepo defenseAssetDomain.DefenseAssetRepositoryInterface
}

// NewReportService создаёт новый ReportService.
func NewReportService(
	projectRepo defenseDomain.DefenseProjectRepositoryInterface,
	budgetSvc BudgetServiceInterface,
	defenseAssetRepo defenseAssetDomain.DefenseAssetRepositoryInterface,
) *ReportService {
	return &ReportService{
		projectRepo:      projectRepo,
		budgetSvc:        budgetSvc,
		defenseAssetRepo: defenseAssetRepo,
	}
}

// GetReport собирает полный отчёт для проекта.
func (s *ReportService) GetReport(ctx context.Context, projectID string, hideCost bool) (*reportDomain.ReportPayload, error) {
	if projectID == "" {
		return nil, reportDomain.ErrInvalidProjectID
	}

	// 1. Загрузить DefenseProject
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, defenseDomain.ErrProjectNotFound) {
			return nil, reportDomain.ErrProjectNotFound
		}
		return nil, fmt.Errorf("load project: %w", err)
	}

	// 2. Вычислить CostCalculation
	calc, err := s.budgetSvc.CalculateCost(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("calculate cost: %w", err)
	}

	// 3. Построить StructuralProfile
	profile := budgetApp.BuildStructuralProfile(project, calc)

	// 4. Преобразовать placedObjects → ReportPlacedObject[]
	objects := s.buildObjectLines(ctx, project, calc)

	// 5. Преобразовать layers → ReportLayer[]
	layers := buildReportLayers(project)

	// 6. Построить baseObject
	baseObj := buildReportBaseObject(project)

	// 7. Если hideCost == true — обнулить все *Mln поля
	if hideCost {
		zeroCostFields(calc)
		profile = zeroProfileCost(profile)
		objects = zeroObjectCostFields(objects)
	}

	payload := reportDomain.NewReportPayload(
		project.ProjectID(),
		project.ProjectName(),
		baseObj,
		layers,
		objects,
		*calc,
		profile,
		hideCost,
	)

	return payload, nil
}

// buildReportLayers преобразует слои проекта в ReportLayer.
func buildReportLayers(project *defenseDomain.DefenseProject) []reportDomain.ReportLayer {
	layers := project.Layers()
	result := make([]reportDomain.ReportLayer, 0, len(layers))

	for _, l := range layers {
		var centerLat, centerLng, radiusM, minRadiusM, maxRadiusM *float64
		geom := l.Geometry()
		if geom.Center() != nil {
			center := *geom.Center()
			lat := center.Lat()
			lng := center.Lng()
			centerLat = &lat
			centerLng = &lng
		}
		if geom.RadiusM() != nil {
			r := *geom.RadiusM()
			radiusM = &r
		}
		if geom.MinRadiusM() != nil {
			r := *geom.MinRadiusM()
			minRadiusM = &r
		}
		if geom.MaxRadiusM() != nil {
			r := *geom.MaxRadiusM()
			maxRadiusM = &r
		}

		rl := reportDomain.NewReportLayer(
			l.ID(),
			l.Name(),
			l.Code(),
			l.Description(),
			string(l.GeometryType()),
			centerLat, centerLng,
			radiusM, minRadiusM, maxRadiusM,
			l.Color(),
			l.Opacity(),
		)
		result = append(result, rl)
	}

	return result
}

// buildReportBaseObject преобразует baseObject проекта в ReportBaseObject.
func buildReportBaseObject(project *defenseDomain.DefenseProject) reportDomain.ReportBaseObject {
	bo := project.BaseObject()
	return reportDomain.NewReportBaseObject(
		bo.ID(),
		bo.Name(),
		bo.Center().Lat(),
		bo.Center().Lng(),
	)
}

// buildObjectLines строит строки размещённых объектов из placedObjects + assetLibrary.
func (s *ReportService) buildObjectLines(ctx context.Context, project *defenseDomain.DefenseProject, calc *budgetDomain.CostCalculation) []reportDomain.ReportPlacedObject {
	placedObjects := project.PlacedObjects()
	projectAssets := project.AssetLibrary()
	layers := project.Layers()

	// Индекс ассетов проекта по ID
	projectAssetMap := make(map[string]defenseDomain.DefenseAsset, len(projectAssets))
	for _, a := range projectAssets {
		projectAssetMap[a.ID()] = a
	}

	// Индекс слоёв по ID
	layerMap := make(map[string]defenseDomain.EditableDefenseLayer, len(layers))
	for _, l := range layers {
		layerMap[l.ID()] = l
	}

	// Индекс расчёта по objectID
	calcMap := make(map[string]budgetDomain.EstimateLine)
	for _, line := range calc.ByObject() {
		calcMap[line.ObjectID()] = line
	}

	// Кэш полных ассетов (defense_asset domain)
	fullAssetCache := make(map[string]*defenseAssetDomain.DefenseAsset)

	result := make([]reportDomain.ReportPlacedObject, 0, len(placedObjects))

	for _, obj := range placedObjects {
		projAsset, hasProjAsset := projectAssetMap[obj.AssetID()]
		calcLine, hasCalc := calcMap[obj.ID()]

		var unitPriceMln, lineTotalMln float64
		if hasCalc {
			unitPriceMln = calcLine.UnitPriceMln()
			lineTotalMln = calcLine.LineTotalMln()
		}

		isCompoundPost := false
		var compSummary *reportDomain.CompositionSummary
		var weaponSummary *reportDomain.WeaponSummary
		var azimuthSummary *reportDomain.AzimuthSectorSummary

		if hasProjAsset {
			cp := projAsset.CompoundProfile()
			if cp != nil {
				isCompoundPost = true
				cs := buildCompositionSummary(cp)
				compSummary = &cs
				azSummary := buildAzimuthSectorSummary(cp, string(projAsset.CoverageType()))
				azimuthSummary = &azSummary
			}

			// Загружаем полный ассет для weaponSpec
			fullAsset, ok := fullAssetCache[obj.AssetID()]
			if !ok {
				fa, err := s.defenseAssetRepo.FindByID(ctx, obj.AssetID())
				if err == nil {
					fullAsset = fa
					fullAssetCache[obj.AssetID()] = fa
				}
			}
			if fullAsset != nil {
				ws := buildWeaponSummary(fullAsset.WeaponSpec())
				if ws != nil {
					weaponSummary = ws
				}
			}
		}

		// Определяем layer code и name
		layerCode := obj.LayerID()
		layerName := obj.LayerID()
		if layer, ok := layerMap[obj.LayerID()]; ok {
			layerCode = layer.Code()
			layerName = layer.Name()
		}

		assetName := obj.AssetID()
		if hasProjAsset {
			assetName = projAsset.Name()
		}

		rpo := reportDomain.NewReportPlacedObject(
			obj.ID(),
			obj.AssetID(),
			assetName,
			obj.LayerID(),
			layerCode,
			layerName,
			obj.Coordinates().Lat(),
			obj.Coordinates().Lng(),
			obj.Quantity(),
			projAsset.ProtectionType(),
			unitPriceMln,
			lineTotalMln,
			isCompoundPost,
			compSummary,
			weaponSummary,
			azimuthSummary,
		)
		result = append(result, rpo)
	}

	return result
}

// buildCompositionSummary строит сводку по составной установке.
func buildCompositionSummary(cp *defenseAssetDomain.DefenseAssetCompoundProfile) reportDomain.CompositionSummary {
	return reportDomain.NewCompositionSummary(
		cp.PostType,
		cp.PersonnelCount,
		cp.Accountability,
		cp.Armament,
		cp.WeaponUnits,
		cp.SectorOrRange,
		cp.Azimuth,
	)
}

// buildWeaponSummary строит сводку по оружию.
func buildWeaponSummary(ws *defenseAssetDomain.WeaponSpecification) *reportDomain.WeaponSummary {
	if ws == nil {
		return nil
	}

	caliber := emptyStr(ws.Caliber)
	ammoType := emptyStr(ws.AmmunitionType)
	opMode := emptyStr(ws.OperationMode)
	modCount := ""
	if ws.ModuleCount != nil {
		modCount = fmt.Sprintf("%d", *ws.ModuleCount)
	}
	isManual := ""
	if ws.IsManual != nil {
		if *ws.IsManual {
			isManual = "ручное"
		} else {
			isManual = "автоматическое"
		}
	}

	wsSummary := reportDomain.NewWeaponSummary(caliber, ammoType, opMode, modCount, isManual)
	return &wsSummary
}

// buildAzimuthSectorSummary строит сводку по азимутальному сектору.
func buildAzimuthSectorSummary(cp *defenseAssetDomain.DefenseAssetCompoundProfile, coverageType string) reportDomain.AzimuthSectorSummary {
	return reportDomain.NewAzimuthSectorSummary(cp.Azimuth, coverageType, nil)
}

// zeroCostFields обнуляет все cost-поля в CostCalculation.
func zeroCostFields(calc *budgetDomain.CostCalculation) {
	newCalc := budgetDomain.NewCostCalculation(0, nil, nil, nil)
	*calc = newCalc
}

// zeroProfileCost обнуляет TotalMln в StructuralProfile.
func zeroProfileCost(profile budgetDomain.StructuralProfile) budgetDomain.StructuralProfile {
	return budgetDomain.NewStructuralProfile(
		profile.ObjectCount(),
		profile.UnitCount(),
		profile.EchelonCount(),
		profile.CategoryCount(),
		profile.ConflictCount(),
		profile.CoveredObjCount(),
		0,
		profile.ByEchelon(),
	)
}

// zeroObjectCostFields обнуляет cost-поля у всех ReportPlacedObject.
func zeroObjectCostFields(objects []reportDomain.ReportPlacedObject) []reportDomain.ReportPlacedObject {
	result := make([]reportDomain.ReportPlacedObject, len(objects))
	for i, o := range objects {
		result[i] = reportDomain.NewReportPlacedObject(
			o.ID(), o.AssetID(), o.AssetName(),
			o.LayerID(), o.LayerCode(), o.LayerName(),
			o.Lat(), o.Lng(),
			o.Quantity(), o.ProtectionType(),
			0, 0,
			o.IsCompoundPost(),
			o.CompositionSummary(),
			o.WeaponSummary(),
			o.AzimuthSectorSummary(),
		)
	}
	return result
}

// emptyStr возвращает пустую строку, если указатель nil, иначе разыменованное значение.
func emptyStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
