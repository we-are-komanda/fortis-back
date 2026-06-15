package application

import (
	"context"
	"errors"
	"testing"
	"time"

	budgetDomain "github.com/fortis/backend/internal/modules/budget/domain"
	defenseAssetDomain "github.com/fortis/backend/internal/modules/defense_asset/domain"
	defenseDomain "github.com/fortis/backend/internal/modules/defense_project/domain"
	reportDomain "github.com/fortis/backend/internal/modules/report/domain"
)

// ---- Mocks ----

type mockProjectRepo struct {
	projects map[string]*defenseDomain.DefenseProject
	err      error
}

func (m *mockProjectRepo) Save(ctx context.Context, project *defenseDomain.DefenseProject) error {
	return m.err
}

func (m *mockProjectRepo) FindByID(ctx context.Context, id string) (*defenseDomain.DefenseProject, error) {
	if m.err != nil {
		return nil, m.err
	}
	p, ok := m.projects[id]
	if !ok {
		return nil, defenseDomain.ErrProjectNotFound
	}
	return p, nil
}

func (m *mockProjectRepo) FindAll(ctx context.Context, limit, offset int) ([]*defenseDomain.DefenseProject, int64, error) {
	return nil, 0, m.err
}

func (m *mockProjectRepo) FindAllByEnterprise(ctx context.Context, enterpriseID string, limit, offset int) ([]*defenseDomain.DefenseProject, int64, error) {
	return nil, 0, m.err
}

func (m *mockProjectRepo) Delete(ctx context.Context, id string) error {
	return m.err
}

type mockBudgetSvc struct {
	calc *budgetDomain.CostCalculation
	err  error
}

func (m *mockBudgetSvc) CalculateCost(ctx context.Context, projectID string) (*budgetDomain.CostCalculation, error) {
	return m.calc, m.err
}

type mockAssetRepo struct {
	assets map[string]*defenseAssetDomain.DefenseAsset
	err    error
}

func (m *mockAssetRepo) Save(ctx context.Context, asset *defenseAssetDomain.DefenseAsset) error {
	return m.err
}

func (m *mockAssetRepo) FindByID(ctx context.Context, id string) (*defenseAssetDomain.DefenseAsset, error) {
	if m.err != nil {
		return nil, m.err
	}
	a, ok := m.assets[id]
	if !ok {
		return nil, errors.New("asset not found")
	}
	return a, nil
}

func (m *mockAssetRepo) FindAll(ctx context.Context, filter defenseAssetDomain.DefenseAssetFilter) ([]*defenseAssetDomain.DefenseAsset, int64, error) {
	return nil, 0, m.err
}

func (m *mockAssetRepo) Update(ctx context.Context, asset *defenseAssetDomain.DefenseAsset) error {
	return m.err
}

func (m *mockAssetRepo) Delete(ctx context.Context, id string) error {
	return m.err
}

// ---- Test helpers ----

