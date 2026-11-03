package ui

import (
	"encoding/json"
	project "github.com/fortis/backend/internal/modules/defense_project/domain"
	"github.com/fortis/backend/internal/modules/report/domain"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

func renderFixture(t *testing.T, raw, view string) (string, error) {
	t.Helper()
	p, err := project.NewDefenseProject("project", "Saved variant", "enterprise", "Original project", project.NewProtectedObject("base", "Base", project.NewCoordinates(0, 0)), nil, nil, nil, nil, nil, nil, project.DefenseProjectModeView, project.DefenseProjectSourceCustom, nil, time.Now())
	require.NoError(t, err)
	p.SetVersion(7)
	p.SetDocument(raw)
	p.SetSnapshot(`{"inputDataVersions":{}}`, "digest")
	p.SetCostCalculationVersion("future-unsupported")
	k, err := domain.NewReportKey("project", 7, true, view)
	require.NoError(t, err)
	draft, err := domain.NewReportDraft("issued", time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC), k, p, nil, nil)
	require.NoError(t, err)
	return RenderReport(draft)
}

const reportFixture = `{"baseObject":{"id":"base","name":"Frozen site","center":{"lat":0,"lng":0,"altitude":0}},"price":"STRUCTURED_FINANCIAL_SENTINEL","layers":[{"id":"l","name":"Polygon","code":"L","isVisible":false,"color":"#123456","geometry":{"type":"polygon","coordinates":[{"lat":0,"lng":0},{"lat":0,"lng":1},{"lat":1,"lng":0}],"isClosed":true,"privateUrl":"STRUCTURED_FINANCIAL_SENTINEL"}},{"id":"f","name":"Freeform","code":"F","geometry":{"type":"freeform","points":[{"lat":1,"lng":2},{"lat":2,"lng":3}]}}],"assetLibrary":[{"id":"a","name":"Frozen item","category":"radar","unitLabel":"unit","description":"Plain text","unitPriceMinor":"987654321012345","provenance":{"sourceUrl":"STRUCTURED_FINANCIAL_SENTINEL"},"iconUrl":"STRUCTURED_FINANCIAL_SENTINEL","documentId":"STRUCTURED_FINANCIAL_SENTINEL","weaponSpec":{"caliber":"Frozen specification","sourceUrl":"STRUCTURED_FINANCIAL_SENTINEL","price":"STRUCTURED_FINANCIAL_SENTINEL"}}],"placedObjects":[{"id":"o","assetId":"a","layerId":"l","quantity":2,"coordinates":{"lat":0,"lng":0},"rotation":0,"scale":1,"isVisibleOnMap":true,"notes":"Text may say 100 RUB: manual review required","customUnitPriceMinor":"987654321012345","metadata":{"price":"STRUCTURED_FINANCIAL_SENTINEL"}}]}`

func TestUnpricedReportAllowlistAndSavedVisibility(t *testing.T) {
	raw, err := renderFixture(t, reportFixture, "all")
	require.NoError(t, err)
	require.NotContains(t, raw, "STRUCTURED_FINANCIAL_SENTINEL")
	require.NotContains(t, raw, "987654321012345")
	var got map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &got))
	for _, key := range []string{"cost", "budget", "sources", "assetLibrary", "structuralProfile"} {
		require.NotContains(t, got, key)
	}
	require.Contains(t, raw, "Frozen specification")
	require.Contains(t, raw, "100 RUB")
	require.Contains(t, raw, "manual_text_review_required")
	require.Nil(t, got["enterprise"].(map[string]any)["name"])
	inventory := got["objects"].([]any)
	require.Len(t, inventory, 1)
	geometry := got["mapSnapshotData"].(map[string]any)["layers"].([]any)[0].(map[string]any)["geometry"].(map[string]any)
	require.Len(t, geometry["coordinates"], 3)
	require.Equal(t, true, geometry["isClosed"])
	require.NotContains(t, geometry, "points")
	visibility := got["mapVisibility"].(map[string]any)
	require.Equal(t, []any{"l", "f"}, visibility["layerIds"])
	require.Equal(t, []any{"o"}, visibility["objectIds"])
	require.Equal(t, false, visibility["showCoverage"])
	selective, err := renderFixture(t, reportFixture, "saved")
	require.NoError(t, err)
	var saved map[string]any
	require.NoError(t, json.Unmarshal([]byte(selective), &saved))
	require.Equal(t, inventory, saved["objects"])
	require.Empty(t, saved["mapVisibility"].(map[string]any)["objectIds"])
	require.Equal(t, []any{"f"}, saved["mapVisibility"].(map[string]any)["layerIds"])
}
func TestReportNeverInventsMalformedGeometry(t *testing.T) {
	for _, geometry := range []string{`{"type":"circle","radiusM":1}`, `{"type":"circle","center":{"lat":0,"lng":0},"radiusM":-1}`, `{"type":"ring","center":{"lat":0,"lng":0},"minRadiusM":2,"maxRadiusM":1}`, `{"type":"polygon","coordinates":[{"lat":91,"lng":0}]}`, `{"type":"circle","center":{"lat":null,"lng":0},"radiusM":1}`, `{"type":"unknown"}`, `{"type":"polygon"}`, `{"type":"freeform"}`} {
		_, err := renderFixture(t, `{"baseObject":{"center":{"lat":0,"lng":0}},"layers":[{"geometry":`+geometry+`}]}`, "all")
		require.Error(t, err, geometry)
	}
}

func TestReportRejectsURLStyleAndPreservesEmptyGeometry(t *testing.T) {
	_, err := renderFixture(t, strings.Replace(reportFixture, "#123456", "url(https://private.example.test/secret)", 1), "all")
	require.ErrorIs(t, err, domain.ErrInvalidSnapshot)
	raw, err := renderFixture(t, `{"baseObject":{"id":"base","center":{"lat":0,"lng":0}},"layers":[{"id":"empty","geometry":{"type":"polygon","coordinates":[],"isClosed":false}},{"id":"empty-freeform","geometry":{"type":"freeform","points":[]}}]}`, "all")
	require.NoError(t, err)
	require.Contains(t, raw, `"coordinates":[]`)
	require.Contains(t, raw, `"points":[]`)
}

func TestReportRetainsSafeRGBStyle(t *testing.T) {
	raw, err := renderFixture(t, strings.Replace(reportFixture, "#123456", "rgba(1, 2, 3, 0.5)", 1), "all")
	require.NoError(t, err)
	require.Contains(t, raw, "rgba(1, 2, 3, 0.5)")
}
