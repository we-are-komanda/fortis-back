package domain

import "time"

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
	projectID       string
	budgetMode      BudgetMode
	budgetAmountMln float64
	createdAt       time.Time
	updatedAt       time.Time
}

// NewBudgetConfig создаёт новый BudgetConfig с валидацией.
func NewBudgetConfig(projectID string, mode BudgetMode, amountMln float64) (*BudgetConfig, error) {
	if projectID == "" {
		return nil, ErrProjectIDRequired
	}
	if mode != BudgetModeLimited && mode != BudgetModeUnlimited {
		return nil, ErrInvalidBudgetMode
	}
	if mode == BudgetModeLimited && amountMln <= 0 {
		return nil, ErrInvalidBudgetAmount
	}

	now := time.Now().UTC()
	return &BudgetConfig{
		projectID:       projectID,
		budgetMode:      mode,
		budgetAmountMln: amountMln,
		createdAt:       now,
		updatedAt:       now,
	}, nil
}

// ProjectID возвращает ID проекта.
func (c *BudgetConfig) ProjectID() string { return c.projectID }

// BudgetMode возвращает режим бюджета.
func (c *BudgetConfig) BudgetMode() BudgetMode { return c.budgetMode }

// BudgetAmountMln возвращает сумму бюджета в млн ₽.
func (c *BudgetConfig) BudgetAmountMln() float64 { return c.budgetAmountMln }

// CreatedAt возвращает время создания.
func (c *BudgetConfig) CreatedAt() time.Time { return c.createdAt }

// UpdatedAt возвращает время последнего обновления.
func (c *BudgetConfig) UpdatedAt() time.Time { return c.updatedAt }

// SetBudget обновляет режим и лимит бюджета.
func (c *BudgetConfig) SetBudget(mode BudgetMode, amountMln float64) error {
	if mode != BudgetModeLimited && mode != BudgetModeUnlimited {
		return ErrInvalidBudgetMode
	}
	if mode == BudgetModeLimited && amountMln <= 0 {
		return ErrInvalidBudgetAmount
	}
	c.budgetMode = mode
	c.budgetAmountMln = amountMln
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
	remaining := c.budgetAmountMln - currentTotalCostMln
	if remaining < 0 {
		return 0, nil
	}
	return remaining, nil
}

// RestoreBudgetConfig восстанавливает BudgetConfig из данных БД (без валидации).
func RestoreBudgetConfig(projectID string, mode BudgetMode, amountMln float64, createdAt, updatedAt time.Time) *BudgetConfig {
	return &BudgetConfig{
		projectID:       projectID,
		budgetMode:      mode,
		budgetAmountMln: amountMln,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}
}
