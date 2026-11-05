package application

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/fortis/backend/internal/modules/defense_project/infrastructure"
)

// Game rules are a frontend-owned document extension, preserved by the existing lossless path.
func TestGameSettingsDocumentRoundTrip(t *testing.T) {
	var payload map[string]any
	if err := json.Unmarshal([]byte(validProjectJSON()), &payload); err != nil {
		t.Fatal(err)
	}
	settings := map[string]any{"intervalMs": float64(5000), "reactionMs": float64(2000), "magazineSize": float64(6), "ammunition": float64(0), "reloadMs": float64(15000), "maxTargets": float64(2)}
	payload["placedObjects"] = []any{map[string]any{"id": "game-unit", "assetId": "game-asset", "layerId": "game-layer", "quantity": 2, "status": "active", "coordinates": map[string]any{"lat": 0, "lng": 0}, "gameSettings": settings}}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	repo := newMockRepo()
	service := NewDefenseProjectService(repo, testAccess{})
	project, err := service.Import(context.Background(), "actor", string(raw))
	if err != nil {
		t.Fatal(err)
	}
	model, err := infrastructure.ToModel(project)
	if err != nil {
		t.Fatal(err)
	}
	reloaded, err := model.ToDomain()
	if err != nil {
		t.Fatal(err)
	}
	repo.projects[project.ProjectID()] = reloaded
	exported, err := service.Export(context.Background(), "actor", project.ProjectID())
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		PlacedObjects []struct{ GameSettings map[string]any }
	}
	if err := json.Unmarshal([]byte(exported), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.PlacedObjects) != 1 || !reflect.DeepEqual(result.PlacedObjects[0].GameSettings, settings) {
		t.Fatalf("game settings lost across import, JSONB mapping and export: %s", exported)
	}
}
