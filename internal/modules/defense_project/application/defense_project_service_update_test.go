package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fortis/backend/internal/modules/defense_project/domain"
)

// newExistingProject создаёт сохранённый проект для тестов UpdateProject.
func newExistingProject(t *testing.T, id, name string, placed []domain.PlacedDefenseObject) *domain.DefenseProject {
	t.Helper()
	project, err := domain.NewDefenseProject(
		id, name, "", "Existing Project",
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

	existing := newExistingProject(t, "p1", "Original", nil)
	repo.projects["p1"] = existing

	projectJSON := `{
		"schemaVersion": 1, "projectId": "ignored", "projectName": "Site Alpha",
		"baseObject": {"id":"o1","name":"Obj","center":{"lat":55.75,"lng":37.61}},
		"layers": [], "assetLibrary": [],
		"placedObjects": [{"id":"pl1","assetId":"a1","layerId":"l1","coordinates":{"lat":55.76,"lng":37.62},"quantity":1,"status":"planned"}],
		"mode": "view", "updatedAt": "2026-06-12T14:00:00.000Z"
	}`

	project, err := service.UpdateProject(context.Background(), "p1", "Renamed", "", projectJSON)
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

	existing := newExistingProject(t, "p1", "Original", nil)
	// Симулируем уже несколько раз сохранённый проект (optimistic-lock версия != 1).
	existing.SetVersion(3)
	repo.projects["p1"] = existing

	projectJSON := `{
		"schemaVersion": 1, "projectId": "ignored", "projectName": "Site Alpha",
		"baseObject": {"id":"o1","name":"Obj","center":{"lat":55.75,"lng":37.61}},
		"layers": [], "assetLibrary": [], "placedObjects": [],
		"mode": "view", "updatedAt": "2026-06-12T14:00:00.000Z"
	}`

	project, err := service.UpdateProject(context.Background(), "p1", "", "", projectJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if project.Version() != 3 {
		t.Errorf("expected returned project version 3, got %d", project.Version())
	}
	// Версия должна сохраниться и в записанном через repo.Save объекте (overwrite
	// не должен сбрасывать её на 1 или вызывать конфликт версий).
	saved, ok := repo.projects["p1"]
	if !ok {
		t.Fatalf("saved project not found in repo")
	}
	if saved.Version() != 3 {
		t.Errorf("expected saved project version 3, got %d", saved.Version())
	}
}

func TestUpdateProjectOverwriteInvalidSchemaVersion(t *testing.T) {
	repo := newMockRepo()
	service := NewDefenseProjectService(repo)

	existing := newExistingProject(t, "p1", "Original", nil)
	repo.projects["p1"] = existing

	projectJSON := `{
		"schemaVersion": 999, "projectName": "Site Alpha",
		"baseObject": {"id":"o1","name":"Obj","center":{"lat":55.75,"lng":37.61}},
		"layers": [], "assetLibrary": [], "placedObjects": []
	}`

	_, err := service.UpdateProject(context.Background(), "p1", "", "", projectJSON)
	if !errors.Is(err, domain.ErrInvalidSchemaVersion) {
		t.Errorf("expected ErrInvalidSchemaVersion, got %v", err)
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
	existing := newExistingProject(t, "p1", "Original", placed)
	repo.projects["p1"] = existing

	project, err := service.UpdateProject(context.Background(), "p1", "OnlyName", "", "")
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
