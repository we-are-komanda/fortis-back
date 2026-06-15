package ui

import (
	budgetDomain "github.com/fortis/backend/internal/modules/budget/domain"
	budgetUi "github.com/fortis/backend/internal/modules/budget/ui"
	"github.com/fortis/backend/internal/modules/report/domain"
)

// domainToReportResponse преобразует доменный ReportPayload в DTO ответа.
func domainToReportResponse(payload *domain.ReportPayload) ReportResponse {
	layers := make([]LayerDTO, len(payload.Layers()))
	for i, l := range payload.Layers() {
		layers[i] = layerToDTO(l)
	}

	objects := make([]ReportObjectLineDTO, len(payload.PlacedObjects()))
	for i, o := range payload.PlacedObjects() {
		objects[i] = objectLineToDTO(o)
	}

	// Преобразуем baseObject
	baseObj := baseObjectToDTO(payload.BaseObject())

	// Преобразуем estimate (CostCalculation) через типы budget DTO
	calc := payload.Estimate()
	estimateDTO := mapCostCalculation(calc)

	// Преобразуем structuralProfile
	profile := payload.StructuralProfile()
	profileDTO := mapStructuralProfile(profile)

	return ReportResponse{
		ProjectID:         payload.ProjectID(),
		ProjectName:       payload.ProjectName(),
		BaseObject:        baseObj,
		Layers:            layers,
		PlacedObjects:     objects,
		Estimate:          estimateDTO,
		StructuralProfile: profileDTO,
		HideCost:          payload.HideCost(),
	}
}

// baseObjectToDTO преобразует доменный ReportBaseObject в DTO.
func baseObjectToDTO(bo domain.ReportBaseObject) BaseObjectDTO {
	return BaseObjectDTO{
		ID:   bo.ID(),
		Name: bo.Name(),
		Center: CenterDTO{
			Lat: bo.Lat(),
			Lng: bo.Lng(),
		},
	}
}

// layerToDTO преобразует доменный ReportLayer в DTO.
func layerToDTO(l domain.ReportLayer) LayerDTO {
	return LayerDTO{
		ID:            l.ID(),
		Name:          l.Name(),
		Code:          l.Code(),
		Description:   l.Description(),
		GeometryType:  l.GeometryType(),
		CenterLat:     l.CenterLat(),
		CenterLng:     l.CenterLng(),
		RadiusM:       l.RadiusM(),
		MinRadiusM:    l.MinRadiusM(),
		MaxRadiusM:    l.MaxRadiusM(),
		Color:         l.Color(),
		Opacity:       l.Opacity(),
	}
}

// objectLineToDTO преобразует доменный ReportPlacedObject в DTO.
func objectLineToDTO(o domain.ReportPlacedObject) ReportObjectLineDTO {
	dto := ReportObjectLineDTO{
		ObjectID:       o.ID(),
		AssetID:        o.AssetID(),
		AssetName:      o.AssetName(),
		LayerID:        o.LayerID(),
		LayerCode:      o.LayerCode(),
		LayerName:      o.LayerName(),
		Quantity:       o.Quantity(),
		ProtectionType: o.ProtectionType(),
		UnitPriceMln:   o.UnitPriceMln(),
		LineTotalMln:   o.LineTotalMln(),
		IsCompoundPost: o.IsCompoundPost(),
	}

	if cs := o.CompositionSummary(); cs != nil {
		dto.CompositionSummary = &CompositionSummaryDTO{
			PostType:       cs.PostType(),
			Personnel:      cs.Personnel(),
			Accountability: cs.Accountability(),
			Armament:       cs.Armament(),
			WeaponUnits:    cs.WeaponUnits(),
			SectorOrRange:  cs.SectorOrRange(),
			Azimuth:        cs.Azimuth(),
		}
	}

	if ws := o.WeaponSummary(); ws != nil {
		dto.WeaponSummary = &WeaponSummaryDTO{
			Caliber:        ws.Caliber(),
			AmmunitionType: ws.AmmunitionType(),
			OperationMode:  ws.OperationMode(),
			ModuleCount:    ws.ModuleCount(),
			IsManual:       ws.IsManual(),
		}
	}

	if az := o.AzimuthSectorSummary(); az != nil {
		dto.AzimuthSectorSummary = &AzimuthSectorSummaryDTO{
			Azimuth:      az.Azimuth(),
			CoverageType: az.CoverageType(),
			Angle:        az.Angle(),
		}
	}

	return dto
}

