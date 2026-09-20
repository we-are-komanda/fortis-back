package application

import (
	"context"
	"errors"
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
	existing, err := s.GetBudgetConfig(ctx, userID, projectID)
	if err != nil {
		if errors.Is(err, budgetDomain.ErrBudgetConfigNotFound) {
			config, createErr := budgetDomain.NewBudgetConfig(projectID, mode, amountMln)
			if createErr != nil {
				return createErr
			}
			return s.budgetRepo.Save(ctx, config)
		}
		return err
	}

	if updateErr := existing.SetBudget(mode, amountMln); updateErr != nil {
		return updateErr
	}
	return s.budgetRepo.Save(ctx, existing)
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

// CompareConfigs сравнивает две конфигурации проектов.
func (s *BudgetService) CompareConfigs(ctx context.Context, userID string, projectID1, projectID2 string) (*budgetDomain.ConfigComparison, error) {
	if projectID1 == "" || projectID2 == "" {
		return nil, budgetDomain.ErrBothIDsRequired
	}

	projectA, err := s.projects.GetProject(ctx, userID, projectID1)
	if err != nil {
		return nil, fmt.Errorf("load project A: %w", err)
	}

	projectB, err := s.projects.GetProject(ctx, userID, projectID2)
	if err != nil {
		return nil, fmt.Errorf("load project B: %w", err)
	}

	calcA, err := s.CalculateCost(ctx, userID, projectID1)
	if err != nil {
		return nil, fmt.Errorf("calculate cost for A: %w", err)
	}

	calcB, err := s.CalculateCost(ctx, userID, projectID2)
	if err != nil {
		return nil, fmt.Errorf("calculate cost for B: %w", err)
	}

	profileA := BuildStructuralProfile(projectA, calcA)
	profileB := BuildStructuralProfile(projectB, calcB)

	snapshotA := budgetDomain.NewConfigSnapshot(projectID1, projectA.ProjectName(), profileA, *calcA)
	snapshotB := budgetDomain.NewConfigSnapshot(projectID2, projectB.ProjectName(), profileB, *calcB)

	diff := computeDiff(profileA, profileB, *calcA, *calcB)

	comparison := budgetDomain.NewConfigComparison(snapshotA, snapshotB, diff)
	return &comparison, nil
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

// computeDiff вычисляет разницу между двумя структурными профилями и расчётами стоимости.
func computeDiff(profileA, profileB budgetDomain.StructuralProfile, calcA, calcB budgetDomain.CostCalculation) budgetDomain.ConfigDiff {
	// Разница по эшелонам — объединяем по LayerID
	echelonMapA := make(map[string]budgetDomain.StructuralEchelonProfile)
	for _, ep := range profileA.ByEchelon() {
		echelonMapA[ep.LayerID()] = ep
	}

	echelonMapB := make(map[string]budgetDomain.StructuralEchelonProfile)
	for _, ep := range profileB.ByEchelon() {
		echelonMapB[ep.LayerID()] = ep
	}

	// Собираем все LayerID из обоих профилей
	allLayerIDs := make(map[string]struct{})
	for _, ep := range profileA.ByEchelon() {
		allLayerIDs[ep.LayerID()] = struct{}{}
	}
	for _, ep := range profileB.ByEchelon() {
		allLayerIDs[ep.LayerID()] = struct{}{}
	}

	byEchelonDiff := make([]budgetDomain.EchelonDiff, 0, len(allLayerIDs))
	for layerID := range allLayerIDs {
		epA := echelonMapA[layerID]
		epB := echelonMapB[layerID]

		layerCode := epA.LayerCode()
		if layerCode == "" {
			layerCode = epB.LayerCode()
		}
		layerName := epA.LayerName()
		if layerName == "" {
			layerName = epB.LayerName()
		}

		diff := budgetDomain.NewEchelonDiff(
			layerID,
			layerCode,
			layerName,
			epB.ObjectCount()-epA.ObjectCount(),
			epB.UnitCount()-epA.UnitCount(),
			epB.CategoryCount()-epA.CategoryCount(),
			epB.ConflictCount()-epA.ConflictCount(),
			epB.CoveredObjCount()-epA.CoveredObjCount(),
		)
		byEchelonDiff = append(byEchelonDiff, diff)
	}

	// Если нет эшелонов — пустой слайс, но не nil
	if byEchelonDiff == nil {
		byEchelonDiff = []budgetDomain.EchelonDiff{}
	}

	return budgetDomain.NewConfigDiff(
		profileB.ObjectCount()-profileA.ObjectCount(),
		profileB.UnitCount()-profileA.UnitCount(),
		profileB.EchelonCount()-profileA.EchelonCount(),
		profileB.CategoryCount()-profileA.CategoryCount(),
		profileB.ConflictCount()-profileA.ConflictCount(),
		profileB.CoveredObjCount()-profileA.CoveredObjCount(),
		round2(calcB.TotalMln()-calcA.TotalMln()),
		byEchelonDiff,
	)
}
