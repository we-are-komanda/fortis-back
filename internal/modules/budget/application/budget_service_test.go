package application

import (
	"context"
	"errors"
	"testing"
	"time"

	budgetDomain "github.com/fortis/backend/internal/modules/budget/domain"
	defenseDomain "github.com/fortis/backend/internal/modules/defense_project/domain"
)

// ---- Mocks ----

type mockBudgetRepo struct {
	configs map[string]*budgetDomain.BudgetConfig
	err     error
}

func (m *mockBudgetRepo) FindByProjectID(_ context.Context, projectID string) (*budgetDomain.BudgetConfig, error) {
	if m.err != nil {
		return nil, m.err
	}
	c, ok := m.configs[projectID]
	if !ok {
		return nil, budgetDomain.ErrBudgetConfigNotFound
	}
	return c, nil
}

func (m *mockBudgetRepo) Save(_ context.Context, config *budgetDomain.BudgetConfig) error {
	if m.err != nil {
		return m.err
	}
	m.configs[config.ProjectID()] = config
	return nil
}

func (m *mockBudgetRepo) Delete(_ context.Context, projectID string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.configs, projectID)
	return nil
}

type mockProjectRepo struct {
	projects map[string]*defenseDomain.DefenseProject
	err      error
}

func (m *mockProjectRepo) FindByID(_ context.Context, id string) (*defenseDomain.DefenseProject, error) {
	if m.err != nil {
		return nil, m.err
	}
	p, ok := m.projects[id]
	if !ok {
		return nil, defenseDomain.ErrProjectNotFound
	}
	return p, nil
}

func (m *mockProjectRepo) Save(_ context.Context, _ *defenseDomain.DefenseProject) error {
	return m.err
}
func (m *mockProjectRepo) FindAll(_ context.Context, _, _ int) ([]*defenseDomain.DefenseProject, int64, error) {
	return nil, 0, m.err
}
func (m *mockProjectRepo) FindAllByEnterprise(_ context.Context, _ string, _, _ int) ([]*defenseDomain.DefenseProject, int64, error) {
	return nil, 0, m.err
}
func (m *mockProjectRepo) Delete(_ context.Context, _ string) error {
	return m.err
}

// ---- Test helpers ----

func makeTestProject(t *testing.T) *defenseDomain.DefenseProject {
	t.Helper()

	now := time.Now().UTC()
	price1 := 20.0
	price2 := 50.0

	nilString := func(s string) *string {
		return &s
	}

	project, err := defenseDomain.NewDefenseProject(
		"proj-1",
		"Test Config",
		"ent-1",
		"Test Project",
		defenseDomain.NewProtectedObject("obj-1", "Object Alpha",
			defenseDomain.NewCoordinates(55.75, 37.62)),
		[]defenseDomain.EditableDefenseLayer{
			defenseDomain.NewEditableDefenseLayer(
				"echelon-1", "Обнаружение", "detection", nil,
				1, nil, nil,
				defenseDomain.LayerGeometryCircle,
				defenseDomain.NewCircleGeometry(defenseDomain.NewCoordinates(55.75, 37.62), 1000),
				nil, nil, true, true, false,
			),
			defenseDomain.NewEditableDefenseLayer(
				"echelon-2", "Подавление", "suppression", nil,
				2, nil, nil,
				defenseDomain.LayerGeometryCircle,
				defenseDomain.NewCircleGeometry(defenseDomain.NewCoordinates(55.75, 37.62), 2000),
				nil, nil, true, true, false,
			),
		},
		[]defenseDomain.DefenseAsset{
			makeTestAsset("asset-1", "Mobile Radar", "radar", &price1),
			makeTestAsset("asset-2", "Jammer System", "jamming", &price2),
		},
		[]defenseDomain.PlacedDefenseObject{
			makeTestPlacedObject("placed-1", "asset-1", "echelon-1", 3, nil, now),
			makeTestPlacedObject("placed-2", "asset-2", "echelon-2", 2, nil, now),
			makeTestPlacedObject("placed-3", "asset-1", "echelon-1", 1, nil, now),
		},
		nil,
		nil,
		nil,
		defenseDomain.DefenseProjectModeView,
		defenseDomain.DefenseProjectSourceCustom,
		nilString("preset-1"),
		now,
	)
	if err != nil {
		t.Fatalf("NewDefenseProject() unexpected error: %v", err)
	}
	return project
}

