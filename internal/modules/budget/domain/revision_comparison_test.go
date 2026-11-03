package domain

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func comparisonSide(t *testing.T, id, value string, version int) ComparedRevision {
	t.Helper()
	price := UnknownUnitPrice()
	if value != "unknown" {
		var err error
		price, err = ParseUnitMinor(value)
		require.NoError(t, err)
	}
	line, err := NewFinancialLine("object", "asset", "layer", "synthetic", "Module", 1, price, "template", "")
	require.NoError(t, err)
	cost := NewCostProjection(NewCalculationIdentity(id, version, CostCalculationV1, "digest", nil), []FinancialLine{line}, map[string]string{"layer": "Layer"}, nil)
	config, err := NewBudgetConfigMinor(id, BudgetModeUnlimited, nil)
	require.NoError(t, err)
	return NewComparedRevision("Configuration "+id, time.Unix(100, 0), NewBudgetProjection(config, cost))
}

func TestRevisionComparisonExactMoney(t *testing.T) {
	for _, tc := range []struct{ a, b, difference, percent string }{
		{"47500000", "55000000", "7500000", "15.79"},
		{"55000000", "47500000", "-7500000", "-13.64"},
		{"0", "1", "1", "null"}, {"0", "0", "0", "null"},
		{"unknown", "1", "null", "null"}, {"1", "unknown", "null", "null"},
		{"20000", "20001", "1", "0.01"}, {"20000", "19999", "-1", "-0.01"},
		{"1000000000000000", "999999999999999", "-1", "0.00"},
	} {
		t.Run(tc.a+"_"+tc.b, func(t *testing.T) {
			got := NewRevisionComparison("enterprise", comparisonSide(t, "A", tc.a, 7), comparisonSide(t, "B", tc.b, 3))
			show := func(v *string) string {
				if v == nil {
					return "null"
				}
				return *v
			}
			require.Equal(t, tc.difference, show(got.DifferenceMinor()))
			require.Equal(t, tc.percent, show(got.DifferencePercent()))
			require.Equal(t, 7, got.A().Budget().Identity().ProjectVersion())
			require.Equal(t, 3, got.B().Budget().Identity().ProjectVersion())
		})
	}
}

func TestRevisionComparisonBeyondJSONNumberPrecision(t *testing.T) {
	maximum, err := ParseUnitMinor("1000000000000000")
	require.NoError(t, err)
	bulk, err := NewFinancialLine("bulk", "asset", "layer", "synthetic", "Bulk", 1000000, maximum, "template", "")
	require.NoError(t, err)
	one, err := ParseUnitMinor("1")
	require.NoError(t, err)
	extra, err := NewFinancialLine("extra", "asset", "layer", "synthetic", "Extra", 1, one, "template", "")
	require.NoError(t, err)
	side := func(id string, lines []FinancialLine) ComparedRevision {
		cost := NewCostProjection(NewCalculationIdentity(id, 1, CostCalculationV1, "digest", nil), lines, nil, nil)
		config, e := NewBudgetConfigMinor(id, BudgetModeUnlimited, nil)
		require.NoError(t, e)
		return NewComparedRevision(id, time.Unix(100, 0), NewBudgetProjection(config, cost))
	}
	got := NewRevisionComparison("enterprise", side("A", []FinancialLine{bulk}), side("B", []FinancialLine{bulk, extra}))
	require.Equal(t, "1000000000000000000000", *got.A().Cost().TotalMinor())
	require.Equal(t, "1", *got.DifferenceMinor())
	require.Equal(t, "0.00", *got.DifferencePercent())
}