func makeTestProject(t *testing.T, placedObjects bool) *defenseDomain.DefenseProject {
	t.Helper()

	baseObj := defenseDomain.NewProtectedObject("base-1", "Test Base",
		defenseDomain.NewCoordinates(55.75, 37.62))

	layer := defenseDomain.NewEditableDefenseLayer(
		"layer-1", "Дальнее обнаружение", "early-warning", nil,
		1, nil, nil,
		defenseDomain.LayerGeometryCircle,
		defenseDomain.NewCircleGeometry(defenseDomain.NewCoordinates(55.75, 37.62), 50000),
		strPtr("#FF0000"), float64Ptr(0.5),
		true, true, false,
	)

	asset := defenseDomain.NewDefenseAsset(
		"asset-1", "РЛС Небо-М", nil, nil,
		defenseDomain.DefenseAssetCategoryDetection,
		[]defenseDomain.DefenseAssetRole{defenseDomain.DefenseAssetRoleDetect},
		float64Ptr(150.0), "RUB", "шт.",
		[]defenseDomain.LayerGeometryType{defenseDomain.LayerGeometryCircle},
		[]string{"early-warning"}, []string{}, []string{},
		"detection", nil, nil,
		defenseDomain.DefenseAssetCoverageCircle, float64Ptr(50000), nil,
		defenseDomain.DefenseAssetDeploymentStatic,
		defenseDomain.DefenseAssetPlacementMapObject,
		nil, nil, nil, nil,
		nil, []string{}, nil, nil, []string{},
	)

	projLayers := []defenseDomain.EditableDefenseLayer{layer}
	projAssets := []defenseDomain.DefenseAsset{asset}
	var projObjects []defenseDomain.PlacedDefenseObject

	if placedObjects {
		obj := defenseDomain.NewPlacedDefenseObject(
			"obj-1", "asset-1", "layer-1", nil,
			defenseDomain.NewCoordinates(55.76, 37.63),
			nil, nil, 2,
			defenseDomain.PlacedObjectStatusActive,
			nil, nil, nil,
			false, false, false, nil,
			nowTime(), nowTime(),
		)
		projObjects = []defenseDomain.PlacedDefenseObject{obj}
	}

	project, err := defenseDomain.NewDefenseProject(
		"proj-1", "variant-1", "ent-1", "Test Project",
		baseObj, projLayers, projAssets, projObjects,
		nil, nil, nil,
		defenseDomain.DefenseProjectModeView,
		defenseDomain.DefenseProjectSourceCustom,
		nil, nowTime(),
	)
	if err != nil {
		t.Fatalf("failed to create test project: %v", err)
	}
	return project
}

func makeTestCalc(t *testing.T, withObjects bool) *budgetDomain.CostCalculation {
	t.Helper()

	if !withObjects {
		calc := budgetDomain.NewCostCalculation(0, nil, nil, nil)
		return &calc
	}

	line := budgetDomain.NewEstimateLine(
		"obj-1", "asset-1", "РЛС Небо-М",
		"layer-1", "Дальнее обнаружение",
		"detection", "detection",
		2, 150.0, 300.0,
	)
	byEchelon := []budgetDomain.EchelonEstimate{
		budgetDomain.NewEchelonEstimate("layer-1", "Дальнее обнаружение", []budgetDomain.EstimateLine{line}, 300.0),
	}
	byType := []budgetDomain.TypeEstimate{
		budgetDomain.NewTypeEstimate("detection", "detection", []budgetDomain.EstimateLine{line}, 300.0),
	}
	calc := budgetDomain.NewCostCalculation(300.0, byEchelon, byType, []budgetDomain.EstimateLine{line})
	return &calc
}

// ---- Tests ----

func TestReportService_GetReport_Success(t *testing.T) {
	t.Parallel()

	project := makeTestProject(t, true)
	calc := makeTestCalc(t, true)

	svc := NewReportService(
		&mockProjectRepo{projects: map[string]*defenseDomain.DefenseProject{"proj-1": project}},
		&mockBudgetSvc{calc: calc},
		&mockAssetRepo{assets: map[string]*defenseAssetDomain.DefenseAsset{}},
	)

	payload, err := svc.GetReport(context.Background(), "proj-1", false)
	if err != nil {
		t.Fatalf("GetReport() unexpected error: %v", err)
	}

	if payload == nil {
		t.Fatal("GetReport() returned nil payload")
	}

	if payload.ProjectID() != "proj-1" {
		t.Errorf("ProjectID() = %s, want proj-1", payload.ProjectID())
	}
	if payload.ProjectName() != "Test Project" {
		t.Errorf("ProjectName() = %s, want Test Project", payload.ProjectName())
	}

	if len(payload.Layers()) != 1 {
		t.Errorf("Layers() length = %d, want 1", len(payload.Layers()))
	}

	if len(payload.PlacedObjects()) != 1 {
		t.Errorf("PlacedObjects() length = %d, want 1", len(payload.PlacedObjects()))
	}

	if payload.Estimate().TotalMln() != 300.0 {
		t.Errorf("Estimate.TotalMln() = %f, want 300.0", payload.Estimate().TotalMln())
	}

	if payload.HideCost() {
		t.Error("HideCost() = true, want false")
	}

	// Проверка структурного профиля
	profile := payload.StructuralProfile()
	if profile.ObjectCount() != 1 {
		t.Errorf("StructuralProfile.ObjectCount() = %d, want 1", profile.ObjectCount())
	}
	if profile.EchelonCount() != 1 {
		t.Errorf("StructuralProfile.EchelonCount() = %d, want 1", profile.EchelonCount())
	}

	// Проверка слоя
	layer := payload.Layers()[0]
	if layer.Name() != "Дальнее обнаружение" {
		t.Errorf("Layer.Name() = %s, want Дальнее обнаружение", layer.Name())
	}

	// Проверка объекта
	obj := payload.PlacedObjects()[0]
	if obj.AssetName() != "РЛС Небо-М" {
		t.Errorf("Object.AssetName() = %s, want РЛС Небо-М", obj.AssetName())
	}
	if obj.Quantity() != 2 {
		t.Errorf("Object.Quantity() = %d, want 2", obj.Quantity())
	}
}

