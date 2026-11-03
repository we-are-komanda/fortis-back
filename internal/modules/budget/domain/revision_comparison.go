package domain

import (
	"errors"
	"math/big"
	"sort"
	"time"
)

var ErrIncompatibleProjects = errors.New("comparison requires distinct revisions of the same enterprise")

type ComparedRevision struct {
	name      string
	updatedAt time.Time
	budget    *BudgetProjection
}

func NewComparedRevision(name string, updatedAt time.Time, budget *BudgetProjection) ComparedRevision {
	return ComparedRevision{name, updatedAt, budget}
}
func (s ComparedRevision) Name() string              { return s.name }
func (s ComparedRevision) UpdatedAt() time.Time      { return s.updatedAt }
func (s ComparedRevision) Budget() *BudgetProjection { return s.budget }
func (s ComparedRevision) Cost() *CostProjection     { return s.budget.cost }

type RevisionComparison struct {
	enterpriseID                       string
	a, b                               ComparedRevision
	differenceMinor, differencePercent *string
	warnings                           []CostIssue
}

func NewRevisionComparison(enterpriseID string, a, b ComparedRevision) *RevisionComparison {
	c := &RevisionComparison{enterpriseID: enterpriseID, a: a, b: b, warnings: []CostIssue{}}
	av, bv := a.Cost().TotalMinor(), b.Cost().TotalMinor()
	if av != nil && bv != nil {
		ai, _ := new(big.Int).SetString(*av, 10)
		bi, _ := new(big.Int).SetString(*bv, 10)
		delta := new(big.Int).Sub(bi, ai)
		value := delta.String()
		c.differenceMinor = &value
		if ai.Sign() > 0 {
			percent := new(big.Rat).SetFrac(new(big.Int).Mul(delta, big.NewInt(100)), ai).FloatString(2)
			if percent == "-0.00" {
				percent = "0.00"
			}
			c.differencePercent = &percent
		}
	} else {
		c.warnings = append(c.warnings, NewCostIssue("incomplete_cost", "A complete cost difference cannot be calculated", uniqueIDs(append(a.Cost().UnknownPriceObjectIDs(), b.Cost().UnknownPriceObjectIDs()...))...))
	}
	keys := func(lines []FinancialLine, selectKey func(FinancialLine) string) map[string]bool {
		result := map[string]bool{}
		for _, line := range lines {
			result[selectKey(line)] = true
		}
		return result
	}
	differs := func(left, right map[string]bool) bool {
		if len(left) != len(right) {
			return true
		}
		for key := range left {
			if !right[key] {
				return true
			}
		}
		return false
	}
	if differs(keys(a.Cost().Lines(), func(l FinancialLine) string { return l.LayerID() }), keys(b.Cost().Lines(), func(l FinancialLine) string { return l.LayerID() })) {
		c.warnings = append(c.warnings, NewCostIssue("unmatched_layer_ids", "Layer IDs differ; rows are not matched by name"))
	}
	if differs(keys(a.Cost().Lines(), func(l FinancialLine) string { return l.AssetID() }), keys(b.Cost().Lines(), func(l FinancialLine) string { return l.AssetID() })) {
		c.warnings = append(c.warnings, NewCostIssue("unmatched_asset_ids", "Asset IDs differ; rows are not matched by name"))
	}
	missing := []string{}
	sources := func(lines []FinancialLine) map[string]map[string]bool {
		result := map[string]map[string]bool{}
		for _, line := range lines {
			if line.UnitPriceMinor() == nil {
				continue
			}
			if line.Provenance() == "" || line.Provenance() == "{}" {
				missing = append(missing, line.ObjectID())
			}
			if result[line.AssetID()] == nil {
				result[line.AssetID()] = map[string]bool{}
			}
			result[line.AssetID()][line.PriceSource()+"\n"+line.Provenance()] = true
		}
		return result
	}
	sa, sb := sources(a.Cost().Lines()), sources(b.Cost().Lines())
	if len(missing) > 0 {
		c.warnings = append(c.warnings, NewCostIssue("price_provenance_missing", "Some priced objects have no recorded price provenance", uniqueIDs(missing)...))
	}
	for asset, left := range sa {
		if right, ok := sb[asset]; ok && differs(left, right) {
			c.warnings = append(c.warnings, NewCostIssue("price_provenance_differs", "Price sources, dates or provenance differ for matching assets"))
			break
		}
	}
	return c
}
func uniqueIDs(ids []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}
	sort.Strings(result)
	return result
}
func (c *RevisionComparison) EnterpriseID() string       { return c.enterpriseID }
func (c *RevisionComparison) A() ComparedRevision        { return c.a }
func (c *RevisionComparison) B() ComparedRevision        { return c.b }
func (c *RevisionComparison) DifferenceMinor() *string   { return copyMinor(c.differenceMinor) }
func (c *RevisionComparison) DifferencePercent() *string { return copyMinor(c.differencePercent) }
func (c *RevisionComparison) Warnings() []CostIssue      { return append([]CostIssue{}, c.warnings...) }
