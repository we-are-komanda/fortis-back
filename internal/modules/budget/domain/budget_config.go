package domain

import (
	"math/big"
	"strconv"
	"time"
)

// BudgetMode — режим бюджета проекта.
type BudgetMode string

const (
	// BudgetModeLimited — бюджет ограничен фиксированной суммой.
	BudgetModeLimited BudgetMode = "limited"
	// BudgetModeUnlimited — бюджет не ограничен.
	BudgetModeUnlimited BudgetMode = "unlimited"
)

// BudgetConfig — агрегат конфигурации бюджета проекта.
type BudgetConfig struct {
	projectID        string
	budgetMode       BudgetMode
	budgetLimitMinor *string
	createdAt        time.Time
	updatedAt        time.Time
}

// NewBudgetConfig создаёт новый BudgetConfig с валидацией.
func NewBudgetConfig(projectID string, mode BudgetMode, amountMln float64) (*BudgetConfig, error) {
	var limit *string
	if mode == BudgetModeLimited {
		value, err := ParseLegacyBudgetMln(strconv.FormatFloat(amountMln, 'g', -1, 64))
		if err != nil {
			return nil, err
		}
		limit = &value
	}
	return NewBudgetConfigMinor(projectID, mode, limit)
}

func NewBudgetConfigMinor(projectID string, mode BudgetMode, limit *string) (*BudgetConfig, error) {
	if projectID == "" {
		return nil, ErrProjectIDRequired
	}
	if mode != BudgetModeLimited && mode != BudgetModeUnlimited {
		return nil, ErrInvalidBudgetMode
	}
	if mode == BudgetModeUnlimited && limit != nil {
		return nil, ErrInvalidBudgetAmount
	}
	if mode == BudgetModeLimited {
		if limit == nil || !minorPattern.MatchString(*limit) {
			return nil, ErrInvalidBudgetAmount
		}
		value, _ := new(big.Int).SetString(*limit, 10)
		canonical := value.String()
		limit = &canonical
	}

	now := time.Now().UTC()
	return &BudgetConfig{
		projectID:        projectID,
		budgetMode:       mode,
		budgetLimitMinor: copyMinor(limit),
		createdAt:        now,
		updatedAt:        now,
	}, nil
}

// ProjectID возвращает ID проекта.
func (c *BudgetConfig) ProjectID() string { return c.projectID }

// BudgetMode возвращает режим бюджета.
func (c *BudgetConfig) BudgetMode() BudgetMode { return c.budgetMode }

// BudgetAmountMln возвращает сумму бюджета в млн ₽.
func (c *BudgetConfig) BudgetAmountMln() float64 {
	if c.budgetLimitMinor == nil {
		return 0
	}
	value, _ := new(big.Rat).SetString(*c.budgetLimitMinor)
	value.Quo(value, big.NewRat(100_000_000, 1))
	amount, _ := value.Float64()
	return amount
}
func (c *BudgetConfig) BudgetLimitMinor() *string { return copyMinor(c.budgetLimitMinor) }

// CreatedAt возвращает время создания.
func (c *BudgetConfig) CreatedAt() time.Time { return c.createdAt }

// UpdatedAt возвращает время последнего обновления.
func (c *BudgetConfig) UpdatedAt() time.Time { return c.updatedAt }

// SetBudget обновляет режим и лимит бюджета.
func (c *BudgetConfig) SetBudget(mode BudgetMode, amountMln float64) error {
	updated, err := NewBudgetConfig(c.projectID, mode, amountMln)
	if err != nil {
		return err
	}
	c.budgetMode, c.budgetLimitMinor = updated.budgetMode, updated.budgetLimitMinor
	c.updatedAt = time.Now().UTC()
	return nil
}

// RemainingBudget возвращает остаток бюджета.
// Если режим unlimited — возвращает 0 и ошибку ErrBudgetNotConfigured.
// Если режим limited — возвращает разницу между лимитом и текущей стоимостью.
func (c *BudgetConfig) RemainingBudget(currentTotalCostMln float64) (float64, error) {
	if c.budgetMode == BudgetModeUnlimited {
		return 0, ErrBudgetNotConfigured
	}
	remaining := c.BudgetAmountMln() - currentTotalCostMln
	if remaining < 0 {
		return 0, nil
	}
	return remaining, nil
}

// RestoreBudgetConfig адаптирует историческое значение в миллионах с сохранением дат.
func RestoreBudgetConfig(projectID string, mode BudgetMode, amountMln float64, createdAt, updatedAt time.Time) *BudgetConfig {
	config, err := NewBudgetConfig(projectID, mode, amountMln)
	if err != nil {
		return nil
	}
	config.createdAt, config.updatedAt = createdAt, updatedAt
	return config
}
