package ui

import (
	"time"

	"github.com/fortis/backend/internal/modules/budget/domain"
)

// budgetConfigToDomain преобразует DTO в запрос на обновление.
func budgetConfigToDTO(config *domain.BudgetConfig) BudgetConfigDTO {
	return BudgetConfigDTO{
		BudgetMode:      string(config.BudgetMode()),
		BudgetAmountMln: config.BudgetAmountMln(),
		ProjectID:       config.ProjectID(),
		CreatedAt:       config.CreatedAt().Format(time.RFC3339Nano),
		UpdatedAt:       config.UpdatedAt().Format(time.RFC3339Nano),
	}
}

// costCalculationToDTO преобразует доменный расчёт в DTO.
func costCalculationToDTO(calc *domain.CostCalculation) CostCalculationDTO {
	byEchelon := make([]EchelonEstimateDTO, len(calc.ByEchelon()))
	for i, e := range calc.ByEchelon() {
		byEchelon[i] = echelonEstimateToDTO(e)
	}

	byType := make([]TypeEstimateDTO, len(calc.ByType()))
	for i, t := range calc.ByType() {
		byType[i] = typeEstimateToDTO(t)
	}

	byObject := make([]EstimateLineDTO, len(calc.ByObject()))
	for i, l := range calc.ByObject() {
		byObject[i] = estimateLineToDTO(l)
	}

	return CostCalculationDTO{
		TotalMln:  calc.TotalMln(),
		ByEchelon: byEchelon,
		ByType:    byType,
		ByObject:  byObject,
	}
}

func echelonEstimateToDTO(e domain.EchelonEstimate) EchelonEstimateDTO {
	lines := make([]EstimateLineDTO, len(e.Lines()))
	for i, l := range e.Lines() {
		lines[i] = estimateLineToDTO(l)
	}
	return EchelonEstimateDTO{
		EchelonID:       e.EchelonID(),
		EchelonName:     e.EchelonName(),
		Lines:           lines,
		EchelonTotalMln: e.EchelonTotalMln(),
	}
}

func typeEstimateToDTO(t domain.TypeEstimate) TypeEstimateDTO {
	lines := make([]EstimateLineDTO, len(t.Lines()))
	for i, l := range t.Lines() {
		lines[i] = estimateLineToDTO(l)
	}
	return TypeEstimateDTO{
		TypeID:       t.TypeID(),
		TypeName:     t.TypeName(),
		Lines:        lines,
		TypeTotalMln: t.TypeTotalMln(),
	}
}

func estimateLineToDTO(l domain.EstimateLine) EstimateLineDTO {
	return EstimateLineDTO{
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

// comparisonToDTO преобразует доменный результат сравнения в DTO.
func comparisonToDTO(comp *domain.ConfigComparison) ConfigComparisonResponse {
	snapA := snapshotToDTO(comp.ProjectA())
	snapB := snapshotToDTO(comp.ProjectB())
	diff := diffToDTO(comp.Diff())

	return ConfigComparisonResponse{
		ProjectA: snapA,
		ProjectB: snapB,
		Diff:     diff,
	}
}

// snapshotToDTO преобразует доменный слепок конфигурации в DTO.
func snapshotToDTO(s domain.ConfigSnapshot) ConfigSnapshotDTO {
	return ConfigSnapshotDTO{
		ProjectID:         s.ProjectID(),
		ProjectName:       s.ProjectName(),
		StructuralProfile: structuralProfileToDTO(s.StructuralProfile()),
		CostCalculation:   costCalculationToDTO(ptr(s.CostCalculation())),
	}
}

// structuralProfileToDTO преобразует доменный структурный профиль в DTO.
func structuralProfileToDTO(sp domain.StructuralProfile) StructuralProfileDTO {
	byEchelon := make([]EchelonProfileDTO, len(sp.ByEchelon()))
	for i, ep := range sp.ByEchelon() {
		byEchelon[i] = echelonProfileToDTO(ep)
	}

	return StructuralProfileDTO{
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

// echelonProfileToDTO преобразует доменный профиль эшелона в DTO.
func echelonProfileToDTO(ep domain.StructuralEchelonProfile) EchelonProfileDTO {
	return EchelonProfileDTO{
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

// diffToDTO преобразует доменную разницу в DTO.
func diffToDTO(d domain.ConfigDiff) ConfigDiffDTO {
	byEchelon := make([]EchelonDiffDTO, len(d.ByEchelon()))
	for i, ed := range d.ByEchelon() {
		byEchelon[i] = echelonDiffToDTO(ed)
	}

	return ConfigDiffDTO{
		ObjectCountDelta:     d.ObjectCountDelta(),
		UnitCountDelta:       d.UnitCountDelta(),
		EchelonCountDelta:    d.EchelonCountDelta(),
		CategoryCountDelta:    d.CategoryCountDelta(),
		ConflictCountDelta:    d.ConflictCountDelta(),
		CoveredObjCountDelta:  d.CoveredObjCountDelta(),
		CostDeltaMln:         d.CostDeltaMln(),
		ByEchelon:            byEchelon,
	}
}

// echelonDiffToDTO преобразует доменную разницу по эшелону в DTO.
func echelonDiffToDTO(ed domain.EchelonDiff) EchelonDiffDTO {
	return EchelonDiffDTO{
		LayerID:            ed.LayerID(),
		LayerCode:          ed.LayerCode(),
		LayerName:          ed.LayerName(),
		ObjectCountDelta:   ed.ObjectCountDelta(),
		UnitCountDelta:     ed.UnitCountDelta(),
		CategoryCountDelta:  ed.CategoryCountDelta(),
		ConflictCountDelta:  ed.ConflictCountDelta(),
		CoveredObjDelta:    ed.CoveredObjDelta(),
	}
}

// ptr возвращает указатель на значение.
func ptr[T any](v T) *T {
	return &v
}

// checkResultToDTO преобразует доменный результат проверки бюджета в DTO.
func checkResultToDTO(result *domain.BudgetCheckResult) BudgetCheckResponse {
	return BudgetCheckResponse{
		Fits:          result.Fits(),
		RemainingMln:  result.RemainingMln(),
		RequiredMln:   result.RequiredMln(),
		BudgetMode:    string(result.BudgetMode()),
	}
}