func makeTestAsset(id, name, protectionType string, price *float64) defenseDomain.DefenseAsset {
	var nilStrPtr *string

	return defenseDomain.NewDefenseAsset(
		id, name, nilStrPtr, nilStrPtr,
		defenseDomain.DefenseAssetCategoryDetection,
		[]defenseDomain.DefenseAssetRole{defenseDomain.DefenseAssetRoleDetect},
		price, "RUB", "unit",
		[]defenseDomain.LayerGeometryType{defenseDomain.LayerGeometryCircle},
		nil, nil, nil,
		protectionType,
		nil, nil,
		defenseDomain.DefenseAssetCoverageCircle, nil, nil,
		defenseDomain.DefenseAssetDeploymentStatic,
		defenseDomain.DefenseAssetPlacementMapObject,
		nilStrPtr, nilStrPtr,
		nil, nil, nil, nil, nilStrPtr, nilStrPtr, nil,
	)
}

func makeTestPlacedObject(id, assetID, layerID string, quantity int, customPrice *float64, now time.Time) defenseDomain.PlacedDefenseObject {
	var nilStrPtr *string
	var nilFloatPtr *float64

	return defenseDomain.NewPlacedDefenseObject(
		id, assetID, layerID, nilStrPtr,
		defenseDomain.NewCoordinates(55.75, 37.62),
		nilFloatPtr, nilFloatPtr, quantity,
		defenseDomain.PlacedObjectStatusActive,
		customPrice, nilFloatPtr, nilFloatPtr,
		false, false, false, nilStrPtr,
		now, now,
	)
}

// ---- Tests ----

func TestBudgetService_GetBudgetConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	budgetConfig, _ := budgetDomain.NewBudgetConfig("proj-1", budgetDomain.BudgetModeLimited, 5000)

	svc := NewBudgetService(
		&mockBudgetRepo{configs: map[string]*budgetDomain.BudgetConfig{"proj-1": budgetConfig}},
		&mockProjectRepo{},
	)

	config, err := svc.GetBudgetConfig(ctx, "proj-1")
	if err != nil {
		t.Fatalf("GetBudgetConfig() unexpected error: %v", err)
	}
	if config.BudgetMode() != budgetDomain.BudgetModeLimited {
		t.Errorf("BudgetMode() = %s, want limited", config.BudgetMode())
	}
}

func TestBudgetService_GetBudgetConfig_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := NewBudgetService(
		&mockBudgetRepo{configs: map[string]*budgetDomain.BudgetConfig{}},
		&mockProjectRepo{},
	)

	_, err := svc.GetBudgetConfig(ctx, "nonexistent")
	if !errors.Is(err, budgetDomain.ErrBudgetConfigNotFound) {
		t.Errorf("expected ErrBudgetConfigNotFound, got %v", err)
	}
}

func TestBudgetService_UpdateBudgetConfig_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := &mockBudgetRepo{configs: map[string]*budgetDomain.BudgetConfig{}}
	svc := NewBudgetService(repo, &mockProjectRepo{})

	err := svc.UpdateBudgetConfig(ctx, "proj-1", budgetDomain.BudgetModeLimited, 10000)
	if err != nil {
		t.Fatalf("UpdateBudgetConfig() unexpected error: %v", err)
	}

	config, err := repo.FindByProjectID(ctx, "proj-1")
	if err != nil {
		t.Fatalf("FindByProjectID() unexpected error: %v", err)
	}
	if config.BudgetAmountMln() != 10000 {
		t.Errorf("BudgetAmountMln() = %f, want 10000", config.BudgetAmountMln())
	}
}

