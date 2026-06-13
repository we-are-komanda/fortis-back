package infrastructure

import (
	"context"
	"errors"

	"github.com/fortis/backend/internal/modules/defense_project/domain"
	"github.com/fortis/backend/internal/rdbms"
	"gorm.io/gorm"
)

type DefenseProjectRepository struct {
	executor rdbms.Executor
}

func NewDefenseProjectRepository(executor rdbms.Executor) domain.DefenseProjectRepositoryInterface {
	return &DefenseProjectRepository{
		executor: executor,
	}
}

func (r *DefenseProjectRepository) Save(ctx context.Context, project *domain.DefenseProject) error {
	model, err := ToModel(project)
	if err != nil {
		return err
	}

	db := r.executor.WithContext(ctx)

	// Upsert: попытка обновить существующую запись или создать новую
	var existing DefenseProjectModel
	result := db.Where("id = ?", model.ID).First(&existing)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Создать новую запись
			return db.Create(model).Error
		}
		return result.Error
	}

	// Обновить существующую с проверкой версии (optimistic locking)
	result = db.Model(&DefenseProjectModel{}).Where("id = ? AND version = ?", model.ID, model.Version).
		Updates(map[string]interface{}{
			"name":          model.Name,
			"enterprise_id": model.EnterpriseID,
			"project_data":  model.ProjectData,
			"version":       model.Version + 1,
			"updated_at":    model.UpdatedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrVersionConflict
	}
	return nil
}

func (r *DefenseProjectRepository) FindByID(ctx context.Context, id string) (*domain.DefenseProject, error) {
	var model DefenseProjectModel
	result := r.executor.WithContext(ctx).Where("id = ?", id).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, result.Error
	}

	return model.ToDomain()
}

func (r *DefenseProjectRepository) FindAll(ctx context.Context, limit, offset int) ([]*domain.DefenseProject, int64, error) {
	var models []DefenseProjectModel
	var total int64

	db := r.executor.WithContext(ctx)

	if err := db.Model(&DefenseProjectModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Limit(limit).Offset(offset).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	projects := make([]*domain.DefenseProject, len(models))
	for i, m := range models {
		p, err := m.ToDomain()
		if err != nil {
			return nil, 0, err
		}
		projects[i] = p
	}

	return projects, total, nil
}

func (r *DefenseProjectRepository) FindAllByEnterprise(ctx context.Context, enterpriseID string, limit, offset int) ([]*domain.DefenseProject, int64, error) {
	var models []DefenseProjectModel
	var total int64

	db := r.executor.WithContext(ctx)

	if err := db.Model(&DefenseProjectModel{}).Where("enterprise_id = ?", enterpriseID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Where("enterprise_id = ?", enterpriseID).Limit(limit).Offset(offset).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	projects := make([]*domain.DefenseProject, len(models))
	for i, m := range models {
		p, err := m.ToDomain()
		if err != nil {
			return nil, 0, err
		}
		projects[i] = p
	}

	return projects, total, nil
}

func (r *DefenseProjectRepository) Delete(ctx context.Context, id string) error {
	result := r.executor.WithContext(ctx).Where("id = ?", id).Delete(&DefenseProjectModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}
