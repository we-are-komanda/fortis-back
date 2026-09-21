package document

import (
	"encoding/json"
	"strings"

	"github.com/fortis/backend/internal/modules/budget/domain"
)

// ReadBudget accepts only an authoritative budget field/table, never project JSON.
func ReadBudget(projectID, raw string) (*domain.BudgetConfig, error) {
	if strings.TrimSpace(raw) == "" || strings.TrimSpace(raw) == "null" {
		return domain.NewBudgetConfigMinor(projectID, domain.BudgetModeUnlimited, nil)
	}
	var input struct {
		Mode   domain.BudgetMode `json:"budgetMode"`
		Limit  json.RawMessage   `json:"budgetLimitMinor"`
		Legacy json.RawMessage   `json:"budgetAmountMln"`
	}
	if err := json.Unmarshal([]byte(raw), &input); err != nil {
		return nil, domain.ErrInvalidBudgetAmount
	}
	var limit *string
	if len(input.Limit) > 0 {
		if err := json.Unmarshal(input.Limit, &limit); err != nil {
			return nil, domain.ErrInvalidBudgetAmount
		}
	} else if input.Mode == domain.BudgetModeLimited {
		value, err := domain.ParseLegacyBudgetMln(string(input.Legacy))
		if err != nil {
			return nil, err
		}
		limit = &value
	}
	return domain.NewBudgetConfigMinor(projectID, input.Mode, limit)
}

func WriteBudget(config *domain.BudgetConfig) (string, error) {
	value, err := json.Marshal(struct {
		Mode  domain.BudgetMode `json:"budgetMode"`
		Limit *string           `json:"budgetLimitMinor"`
	}{config.BudgetMode(), config.BudgetLimitMinor()})
	return string(value), err
}
