package application

import (
	"context"
	"encoding/json"
	"github.com/fortis/backend/internal/modules/defense_asset/domain"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCatalogServerStampAndDemoPreservation(t *testing.T) {
	s := NewDefenseAssetService(newMockRepo(), testAccess{})
	date := "2026-09-20"
	in := CreateInput{EnterpriseID: testString("ent-1"), Name: "Synthetic", Category: "detection", CoverageType: "none", CatalogMetadataInput: CatalogMetadataInput{UnitPriceMinor: json.RawMessage("null"), Provenance: &ProvenanceInput{SourceLabel: "Fixture", SourceDate: &date, Quality: "demo"}}}
	a, err := s.Create(context.Background(), "actor", in)
	require.NoError(t, err)
	require.True(t, a.HasUnitPriceMinor())
	require.NotNil(t, a.Provenance())
	require.Equal(t, "actor", a.Provenance().RecordedBy())
	require.NotEmpty(t, a.Provenance().Revision())
	_, err = s.Update(context.Background(), "actor", UpdateInput{ID: a.ID(), CatalogMetadataInput: CatalogMetadataInput{Provenance: &ProvenanceInput{SourceLabel: "replacement", SourceDate: &date, Quality: "confirmed"}}})
	require.Error(t, err)
}

func TestCatalogRejectedUpdatePreservesLoadedAggregate(t *testing.T) {
	repo := newMockRepo()
	s := NewDefenseAssetService(repo, testAccess{})
	a, err := s.Create(context.Background(), "actor", CreateInput{EnterpriseID: testString("ent-1"), Name: "Original", Category: "detection", CoverageType: "none", DetectionSpec: &domain.DetectionSpecification{}})
	require.NoError(t, err)
	before := a.UpdatedAt()
	category := domain.DefenseAssetCategory("kinetic")
	name := "Rejected rename"
	_, err = s.Update(context.Background(), "actor", UpdateInput{ID: a.ID(), Name: &name, Category: &category})
	require.ErrorIs(t, err, domain.ErrDefenseAssetInvalidSpecification)
	require.Equal(t, "Original", a.Name())
	require.Equal(t, domain.DefenseAssetCategory("detection"), a.Category())
	require.Equal(t, before, a.UpdatedAt())
	require.Same(t, a, repo.assets[a.ID()])
}
