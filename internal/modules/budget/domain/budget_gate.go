package domain

import (
	"math/big"
	"strconv"
)

// ParseLegacyBudgetMln converts the original JSON decimal directly to kopecks.
// The exponent bound limits pathological legacy input allocation, not budget size.
func ParseLegacyBudgetMln(value string) (string, error) {
	parts := decimalPattern.FindStringSubmatch(value)
	if parts == nil || len(value) > 10000 {
		return "", ErrInvalidBudgetAmount
	}
	if parts[4] != "" {
		exponent, err := strconv.ParseInt(parts[4], 10, 64)
		if err != nil || exponent > 10000 || exponent < -10000 {
			return "", ErrInvalidBudgetAmount
		}
	}
	amount, ok := new(big.Rat).SetString(value)
	if !ok || amount.Sign() < 0 {
		return "", ErrInvalidBudgetAmount
	}
	amount.Mul(amount, big.NewRat(100_000_000, 1))
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(amount.Num(), amount.Denom(), remainder)
	if remainder.Mul(remainder, big.NewInt(2)).Cmp(amount.Denom()) >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	return quotient.String(), nil
}

type BudgetProjection struct {
	config    *BudgetConfig
	cost      *CostProjection
	remaining *string
	status    string
}

func NewBudgetProjection(config *BudgetConfig, cost *CostProjection) *BudgetProjection {
	p := &BudgetProjection{config: config, cost: cost, status: "indeterminate"}
	if !cost.IsComplete() {
		return p
	}
	if config.BudgetMode() == BudgetModeUnlimited {
		p.status = "unlimited"
		return p
	}
	limit, _ := new(big.Int).SetString(*config.BudgetLimitMinor(), 10)
	spent, _ := new(big.Int).SetString(*cost.TotalMinor(), 10)
	remaining := limit.Sub(limit, spent).String()
	p.remaining = &remaining
	p.status = "within_limit"
	if limit.Sign() < 0 {
		p.status = "exceeded"
	}
	return p
}
func (p *BudgetProjection) Identity() CalculationIdentity { return p.cost.Identity() }
func (p *BudgetProjection) Config() *BudgetConfig         { return p.config }
func (p *BudgetProjection) SpentMinor() *string           { return p.cost.TotalMinor() }
func (p *BudgetProjection) RemainingMinor() *string       { return copyMinor(p.remaining) }
func (p *BudgetProjection) Status() string                { return p.status }

type CandidateBudgetCheck struct {
	projection *BudgetProjection
	current    *string
	reason     string
}

func CheckCandidateBudget(config *BudgetConfig, candidate, current *CostProjection) *CandidateBudgetCheck {
	check := &CandidateBudgetCheck{projection: NewBudgetProjection(config, candidate), reason: "within_budget"}
	if current != nil {
		check.current = current.TotalMinor()
	}
	switch {
	case config.BudgetMode() == BudgetModeUnlimited:
		check.reason = "unlimited"
	case !candidate.IsComplete():
		check.reason = "price_unknown"
	case check.projection.Status() == "exceeded":
		check.reason = "budget_exceeded"
		if check.current != nil {
			before, _ := new(big.Int).SetString(*check.current, 10)
			after, _ := new(big.Int).SetString(*candidate.TotalMinor(), 10)
			if after.Cmp(before) <= 0 {
				check.reason = "non_increasing_overage"
			}
		}
	}
	return check
}
func (c *CandidateBudgetCheck) Projection() *BudgetProjection { return c.projection }
func (c *CandidateBudgetCheck) ReasonCode() string            { return c.reason }
func (c *CandidateBudgetCheck) Allowed() bool {
	return c.reason != "budget_exceeded" && c.reason != "price_unknown"
}
func (c *CandidateBudgetCheck) Err() error {
	if c.Allowed() {
		return nil
	}
	return &BudgetViolation{check: c}
}

type BudgetViolation struct{ check *CandidateBudgetCheck }

func (e *BudgetViolation) Error() string                 { return e.Code() }
func (e *BudgetViolation) Code() string                  { return e.check.reason }
func (e *BudgetViolation) Projection() *BudgetProjection { return e.check.projection }
func (e *BudgetViolation) CurrentSpentMinor() *string    { return copyMinor(e.check.current) }
func (e *BudgetViolation) ObjectIDs() []string {
	return e.check.projection.cost.UnknownPriceObjectIDs()
}