func TestBudgetService_UpdateBudgetConfig_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	existing, _ := budgetDomain.NewBudgetConfig("proj-1", budgetDomain.BudgetModeLimited, 5000)
	repo := &mockBudgetRepo{configs: map[string]*budgetDomain.BudgetConfig{"proj-1": existing}}
	svc := NewBudgetService(repo, &mockProjectRepo{})

	err := svc.UpdateBudgetConfig(ctx, "proj-1", budgetDomain.BudgetModeLimited, 15000)
	if err != nil {
		t.Fatalf("UpdateBudgetConfig() unexpected error: %v", err)
	}

	config, _ := repo.FindByProjectID(ctx, "proj-1")
	if config.BudgetAmountMln() != 15000 {
		t.Errorf("BudgetAmountMln() = %f, want 15000", config.BudgetAmountMln())
	}
}

func TestBudgetService_CalculateCost(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	project := makeTestProject(t)

	svc := NewBudgetService(
		&mockBudgetRepo{configs: map[string]*budgetDomain.BudgetConfig{}},
		&mockProjectRepo{projects: map[string]*defenseDomain.DefenseProject{"proj-1": project}},
	)

	calc, err := svc.CalculateCost(ctx, "proj-1")
	if err != nil {
		t.Fatalf("CalculateCost() unexpected error: %v", err)
	}

	// placed-1: asset-1 (20) x 3 = 60 (echelon-1)
	// placed-2: asset-2 (50) x 2 = 100 (echelon-2)
	// placed-3: asset-1 (20) x 1 = 20 (echelon-1)
	// Total: 60 + 100 + 20 = 180

	if calc.TotalMln() != 180 {
		t.Errorf("TotalMln() = %f, want 180", calc.TotalMln())
	}

	if len(calc.ByEchelon()) != 2 {
		t.Fatalf("ByEchelon() length = %d, want 2", len(calc.ByEchelon()))
	}

	if len(calc.ByObject()) != 3 {
		t.Errorf("ByObject() length = %d, want 3", len(calc.ByObject()))
	}

	// Check echelon-1 total: 60 + 20 = 80
	foundEchelon1 := false
	for _, e := range calc.ByEchelon() {
		if e.EchelonID() == "echelon-1" {
			foundEchelon1 = true
			if e.EchelonTotalMln() != 80 {
				t.Errorf("echelon-1 total = %f, want 80", e.EchelonTotalMln())
			}
		}
	}
	if !foundEchelon1 {
		t.Error("echelon-1 not found in byEchelon")
	}

	// Check type groups
	if len(calc.ByType()) != 2 {
		t.Fatalf("ByType() length = %d, want 2", len(calc.ByType()))
	}
}

func TestBudgetService_CheckBudget_Limited_Fits(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	project := makeTestProject(t)
	budgetConfig, _ := budgetDomain.NewBudgetConfig("proj-1", budgetDomain.BudgetModeLimited, 1000)

	svc := NewBudgetService(
		&mockBudgetRepo{configs: map[string]*budgetDomain.BudgetConfig{"proj-1": budgetConfig}},
		&mockProjectRepo{projects: map[string]*defenseDomain.DefenseProject{"proj-1": project}},
	)

	input := budgetDomain.NewBudgetCheckInput("asset-1", 1, "echelon-1")
	result, err := svc.CheckBudget(ctx, "proj-1", input)
	if err != nil {
		t.Fatalf("CheckBudget() unexpected error: %v", err)
	}

	// Total cost = 180, budget = 1000, remaining = 820, adding 1x20 = 20, fits
	if !result.Fits() {
		t.Error("Fits() = false, want true")
	}
	if result.RemainingMln() != 820 {
		t.Errorf("RemainingMln() = %f, want 820", result.RemainingMln())
	}
	if result.RequiredMln() != 20 {
		t.Errorf("RequiredMln() = %f, want 20", result.RequiredMln())
	}
	if result.BudgetMode() != budgetDomain.BudgetModeLimited {
		t.Errorf("BudgetMode() = %s, want limited", result.BudgetMode())
	}
}

func TestBudgetService_CheckBudget_Limited_DoesNotFit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	project := makeTestProject(t)
	budgetConfig, _ := budgetDomain.NewBudgetConfig("proj-1", budgetDomain.BudgetModeLimited, 150)

	svc := NewBudgetService(
		&mockBudgetRepo{configs: map[string]*budgetDomain.BudgetConfig{"proj-1": budgetConfig}},
		&mockProjectRepo{projects: map[string]*defenseDomain.DefenseProject{"proj-1": project}},
	)

	input := budgetDomain.NewBudgetCheckInput("asset-1", 1, "echelon-1")
	result, err := svc.CheckBudget(ctx, "proj-1", input)
	if err != nil {
		t.Fatalf("CheckBudget() unexpected error: %v", err)
	}

	if result.Fits() {
		t.Error("Fits() = true, want false (total 180 > budget 150)")
	}
}

