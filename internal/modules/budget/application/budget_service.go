package application

import (
	"context"
	"fmt"
	defenseApp "github.com/fortis/backend/internal/modules/defense_project/application"
	"math"

	budgetDomain "github.com/fortis/backend/internal/modules/budget/domain"
	defenseDomain "github.com/fortis/backend/internal/modules/defense_project/domain"
)

// BudgetService — сервис для управления бюджетом и расчёта стоимости.
type BudgetService struct {
	budgetRepo  budgetDomain.BudgetConfigRepositoryInterface
	projects    *defenseApp.DefenseProjectService
	projections budgetDomain.CostProjectionRepositoryInterface
}

// NewBudgetService создаёт новый BudgetService.
func NewBudgetService(
	budgetRepo budgetDomain.BudgetConfigRepositoryInterface,
	projects *defenseApp.DefenseProjectService,
	projections budgetDomain.CostProjectionRepositoryInterface,
) *BudgetService {
	return &BudgetService{
		budgetRepo:  budgetRepo,
		projects:    projects,
		projections: projections,
	}
}

// GetBudgetConfig возвращает конфигурацию бюджета проекта.
func (s *BudgetService) GetBudgetConfig(ctx context.Context, userID string, projectID string) (*budgetDomain.BudgetConfig, error) {
	if _, err := s.projects.GetProject(ctx, userID, projectID); err != nil {
		return nil, err
	}
	return s.budgetRepo.FindByProjectID(ctx, projectID)
}

// UpdateBudgetConfig обновляет конфигурацию бюджета проекта.
func (s *BudgetService) UpdateBudgetConfig(ctx context.Context, userID string, projectID string, mode budgetDomain.BudgetMode, amountMln float64) error {
	if _, err := s.projects.GetProject(ctx, userID, projectID); err != nil {
		return err
	}
	// Legacy callers cannot bypass the versioned transaction.
	return defenseDomain.ErrVersionRequired
}

// CalculateCost выполняет полный расчёт стоимости конфигурации проекта.
func (s *BudgetService) CalculateCost(ctx context.Context, userID string, projectID string) (*budgetDomain.CostCalculation, error) {
	projection, err := s.ProjectCost(ctx, userID, projectID, nil)
	if err != nil {
		return nil, err
	}
	return legacyCost(projection)
}

// CheckBudget проверяет, помещается ли добавление средства в остаток бюджета.
func (s *BudgetService) CheckBudget(ctx context.Context, userID string, projectID string, input budgetDomain.BudgetCheckInput) (*budgetDomain.BudgetCheckResult, error) {
	config, err := s.GetBudgetConfig(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}

	if config.BudgetMode() == budgetDomain.BudgetModeUnlimited {
		result := budgetDomain.NewBudgetCheckResult(true, 0, 0, budgetDomain.BudgetModeUnlimited)
		return &result, nil
	}

	// Текущая стоимость конфигурации
	calc, err := s.CalculateCost(ctx, userID, projectID)
	if err != nil {
		return nil, fmt.Errorf("calculate cost: %w", err)
	}

	remaining, err := config.RemainingBudget(calc.TotalMln())
	if err != nil {
		return nil, err
	}

	// Стоимость добавляемого средства
	requiredMln, err := s.calculateAdditionCost(ctx, userID, projectID, input)
	if err != nil {
		return nil, fmt.Errorf("calculate addition cost: %w", err)
	}

	fits := requiredMln <= remaining
	result := budgetDomain.NewBudgetCheckResult(fits, remaining, requiredMln, config.BudgetMode())
	return &result, nil
}

// calculateAdditionCost вычисляет стоимость добавления средства в бюджет.
func (s *BudgetService) calculateAdditionCost(ctx context.Context, userID string, projectID string, input budgetDomain.BudgetCheckInput) (float64, error) {
	project, err := s.projects.GetProject(ctx, userID, projectID)
	if err != nil {
		return 0, fmt.Errorf("load project: %w", err)
	}

	for _, a := range project.AssetLibrary() {
		if a.ID() == input.AssetID() {
			var unitPrice float64
			if a.PricePerUnitMln() != nil {
				unitPrice = *a.PricePerUnitMln()
			}
			return round2(unitPrice * float64(input.Quantity())), nil
		}
	}

	return 0, fmt.Errorf("asset %s not found in project library", input.AssetID())
}

