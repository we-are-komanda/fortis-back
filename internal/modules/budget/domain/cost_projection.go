package domain

import (
	"errors"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

const CostCalculationV1 = "cost-rub-v1"

var ErrUnsupportedCalculationVersion = errors.New("unsupported calculation version")
var ErrIncompleteCost = errors.New("project cost is incomplete")
var ErrInvalidCost = errors.New("invalid financial input")

type CostValidationError struct {
	Code, Field, Message string
	ObjectIDs            []string
}

func (e *CostValidationError) Error() string { return e.Field + ": " + e.Message }
func (e *CostValidationError) Unwrap() error { return ErrInvalidCost }

type CalculationIdentity struct {
	projectID                          string
	projectVersion                     int
	calculationVersion, snapshotDigest string
	inputDataVersions                  map[string]string
}

func NewCalculationIdentity(id string, version int, calculationVersion, digest string, inputs map[string]string) CalculationIdentity {
	if inputs == nil {
		inputs = map[string]string{}
	}
	return CalculationIdentity{id, version, calculationVersion, digest, inputs}
}
func (i CalculationIdentity) ProjectID() string          { return i.projectID }
func (i CalculationIdentity) ProjectVersion() int        { return i.projectVersion }
func (i CalculationIdentity) CalculationVersion() string { return i.calculationVersion }
func (i CalculationIdentity) SnapshotDigest() string     { return i.snapshotDigest }
func (i CalculationIdentity) InputDataVersions() map[string]string {
	copy := make(map[string]string, len(i.inputDataVersions))
	for k, v := range i.inputDataVersions {
		copy[k] = v
	}
	return copy
}

// UnitPrice distinguishes an unknown price from an explicit zero.
type UnitPrice struct {
	minor string
	known bool
}

func UnknownUnitPrice() UnitPrice { return UnitPrice{} }
func (p UnitPrice) Minor() *string {
	if !p.known {
		return nil
	}
	s := p.minor
	return &s
}

var minorPattern = regexp.MustCompile(`^[0-9]+$`)
var decimalPattern = regexp.MustCompile(`^(-?)(0|[1-9][0-9]*)(?:\.([0-9]+))?(?:[eE]([+-]?[0-9]+))?$`)
var maxUnitMinor = big.NewInt(1_000_000_000_000_000)

func ParseUnitMinor(value string) (UnitPrice, error) {
	if !minorPattern.MatchString(value) {
		return UnitPrice{}, ErrInvalidCost
	}
	value = strings.TrimLeft(value, "0")
	if value == "" {
		value = "0"
	}
	if len(value) > 16 {
		return UnitPrice{}, ErrInvalidCost
	}
	n, ok := new(big.Int).SetString(value, 10)
	if !ok || n.Cmp(maxUnitMinor) > 0 {
		return UnitPrice{}, ErrInvalidCost
	}
	return UnitPrice{n.String(), true}, nil
}

// ParseLegacyMln consumes the JSON decimal spelling and rounds half-up exactly
// once at the unit boundary. It never multiplies a binary floating point value.
func ParseLegacyMln(value string) (UnitPrice, bool, error) {
	m := decimalPattern.FindStringSubmatch(value)
	if m == nil {
		return UnitPrice{}, false, ErrInvalidCost
	}
	digits := strings.TrimLeft(m[2]+m[3], "0")
	if digits == "" {
		return UnitPrice{"0", true}, false, nil
	}
	if m[1] == "-" {
		return UnitPrice{}, false, ErrInvalidCost
	}
	exp := int64(0)
	if m[4] != "" {
		var err error
		exp, err = strconv.ParseInt(m[4], 10, 64)
		if err != nil {
			if strings.HasPrefix(m[4], "-") {
				return UnitPrice{"0", true}, true, nil
			}
			return UnitPrice{}, false, ErrInvalidCost
		}
	}
	// Huge exponents can be resolved without allocating huge powers of ten.
	if exp > 10000000 {
		return UnitPrice{}, false, ErrInvalidCost
	}
	if exp < -10000000 {
		return UnitPrice{"0", true}, true, nil
	}
	shift := int64(8-len(m[3])) + exp
	var integer, discarded string
	if shift >= 0 {
		if int64(len(digits))+shift > 16 {
			return UnitPrice{}, false, ErrInvalidCost
		}
		integer = digits + strings.Repeat("0", int(shift))
	} else {
		cut := int64(len(digits)) + shift
		if cut < 0 {
			return UnitPrice{"0", true}, true, nil
		}
		if cut > 16 {
			return UnitPrice{}, false, ErrInvalidCost
		}
		integer = digits[:int(cut)]
		discarded = digits[int(cut):]
		if integer == "" {
			integer = "0"
		}
	}
	n, _ := new(big.Int).SetString(integer, 10)
	rounded := strings.Trim(discarded, "0") != ""
	if discarded != "" && discarded[0] >= '5' {
		n.Add(n, big.NewInt(1))
	}
	if n.Cmp(maxUnitMinor) > 0 {
		return UnitPrice{}, false, ErrInvalidCost
	}
	return UnitPrice{n.String(), true}, rounded, nil
}

func SumComponentPrices(prices []UnitPrice, quantities []int) (UnitPrice, error) {
	if len(prices) == 0 {
		return UnknownUnitPrice(), nil
	}
	sum := new(big.Int)
	unknown := false
	for i, p := range prices {
		if quantities[i] < 1 || quantities[i] > 1_000_000 {
			return UnitPrice{}, ErrInvalidCost
		}
		if !p.known {
			unknown = true
			continue
		}
		n, _ := new(big.Int).SetString(p.minor, 10)
		n.Mul(n, big.NewInt(int64(quantities[i])))
		sum.Add(sum, n)
	}
	if sum.Cmp(maxUnitMinor) > 0 {
		return UnitPrice{}, ErrInvalidCost
	}
	if unknown {
		return UnknownUnitPrice(), nil
	}
	return UnitPrice{sum.String(), true}, nil
}

type CostIssue struct {
	code, severity, message string
	objectIDs               []string
}

func NewCostIssue(code, message string, objectIDs ...string) CostIssue {
	return CostIssue{code, "warning", message, objectIDs}
}
func (i CostIssue) Code() string        { return i.code }
func (i CostIssue) Severity() string    { return i.severity }
func (i CostIssue) Message() string     { return i.message }
func (i CostIssue) ObjectIDs() []string { return append([]string{}, i.objectIDs...) }

type FinancialLine struct {
	objectID, assetID, layerID, category, name string
	quantity                                   int
	unitPriceMinor, lineTotalMinor             *string
	priceSource, provenance                    string
}

func NewFinancialLine(objectID, assetID, layerID, category, name string, quantity int, price UnitPrice, source, provenance string) (FinancialLine, error) {
	if quantity < 1 || quantity > 1_000_000 {
		return FinancialLine{}, ErrInvalidCost
	}
	line := FinancialLine{objectID: objectID, assetID: assetID, layerID: layerID, category: category, name: name, quantity: quantity, unitPriceMinor: price.Minor(), priceSource: source, provenance: provenance}
	if price.known {
		n, _ := new(big.Int).SetString(price.minor, 10)
		n.Mul(n, big.NewInt(int64(quantity)))
		s := n.String()
		line.lineTotalMinor = &s
	} else {
		line.priceSource = "unknown"
	}
	return line, nil
}
func (l FinancialLine) ObjectID() string        { return l.objectID }
func (l FinancialLine) AssetID() string         { return l.assetID }
func (l FinancialLine) LayerID() string         { return l.layerID }
func (l FinancialLine) Category() string        { return l.category }
func (l FinancialLine) Name() string            { return l.name }
func (l FinancialLine) Quantity() int           { return l.quantity }
func (l FinancialLine) UnitPriceMinor() *string { return copyMinor(l.unitPriceMinor) }
func (l FinancialLine) LineTotalMinor() *string { return copyMinor(l.lineTotalMinor) }
func (l FinancialLine) PriceSource() string     { return l.priceSource }
func (l FinancialLine) Provenance() string      { return l.provenance }
func copyMinor(p *string) *string {
	if p == nil {
		return nil
	}
	s := *p
	return &s
}

type CostGroup struct {
	id, name               string
	objectCount, unitCount int
	knownSubtotalMinor     string
	totalMinor             *string
}

func (g CostGroup) ID() string                 { return g.id }
func (g CostGroup) Name() string               { return g.name }
func (g CostGroup) ObjectCount() int           { return g.objectCount }
func (g CostGroup) UnitCount() int             { return g.unitCount }
func (g CostGroup) KnownSubtotalMinor() string { return g.knownSubtotalMinor }
func (g CostGroup) TotalMinor() *string        { return copyMinor(g.totalMinor) }

type CostProjection struct {
	identity              CalculationIdentity
	lines                 []FinancialLine
	byLayer, byType       []CostGroup
	knownSubtotalMinor    string
	totalMinor            *string
	unknownPriceObjectIDs []string
	warnings              []CostIssue
}

func NewCostProjection(identity CalculationIdentity, lines []FinancialLine, layerNames map[string]string, warnings []CostIssue) *CostProjection {
	p := &CostProjection{identity: identity, lines: lines, warnings: warnings, unknownPriceObjectIDs: []string{}}
	if p.lines == nil {
		p.lines = []FinancialLine{}
	}
	if p.warnings == nil {
		p.warnings = []CostIssue{}
	}
	sum := new(big.Int)
	for _, line := range lines {
		if line.lineTotalMinor == nil {
			p.unknownPriceObjectIDs = append(p.unknownPriceObjectIDs, line.objectID)
		} else {
			n, _ := new(big.Int).SetString(*line.lineTotalMinor, 10)
			sum.Add(sum, n)
		}
	}
	p.knownSubtotalMinor = sum.String()
	if len(p.unknownPriceObjectIDs) == 0 {
		p.totalMinor = copyMinor(&p.knownSubtotalMinor)
	}
	p.byLayer = costGroups(lines, func(l FinancialLine) (string, string) {
		name, ok := layerNames[l.layerID]
		if !ok {
			name = l.layerID
		}
		return l.layerID, name
	})
	p.byType = costGroups(lines, func(l FinancialLine) (string, string) { return l.category, l.category })
	return p
}
func costGroups(lines []FinancialLine, key func(FinancialLine) (string, string)) []CostGroup {
	groups := []CostGroup{}
	indices := map[string]int{}
	for _, line := range lines {
		id, name := key(line)
		index, ok := indices[id]
		if !ok {
			index = len(groups)
			indices[id] = index
			zero := "0"
			groups = append(groups, CostGroup{id: id, name: name, knownSubtotalMinor: "0", totalMinor: &zero})
		}
		g := &groups[index]
		g.objectCount++
		g.unitCount += line.quantity
		if line.lineTotalMinor == nil {
			g.totalMinor = nil
			continue
		}
		sum, _ := new(big.Int).SetString(g.knownSubtotalMinor, 10)
		n, _ := new(big.Int).SetString(*line.lineTotalMinor, 10)
		sum.Add(sum, n)
		g.knownSubtotalMinor = sum.String()
		if g.totalMinor != nil {
			g.totalMinor = copyMinor(&g.knownSubtotalMinor)
		}
	}
	return groups
}
func (p *CostProjection) Identity() CalculationIdentity { return p.identity }
func (p *CostProjection) Lines() []FinancialLine        { return append([]FinancialLine{}, p.lines...) }
func (p *CostProjection) ByLayer() []CostGroup          { return append([]CostGroup{}, p.byLayer...) }
func (p *CostProjection) ByType() []CostGroup           { return append([]CostGroup{}, p.byType...) }
func (p *CostProjection) KnownSubtotalMinor() string    { return p.knownSubtotalMinor }
func (p *CostProjection) TotalMinor() *string           { return copyMinor(p.totalMinor) }
func (p *CostProjection) IsComplete() bool              { return p.totalMinor != nil }
func (p *CostProjection) UnknownPriceObjectIDs() []string {
	return append([]string{}, p.unknownPriceObjectIDs...)
}
func (p *CostProjection) Warnings() []CostIssue { return append([]CostIssue{}, p.warnings...) }

// Restore functions read the immutable persisted result without recalculation.
func RestoreFinancialLine(objectID, assetID, layerID, category, name string, quantity int, unit, total *string, source, provenance string) FinancialLine {
	return FinancialLine{objectID, assetID, layerID, category, name, quantity, copyMinor(unit), copyMinor(total), source, provenance}
}
func RestoreCostGroup(id, name string, objects, units int, known string, total *string) CostGroup {
	return CostGroup{id, name, objects, units, known, copyMinor(total)}
}
func RestoreCostIssue(code, severity, message string, ids []string) CostIssue {
	return CostIssue{code, severity, message, append([]string{}, ids...)}
}
func RestoreCostProjection(identity CalculationIdentity, lines []FinancialLine, layers, types []CostGroup, known string, total *string, unknown []string, warnings []CostIssue) *CostProjection {
	return &CostProjection{identity, lines, layers, types, known, copyMinor(total), unknown, warnings}
}
