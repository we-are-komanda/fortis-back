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

	// Обновить существующую
	return db.Model(&DefenseProjectModel{}).Where("id = ?", model.ID).Update("project_data", model.ProjectData).Error
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
