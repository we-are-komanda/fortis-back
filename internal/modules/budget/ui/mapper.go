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

// checkResultToDTO преобразует доменный результат проверки бюджета в DTO.
func checkResultToDTO(result *domain.BudgetCheckResult) BudgetCheckResponse {
	return BudgetCheckResponse{
		Fits:          result.Fits(),
		RemainingMln:  result.RemainingMln(),
		RequiredMln:   result.RequiredMln(),
		BudgetMode:    string(result.BudgetMode()),
	}
}
