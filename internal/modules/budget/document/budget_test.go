package document

import (
	"testing"

	"github.com/fortis/backend/internal/modules/budget/domain"
	"github.com/stretchr/testify/require"
)

func TestBudgetDocumentAuthorityAndLegacy(t *testing.T) {
	for _, test := range []struct {
		raw, mode string
		limit     *string
	}{
		{"null", "unlimited", nil},
		{`{"budgetMode":"unlimited","budgetLimitMinor":null}`, "unlimited", nil},
		{`{"budgetMode":"limited","budgetLimitMinor":"0","budgetAmountMln":999}`, "limited", stringPointer("0")},
		{`{"budgetMode":"limited","budgetAmountMln":0.000000005}`, "limited", stringPointer("1")},
		{`{"budgetMode":"limited","budgetAmountMln":9007199254.740993}`, "limited", stringPointer("900719925474099300")},
	} {
		config, err := ReadBudget("project", test.raw)
		require.NoError(t, err)
		require.Equal(t, domain.BudgetMode(test.mode), config.BudgetMode())
		require.Equal(t, test.limit, config.BudgetLimitMinor())
		encoded, err := WriteBudget(config)
		require.NoError(t, err)
		require.NotContains(t, encoded, "budgetAmountMln")
	}
	for _, raw := range []string{`{}`, `{"budgetMode":"limited"}`, `{"budgetMode":"limited","budgetLimitMinor":null,"budgetAmountMln":10}`, `{"budgetMode":"limited","budgetLimitMinor":1}`, `{"budgetMode":"limited","budgetAmountMln":-1}`} {
		_, err := ReadBudget("project", raw)
		require.Error(t, err, raw)
	}
}

func stringPointer(value string) *string { return &value }