// round2 округляет float64 до 2 знаков после запятой.
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// CompareConfigs authorizes both pinned revisions before evaluating either side.
func (s *BudgetService) CompareConfigs(ctx context.Context, actor string, a, b defenseDomain.ProjectRef) (*budgetDomain.RevisionComparison, error) {
	if a.ProjectID == "" || b.ProjectID == "" {
		return nil, budgetDomain.ErrBothIDsRequired
	}
	if a.ProjectVersion <= 0 || b.ProjectVersion <= 0 {
		return nil, defenseDomain.ErrVersionRequired
	}
	pa, err := s.projects.GetRevision(ctx, actor, a.ProjectID, &a.ProjectVersion)
	if err != nil {
		return nil, err
	}
	pb, err := s.projects.GetRevision(ctx, actor, b.ProjectID, &b.ProjectVersion)
	if err != nil {
		return nil, err
	}
	if pa.EnterpriseID() != pb.EnterpriseID() || (a.ProjectID == b.ProjectID && a.ProjectVersion == b.ProjectVersion) {
		return nil, budgetDomain.ErrIncompatibleProjects
	}
	side := func(p *defenseDomain.DefenseProject) (budgetDomain.ComparedRevision, error) {
		cost, err := s.projectCost(ctx, p)
		if err != nil {
			return budgetDomain.ComparedRevision{}, err
		}
		config, _, err := budgetFromSnapshot(p)
		if err != nil {
			return budgetDomain.ComparedRevision{}, err
		}
		return budgetDomain.NewComparedRevision(p.Name(), p.UpdatedAt(), budgetDomain.NewBudgetProjection(config, cost)), nil
	}
	ca, err := side(pa)
	if err != nil {
		return nil, err
	}
	cb, err := side(pb)
	if err != nil {
		return nil, err
	}
	return budgetDomain.NewRevisionComparison(pa.EnterpriseID(), ca, cb), nil
}

// BuildStructuralProfile строит структурный профиль из DefenseProject.
func BuildStructuralProfile(project *defenseDomain.DefenseProject, calc *budgetDomain.CostCalculation) budgetDomain.StructuralProfile {
	placedObjects := project.PlacedObjects()
	layers := project.Layers()

	// Строим индекс слоёв
	layerMap := make(map[string]defenseDomain.EditableDefenseLayer)
	for _, l := range layers {
		layerMap[l.ID()] = l
	}

	// Строим индекс категорий ассетов
	assetMap := make(map[string]defenseDomain.DefenseAsset)
	for _, a := range project.AssetLibrary() {
		assetMap[a.ID()] = a
	}

	// Считаем метрики по слоям
	type echelonStats struct {
		layerID       string
		objectCount   int
		unitCount     int
		categories    map[string]struct{}
		conflictCount int
		coveredCount  int
	}

	echelonStatsMap := make(map[string]*echelonStats)
	echelonOrder := make([]string, 0)

	for _, obj := range placedObjects {
		layerID := obj.LayerID()
		stats, exists := echelonStatsMap[layerID]
		if !exists {
			stats = &echelonStats{
				layerID:    layerID,
				categories: make(map[string]struct{}),
			}
			echelonStatsMap[layerID] = stats
			echelonOrder = append(echelonOrder, layerID)
		}

		stats.objectCount++
		stats.unitCount += obj.Quantity()
		if obj.HasGeometryConflict() || obj.HasCoverageConflict() || obj.HasTerrainConflict() {
			stats.conflictCount++
		}

		// Определяем категорию ассета
		if asset, ok := assetMap[obj.AssetID()]; ok {
			stats.categories[string(asset.Category())] = struct{}{}
		}

		// covered: если у объекта нет конфликтов покрытия — считаем покрытым
		if !obj.HasCoverageConflict() {
			stats.coveredCount++
		}
	}

	// Собираем глобальные метрики
	var totalObjects, totalUnits, totalConflicts, totalCovered int
	allCategories := make(map[string]struct{})
	byEchelon := make([]budgetDomain.StructuralEchelonProfile, 0, len(echelonOrder))

	for _, layerID := range echelonOrder {
		stats := echelonStatsMap[layerID]
		totalObjects += stats.objectCount
		totalUnits += stats.unitCount
		totalConflicts += stats.conflictCount
		totalCovered += stats.coveredCount

		for c := range stats.categories {
			allCategories[c] = struct{}{}
		}

		layer := layerMap[layerID]
		ep, _ := budgetDomain.NewStructuralEchelonProfile(
			layerID,
			layer.Code(),
			layer.Name(),
			stats.objectCount,
			stats.unitCount,
			len(stats.categories),
			stats.conflictCount,
			stats.coveredCount,
		)
		byEchelon = append(byEchelon, ep)
	}

	totalEchelons := len(echelonOrder)
	totalCategories := len(allCategories)
	totalMln := 0.0
	if calc != nil {
		totalMln = calc.TotalMln()
	}

	return budgetDomain.NewStructuralProfile(
		totalObjects,
		totalUnits,
		totalEchelons,
		totalCategories,
		totalConflicts,
		totalCovered,
		totalMln,
		byEchelon,
	)
}