func TestReportService_GetReport_EmptyConfig(t *testing.T) {
	t.Parallel()

	project := makeTestProject(t, false)
	calc := makeTestCalc(t, false)

	svc := NewReportService(
		&mockProjectRepo{projects: map[string]*defenseDomain.DefenseProject{"proj-1": project}},
		&mockBudgetSvc{calc: calc},
		&mockAssetRepo{},
	)

	payload, err := svc.GetReport(context.Background(), "proj-1", false)
	if err != nil {
		t.Fatalf("GetReport() unexpected error: %v", err)
	}

	if len(payload.PlacedObjects()) != 0 {
		t.Errorf("PlacedObjects() length = %d, want 0", len(payload.PlacedObjects()))
	}

	if payload.Estimate().TotalMln() != 0 {
		t.Errorf("Estimate.TotalMln() = %f, want 0", payload.Estimate().TotalMln())
	}

	profile := payload.StructuralProfile()
	if profile.ObjectCount() != 0 {
		t.Errorf("StructuralProfile.ObjectCount() = %d, want 0", profile.ObjectCount())
	}
	if profile.TotalMln() != 0 {
		t.Errorf("StructuralProfile.TotalMln() = %f, want 0", profile.TotalMln())
	}

	// Layers всё равно должны быть (проект может иметь слои без объектов)
	if len(payload.Layers()) != 1 {
		t.Errorf("Layers() length = %d, want 1", len(payload.Layers()))
	}
}

func TestReportService_GetReport_HideCost(t *testing.T) {
	t.Parallel()

	project := makeTestProject(t, true)
	calc := makeTestCalc(t, true)

	svc := NewReportService(
		&mockProjectRepo{projects: map[string]*defenseDomain.DefenseProject{"proj-1": project}},
		&mockBudgetSvc{calc: calc},
		&mockAssetRepo{assets: map[string]*defenseAssetDomain.DefenseAsset{}},
	)

	payload, err := svc.GetReport(context.Background(), "proj-1", true)
	if err != nil {
		t.Fatalf("GetReport() unexpected error: %v", err)
	}

	if !payload.HideCost() {
		t.Error("HideCost() = false, want true")
	}

	// Все cost-поля должны быть 0
	if payload.Estimate().TotalMln() != 0 {
		t.Errorf("Estimate.TotalMln() = %f, want 0", payload.Estimate().TotalMln())
	}

	for _, obj := range payload.PlacedObjects() {
		if obj.UnitPriceMln() != 0 {
			t.Errorf("Object.UnitPriceMln() = %f, want 0", obj.UnitPriceMln())
		}
		if obj.LineTotalMln() != 0 {
			t.Errorf("Object.LineTotalMln() = %f, want 0", obj.LineTotalMln())
		}
	}

	profile := payload.StructuralProfile()
	if profile.TotalMln() != 0 {
		t.Errorf("StructuralProfile.TotalMln() = %f, want 0", profile.TotalMln())
	}

	// Не-cost поля должны сохраниться
	if len(payload.PlacedObjects()) != 1 {
		t.Errorf("PlacedObjects() length = %d, want 1", len(payload.PlacedObjects()))
	}
	if payload.PlacedObjects()[0].Quantity() != 2 {
		t.Errorf("Object.Quantity() = %d, want 2", payload.PlacedObjects()[0].Quantity())
	}
}