func TestBudgetService_CheckBudget_Unlimited(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	budgetConfig, _ := budgetDomain.NewBudgetConfig("proj-1", budgetDomain.BudgetModeUnlimited, 0)

	svc := NewBudgetService(
		&mockBudgetRepo{configs: map[string]*budgetDomain.BudgetConfig{"proj-1": budgetConfig}},
		&mockProjectRepo{},
	)

	input := budgetDomain.NewBudgetCheckInput("asset-1", 10, "echelon-1")
	result, err := svc.CheckBudget(ctx, "proj-1", input)
	if err != nil {
		t.Fatalf("CheckBudget() unexpected error: %v", err)
	}

	if !result.Fits() {
		t.Error("Fits() = false, want true (unlimited)")
	}
	if result.BudgetMode() != budgetDomain.BudgetModeUnlimited {
		t.Errorf("BudgetMode() = %s, want unlimited", result.BudgetMode())
	}
}

func TestBudgetService_CompareConfigs_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	projectA := makeTestProject(t)
	projectB := makeTestProjectWithDifferentData(t)

	svc := NewBudgetService(
		&mockBudgetRepo{configs: map[string]*budgetDomain.BudgetConfig{}},
		&mockProjectRepo{
			projects: map[string]*defenseDomain.DefenseProject{
				"proj-1": projectA,
				"proj-2": projectB,
			},
		},
	)

	comp, err := svc.CompareConfigs(ctx, "proj-1", "proj-2")
	if err != nil {
		t.Fatalf("CompareConfigs() unexpected error: %v", err)
	}

	// Check snapshots
	if comp.ProjectA().ProjectID() != "proj-1" {
		t.Errorf("ProjectA().ProjectID() = %s, want proj-1", comp.ProjectA().ProjectID())
	}
	if comp.ProjectB().ProjectID() != "proj-2" {
		t.Errorf("ProjectB().ProjectID() = %s, want proj-2", comp.ProjectB().ProjectID())
	}

	// Check structural profiles exist
	if comp.ProjectA().StructuralProfile().ObjectCount() == 0 && comp.ProjectB().StructuralProfile().ObjectCount() == 0 {
		t.Error("both structural profiles are empty")
	}

	// Check diff exists
	diff := comp.Diff()
	if diff.ObjectCountDelta() == 0 && diff.UnitCountDelta() == 0 {
		t.Log("diff is zero — projects may be identical in structure")
	}

	// Check cost calculation exists
	if comp.ProjectA().CostCalculation().TotalMln() <= 0 {
		t.Errorf("ProjectA cost should be > 0, got %f", comp.ProjectA().CostCalculation().TotalMln())
	}
	if comp.ProjectB().CostCalculation().TotalMln() <= 0 {
		t.Errorf("ProjectB cost should be > 0, got %f", comp.ProjectB().CostCalculation().TotalMln())
	}
}

func TestBudgetService_CompareConfigs_EmptyIDs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := NewBudgetService(&mockBudgetRepo{}, &mockProjectRepo{})

	_, err := svc.CompareConfigs(ctx, "", "")
	if !errors.Is(err, budgetDomain.ErrBothIDsRequired) {
		t.Errorf("error = %v, want ErrBothIDsRequired", err)
	}

	_, err = svc.CompareConfigs(ctx, "proj-1", "")
	if !errors.Is(err, budgetDomain.ErrBothIDsRequired) {
		t.Errorf("error = %v, want ErrBothIDsRequired", err)
	}

	_, err = svc.CompareConfigs(ctx, "", "proj-2")
	if !errors.Is(err, budgetDomain.ErrBothIDsRequired) {
		t.Errorf("error = %v, want ErrBothIDsRequired", err)
	}
}

