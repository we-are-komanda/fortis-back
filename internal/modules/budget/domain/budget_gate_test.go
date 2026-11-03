package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func minor(value string) *string { return &value }

func gateCost(value *string) *CostProjection {
	price := UnknownUnitPrice()
	if value != nil {
		price, _ = ParseUnitMinor(*value)
	}
	line, err := NewFinancialLine("object", "asset", "layer", "radar", "Synthetic", 1, price, "template", "")
	if err != nil {
		panic(err)
	}
	return NewCostProjection(NewCalculationIdentity("project", 1, CostCalculationV1, "digest", nil), []FinancialLine{line}, nil, nil)
}

func TestExactBudgetGate(t *testing.T) {
	config, err := NewBudgetConfigMinor("project", BudgetModeLimited, minor("50000000"))
	require.NoError(t, err)
	for _, test := range []struct {
		name               string
		candidate, current *string
		reason             string
	}{
		{"equal", minor("50000000"), minor("47500000"), "within_budget"},
		{"one kopeck over", minor("50000001"), minor("47500000"), "budget_exceeded"},
		{"unknown", nil, minor("47500000"), "price_unknown"},
		{"overage same", minor("55000000"), minor("55000000"), "non_increasing_overage"},
		{"overage reduced", minor("52000000"), minor("55000000"), "non_increasing_overage"},
		{"overage increased", minor("55000001"), minor("55000000"), "budget_exceeded"},
		{"unknown baseline overage", minor("55000000"), nil, "budget_exceeded"},
	} {
		t.Run(test.name, func(t *testing.T) {
			check := CheckCandidateBudget(config, gateCost(test.candidate), gateCost(test.current))
			require.Equal(t, test.reason, check.ReasonCode())
			require.Equal(t, test.reason != "budget_exceeded" && test.reason != "price_unknown", check.Allowed())
			if !check.Allowed() {
				var violation *BudgetViolation
				require.True(t, errors.As(check.Err(), &violation))
				require.Equal(t, test.reason, violation.Code())
			}
		})
	}
	zero, err := NewBudgetConfigMinor("project", BudgetModeLimited, minor("0"))
	require.NoError(t, err)
	require.True(t, CheckCandidateBudget(zero, gateCost(minor("0")), nil).Allowed())
	require.False(t, CheckCandidateBudget(zero, gateCost(minor("1")), nil).Allowed())
	unlimited, err := NewBudgetConfigMinor("project", BudgetModeUnlimited, nil)
	require.NoError(t, err)
	check := CheckCandidateBudget(unlimited, gateCost(nil), nil)
	require.True(t, check.Allowed())
	require.Equal(t, "indeterminate", check.Projection().Status())
	require.Nil(t, check.Projection().SpentMinor())
	projection := NewBudgetProjection(config, gateCost(minor("50000001")))
	require.Equal(t, "exceeded", projection.Status())
	require.Equal(t, minor("-1"), projection.RemainingMinor())
}

func TestBudgetMinorValidation(t *testing.T) {
	for _, value := range []*string{nil, minor(""), minor("-1"), minor("0.1"), minor("NaN"), minor("1e3")} {
		_, err := NewBudgetConfigMinor("project", BudgetModeLimited, value)
		require.ErrorIs(t, err, ErrInvalidBudgetAmount)
	}
	_, err := NewBudgetConfigMinor("project", BudgetModeUnlimited, minor("0"))
	require.ErrorIs(t, err, ErrInvalidBudgetAmount)
	config, err := NewBudgetConfigMinor("project", BudgetModeLimited, minor("0001000000000000000000"))
	require.NoError(t, err)
	require.Equal(t, minor("1000000000000000000"), config.BudgetLimitMinor())
}
