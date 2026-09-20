package application

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/fortis/backend/internal/modules/defense_asset/domain"
	"github.com/google/uuid"
	"time"
)

type ProvenanceInput struct {
	SourceLabel                             string
	SourceDocumentID, SourceURL, SourceDate *string
	Quality                                 string
}
type PriceComponentInput struct {
	ID, Name       string
	Quantity       int
	UnitPriceMinor *string
}
type CatalogMetadataInput struct {
	UnitPriceMinor  json.RawMessage
	PricingMode     string
	Components      []PriceComponentInput
	Provenance      *ProvenanceInput
	FieldProvenance map[string]*ProvenanceInput
}

func optionalString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
func (s *DefenseAssetService) catalogProvenance(ctx context.Context, actor string, in *ProvenanceInput) (*domain.Provenance, error) {
	if in == nil {
		return nil, nil
	}
	docID, revision, checksum := optionalString(in.SourceDocumentID), "", ""
	if docID != "" {
		if _, err := uuid.Parse(docID); err != nil {
			return nil, domain.ErrInvalidCatalogMetadata
		}
		if s.documents == nil {
			return nil, domain.ErrInvalidCatalogMetadata
		}
		doc, err := s.documents.FindByID(ctx, docID)
		if err != nil {
			if errors.Is(err, domain.ErrDocumentNotFound) {
				return nil, domain.ErrInvalidCatalogMetadata
			}
			return nil, err
		}
		if _, err = s.GetByID(ctx, actor, doc.AssetID()); err != nil {
			return nil, err
		}
		if doc.Status() != "ready" || len(doc.Checksum()) != 64 {
			return nil, domain.ErrInvalidCatalogMetadata
		}
		revision, checksum = doc.Revision(), doc.Checksum()
	}
	return domain.NewProvenance(in.SourceLabel, docID, optionalString(in.SourceURL), optionalString(in.SourceDate), time.Now().UTC().Format(time.RFC3339Nano), actor, in.Quality, uuid.NewString(), revision, checksum)
}
func (s *DefenseAssetService) applyCatalogMetadata(ctx context.Context, actor string, a *domain.DefenseAsset, in CatalogMetadataInput) error {
	var price *string
	if len(in.UnitPriceMinor) > 0 {
		if err := json.Unmarshal(in.UnitPriceMinor, &price); err != nil {
			return domain.ErrInvalidCatalogMetadata
		}
	}
	var components []domain.PriceComponent
	if in.Components != nil {
		components = []domain.PriceComponent{}
	}
	for _, c := range in.Components {
		v, err := domain.NewPriceComponent(c.ID, c.Name, c.Quantity, c.UnitPriceMinor)
		if err != nil {
			return err
		}
		components = append(components, v)
	}
	if err := a.SetExactPrice(len(in.UnitPriceMinor) > 0, price, in.PricingMode, components); err != nil {
		return err
	}
	if in.Provenance == nil && in.FieldProvenance == nil {
		return nil
	}
	p := a.Provenance()
	var err error
	if in.Provenance != nil {
		p, err = s.catalogProvenance(ctx, actor, in.Provenance)
		if err != nil {
			return err
		}
	}
	var fields map[string]*domain.Provenance
	if in.FieldProvenance != nil {
		fields = map[string]*domain.Provenance{}
	}
	for k, v := range in.FieldProvenance {
		p, err := s.catalogProvenance(ctx, actor, v)
		if err != nil {
			return err
		}
		fields[k] = p
	}
	return a.SetProvenance(p, fields)
}