func TestBudgetService_CompareConfigs_ProjectNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	projectA := makeTestProject(t)

	svc := NewBudgetService(
		&mockBudgetRepo{configs: map[string]*budgetDomain.BudgetConfig{}},
		&mockProjectRepo{
			projects: map[string]*defenseDomain.DefenseProject{
				"proj-1": projectA,
			},
		},
	)

	_, err := svc.CompareConfigs(ctx, "proj-1", "nonexistent")
	if err == nil {
		t.Fatal("CompareConfigs() expected error, got nil")
	}
}

func TestBudgetService_CompareConfigs_SameProject(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	project := makeTestProject(t)

	svc := NewBudgetService(
		&mockBudgetRepo{configs: map[string]*budgetDomain.BudgetConfig{}},
		&mockProjectRepo{
			projects: map[string]*defenseDomain.DefenseProject{
				"proj-1": project,
			},
		},
	)

	comp, err := svc.CompareConfigs(ctx, "proj-1", "proj-1")
	if err != nil {
		t.Fatalf("CompareConfigs() unexpected error: %v", err)
	}

	// Comparing same project should produce zero diff
	diff := comp.Diff()
	if diff.ObjectCountDelta() != 0 {
		t.Errorf("ObjectCountDelta() = %d, want 0", diff.ObjectCountDelta())
	}
	if diff.UnitCountDelta() != 0 {
		t.Errorf("UnitCountDelta() = %d, want 0", diff.UnitCountDelta())
	}
	if diff.CostDeltaMln() != 0 {
		t.Errorf("CostDeltaMln() = %f, want 0", diff.CostDeltaMln())
	}
}

func TestBuildStructuralProfile(t *testing.T) {
	t.Parallel()

	project := makeTestProject(t)
	calc := budgetDomain.NewCostCalculation(180, nil, nil, nil)
	profile := buildStructuralProfile(project, &calc)

	if profile.ObjectCount() <= 0 {
		t.Errorf("ObjectCount() = %d, want > 0", profile.ObjectCount())
	}
	if profile.EchelonCount() <= 0 {
		t.Errorf("EchelonCount() = %d, want > 0", profile.EchelonCount())
	}
	if profile.UnitCount() <= 0 {
		t.Errorf("UnitCount() = %d, want > 0", profile.UnitCount())
	}
	if len(profile.ByEchelon()) == 0 {
		t.Errorf("ByEchelon() length = %d, want > 0", len(profile.ByEchelon()))
	}
}

func TestComputeDiff(t *testing.T) {
	t.Parallel()

	profileA := budgetDomain.NewStructuralProfile(5, 10, 2, 3, 1, 4, 100.0, nil)
	profileB := budgetDomain.NewStructuralProfile(8, 16, 3, 4, 2, 6, 250.0, nil)
	calcA := budgetDomain.NewCostCalculation(100, nil, nil, nil)
	calcB := budgetDomain.NewCostCalculation(250, nil, nil, nil)

	diff := computeDiff(profileA, profileB, calcA, calcB)

	if diff.ObjectCountDelta() != 3 {
		t.Errorf("ObjectCountDelta() = %d, want 3", diff.ObjectCountDelta())
	}
	if diff.UnitCountDelta() != 6 {
		t.Errorf("UnitCountDelta() = %d, want 6", diff.UnitCountDelta())
	}
	if diff.EchelonCountDelta() != 1 {
		t.Errorf("EchelonCountDelta() = %d, want 1", diff.EchelonCountDelta())
	}
	if diff.CostDeltaMln() != 150.0 {
		t.Errorf("CostDeltaMln() = %f, want 150.0", diff.CostDeltaMln())
	}
}

