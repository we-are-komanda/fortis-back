package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fortis/backend/internal/modules/defense_project/domain"
)

// newExistingProject создаёт сохранённый проект для тестов UpdateProject.
//
//nolint:unparam // name and id intentionally hardcoded for tests
func newExistingProject(t *testing.T, name string, placed []domain.PlacedDefenseObject) *domain.DefenseProject {
	t.Helper()
	project, err := domain.NewDefenseProject(
		"p1", name, "", "Existing Project",
		domain.NewProtectedObject("obj-0", "Obj0", domain.NewCoordinates(55.0, 37.0)),
		nil, nil, placed,
		nil, nil, nil,
		domain.DefenseProjectModeView, domain.DefenseProjectSourceCustom,
		nil, time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("failed to build existing project: %v", err)
	}
	return project
}

func TestUpdateProjectOverwritesContentWhenProjectJSONProvided(t *testing.T) {
	repo := newMockRepo()
	service := NewDefenseProjectService(repo)

	existing := newExistingProject(t, "Original", nil)
	repo.projects["p1"] = existing

	projectJSON := `{
		"schemaVersion": 1, "projectId": "ignored", "projectName": "Site Alpha",
		"baseObject": {"id":"o1","name":"Obj","center":{"lat":55.75,"lng":37.61}},
		"layers": [], "assetLibrary": [],
		"placedObjects": [{"id":"pl1","assetId":"a1","layerId":"l1","coordinates":{"lat":55.76,"lng":37.62},"quantity":1,"status":"planned"}],
		"mode": "view", "updatedAt": "2026-06-12T14:00:00.000Z"
	}`

	project, err := service.UpdateProject(context.Background(), "p1", "Renamed", "", projectJSON, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if project.ProjectID() != "p1" {
		t.Errorf("expected project ID 'p1', got %q", project.ProjectID())
	}
	if got := len(project.PlacedObjects()); got != 1 {
		t.Errorf("expected 1 placed object, got %d", got)
	}
	if project.Name() != "Renamed" {
		t.Errorf("expected name 'Renamed', got %q", project.Name())
	}
}

func TestUpdateProjectOverwritePreservesVersion(t *testing.T) {
	repo := newMockRepo()
	service := NewDefenseProjectService(repo)

	existing := newExistingProject(t, "Original", nil)
	// Симулируем уже несколько раз сохранённый проект (optimistic-lock версия != 1).
	existing.SetVersion(3)
	repo.projects["p1"] = existing

	projectJSON := `{
		"schemaVersion": 1, "projectId": "ignored", "projectName": "Site Alpha",
		"baseObject": {"id":"o1","name":"Obj","center":{"lat":55.75,"lng":37.61}},
		"layers": [], "assetLibrary": [], "placedObjects": [],
		"mode": "view", "updatedAt": "2026-06-12T14:00:00.000Z"
	}`

	project, err := service.UpdateProject(context.Background(), "p1", "", "", projectJSON, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if project.Version() != 4 {
		t.Errorf("expected returned project version 4, got %d", project.Version())
	}
	// Версия должна сохраниться и в записанном через repo.Save объекте как
	// следующая optimistic-lock версия (overwrite не должен сбрасывать её на 1).
	saved, ok := repo.projects["p1"]
	if !ok {
		t.Fatalf("saved project not found in repo")
	}
	if saved.Version() != 4 {
		t.Errorf("expected saved project version 4, got %d", saved.Version())
	}
}

func TestUpdateProjectOverwriteInvalidSchemaVersion(t *testing.T) {
	repo := newMockRepo()
	service := NewDefenseProjectService(repo)

	existing := newExistingProject(t, "Original", nil)
	repo.projects["p1"] = existing

	projectJSON := `{
		"schemaVersion": 999, "projectName": "Site Alpha",
		"baseObject": {"id":"o1","name":"Obj","center":{"lat":55.75,"lng":37.61}},
		"layers": [], "assetLibrary": [], "placedObjects": []
	}`

	_, err := service.UpdateProject(context.Background(), "p1", "", "", projectJSON, nil)
	if !errors.Is(err, domain.ErrInvalidSchemaVersion) {
		t.Errorf("expected ErrInvalidSchemaVersion, got %v", err)
	}
}

func TestUpdateProject_WithExplicitVersion_Success(t *testing.T) {
	repo := newMockRepo()
	service := NewDefenseProjectService(repo)

	existing := newExistingProject(t, "Original", nil)
	existing.SetVersion(3)
	repo.projects["p1"] = existing

	version := 3
	project, err := service.UpdateProject(context.Background(), "p1", "Renamed", "", "", &version)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if project.Name() != "Renamed" {
		t.Errorf("expected name 'Renamed', got %q", project.Name())
	}
	if project.Version() != 4 {
		t.Errorf("expected version 4, got %d", project.Version())
	}
}

func TestUpdateProject_WithExplicitVersion_Conflict(t *testing.T) {
	repo := newMockRepo()
	service := NewDefenseProjectService(repo)

	existing := newExistingProject(t, "Original", nil)
	existing.SetVersion(3)
	repo.projects["p1"] = existing

	// Клиент указывает неверную версию (ожидал 999, а актуальна 3)
	version := 999
	_, err := service.UpdateProject(context.Background(), "p1", "Renamed", "", "", &version)
	if !errors.Is(err, domain.ErrVersionConflict) {
		t.Errorf("expected ErrVersionConflict, got %v", err)
	}
}

func TestUpdateProject_WithoutVersion_StillWorks(t *testing.T) {
	repo := newMockRepo()
	service := NewDefenseProjectService(repo)

	existing := newExistingProject(t, "Original", nil)
	existing.SetVersion(3)
	repo.projects["p1"] = existing

	// Не передаём version (nil) — должно работать как раньше
	project, err := service.UpdateProject(context.Background(), "p1", "Renamed", "", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if project.Name() != "Renamed" {
		t.Errorf("expected name 'Renamed', got %q", project.Name())
	}
	if project.Version() != 4 {
		t.Errorf("expected version 4, got %d", project.Version())
	}
}

func TestUpdateProject_InvalidProjectJson_ReturnsError(t *testing.T) {
	repo := newMockRepo()
	service := NewDefenseProjectService(repo)

	existing := newExistingProject(t, "Original", nil)
	repo.projects["p1"] = existing

	// Существующий тест на некорректный projectJSON не сломан
	_, err := service.UpdateProject(context.Background(), "p1", "", "", "not json", nil)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestUpdateProjectMetadataOnlyWhenNoProjectJSON(t *testing.T) {
	repo := newMockRepo()
	service := NewDefenseProjectService(repo)

	placed := []domain.PlacedDefenseObject{
		domain.NewPlacedDefenseObject(
			"pl-existing", "a1", "l1", nil,
			domain.NewCoordinates(55.1, 37.1),
			nil, nil, 1, domain.PlacedObjectStatusPlanned,
			nil, nil, nil, false, false, false, nil,
			time.Now().UTC(), time.Now().UTC(),
		),
	}
	existing := newExistingProject(t, "Original", placed)
	repo.projects["p1"] = existing

	project, err := service.UpdateProject(context.Background(), "p1", "OnlyName", "", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := len(project.PlacedObjects()); got != 1 {
		t.Errorf("expected content untouched (1 placed object), got %d", got)
	}
	if project.Name() != "OnlyName" {
		t.Errorf("expected name 'OnlyName', got %q", project.Name())
	}
}
