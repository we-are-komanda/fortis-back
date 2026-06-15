package application

import (
	"context"
	"errors"
	"fmt"
	"math"

	budgetDomain "github.com/fortis/backend/internal/modules/budget/domain"
	defenseDomain "github.com/fortis/backend/internal/modules/defense_project/domain"
)

// BudgetService — сервис для управления бюджетом и расчёта стоимости.
type BudgetService struct {
	budgetRepo  budgetDomain.BudgetConfigRepositoryInterface
	projectRepo defenseDomain.DefenseProjectRepositoryInterface
}

// NewBudgetService создаёт новый BudgetService.
func NewBudgetService(
	budgetRepo budgetDomain.BudgetConfigRepositoryInterface,
	projectRepo defenseDomain.DefenseProjectRepositoryInterface,
) *BudgetService {
	return &BudgetService{
		budgetRepo:  budgetRepo,
		projectRepo: projectRepo,
	}
}

// GetBudgetConfig возвращает конфигурацию бюджета проекта.
func (s *BudgetService) GetBudgetConfig(ctx context.Context, projectID string) (*budgetDomain.BudgetConfig, error) {
	return s.budgetRepo.FindByProjectID(ctx, projectID)
}

// UpdateBudgetConfig обновляет конфигурацию бюджета проекта.
func (s *BudgetService) UpdateBudgetConfig(ctx context.Context, projectID string, mode budgetDomain.BudgetMode, amountMln float64) error {
	existing, err := s.budgetRepo.FindByProjectID(ctx, projectID)
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
func (s *BudgetService) CalculateCost(ctx context.Context, projectID string) (*budgetDomain.CostCalculation, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("load project: %w", err)
	}

	placedObjects := project.PlacedObjects()
	assetLibrary := project.AssetLibrary()
	layers := project.Layers()

	// Индекс ассетов по ID
	assetMap := make(map[string]defenseDomain.DefenseAsset, len(assetLibrary))
	for _, a := range assetLibrary {
		assetMap[a.ID()] = a
	}

	// Индекс слоёв по ID
	layerMap := make(map[string]defenseDomain.EditableDefenseLayer, len(layers))
	for _, l := range layers {
		layerMap[l.ID()] = l
	}

	// Расчёт строк (lines) по каждому размещённому объекту
	lines := make([]budgetDomain.EstimateLine, 0, len(placedObjects))
	for _, obj := range placedObjects {
		asset, ok := assetMap[obj.AssetID()]
		if !ok {
			continue
		}

		unitPriceMln := s.resolveUnitPrice(obj, asset)
		if unitPriceMln == 0 {
			continue
		}

		quantity := obj.Quantity()
		if quantity <= 0 {
			quantity = 1
		}

		lineTotal := round2(unitPriceMln * float64(quantity))

		// Определяем название эшелона (слоя)
		echelonName := obj.LayerID()
		if layer, ok := layerMap[obj.LayerID()]; ok {
			echelonName = layer.Name()
		}

		line := budgetDomain.NewEstimateLine(
			obj.ID(),
			asset.ID(),
			asset.Name(),
			obj.LayerID(),
			echelonName,
			asset.ProtectionType(),
			asset.ProtectionType(),
			quantity,
			unitPriceMln,
			lineTotal,
		)
		lines = append(lines, line)
	}

	// Группировка по эшелонам
	byEchelon := groupByEchelon(lines)

	// Группировка по типам
	byType := groupByType(lines)

	// Общий итог
	var totalMln float64
	for _, e := range byEchelon {
		totalMln += e.EchelonTotalMln()
	}
	totalMln = round2(totalMln)

	calc := budgetDomain.NewCostCalculation(totalMln, byEchelon, byType, lines)
	return &calc, nil
}

// CheckBudget проверяет, помещается ли добавление средства в остаток бюджета.
func (s *BudgetService) CheckBudget(ctx context.Context, projectID string, input budgetDomain.BudgetCheckInput) (*budgetDomain.BudgetCheckResult, error) {
	config, err := s.budgetRepo.FindByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if config.BudgetMode() == budgetDomain.BudgetModeUnlimited {
		result := budgetDomain.NewBudgetCheckResult(true, 0, 0, budgetDomain.BudgetModeUnlimited)
		return &result, nil
	}

	// Текущая стоимость конфигурации
	calc, err := s.CalculateCost(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("calculate cost: %w", err)
	}

	remaining, err := config.RemainingBudget(calc.TotalMln())
	if err != nil {
		return nil, err
	}

	// Стоимость добавляемого средства
	requiredMln, err := s.calculateAdditionCost(ctx, projectID, input)
	if err != nil {
		return nil, fmt.Errorf("calculate addition cost: %w", err)
	}

	fits := requiredMln <= remaining
	result := budgetDomain.NewBudgetCheckResult(fits, remaining, requiredMln, config.BudgetMode())
	return &result, nil
}

