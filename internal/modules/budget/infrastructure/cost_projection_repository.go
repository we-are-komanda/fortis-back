package infrastructure

import (
	"context"
	"errors"

	"github.com/fortis/backend/internal/modules/budget/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *BudgetConfigRepository) FindCostProjection(ctx context.Context, id string, version int, calculator string) (*domain.CostProjection, error) {
	var model CostProjectionModel
	err := r.db.WithContext(ctx).Where("project_id = ? AND project_version = ? AND calculation_version = ?", id, version, calculator).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrCostProjectionNotFound
	}
	if err != nil {
		return nil, err
	}
	return model.ToDomain()
}

func (r *BudgetConfigRepository) SaveCostProjection(ctx context.Context, p *domain.CostProjection) (*domain.CostProjection, error) {
	model, err := costProjectionToModel(p)
	if err != nil {
		return nil, err
	}
	var saved *domain.CostProjection
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(model).Error; err != nil {
			return err
		}
		var stored CostProjectionModel
		if err := tx.Where("project_id = ? AND project_version = ? AND calculation_version = ?", model.ProjectID, model.ProjectVersion, model.CalculationVersion).First(&stored).Error; err != nil {
			return err
		}
		var err error
		saved, err = stored.ToDomain()
		return err
	})
	return saved, err
}
