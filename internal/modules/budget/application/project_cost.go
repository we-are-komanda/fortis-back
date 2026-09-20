package application

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/fortis/backend/internal/modules/budget/document"
	"github.com/fortis/backend/internal/modules/budget/domain"
)

func (s *BudgetService) ProjectCost(ctx context.Context, actorID, projectID string, version *int) (*domain.CostProjection, error) {
	project, err := s.projects.GetRevision(ctx, actorID, projectID, version)
	if err != nil {
		return nil, err
	}
	calculator := project.CostCalculationVersion()
	if calculator != domain.CostCalculationV1 {
		return nil, domain.ErrUnsupportedCalculationVersion
	}
	if saved, err := s.projections.FindCostProjection(ctx, project.ProjectID(), project.Version(), calculator); err == nil {
		return saved, nil
	} else if !errors.Is(err, domain.ErrCostProjectionNotFound) {
		return nil, err
	}
	var snapshot struct {
		InputDataVersions map[string]string `json:"inputDataVersions"`
	}
	if project.Snapshot() != "" {
		if err := json.Unmarshal([]byte(project.Snapshot()), &snapshot); err != nil {
			return nil, err
		}
	}
	identity := domain.NewCalculationIdentity(project.ProjectID(), project.Version(), calculator, project.SnapshotDigest(), snapshot.InputDataVersions)
	projection, err := document.Calculate(project.Document(), identity)
	if err != nil {
		return nil, err
	}
	return s.projections.SaveCostProjection(ctx, projection)
}

// Legacy fields are derived from the exact projection only. Callers unable to
// represent null totals must surface incompleteness instead of showing zero.
func legacyCost(projection *domain.CostProjection) (*domain.CostCalculation, error) {
	if !projection.IsComplete() {
		return nil, domain.ErrIncompleteCost
	}
	lineGroups := map[string][]domain.EstimateLine{}
	typeGroups := map[string][]domain.EstimateLine{}
	layerNames := map[string]string{}
	for _, g := range projection.ByLayer() {
		layerNames[g.ID()] = g.Name()
	}
	lines := []domain.EstimateLine{}
	for _, l := range projection.Lines() {
		line := domain.NewEstimateLine(l.ObjectID(), l.AssetID(), l.Name(), l.LayerID(), layerNames[l.LayerID()], l.Category(), l.Category(), l.Quantity(), minorMln(*l.UnitPriceMinor()), minorMln(*l.LineTotalMinor()))
		lines = append(lines, line)
		lineGroups[l.LayerID()] = append(lineGroups[l.LayerID()], line)
		typeGroups[l.Category()] = append(typeGroups[l.Category()], line)
	}
	layers := []domain.EchelonEstimate{}
	for _, g := range projection.ByLayer() {
		layers = append(layers, domain.NewEchelonEstimate(g.ID(), g.Name(), lineGroups[g.ID()], minorMln(*g.TotalMinor())))
	}
	types := []domain.TypeEstimate{}
	for _, g := range projection.ByType() {
		types = append(types, domain.NewTypeEstimate(g.ID(), g.Name(), typeGroups[g.ID()], minorMln(*g.TotalMinor())))
	}
	calculation := domain.NewCostCalculation(minorMln(*projection.TotalMinor()), layers, types, lines)
	return &calculation, nil
}
func minorMln(minor string) float64 {
	n, _ := new(big.Rat).SetString(minor)
	n.Quo(n, big.NewRat(100_000_000, 1))
	value, _ := n.Float64()
	return value
}