// resolveUnitPrice определяет цену за единицу: сначала кастомная, затем из библиотеки.
func (s *BudgetService) resolveUnitPrice(obj defenseDomain.PlacedDefenseObject, asset defenseDomain.DefenseAsset) float64 {
	if obj.CustomPricePerUnitMln() != nil {
		return *obj.CustomPricePerUnitMln()
	}
	if asset.PricePerUnitMln() != nil {
		return *asset.PricePerUnitMln()
	}
	return 0
}

// calculateAdditionCost вычисляет стоимость добавления средства в бюджет.
func (s *BudgetService) calculateAdditionCost(ctx context.Context, projectID string, input budgetDomain.BudgetCheckInput) (float64, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
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

// groupByEchelon группирует строки расчёта по эшелонам.
func groupByEchelon(lines []budgetDomain.EstimateLine) []budgetDomain.EchelonEstimate {
	groups := make(map[string][]budgetDomain.EstimateLine)
	keys := make([]string, 0)
	for _, line := range lines {
		key := line.EchelonID()
		if _, ok := groups[key]; !ok {
			keys = append(keys, key)
		}
		groups[key] = append(groups[key], line)
	}

	result := make([]budgetDomain.EchelonEstimate, 0, len(groups))
	for _, key := range keys {
		group := groups[key]
		var total float64
		for _, line := range group {
			total += line.LineTotalMln()
		}
		echelon := budgetDomain.NewEchelonEstimate(
			group[0].EchelonID(),
			group[0].EchelonName(),
			group,
			round2(total),
		)
		result = append(result, echelon)
	}
	return result
}

// groupByType группирует строки расчёта по типам защиты.
func groupByType(lines []budgetDomain.EstimateLine) []budgetDomain.TypeEstimate {
	groups := make(map[string][]budgetDomain.EstimateLine)
	keys := make([]string, 0)
	for _, line := range lines {
		key := line.TypeID()
		if _, ok := groups[key]; !ok {
			keys = append(keys, key)
		}
		groups[key] = append(groups[key], line)
	}

	result := make([]budgetDomain.TypeEstimate, 0, len(groups))
	for _, key := range keys {
		group := groups[key]
		var total float64
		for _, line := range group {
			total += line.LineTotalMln()
		}
		te := budgetDomain.NewTypeEstimate(
			group[0].TypeID(),
			group[0].TypeName(),
			group,
			round2(total),
		)
		result = append(result, te)
	}
	return result
}

// round2 округляет float64 до 2 знаков после запятой.
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// CompareConfigs сравнивает две конфигурации проектов.
func (s *BudgetService) CompareConfigs(ctx context.Context, projectID1, projectID2 string) (*budgetDomain.ConfigComparison, error) {
	if projectID1 == "" || projectID2 == "" {
		return nil, budgetDomain.ErrBothIDsRequired
	}

	projectA, err := s.projectRepo.FindByID(ctx, projectID1)
	if err != nil {
		return nil, fmt.Errorf("load project A: %w", err)
	}

	projectB, err := s.projectRepo.FindByID(ctx, projectID2)
	if err != nil {
		return nil, fmt.Errorf("load project B: %w", err)
	}

	calcA, err := s.CalculateCost(ctx, projectID1)
	if err != nil {
		return nil, fmt.Errorf("calculate cost for A: %w", err)
	}

	calcB, err := s.CalculateCost(ctx, projectID2)
	if err != nil {
		return nil, fmt.Errorf("calculate cost for B: %w", err)
	}

	profileA := buildStructuralProfile(projectA, calcA)
	profileB := buildStructuralProfile(projectB, calcB)

	snapshotA := budgetDomain.NewConfigSnapshot(projectID1, projectA.ProjectName(), profileA, *calcA)
	snapshotB := budgetDomain.NewConfigSnapshot(projectID2, projectB.ProjectName(), profileB, *calcB)

	diff := computeDiff(profileA, profileB, *calcA, *calcB)

	comparison := budgetDomain.NewConfigComparison(snapshotA, snapshotB, diff)
	return &comparison, nil
}

// buildStructuralProfile строит структурный профиль из DefenseProject.
func buildStructuralProfile(project *defenseDomain.DefenseProject, calc *budgetDomain.CostCalculation) budgetDomain.StructuralProfile {
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
		layerID      string
		objectCount  int
		unitCount    int
		categories   map[string]struct{}
		conflictCount int
		coveredCount int
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
