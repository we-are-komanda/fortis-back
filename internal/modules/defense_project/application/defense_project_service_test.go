package application

import (
	"context"
	"errors"
	"testing"

	"github.com/fortis/backend/internal/modules/defense_project/domain"
)

// mockRepo — мок репозитория для тестирования сервиса.
type mockRepo struct {
	projects map[string]*domain.DefenseProject
	err      error
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		projects: make(map[string]*domain.DefenseProject),
	}
}

func (m *mockRepo) Save(ctx context.Context, project *domain.DefenseProject) error {
	if m.err != nil {
		return m.err
	}
	m.projects[project.ProjectID()] = project
	return nil
}

func (m *mockRepo) FindByID(ctx context.Context, id string) (*domain.DefenseProject, error) {
	if m.err != nil {
		return nil, m.err
	}
	p, ok := m.projects[id]
	if !ok {
		return nil, domain.ErrProjectNotFound
	}
	return p, nil
}

func (m *mockRepo) Delete(ctx context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.projects, id)
	return nil
}

func validProjectJSON() string {
	return `{
		"schemaVersion": 1,
		"projectId": "current",
		"projectName": "Тестовый проект",
		"baseObject": {
			"id": "obj-1",
			"name": "Объект Альфа",
			"center": { "lat": 55.75, "lng": 37.62 }
		},
		"layers": [],
		"assetLibrary": [],
		"placedObjects": [],
		"mode": "view",
		"source": "custom",
		"updatedAt": "2026-06-12T14:00:00.000Z"
	}`
}

func TestDefenseProjectService_Import_Success(t *testing.T) {
	repo := newMockRepo()
	service := NewDefenseProjectService(repo)

	project, err := service.Import(context.Background(), validProjectJSON())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if project.ProjectName() != "Тестовый проект" {
		t.Errorf("expected project name 'Тестовый проект', got %q", project.ProjectName())
	}

	if project.SchemaVersion() != 1 {
		t.Errorf("expected schema version 1, got %d", project.SchemaVersion())
	}

	if project.ProjectID() == "" || project.ProjectID() == "current" {
		t.Errorf("expected new UUID project ID, got %q", project.ProjectID())
	}

	if project.Mode() != domain.DefenseProjectModeView {
		t.Errorf("expected mode 'view', got %q", project.Mode())
	}

	if project.Source() != domain.DefenseProjectSourceCustom {
		t.Errorf("expected source 'custom', got %q", project.Source())
	}

	if project.BaseObject().Name() != "Объект Альфа" {
		t.Errorf("expected base object name 'Объект Альфа', got %q", project.BaseObject().Name())
	}

	// Проверяем, что проект сохранён в репозитории
	saved, err := repo.FindByID(context.Background(), project.ProjectID())
	if err != nil {
		t.Fatalf("failed to find saved project: %v", err)
	}
	if saved.ProjectName() != "Тестовый проект" {
		t.Errorf("expected saved project name 'Тестовый проект', got %q", saved.ProjectName())
	}
}

func TestDefenseProjectService_Import_InvalidSchemaVersion(t *testing.T) {
	repo := newMockRepo()
	service := NewDefenseProjectService(repo)

	jsonStr := `{
		"schemaVersion": 999,
		"projectId": "test",
		"projectName": "Test",
		"baseObject": { "id": "o1", "name": "Obj", "center": { "lat": 0, "lng": 0 } },
		"layers": [],
		"assetLibrary": [],
		"placedObjects": [],
		"mode": "view",
		"updatedAt": "2026-06-12T14:00:00.000Z"
	}`

	_, err := service.Import(context.Background(), jsonStr)
	if !errors.Is(err, domain.ErrInvalidSchemaVersion) {
		t.Errorf("expected ErrInvalidSchemaVersion, got %v", err)
	}
}

