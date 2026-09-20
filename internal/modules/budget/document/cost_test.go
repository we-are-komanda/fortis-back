package document

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/fortis/backend/internal/modules/budget/domain"
	"github.com/stretchr/testify/require"
)

func TestCostFinanceGolden(t *testing.T) {
	raw, err := os.ReadFile("testdata/finance-v1.json")
	require.NoError(t, err)
	var fixture struct {
		Layers   []map[string]any `json:"layers"`
		Assets   []map[string]any `json:"assets"`
		VariantA struct {
			Objects  []map[string]any  `json:"objects"`
			Expected string            `json:"expectedTotalMinor"`
			Groups   map[string]string `json:"expectedByLayer"`
		} `json:"variantA"`
		VariantB struct {
			Objects  []map[string]any  `json:"objects"`
			Expected string            `json:"expectedTotalMinor"`
			Groups   map[string]string `json:"expectedByLayer"`
		} `json:"variantB"`
	}
	require.NoError(t, json.Unmarshal(raw, &fixture))
	for _, asset := range fixture.Assets {
		asset["currency"] = "RUB"
		asset["category"] = "synthetic"
	}
	for _, variant := range []struct {
		name    string
		objects []map[string]any
		total   string
		groups  map[string]string
	}{
		{"A", fixture.VariantA.Objects, fixture.VariantA.Expected, fixture.VariantA.Groups},
		{"B", fixture.VariantB.Objects, fixture.VariantB.Expected, fixture.VariantB.Groups},
	} {
		t.Run(variant.name, func(t *testing.T) {
			body, err := json.Marshal(map[string]any{"assetLibrary": fixture.Assets, "layers": fixture.Layers, "placedObjects": variant.objects})
			require.NoError(t, err)
			projection, err := Calculate(string(body), domain.CalculationIdentity{})
			require.NoError(t, err)
			require.True(t, projection.IsComplete())
			require.Equal(t, variant.total, *projection.TotalMinor())
			require.Len(t, projection.Lines(), 3)
			for _, group := range projection.ByLayer() {
				require.Equal(t, variant.groups[group.ID()], *group.TotalMinor())
			}
		})
	}
}

func costInput() (map[string]any, map[string]any, map[string]any) {
	asset := map[string]any{"id": "asset", "name": "Module", "category": "synthetic", "currency": "RUB", "unitPriceMinor": "12500000"}
	object := map[string]any{"id": "object", "assetId": "asset", "layerId": "custom", "quantity": 2}
	return map[string]any{"layers": []any{map[string]any{"id": "custom", "name": "Custom layer"}}, "assetLibrary": []any{asset}, "placedObjects": []any{object}}, asset, object
}

func TestCostPriceBoundaries(t *testing.T) {
	cases := []struct {
		name, total, source string
		change              func(map[string]any, map[string]any, map[string]any)
	}{
		{"zero override", "0", "instance_override", func(p, a, o map[string]any) { o["customPriceMinor"] = "0" }},
		{"unknown override", "", "unknown", func(p, a, o map[string]any) { o["customPriceMinor"] = nil; o["customPricePerUnitMln"] = 99 }},
		{"unknown template", "", "unknown", func(p, a, o map[string]any) { a["unitPriceMinor"] = nil; a["pricePerUnitMln"] = 99 }},
		{"legacy decimal", "25000000", "template", func(p, a, o map[string]any) { delete(a, "unitPriceMinor"); a["pricePerUnitMln"] = json.Number("0.125") }},
		{"half up before quantity", "2", "template", func(p, a, o map[string]any) {
			delete(a, "unitPriceMinor")
			a["pricePerUnitMln"] = json.Number("0.000000005")
		}},
		{"scientific integer quantity", "25000000", "template", func(p, a, o map[string]any) { o["quantity"] = json.Number("2e0") }},
		{"decimal integer quantity", "25000000", "template", func(p, a, o map[string]any) { o["quantity"] = json.Number("2.0") }},
		{"missing asset override", "6", "instance_override", func(p, a, o map[string]any) { p["assetLibrary"] = []any{}; o["customPriceMinor"] = "3" }},
		{"missing asset unknown", "", "unknown", func(p, a, o map[string]any) { p["assetLibrary"] = []any{} }},
		{"components missing", "", "unknown", func(p, a, o map[string]any) { a["pricingMode"] = "components" }},
		{"override skips invalid bundle", "0", "instance_override", func(p, a, o map[string]any) {
			a["pricingMode"] = "invalid"
			a["unitPriceMinor"] = "-1"
			o["customPriceMinor"] = "0"
		}},
		{"override skips malformed components", "0", "instance_override", func(p, a, o map[string]any) { a["components"] = "ignored-malformed"; o["customPriceMinor"] = "0" }},
		{"hidden inactive remains", "25000000", "template", func(p, a, o map[string]any) { o["status"] = "inactive"; o["isVisibleOnMap"] = false }},
		{"maintenance remains", "25000000", "template", func(p, a, o map[string]any) { o["status"] = "maintenance" }},
		{"beyond JSON safe integer", "1000000000000000000000", "template", func(p, a, o map[string]any) { a["unitPriceMinor"] = "1000000000000000"; o["quantity"] = 1_000_000 }},
		{"leading zeros", "6", "template", func(p, a, o map[string]any) { a["unitPriceMinor"] = "0003" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, a, o := costInput()
			tc.change(p, a, o)
			raw, err := json.Marshal(p)
			require.NoError(t, err)
			got, err := Calculate(string(raw), domain.CalculationIdentity{})
			require.NoError(t, err)
			require.Len(t, got.Lines(), 1)
			require.Equal(t, tc.source, got.Lines()[0].PriceSource())
			if tc.total == "" {
				require.Nil(t, got.TotalMinor())
				require.False(t, got.IsComplete())
				require.Equal(t, "0", got.KnownSubtotalMinor())
			} else {
				require.Equal(t, tc.total, *got.TotalMinor())
			}
		})
	}
}

