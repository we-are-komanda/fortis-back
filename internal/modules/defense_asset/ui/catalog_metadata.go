package ui

import (
	"encoding/json"
	"github.com/fortis/backend/internal/modules/defense_asset/application"
	"github.com/fortis/backend/internal/modules/defense_asset/domain"
)

type ProvenanceDTO struct {
	SourceLabel            string  `json:"sourceLabel"`
	SourceDocumentID       *string `json:"sourceDocumentId"`
	SourceURL              *string `json:"sourceUrl"`
	SourceDate             *string `json:"sourceDate"`
	RecordedAt             string  `json:"recordedAt"`
	RecordedBy             string  `json:"recordedBy"`
	Quality                string  `json:"quality"`
	Revision               string  `json:"revision"`
	SourceDocumentRevision *string `json:"sourceDocumentRevision"`
	SourceDocumentChecksum *string `json:"sourceDocumentChecksum"`
}
type PriceComponentDTO struct {
	ID             string  `json:"id,omitempty"`
	Name           string  `json:"name"`
	Quantity       int     `json:"quantity"`
	UnitPriceMinor *string `json:"unitPriceMinor"`
}
type CatalogMetadataDTO struct {
	// Цена в копейках RUB: отсутствие сохраняет legacy fallback, null означает неизвестную цену.
	// swagger:type string
	UnitPriceMinor  json.RawMessage           `json:"unitPriceMinor,omitempty"`
	PricingMode     string                    `json:"pricingMode,omitempty"`
	Components      []PriceComponentDTO       `json:"components,omitempty"`
	Provenance      *ProvenanceDTO            `json:"provenance"`
	FieldProvenance map[string]*ProvenanceDTO `json:"fieldProvenance,omitempty"`
}

func nullable(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}
func provenanceToData(p *domain.Provenance) *ProvenanceDTO {
	if p == nil {
		return nil
	}
	return &ProvenanceDTO{p.SourceLabel(), nullable(p.SourceDocumentID()), nullable(p.SourceURL()), nullable(p.SourceDate()), p.RecordedAt(), p.RecordedBy(), p.Quality(), p.Revision(), nullable(p.DocumentRevision()), nullable(p.DocumentChecksum())}
}
func metadataFromAsset(a *domain.DefenseAsset) CatalogMetadataDTO {
	d := CatalogMetadataDTO{PricingMode: a.PricingMode(), Provenance: provenanceToData(a.Provenance()), FieldProvenance: map[string]*ProvenanceDTO{}}
	if a.HasUnitPriceMinor() {
		d.UnitPriceMinor, _ = json.Marshal(a.UnitPriceMinor())
	}
	if a.Components() != nil {
		d.Components = []PriceComponentDTO{}
	}
	for _, c := range a.Components() {
		d.Components = append(d.Components, PriceComponentDTO{c.ID(), c.Name(), c.Quantity(), c.UnitPriceMinor()})
	}
	for k, v := range a.FieldProvenance() {
		d.FieldProvenance[k] = provenanceToData(v)
	}
	return d
}

func provenanceInput(p *ProvenanceDTO) *application.ProvenanceInput {
	if p == nil {
		return nil
	}
	return &application.ProvenanceInput{SourceLabel: p.SourceLabel, SourceDocumentID: p.SourceDocumentID, SourceURL: p.SourceURL, SourceDate: p.SourceDate, Quality: p.Quality}
}
func catalogInput(d CatalogMetadataDTO) application.CatalogMetadataInput {
	in := application.CatalogMetadataInput{UnitPriceMinor: d.UnitPriceMinor, PricingMode: d.PricingMode, Provenance: provenanceInput(d.Provenance)}
	if d.Components != nil {
		in.Components = []application.PriceComponentInput{}
	}
	for _, c := range d.Components {
		in.Components = append(in.Components, application.PriceComponentInput{ID: c.ID, Name: c.Name, Quantity: c.Quantity, UnitPriceMinor: c.UnitPriceMinor})
	}
	if d.FieldProvenance != nil {
		in.FieldProvenance = map[string]*application.ProvenanceInput{}
	}
	for k, v := range d.FieldProvenance {
		in.FieldProvenance[k] = provenanceInput(v)
	}
	return in
}
