package infrastructure

import (
	"encoding/json"
	"github.com/fortis/backend/internal/modules/defense_asset/domain"
)

type provenanceData struct {
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
type componentData struct {
	ID             string  `json:"id,omitempty"`
	Name           string  `json:"name"`
	Quantity       int     `json:"quantity"`
	UnitPriceMinor *string `json:"unitPriceMinor"`
}
type catalogMetadataData struct {
	UnitPriceMinor  json.RawMessage            `json:"unitPriceMinor,omitempty"`
	PricingMode     string                     `json:"pricingMode,omitempty"`
	Components      []componentData            `json:"components,omitempty"`
	Provenance      *provenanceData            `json:"provenance"`
	FieldProvenance map[string]*provenanceData `json:"fieldProvenance,omitempty"`
}

func stringValue(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
func nullable(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}
func provenanceToData(p *domain.Provenance) *provenanceData {
	if p == nil {
		return nil
	}
	return &provenanceData{p.SourceLabel(), nullable(p.SourceDocumentID()), nullable(p.SourceURL()), nullable(p.SourceDate()), p.RecordedAt(), p.RecordedBy(), p.Quality(), p.Revision(), nullable(p.DocumentRevision()), nullable(p.DocumentChecksum())}
}
func (d *provenanceData) domain() (*domain.Provenance, error) {
	if d == nil {
		return nil, nil
	}
	return domain.NewProvenance(d.SourceLabel, stringValue(d.SourceDocumentID), stringValue(d.SourceURL), stringValue(d.SourceDate), d.RecordedAt, d.RecordedBy, d.Quality, d.Revision, stringValue(d.SourceDocumentRevision), stringValue(d.SourceDocumentChecksum))
}
func metadataFromAsset(a *domain.DefenseAsset) catalogMetadataData {
	d := catalogMetadataData{PricingMode: a.PricingMode(), Provenance: provenanceToData(a.Provenance()), FieldProvenance: map[string]*provenanceData{}}
	if a.HasUnitPriceMinor() {
		d.UnitPriceMinor, _ = json.Marshal(a.UnitPriceMinor())
	}
	if a.Components() != nil {
		d.Components = []componentData{}
	}
	for _, c := range a.Components() {
		d.Components = append(d.Components, componentData{c.ID(), c.Name(), c.Quantity(), c.UnitPriceMinor()})
	}
	for k, v := range a.FieldProvenance() {
		d.FieldProvenance[k] = provenanceToData(v)
	}
	return d
}
func (d catalogMetadataData) apply(a *domain.DefenseAsset) error {
	var price *string
	if len(d.UnitPriceMinor) > 0 {
		if err := json.Unmarshal(d.UnitPriceMinor, &price); err != nil {
			return domain.ErrInvalidCatalogMetadata
		}
	}
	var components []domain.PriceComponent
	if d.Components != nil {
		components = []domain.PriceComponent{}
	}
	for _, c := range d.Components {
		v, err := domain.NewPriceComponent(c.ID, c.Name, c.Quantity, c.UnitPriceMinor)
		if err != nil {
			return err
		}
		components = append(components, v)
	}
	if err := a.SetExactPrice(len(d.UnitPriceMinor) > 0, price, d.PricingMode, components); err != nil {
		return err
	}
	p, err := d.Provenance.domain()
	if err != nil {
		return err
	}
	fields := map[string]*domain.Provenance{}
	for k, v := range d.FieldProvenance {
		p, err := v.domain()
		if err != nil {
			return err
		}
		fields[k] = p
	}
	return a.SetProvenance(p, fields)
}