// makeTestProjectWithDifferentData создаёт проект с другими данными для сравнения.
func makeTestProjectWithDifferentData(t *testing.T) *defenseDomain.DefenseProject {
	t.Helper()

	now := time.Now().UTC()
	price1 := 20.0
	price2 := 50.0
	price3 := 100.0

	nilString := func(s string) *string {
		return &s
	}

	project, err := defenseDomain.NewDefenseProject(
		"proj-2",
		"Test Config B",
		"ent-1",
		"Test Project B",
		defenseDomain.NewProtectedObject("obj-1", "Object Alpha",
			defenseDomain.NewCoordinates(55.75, 37.62)),
		[]defenseDomain.EditableDefenseLayer{
			defenseDomain.NewEditableDefenseLayer(
				"echelon-1", "Обнаружение", "detection", nil,
				1, nil, nil,
				defenseDomain.LayerGeometryCircle,
				defenseDomain.NewCircleGeometry(defenseDomain.NewCoordinates(55.75, 37.62), 1000),
				nil, nil, true, true, false,
			),
			defenseDomain.NewEditableDefenseLayer(
				"echelon-2", "Подавление", "suppression", nil,
				2, nil, nil,
				defenseDomain.LayerGeometryCircle,
				defenseDomain.NewCircleGeometry(defenseDomain.NewCoordinates(55.75, 37.62), 2000),
				nil, nil, true, true, false,
			),
			defenseDomain.NewEditableDefenseLayer(
				"echelon-3", "Поражение", "engagement", nil,
				3, nil, nil,
				defenseDomain.LayerGeometryCircle,
				defenseDomain.NewCircleGeometry(defenseDomain.NewCoordinates(55.75, 37.62), 3000),
				nil, nil, true, true, false,
			),
		},
		[]defenseDomain.DefenseAsset{
			makeTestAsset("asset-1", "Mobile Radar", "radar", &price1),
			makeTestAsset("asset-2", "Jammer System", "jamming", &price2),
			makeTestAsset("asset-3", "Intercept Missile", "kinetic", &price3),
		},
		[]defenseDomain.PlacedDefenseObject{
			makeTestPlacedObject("placed-1", "asset-1", "echelon-1", 4, nil, now),
			makeTestPlacedObject("placed-2", "asset-2", "echelon-2", 3, nil, now),
			makeTestPlacedObject("placed-3", "asset-3", "echelon-3", 2, nil, now),
		},
		nil,
		nil,
		nil,
		defenseDomain.DefenseProjectModeView,
		defenseDomain.DefenseProjectSourceCustom,
		nilString("preset-2"),
		now,
	)
	if err != nil {
		t.Fatalf("NewDefenseProject() unexpected error: %v", err)
	}
	return project
}

func TestBudgetService_CalculateCost_WithCustomPrice(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	now := time.Now().UTC()
	price := 20.0
	customPrice := 25.0

	nilString := func(s string) *string {
		return &s
	}

	project, err := defenseDomain.NewDefenseProject(
		"proj-2", "Test Config", "ent-1", "Test",
		defenseDomain.NewProtectedObject("obj-1", "Object",
			defenseDomain.NewCoordinates(55.75, 37.62)),
		[]defenseDomain.EditableDefenseLayer{
			defenseDomain.NewEditableDefenseLayer(
				"echelon-1", "Обнаружение", "detection", nil,
				1, nil, nil,
				defenseDomain.LayerGeometryCircle,
				defenseDomain.NewCircleGeometry(defenseDomain.NewCoordinates(55.75, 37.62), 1000),
				nil, nil, true, true, false,
			),
		},
		[]defenseDomain.DefenseAsset{
			makeTestAsset("asset-1", "Mobile Radar", "radar", &price),
		},
		[]defenseDomain.PlacedDefenseObject{
			makeTestPlacedObject("placed-1", "asset-1", "echelon-1", 2, &customPrice, now),
		},
		nil, nil, nil,
		defenseDomain.DefenseProjectModeView,
		defenseDomain.DefenseProjectSourceCustom,
		nilString("preset-1"),
		now,
	)
	if err != nil {
		t.Fatalf("NewDefenseProject() unexpected error: %v", err)
	}

	svc := NewBudgetService(
		&mockBudgetRepo{},
		&mockProjectRepo{projects: map[string]*defenseDomain.DefenseProject{"proj-2": project}},
	)

	calc, err := svc.CalculateCost(ctx, "proj-2")
	if err != nil {
		t.Fatalf("CalculateCost() unexpected error: %v", err)
	}

	// customPrice=25 x 2 = 50
	if calc.TotalMln() != 50 {
		t.Errorf("TotalMln() = %f, want 50 (custom price)", calc.TotalMln())
	}
}
