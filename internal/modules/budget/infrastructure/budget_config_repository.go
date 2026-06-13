package infrastructure

import (
	"context"
	"errors"

	"github.com/fortis/backend/internal/modules/budget/domain"
	"github.com/fortis/backend/internal/rdbms"
	"gorm.io/gorm"
)

// BudgetConfigRepository — реализация BudgetConfigRepositoryInterface.
type BudgetConfigRepository struct {
	db rdbms.Executor
}

// NewBudgetConfigRepository создаёт новый репозиторий конфигурации бюджета.
func NewBudgetConfigRepository(db rdbms.Executor) *BudgetConfigRepository {
	return &BudgetConfigRepository{db: db}
}

// FindByProjectID находит конфигурацию бюджета по ID проекта.
func (r *BudgetConfigRepository) FindByProjectID(ctx context.Context, projectID string) (*domain.BudgetConfig, error) {
	var model BudgetConfigModel
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		First(&model).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrBudgetConfigNotFound
		}
		return nil, err
	}

	return model.ToDomain()
}

// Save сохраняет конфигурацию бюджета (создаёт или обновляет по ProjectID).
func (r *BudgetConfigRepository) Save(ctx context.Context, config *domain.BudgetConfig) error {
	model, err := ToModel(config)
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).
		Save(model).
		Error
}

// Delete удаляет конфигурацию бюджета по ID проекта.
func (r *BudgetConfigRepository) Delete(ctx context.Context, projectID string) error {
	return r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Delete(&BudgetConfigModel{}).
		Error
}