func TestCostCompoundGolden(t *testing.T) {
	raw, err := os.ReadFile("testdata/finance-v1.json")
	require.NoError(t, err)
	var fixture struct {
		Compound struct {
			Quantity   int              `json:"quantity"`
			Components []map[string]any `json:"components"`
			Bundle     string           `json:"ignoredBundlePriceMinor"`
			Expected   string           `json:"expectedTotalMinor"`
		} `json:"compound"`
	}
	require.NoError(t, json.Unmarshal(raw, &fixture))
	p, a, o := costInput()
	a["pricingMode"] = "components"
	a["components"] = fixture.Compound.Components
	a["unitPriceMinor"] = fixture.Compound.Bundle
	o["quantity"] = fixture.Compound.Quantity
	body, err := json.Marshal(p)
	require.NoError(t, err)
	got, err := Calculate(string(body), domain.CalculationIdentity{})
	require.NoError(t, err)
	require.Equal(t, fixture.Compound.Expected, *got.TotalMinor())
	require.Equal(t, "components", got.Lines()[0].PriceSource())
}

func TestCostRejectsInvalidFields(t *testing.T) {
	for _, tc := range []struct {
		name, code, field string
		change            func(map[string]any, map[string]any)
	}{
		{"negative price", "invalid_price", "assetLibrary.asset.unitPriceMinor", func(a, o map[string]any) { a["unitPriceMinor"] = "-1" }},
		{"nonnumeric price", "invalid_price", "assetLibrary.asset.unitPriceMinor", func(a, o map[string]any) { a["unitPriceMinor"] = "NaN" }},
		{"fractional price", "invalid_price", "assetLibrary.asset.unitPriceMinor", func(a, o map[string]any) { a["unitPriceMinor"] = "1.5" }},
		{"excess price", "invalid_price", "assetLibrary.asset.unitPriceMinor", func(a, o map[string]any) { a["unitPriceMinor"] = "1000000000000001" }},
		{"fractional quantity", "invalid_quantity", "placedObjects.object.quantity", func(a, o map[string]any) { o["quantity"] = json.Number("1.000000000000000000001") }},
		{"zero quantity", "invalid_quantity", "placedObjects.object.quantity", func(a, o map[string]any) { o["quantity"] = 0 }},
		{"wrong currency even override", "invalid_currency", "assetLibrary.asset.currency", func(a, o map[string]any) { a["currency"] = "USD"; o["customPriceMinor"] = "0" }},
		{"unknown pricing mode", "invalid_pricing_mode", "assetLibrary.asset.pricingMode", func(a, o map[string]any) { a["pricingMode"] = "template" }},
		{"null pricing mode", "invalid_pricing_mode", "assetLibrary.asset.pricingMode", func(a, o map[string]any) { a["pricingMode"] = nil }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, a, o := costInput()
			tc.change(a, o)
			raw, err := json.Marshal(p)
			require.NoError(t, err)
			_, err = Calculate(string(raw), domain.CalculationIdentity{})
			var fieldError *domain.CostValidationError
			require.ErrorAs(t, err, &fieldError)
			require.Equal(t, tc.code, fieldError.Code)
			require.Equal(t, tc.field, fieldError.Field)
		})
	}
}

func TestCostEmptyAndUnsupportedVersion(t *testing.T) {
	got, err := Calculate(`{"placedObjects":[]}`, domain.CalculationIdentity{})
	require.NoError(t, err)
	require.Empty(t, got.Lines())
	require.Empty(t, got.ByLayer())
	require.Empty(t, got.ByType())
	require.Equal(t, "0", *got.TotalMinor())
	_, err = Calculate(`{"placedObjects":[]}`, domain.NewCalculationIdentity("p", 1, "future-unsupported", "digest", nil))
	require.ErrorIs(t, err, domain.ErrUnsupportedCalculationVersion)
}
