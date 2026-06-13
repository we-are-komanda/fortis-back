package domain

import "context"

// BudgetConfigRepositoryInterface — интерфейс репозитория для BudgetConfig.
type BudgetConfigRepositoryInterface interface {
	// FindByProjectID находит конфигурацию бюджета по ID проекта.
	FindByProjectID(ctx context.Context, projectID string) (*BudgetConfig, error)

	// Save сохраняет конфигурацию бюджета (создаёт или обновляет по ProjectID).
	Save(ctx context.Context, config *BudgetConfig) error

	// Delete удаляет конфигурацию бюджета по ID проекта.
	Delete(ctx context.Context, projectID string) error
}
