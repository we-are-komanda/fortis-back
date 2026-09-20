// Package document adapts the preserved project document to exact cost values.
package document

import (
	"encoding/json"
	"math/big"
	"strconv"
	"strings"

	"github.com/fortis/backend/internal/modules/budget/domain"
)

type componentInput struct {
	Name      string          `json:"name"`
	Quantity  json.RawMessage `json:"quantity"`
	UnitPrice json.RawMessage `json:"unitPriceMinor"`
}
type assetInput struct {
	ID              string                     `json:"id"`
	Name            string                     `json:"name"`
	Category        string                     `json:"category"`
	Currency        string                     `json:"currency"`
	UnitPrice       json.RawMessage            `json:"unitPriceMinor"`
	LegacyPrice     json.RawMessage            `json:"pricePerUnitMln"`
	PricingMode     json.RawMessage            `json:"pricingMode"`
	Components      json.RawMessage            `json:"components"`
	Provenance      json.RawMessage            `json:"provenance"`
	FieldProvenance map[string]json.RawMessage `json:"fieldProvenance"`
}
type objectInput struct {
	ID              string                     `json:"id"`
	AssetID         string                     `json:"assetId"`
	LayerID         string                     `json:"layerId"`
	Name            string                     `json:"name"`
	Quantity        json.RawMessage            `json:"quantity"`
	CustomPrice     json.RawMessage            `json:"customPriceMinor"`
	LegacyPrice     json.RawMessage            `json:"customPricePerUnitMln"`
	FieldProvenance map[string]json.RawMessage `json:"fieldProvenance"`
}
type projectInput struct {
	Layers []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"layers"`
	Assets  []assetInput  `json:"assetLibrary"`
	Objects []objectInput `json:"placedObjects"`
}

func Calculate(raw string, identity domain.CalculationIdentity) (*domain.CostProjection, error) {
	switch identity.CalculationVersion() {
	case "", domain.CostCalculationV1:
		return calculateV1(raw, identity)
	default:
		return nil, domain.ErrUnsupportedCalculationVersion
	}
}

// Keep this dispatch target stable when a later algorithm is introduced.
func calculateV1(raw string, identity domain.CalculationIdentity) (*domain.CostProjection, error) {
	var project projectInput
	if err := json.Unmarshal([]byte(raw), &project); err != nil {
		return nil, invalid("invalid_price", "projectJson", "invalid financial project document", "")
	}
	assets := map[string]assetInput{}
	layers := map[string]string{}
	for _, asset := range project.Assets {
		assets[asset.ID] = asset
	}
	for _, layer := range project.Layers {
		layers[layer.ID] = layer.Name
	}
	lines := []domain.FinancialLine{}
	warnings := []domain.CostIssue{}
	for _, object := range project.Objects {
		objectField := "placedObjects." + object.ID
		quantity, err := readQuantity(object.Quantity, objectField+".quantity", object.ID)
		if err != nil {
			return nil, err
		}
		asset, exists := assets[object.AssetID]
		name, category := object.Name, asset.Category
		if name == "" {
			name = asset.Name
		}
		if name == "" {
			name = object.AssetID
		}
		if category == "" {
			category = "unknown"
		}
		if _, ok := layers[object.LayerID]; !ok {
			warnings = append(warnings, domain.NewCostIssue("missing_layer", "Referenced project layer is missing", object.ID))
		}
		price := domain.UnknownUnitPrice()
		source, provenance := "unknown", ""
		assetField := "assetLibrary." + asset.ID
		if !exists {
			warnings = append(warnings, domain.NewCostIssue("missing_asset", "Referenced project asset is missing", object.ID))
		}
		if exists && asset.Currency != "RUB" {
			return nil, invalid("invalid_currency", assetField+".currency", "currency must be RUB", object.ID)
		}
		if len(object.CustomPrice) > 0 || len(object.LegacyPrice) > 0 {
			field := objectField + ".customPriceMinor"
			provenance = readProvenance(object.FieldProvenance["customPriceMinor"])
			if len(object.CustomPrice) == 0 {
				field = objectField + ".customPricePerUnitMln"
				provenance = readProvenance(object.FieldProvenance["customPricePerUnitMln"])
			}
			price, err = readPrice(object.CustomPrice, object.LegacyPrice, field, object.ID, &warnings)
			source = "instance_override"
		} else if exists {
			mode := ""
			if len(asset.PricingMode) > 0 {
				if err := json.Unmarshal(asset.PricingMode, &mode); err != nil || (mode != "bundle" && mode != "components") {
					return nil, invalid("invalid_pricing_mode", assetField+".pricingMode", "pricingMode must be bundle or components", object.ID)
				}
			}
			switch mode {
			case "", "bundle":
				field := assetField + ".unitPriceMinor"
				provenance = readProvenance(asset.FieldProvenance["unitPriceMinor"])
				if len(asset.UnitPrice) == 0 {
					field = assetField + ".pricePerUnitMln"
					provenance = readProvenance(asset.FieldProvenance["pricePerUnitMln"])
				}
				price, err = readPrice(asset.UnitPrice, asset.LegacyPrice, field, object.ID, &warnings)
				source = "template"
			case "components":
				provenance = readProvenance(asset.FieldProvenance["components"])
				source = "components"
				var components []*componentInput
				if len(asset.Components) > 0 {
					if err := json.Unmarshal(asset.Components, &components); err != nil {
						return nil, invalid("invalid_price", assetField+".components", "components must be an array", object.ID)
					}
				}
				if len(components) == 0 {
					warnings = append(warnings, domain.NewCostIssue("missing_components", "Component prices are not defined", object.ID))
				}
				prices := make([]domain.UnitPrice, 0, len(components))
				quantities := make([]int, 0, len(components))
				for index, component := range components {
					field := assetField + ".components." + strconv.Itoa(index)
					if component == nil {
						return nil, invalid("invalid_price", field, "component must be an object", object.ID)
					}
					q, e := readQuantity(component.Quantity, field+".quantity", object.ID)
					if e != nil {
						return nil, e
					}
					p, e := readPrice(component.UnitPrice, nil, field+".unitPriceMinor", object.ID, &warnings)
					if e != nil {
						return nil, e
					}
					prices = append(prices, p)
					quantities = append(quantities, q)
				}
				price, err = domain.SumComponentPrices(prices, quantities)
				if err != nil {
					return nil, invalid("invalid_price", assetField+".components", "component unit price exceeds supported range", object.ID)
				}
			default:
				return nil, invalid("invalid_pricing_mode", assetField+".pricingMode", "pricingMode must be bundle or components", object.ID)
			}
			if provenance == "" {
				provenance = readProvenance(asset.Provenance)
			}
		}
		if err != nil {
			return nil, err
		}
		if price.Minor() == nil {
			warnings = append(warnings, domain.NewCostIssue("missing_price", "Unit price is unknown", object.ID))
			source = "unknown"
		}
		line, err := domain.NewFinancialLine(object.ID, object.AssetID, object.LayerID, category, name, quantity, price, source, provenance)
		if err != nil {
			return nil, invalid("invalid_quantity", objectField+".quantity", "quantity is outside supported range", object.ID)
		}
		lines = append(lines, line)
	}
	return domain.NewCostProjection(identity, lines, layers, warnings), nil
}

func readQuantity(raw json.RawMessage, field, objectID string) (int, error) {
	value := string(raw)
	// Bound the exponent before big.Rat so untrusted JSON cannot allocate a
	// huge power of ten. Larger powers cannot represent a supported quantity.
	if pos := strings.IndexAny(value, "eE"); pos >= 0 {
		exponent, err := strconv.ParseInt(value[pos+1:], 10, 64)
		if err != nil || exponent > int64(len(value))+6 || exponent < -int64(len(value)) {
			return 0, invalid("invalid_quantity", field, "quantity must be an integer from 1 to 1000000", objectID)
		}
	}
	n, ok := new(big.Rat).SetString(value)
	if !ok || !n.IsInt() || !n.Num().IsInt64() || n.Num().Int64() < 1 || n.Num().Int64() > 1_000_000 {
		return 0, invalid("invalid_quantity", field, "quantity must be an integer from 1 to 1000000", objectID)
	}
	return int(n.Num().Int64()), nil
}
func readPrice(minor, legacy json.RawMessage, field, objectID string, warnings *[]domain.CostIssue) (domain.UnitPrice, error) {
	if len(minor) > 0 {
		if string(minor) == "null" {
			return domain.UnknownUnitPrice(), nil
		}
		var value string
		if err := json.Unmarshal(minor, &value); err != nil {
			return domain.UnitPrice{}, invalid("invalid_price", field, "minor price must be a decimal integer string or null", objectID)
		}
		p, err := domain.ParseUnitMinor(value)
		if err != nil {
			return domain.UnitPrice{}, invalid("invalid_price", field, "unit price must be from 0 to 1000000000000000 minor", objectID)
		}
		return p, nil
	}
	if len(legacy) == 0 || string(legacy) == "null" {
		return domain.UnknownUnitPrice(), nil
	}
	p, rounded, err := domain.ParseLegacyMln(string(legacy))
	if err != nil {
		return domain.UnitPrice{}, invalid("invalid_price", field, "legacy price must be a nonnegative decimal within supported range", objectID)
	}
	*warnings = append(*warnings, domain.NewCostIssue("legacy_price_converted", "Legacy million-ruble price converted to minor units", objectID))
	if rounded {
		*warnings = append(*warnings, domain.NewCostIssue("legacy_price_rounded", "Legacy unit price rounded half-up to one minor unit", objectID))
	}
	return p, nil
}
func invalid(code, field, message, objectID string) error {
	ids := []string{}
	if objectID != "" {
		ids = append(ids, objectID)
	}
	return &domain.CostValidationError{Code: code, Field: field, Message: message, ObjectIDs: ids}
}
func readProvenance(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var value map[string]json.RawMessage
	if err := json.Unmarshal(raw, &value); err != nil || value == nil {
		return ""
	}
	filtered := map[string]json.RawMessage{}
	for _, key := range []string{"sourceLabel", "sourceDocumentId", "sourceDocumentRevision", "sourceDocumentChecksum", "sourceUrl", "sourceDate", "recordedAt", "recordedBy", "quality", "revision"} {
		if v, ok := value[key]; ok {
			filtered[key] = v
		}
	}
	encoded, err := json.Marshal(filtered)
	if err != nil {
		return ""
	}
	return string(encoded)
}
