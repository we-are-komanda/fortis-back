package ui

import (
	"encoding/json"
	"time"

	"github.com/fortis/backend/internal/modules/budget/domain"
)

func costProjectionToDTO(p *domain.CostProjection) CostProjectionDTO {
	i := p.Identity()
	dto := CostProjectionDTO{Identity: CalculationIdentityDTO{i.ProjectID(), i.ProjectVersion(), i.CalculationVersion(), i.InputDataVersions(), i.SnapshotDigest()}, Currency: "RUB", Lines: []FinancialLineDTO{}, ByLayer: []CostGroupDTO{}, ByType: []CostGroupDTO{}, KnownSubtotalMinor: p.KnownSubtotalMinor(), TotalMinor: p.TotalMinor(), IsComplete: p.IsComplete(), UnknownPriceObjectIDs: p.UnknownPriceObjectIDs(), Warnings: []CostIssueDTO{}}
	for _, l := range p.Lines() {
		var provenance *CostDataProvenanceDTO
		if l.Provenance() != "" {
			var value CostDataProvenanceDTO
			if err := json.Unmarshal([]byte(l.Provenance()), &value); err == nil {
				provenance = &value
			}
		}
		dto.Lines = append(dto.Lines, FinancialLineDTO{l.ObjectID(), l.AssetID(), l.LayerID(), l.Category(), l.Name(), l.Quantity(), l.UnitPriceMinor(), l.LineTotalMinor(), l.PriceSource(), provenance})
	}
	groups := func(source []domain.CostGroup) []CostGroupDTO {
		result := []CostGroupDTO{}
		for _, g := range source {
			result = append(result, CostGroupDTO{g.ID(), g.Name(), g.ObjectCount(), g.UnitCount(), g.KnownSubtotalMinor(), g.TotalMinor()})
		}
		return result
	}
	dto.ByLayer, dto.ByType = groups(p.ByLayer()), groups(p.ByType())
	for _, w := range p.Warnings() {
		dto.Warnings = append(dto.Warnings, CostIssueDTO{w.Code(), w.Severity(), w.ObjectIDs(), w.Message()})
	}
	return dto
}

func comparisonToDTO(comp *domain.RevisionComparison) ConfigComparisonResponse {
	side := func(s domain.ComparedRevision) ComparedRevisionDTO {
		cost := costProjectionToDTO(s.Cost())
		return ComparedRevisionDTO{Identity: cost.Identity, Name: s.Name(), UpdatedAt: s.UpdatedAt().UTC().Format(time.RFC3339Nano), Cost: cost, Budget: budgetProjectionToDTO(s.Budget())}
	}
	warnings := []CostIssueDTO{}
	for _, w := range comp.Warnings() {
		warnings = append(warnings, CostIssueDTO{w.Code(), w.Severity(), w.ObjectIDs(), w.Message()})
	}
	return ConfigComparisonResponse{EnterpriseID: comp.EnterpriseID(), A: side(comp.A()), B: side(comp.B()), DifferenceMinor: comp.DifferenceMinor(), DifferencePercent: comp.DifferencePercent(), Warnings: warnings}
}

func budgetProjectionToDTO(p *domain.BudgetProjection) BudgetProjectionDTO {
	i := p.Identity()
	return BudgetProjectionDTO{Identity: CalculationIdentityDTO{ProjectID: i.ProjectID(), ProjectVersion: i.ProjectVersion(), CalculationVersion: i.CalculationVersion(), InputDataVersions: i.InputDataVersions(), SnapshotDigest: i.SnapshotDigest()}, Config: BudgetSettingsDTO{BudgetMode: string(p.Config().BudgetMode()), BudgetLimitMinor: p.Config().BudgetLimitMinor()}, SpentMinor: p.SpentMinor(), RemainingMinor: p.RemainingMinor(), Status: p.Status()}
}

// CostProjectionToDTO and BudgetProjectionToDTO share the canonical financial wire format with issued reports.
func CostProjectionToDTO(p *domain.CostProjection) CostProjectionDTO { return costProjectionToDTO(p) }
func BudgetProjectionToDTO(p *domain.BudgetProjection) BudgetProjectionDTO {
	return budgetProjectionToDTO(p)
}