// mapCostCalculation преобразует доменный CostCalculation в DTO.
func mapCostCalculation(calc budgetDomain.CostCalculation) budgetUi.CostCalculationDTO {
	byEchelon := make([]budgetUi.EchelonEstimateDTO, len(calc.ByEchelon()))
	for i, e := range calc.ByEchelon() {
		byEchelon[i] = mapEchelonEstimate(e)
	}

	byType := make([]budgetUi.TypeEstimateDTO, len(calc.ByType()))
	for i, t := range calc.ByType() {
		byType[i] = mapTypeEstimate(t)
	}

	byObject := make([]budgetUi.EstimateLineDTO, len(calc.ByObject()))
	for i, l := range calc.ByObject() {
		byObject[i] = mapEstimateLine(l)
	}

	return budgetUi.CostCalculationDTO{
		TotalMln:  calc.TotalMln(),
		ByEchelon: byEchelon,
		ByType:    byType,
		ByObject:  byObject,
	}
}

// mapEchelonEstimate преобразует EchelonEstimate в DTO.
func mapEchelonEstimate(e budgetDomain.EchelonEstimate) budgetUi.EchelonEstimateDTO {
	lines := make([]budgetUi.EstimateLineDTO, len(e.Lines()))
	for i, l := range e.Lines() {
		lines[i] = mapEstimateLine(l)
	}
	return budgetUi.EchelonEstimateDTO{
		EchelonID:       e.EchelonID(),
		EchelonName:     e.EchelonName(),
		Lines:           lines,
		EchelonTotalMln: e.EchelonTotalMln(),
	}
}

// mapTypeEstimate преобразует TypeEstimate в DTO.
func mapTypeEstimate(t budgetDomain.TypeEstimate) budgetUi.TypeEstimateDTO {
	lines := make([]budgetUi.EstimateLineDTO, len(t.Lines()))
	for i, l := range t.Lines() {
		lines[i] = mapEstimateLine(l)
	}
	return budgetUi.TypeEstimateDTO{
		TypeID:       t.TypeID(),
		TypeName:     t.TypeName(),
		Lines:        lines,
		TypeTotalMln: t.TypeTotalMln(),
	}
}

// mapEstimateLine преобразует EstimateLine в DTO.
func mapEstimateLine(l budgetDomain.EstimateLine) budgetUi.EstimateLineDTO {
	return budgetUi.EstimateLineDTO{
		ObjectID:     l.ObjectID(),
		AssetID:      l.AssetID(),
		AssetName:    l.AssetName(),
		EchelonID:    l.EchelonID(),
		EchelonName:  l.EchelonName(),
		TypeID:       l.TypeID(),
		TypeName:     l.TypeName(),
		Quantity:     l.Quantity(),
		UnitPriceMln: l.UnitPriceMln(),
		LineTotalMln: l.LineTotalMln(),
	}
}

// mapStructuralProfile преобразует StructuralProfile в DTO.
func mapStructuralProfile(sp budgetDomain.StructuralProfile) budgetUi.StructuralProfileDTO {
	byEchelon := make([]budgetUi.EchelonProfileDTO, len(sp.ByEchelon()))
	for i, ep := range sp.ByEchelon() {
		byEchelon[i] = mapEchelonProfile(ep)
	}

	return budgetUi.StructuralProfileDTO{
		ObjectCount:     sp.ObjectCount(),
		UnitCount:       sp.UnitCount(),
		EchelonCount:    sp.EchelonCount(),
		CategoryCount:   sp.CategoryCount(),
		ConflictCount:   sp.ConflictCount(),
		CoveredObjCount: sp.CoveredObjCount(),
		TotalMln:        sp.TotalMln(),
		ByEchelon:       byEchelon,
	}
}

// mapEchelonProfile преобразует StructuralEchelonProfile в DTO.
func mapEchelonProfile(ep budgetDomain.StructuralEchelonProfile) budgetUi.EchelonProfileDTO {
	return budgetUi.EchelonProfileDTO{
		LayerID:         ep.LayerID(),
		LayerCode:       ep.LayerCode(),
		LayerName:       ep.LayerName(),
		ObjectCount:     ep.ObjectCount(),
		UnitCount:       ep.UnitCount(),
		CategoryCount:   ep.CategoryCount(),
		ConflictCount:   ep.ConflictCount(),
		CoveredObjCount: ep.CoveredObjCount(),
	}
}
