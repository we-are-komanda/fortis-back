package infrastructure

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCatalogMetadataRoundTrip(t *testing.T) {
	id := uuid.New()
	raw := `{"id":"` + id.String() + `","name":"Synthetic","category":"radar","coverageType":"none","pricePerUnitMln":123,"unitPriceMinor":null,"pricingMode":"components","components":[{"id":"c","name":"Part","quantity":2,"unitPriceMinor":"125"}],"provenance":{"sourceLabel":"fixture","sourceDocumentId":null,"sourceUrl":null,"sourceDate":null,"recordedAt":"2026-09-20T00:00:00Z","recordedBy":"actor","quality":"demo","revision":"1"},"fieldProvenance":{}}`
	model := DefenseAssetModel{ID: id, AssetData: raw}
	a, err := model.ToDomain()
	require.NoError(t, err)
	require.True(t, a.HasUnitPriceMinor())
	require.Nil(t, a.UnitPriceMinor())
	require.Equal(t, "components", a.PricingMode())
	require.Len(t, a.Components(), 1)
	require.Equal(t, "demo", a.Provenance().Quality())
	saved, err := ToModel(a)
	require.NoError(t, err)
	b, err := saved.ToDomain()
	require.NoError(t, err)
	require.True(t, b.HasUnitPriceMinor())
	require.Nil(t, b.UnitPriceMinor())
	require.Equal(t, "125", *b.Components()[0].UnitPriceMinor())
	require.Equal(t, a.Provenance().RecordedBy(), b.Provenance().RecordedBy())
}
