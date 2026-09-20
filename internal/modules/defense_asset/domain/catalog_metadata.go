package domain

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrInvalidCatalogMetadata = errors.New("invalid catalog provenance or price")

func ValidateMinor(value *string) error {
	if value == nil {
		return nil
	}
	if *value == "" || len(*value) > 16 {
		return ErrInvalidCatalogMetadata
	}
	for _, r := range *value {
		if r < '0' || r > '9' {
			return ErrInvalidCatalogMetadata
		}
	}
	n, err := strconv.ParseUint(*value, 10, 64)
	if err != nil || n > 1000000000000000 {
		return ErrInvalidCatalogMetadata
	}
	return nil
}

type Provenance struct {
	sourceLabel, sourceDocumentID, sourceURL, sourceDate, recordedAt, recordedBy, quality, revision, documentRevision, documentChecksum string
}

func NewProvenance(label, documentID, sourceURL, sourceDate, recordedAt, recordedBy, quality, revision, documentRevision, documentChecksum string) (*Provenance, error) {
	label = strings.TrimSpace(label)
	if utf8.RuneCountInString(label) > 1000 || len(sourceURL) > 2048 {
		return nil, ErrInvalidCatalogMetadata
	}
	if quality != "demo" && quality != "estimated" && quality != "confirmed" {
		return nil, ErrInvalidCatalogMetadata
	}
	if sourceURL != "" {
		u, err := url.Parse(sourceURL)
		if err != nil || u.Host == "" || u.User != nil || (u.Scheme != "https" && u.Scheme != "http") || strings.ContainsAny(sourceURL, "\r\n") {
			return nil, ErrInvalidCatalogMetadata
		}
	}
	if sourceDate != "" {
		if _, err := time.Parse("2006-01-02", sourceDate); err != nil {
			return nil, ErrInvalidCatalogMetadata
		}
	}
	if quality == "confirmed" && (sourceDate == "" || (label == "" && sourceURL == "" && documentID == "")) {
		return nil, ErrInvalidCatalogMetadata
	}
	return &Provenance{label, documentID, sourceURL, sourceDate, recordedAt, recordedBy, quality, revision, documentRevision, documentChecksum}, nil
}
func (p *Provenance) Quality() string          { return p.quality }
func (p *Provenance) SourceLabel() string      { return p.sourceLabel }
func (p *Provenance) SourceDocumentID() string { return p.sourceDocumentID }
func (p *Provenance) SourceURL() string        { return p.sourceURL }
func (p *Provenance) SourceDate() string       { return p.sourceDate }
func (p *Provenance) RecordedAt() string       { return p.recordedAt }
func (p *Provenance) RecordedBy() string       { return p.recordedBy }
func (p *Provenance) Revision() string         { return p.revision }
func (p *Provenance) DocumentRevision() string { return p.documentRevision }
func (p *Provenance) DocumentChecksum() string { return p.documentChecksum }

type PriceComponent struct {
	id, name       string
	quantity       int
	unitPriceMinor *string
}

func NewPriceComponent(id, name string, quantity int, price *string) (PriceComponent, error) {
	if strings.TrimSpace(name) == "" || len(name) > 500 || quantity < 1 || quantity > 1000000 {
		return PriceComponent{}, ErrInvalidCatalogMetadata
	}
	if err := ValidateMinor(price); err != nil {
		return PriceComponent{}, err
	}
	return PriceComponent{id, name, quantity, price}, nil
}
func (c PriceComponent) ID() string              { return c.id }
func (c PriceComponent) Name() string            { return c.name }
func (c PriceComponent) Quantity() int           { return c.quantity }
func (c PriceComponent) UnitPriceMinor() *string { return c.unitPriceMinor }

func (a *DefenseAsset) HasUnitPriceMinor() bool                 { return a.hasUnitPriceMinor }
func (a *DefenseAsset) UnitPriceMinor() *string                 { return a.unitPriceMinor }
func (a *DefenseAsset) PricingMode() string                     { return a.pricingMode }
func (a *DefenseAsset) Components() []PriceComponent            { return a.components }
func (a *DefenseAsset) Provenance() *Provenance                 { return a.provenance }
func (a *DefenseAsset) FieldProvenance() map[string]*Provenance { return a.fieldProvenance }
func (a *DefenseAsset) SetExactPrice(present bool, price *string, mode string, components []PriceComponent) error {
	if err := ValidateMinor(price); err != nil {
		return err
	}
	if mode != "" && mode != "bundle" && mode != "components" {
		return ErrInvalidCatalogMetadata
	}
	if len(components) > 1000 {
		return ErrInvalidCatalogMetadata
	}
	if present {
		a.hasUnitPriceMinor = true
		a.unitPriceMinor = price
	}
	if mode != "" {
		a.pricingMode = mode
	}
	if components != nil {
		a.components = components
	}
	return nil
}
func (a *DefenseAsset) SetProvenance(p *Provenance, fields map[string]*Provenance) error {
	if a.provenance != nil && a.provenance.Quality() == "demo" && (p == nil || p.Quality() != "demo") {
		return ErrInvalidCatalogMetadata
	}
	if len(fields) > 100 {
		return ErrInvalidCatalogMetadata
	}
	for k, v := range fields {
		if k == "" || len(k) > 100 || v == nil {
			return ErrInvalidCatalogMetadata
		}
		if old := a.fieldProvenance[k]; old != nil && old.Quality() == "demo" && v.Quality() != "demo" {
			return ErrInvalidCatalogMetadata
		}
	}
	a.provenance = p
	if fields != nil {
		for k, v := range a.fieldProvenance {
			if v.Quality() == "demo" && fields[k] == nil {
				fields[k] = v
			}
		}
		a.fieldProvenance = fields
	}
	return nil
}
