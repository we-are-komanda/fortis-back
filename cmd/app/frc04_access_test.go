package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// These assertions fail if the current persistence adapters drop runtime data,
// or if an update can bypass the client's expected version.
func TestFRC04LosslessDocument(t *testing.T) {
	f := newAccessFixture(t)
	a := f.people[0]
	raw := frc04Document(a.e)
	var created struct {
		ProjectID string `json:"projectId"`
	}
	require.NoError(t, json.Unmarshal([]byte(f.request(t, a, "POST", "/projects", map[string]string{"name": "Lossless", "enterpriseId": a.e, "projectJson": raw}, 200)), &created))
	var original, loaded map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &original))
	require.NoError(t, json.Unmarshal([]byte(f.request(t, a, "GET", "/projects/export?id="+created.ProjectID, nil, 200)), &loaded))
	for _, key := range []string{"baseObject", "layers", "assetLibrary", "placedObjects", "customFields"} {
		require.Equal(t, original[key], loaded[key], key)
	}
}

func TestFRC04VersionRequired(t *testing.T) {
	f := newAccessFixture(t)
	a := f.people[0]
	body := f.request(t, a, "PUT", "/projects/update?id="+a.p, map[string]string{"name": "unconditional"}, 400)
	require.Contains(t, body, "version_required")
	require.Equal(t, 1, f.projects.projects[a.p].Version())
}

func frc04Document(enterpriseID string) string {
	return js(map[string]any{
		"schemaVersion": 1, "enterpriseId": enterpriseID, "projectName": "Synthetic revision", "mode": "view", "source": "custom",
		"baseObject":    map[string]any{"id": "facility", "name": "Synthetic facility", "center": map[string]any{"lat": 0.0, "lng": 0.0, "altitude": 17}, "customFields": map[string]any{"label": "preserve"}},
		"layers":        []any{map[string]any{"id": "custom-layer", "name": "Polygon", "code": "CUSTOM", "geometryType": "polygon", "geometry": map[string]any{"type": "polygon", "points": []any{map[string]any{"lat": 0, "lng": 0}, map[string]any{"lat": 0, "lng": 1}, map[string]any{"lat": 1, "lng": 0}}}, "isVisible": false, "isLocked": true, "notes": "layer note", "customFields": map[string]any{"number": 9007199254740993}}},
		"assetLibrary":  []any{map[string]any{"id": "asset", "name": "Synthetic asset", "category": "radar", "roles": []string{}, "currency": "RUB", "unitLabel": "piece", "coverageType": "circle", "coverageRadius": 1250, "maxEffectiveDistance": 2500, "compoundProfile": map[string]any{"personnel": []any{}, "customField": "compound extension"}, "customFields": map[string]any{"note": "asset note"}}},
		"placedObjects": []any{map[string]any{"id": "placement", "assetId": "asset", "layerId": "custom-layer", "coordinates": map[string]any{"lat": 0, "lng": 0}, "quantity": 2, "status": "planned", "notes": "note", "isVisible": false, "customPricePerUnitMln": 0, "instanceOverrides": map[string]any{"label": "instance"}}},
		"customFields":  map[string]any{"role": "ignored", "price": 999, "nested": []any{nil, false, "value"}},
	})
}