func TestReportService_GetReport_InvalidProjectID(t *testing.T) {
	t.Parallel()

	svc := NewReportService(
		&mockProjectRepo{},
		&mockBudgetSvc{},
		&mockAssetRepo{},
	)

	_, err := svc.GetReport(context.Background(), "", false)
	if err == nil {
		t.Fatal("GetReport() expected error for empty project ID")
	}
	if !errors.Is(err, reportDomain.ErrInvalidProjectID) {
		t.Errorf("GetReport() error = %v, want ErrInvalidProjectID", err)
	}
}

func TestReportService_GetReport_ProjectNotFound(t *testing.T) {
	t.Parallel()

	svc := NewReportService(
		&mockProjectRepo{
			projects: map[string]*defenseDomain.DefenseProject{},
		},
		&mockBudgetSvc{},
		&mockAssetRepo{},
	)

	_, err := svc.GetReport(context.Background(), "nonexistent", false)
	if err == nil {
		t.Fatal("GetReport() expected error for nonexistent project")
	}
	if !errors.Is(err, reportDomain.ErrProjectNotFound) {
		t.Errorf("GetReport() error = %v, want ErrProjectNotFound", err)
	}
}

func TestReportService_GetReport_CompoundPost(t *testing.T) {
	t.Parallel()

	baseObj := defenseDomain.NewProtectedObject("base-1", "Test Base",
		defenseDomain.NewCoordinates(55.75, 37.62))

	layer := defenseDomain.NewEditableDefenseLayer(
		"layer-1", "Дальнее обнаружение", "early-warning", nil,
		1, nil, nil,
		defenseDomain.LayerGeometryCircle,
		defenseDomain.NewCircleGeometry(defenseDomain.NewCoordinates(55.75, 37.62), 50000),
		strPtr("#FF0000"), float64Ptr(0.5),
		true, true, false,
	)

	compoundProfile := &defenseAssetDomain.DefenseAssetCompoundProfile{
		Kind:           "МОГ",
		PostType:       "МОГ",
		PersonnelCount: "4 чел.",
		Accountability: "командир роты",
		Armament:       "ПЗРК Игла",
		WeaponUnits:    "2",
		SectorOrRange:  "120°",
		Azimuth:        180,
	}

	// Используем DefenseAsset из defense_project domain (он принимает compoundProfile)
	asset := defenseDomain.NewDefenseAsset(
		"asset-compound", "ЗУ-23", nil, nil,
		defenseDomain.DefenseAssetCategoryDetection,
		[]defenseDomain.DefenseAssetRole{defenseDomain.DefenseAssetRoleDestroy},
		float64Ptr(50.0), "RUB", "шт.",
		[]defenseDomain.LayerGeometryType{defenseDomain.LayerGeometryCircle},
		[]string{"early-warning"}, []string{}, []string{},
		"artillery", nil, nil,
		defenseDomain.DefenseAssetCoverageSector, float64Ptr(1000), float64Ptr(60),
		defenseDomain.DefenseAssetDeploymentStatic,
		defenseDomain.DefenseAssetPlacementMapObject,
		nil, nil, nil, nil,
		compoundProfile, []string{}, nil, nil, []string{},
	)

	obj := defenseDomain.NewPlacedDefenseObject(
		"obj-compound", "asset-compound", "layer-1", nil,
		defenseDomain.NewCoordinates(55.76, 37.63),
		nil, nil, 1,
		defenseDomain.PlacedObjectStatusActive,
		nil, nil, nil,
		false, false, false, nil,
		nowTime(), nowTime(),
	)

	project, err := defenseDomain.NewDefenseProject(
		"proj-compound", "variant-1", "ent-1", "Test Compound",
		baseObj,
		[]defenseDomain.EditableDefenseLayer{layer},
		[]defenseDomain.DefenseAsset{asset},
		[]defenseDomain.PlacedDefenseObject{obj},
		nil, nil, nil,
		defenseDomain.DefenseProjectModeView,
		defenseDomain.DefenseProjectSourceCustom,
		nil, nowTime(),
	)
	if err != nil {
		t.Fatalf("failed to create test project: %v", err)
	}

	calc := makeTestCalc(t, true)

	// Создаем полный asset с weaponSpec для проверки weaponSummary
	fullAsset, err := defenseAssetDomain.NewDefenseAsset(
		"asset-compound", "ЗУ-23", "", "",
		defenseAssetDomain.DefenseAssetCategoryArtillery,
		[]defenseAssetDomain.DefenseAssetRole{defenseAssetDomain.DefenseAssetRoleDestruction},
		float64Ptr(50.0), "RUB", "шт.",
		[]defenseAssetDomain.LayerType{},
		[]string{}, []string{}, []string{},
		"artillery", nil, nil,
		defenseAssetDomain.DefenseAssetCoverageSector,
		float64Ptr(1000), float64Ptr(60),
		defenseAssetDomain.DeploymentTypeStatic,
		defenseAssetDomain.PlacementTypeMapObject,
		"", "",
		nil, nil,
		compoundProfile,
		&defenseAssetDomain.WeaponSpecification{
			Caliber:        strPtr("23 мм"),
			AmmunitionType: strPtr("ОФ"),
			OperationMode:  strPtr("автомат"),
			ModuleCount:    intPtr(2),
			IsManual:       boolPtr(false),
		},
		nil, nil,
		[]string{}, "",
		nil, []string{}, nil, false,
		nowTime(), nowTime(),
	)
	if err != nil {
		t.Fatalf("failed to create full asset: %v", err)
	}

	svc := NewReportService(
		&mockProjectRepo{projects: map[string]*defenseDomain.DefenseProject{"proj-compound": project}},
		&mockBudgetSvc{calc: calc},
		&mockAssetRepo{
			assets: map[string]*defenseAssetDomain.DefenseAsset{"asset-compound": fullAsset},
		},
	)

	payload, err := svc.GetReport(context.Background(), "proj-compound", false)
	if err != nil {
		t.Fatalf("GetReport() unexpected error: %v", err)
	}

	if len(payload.PlacedObjects()) != 1 {
		t.Fatalf("PlacedObjects() length = %d, want 1", len(payload.PlacedObjects()))
	}

	objResult := payload.PlacedObjects()[0]

	// Проверка compound
	if !objResult.IsCompoundPost() {
		t.Error("IsCompoundPost() = false, want true")
	}

	cs := objResult.CompositionSummary()
	if cs == nil {
		t.Fatal("CompositionSummary() = nil, want non-nil")
	}
	if cs.PostType() != "МОГ" {
		t.Errorf("CompositionSummary.PostType() = %s, want МОГ", cs.PostType())
	}
	if cs.Personnel() != "4 чел." {
		t.Errorf("CompositionSummary.Personnel() = %s, want 4 чел.", cs.Personnel())
	}
	if cs.Armament() != "ПЗРК Игла" {
		t.Errorf("CompositionSummary.Armament() = %s, want ПЗРК Игла", cs.Armament())
	}
	if cs.Azimuth() != 180 {
		t.Errorf("CompositionSummary.Azimuth() = %f, want 180", cs.Azimuth())
	}

	// Проверка weapon summary
	ws := objResult.WeaponSummary()
	if ws == nil {
		t.Fatal("WeaponSummary() = nil, want non-nil")
	}
	if ws.Caliber() != "23 мм" {
		t.Errorf("WeaponSummary.Caliber() = %s, want 23 мм", ws.Caliber())
	}
	if ws.ModuleCount() != "2" {
		t.Errorf("WeaponSummary.ModuleCount() = %s, want 2", ws.ModuleCount())
	}

	// Проверка azimuth sector summary
	az := objResult.AzimuthSectorSummary()
	if az == nil {
		t.Fatal("AzimuthSectorSummary() = nil, want non-nil")
	}
	if az.Azimuth() != 180 {
		t.Errorf("AzimuthSectorSummary.Azimuth() = %f, want 180", az.Azimuth())
	}
}

// ---- Helpers ----

func strPtr(s string) *string { return &s }

func float64Ptr(f float64) *float64 { return &f }

func intPtr(i int) *int { return &i }

func boolPtr(b bool) *bool { return &b }

func nowTime() time.Time { return time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC) }
