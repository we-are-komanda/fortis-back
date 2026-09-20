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
	var existing DefenseProjectModel
	err := r.executor.WithContext(ctx).Where("id = ?", project.ProjectID()).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	saved, err := r.Commit(ctx, project, domain.ProjectWrite{Create: errors.Is(err, gorm.ErrRecordNotFound)})
	if err == nil {
		*project = *saved
	}
	return err
}

func nullableEnterpriseID(enterpriseID string) interface{} {
	if enterpriseID == "" {
		return nil
	}
	return enterpriseID
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

// FindAllByUserID scopes both rows and count before applying pagination.
func (r *DefenseProjectRepository) FindAllByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.DefenseProject, int64, error) {
	query := r.executor.WithContext(ctx).Model(&DefenseProjectModel{}).
		Where("enterprise_id IN (SELECT enterprise_id FROM user_enterprises WHERE user_id = ?)", userID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var models []DefenseProjectModel
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	projects := make([]*domain.DefenseProject, 0, len(models))
	for _, model := range models {
		project, err := model.ToDomain()
		if err != nil {
			return nil, 0, err
		}
		projects = append(projects, project)
	}
	return projects, total, nil
}