func TestDefenseProjectService_Import_InvalidJSON(t *testing.T) {
	repo := newMockRepo()
	service := NewDefenseProjectService(repo)

	_, err := service.Import(context.Background(), "not json at all")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDefenseProjectService_Import_EmptyProjectName(t *testing.T) {
	repo := newMockRepo()
	service := NewDefenseProjectService(repo)

	jsonStr := `{
		"schemaVersion": 1,
		"projectId": "test",
		"projectName": "",
		"baseObject": { "id": "o1", "name": "Obj", "center": { "lat": 0, "lng": 0 } },
		"layers": [],
		"assetLibrary": [],
		"placedObjects": [],
		"mode": "view",
		"updatedAt": "2026-06-12T14:00:00.000Z"
	}`

	_, err := service.Import(context.Background(), jsonStr)
	if !errors.Is(err, domain.ErrInvalidProjectData) {
		t.Errorf("expected ErrInvalidProjectData, got %v", err)
	}
}

func TestDefenseProjectService_Export_Success(t *testing.T) {
	repo := newMockRepo()
	service := NewDefenseProjectService(repo)

	// Сначала импортируем проект
	project, err := service.Import(context.Background(), validProjectJSON())
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	// Экспортируем
	jsonStr, err := service.Export(context.Background(), project.ProjectID())
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	if jsonStr == "" {
		t.Fatal("expected non-empty JSON string")
	}
}

func TestDefenseProjectService_Export_NotFound(t *testing.T) {
	repo := newMockRepo()
	service := NewDefenseProjectService(repo)

	_, err := service.Export(context.Background(), "nonexistent-id")
	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Errorf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestDefenseProjectService_Import_WithLayers(t *testing.T) {
	repo := newMockRepo()
	service := NewDefenseProjectService(repo)

	jsonStr := `{
		"schemaVersion": 1,
		"projectId": "test",
		"projectName": "Слоистый проект",
		"baseObject": {
			"id": "obj-1",
			"name": "Объект",
			"center": { "lat": 55.75, "lng": 37.62 }
		},
		"layers": [
			{
				"id": "layer-1",
				"name": "Внешний периметр",
				"code": "outer",
				"order": 1,
				"geometryType": "ring",
				"geometry": {
					"type": "ring",
					"center": { "lat": 55.75, "lng": 37.62 },
					"minRadiusM": 100,
					"maxRadiusM": 5000
				},
				"isActive": true,
				"isVisible": true,
				"isLocked": false
			},
			{
				"id": "layer-2",
				"name": "Внутренний периметр",
				"code": "inner",
				"order": 2,
				"geometryType": "circle",
				"geometry": {
					"type": "circle",
					"center": { "lat": 55.75, "lng": 37.62 },
					"radiusM": 100
				},
				"isActive": true
			}
		],
		"assetLibrary": [],
		"placedObjects": [],
		"mode": "view",
		"updatedAt": "2026-06-12T14:00:00.000Z"
	}`

	project, err := service.Import(context.Background(), jsonStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(project.Layers()) != 2 {
		t.Fatalf("expected 2 layers, got %d", len(project.Layers()))
	}

	layer1 := project.Layers()[0]
	if layer1.Name() != "Внешний периметр" {
		t.Errorf("expected layer name 'Внешний периметр', got %q", layer1.Name())
	}
	if layer1.GeometryType() != domain.LayerGeometryRing {
		t.Errorf("expected ring geometry, got %q", layer1.GeometryType())
	}
	if layer1.IsVisible() != true {
		t.Errorf("expected layer isVisible=true")
	}

	layer2 := project.Layers()[1]
	if layer2.GeometryType() != domain.LayerGeometryCircle {
		t.Errorf("expected circle geometry, got %q", layer2.GeometryType())
	}
	// IsVisible should default to true when not specified
	if layer2.IsVisible() != true {
		t.Errorf("expected layer isVisible to default to true")
	}
}

func TestDefenseProjectService_Import_WithAssets(t *testing.T) {
	repo := newMockRepo()
	service := NewDefenseProjectService(repo)

	jsonStr := `{
		"schemaVersion": 1,
		"projectId": "test",
		"projectName": "Проект с ассетами",
		"baseObject": {
			"id": "obj-1",
			"name": "Объект",
			"center": { "lat": 55.75, "lng": 37.62 }
		},
		"layers": [],
		"assetLibrary": [
			{
				"id": "asset-1",
				"name": "РЛС-1",
				"category": "detection",
				"roles": ["detect", "track"],
				"pricePerUnitMln": 10.5,
				"currency": "RUB",
				"unitLabel": "шт.",
				"coverageType": "circle",
				"coverageRadius": 5000,
				"deploymentType": "static",
				"placementType": "map-object"
			}
		],
		"placedObjects": [],
		"mode": "view",
		"updatedAt": "2026-06-12T14:00:00.000Z"
	}`

	project, err := service.Import(context.Background(), jsonStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(project.AssetLibrary()) != 1 {
		t.Fatalf("expected 1 asset, got %d", len(project.AssetLibrary()))
	}

	asset := project.AssetLibrary()[0]
	if asset.Name() != "РЛС-1" {
		t.Errorf("expected asset name 'РЛС-1', got %q", asset.Name())
	}
	if asset.Category() != domain.DefenseAssetCategoryDetection {
		t.Errorf("expected category 'detection', got %q", asset.Category())
	}
	if len(asset.Roles()) != 2 {
		t.Errorf("expected 2 roles, got %d", len(asset.Roles()))
	}
}

func TestDefenseProjectService_Import_WithPlacedObjects(t *testing.T) {
	repo := newMockRepo()
	service := NewDefenseProjectService(repo)

	jsonStr := `{
		"schemaVersion": 1,
		"projectId": "test",
		"projectName": "Проект с размещениями",
		"baseObject": {
			"id": "obj-1",
			"name": "Объект",
			"center": { "lat": 55.75, "lng": 37.62 }
		},
		"layers": [],
		"assetLibrary": [],
		"placedObjects": [
			{
				"id": "placed-1",
				"assetId": "asset-1",
				"layerId": "layer-1",
				"coordinates": { "lat": 55.76, "lng": 37.63 },
				"quantity": 2,
				"status": "active",
				"hasGeometryConflict": false,
				"hasCoverageConflict": false,
				"hasTerrainConflict": false,
				"createdAt": "2026-06-12T14:00:00.000Z",
				"updatedAt": "2026-06-12T14:00:00.000Z"
			}
		],
		"mode": "view",
		"updatedAt": "2026-06-12T14:00:00.000Z"
	}`

	project, err := service.Import(context.Background(), jsonStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(project.PlacedObjects()) != 1 {
		t.Fatalf("expected 1 placed object, got %d", len(project.PlacedObjects()))
	}

	obj := project.PlacedObjects()[0]
	if obj.AssetID() != "asset-1" {
		t.Errorf("expected assetId 'asset-1', got %q", obj.AssetID())
	}
	if obj.Quantity() != 2 {
		t.Errorf("expected quantity 2, got %d", obj.Quantity())
	}
	if obj.Status() != domain.PlacedObjectStatusActive {
		t.Errorf("expected status 'active', got %q", obj.Status())
	}
}

func TestDefenseProjectService_Export_RoundTrip(t *testing.T) {
	repo := newMockRepo()
	service := NewDefenseProjectService(repo)

	// Импортируем
	project, err := service.Import(context.Background(), validProjectJSON())
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	// Экспортируем
	exportedJSON, err := service.Export(context.Background(), project.ProjectID())
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	// Импортируем экспортированное (round-trip)
	reimported, err := service.Import(context.Background(), exportedJSON)
	if err != nil {
		t.Fatalf("reimport failed: %v", err)
	}

	if reimported.ProjectName() != project.ProjectName() {
		t.Errorf("round-trip: expected project name %q, got %q",
			project.ProjectName(), reimported.ProjectName())
	}
}
