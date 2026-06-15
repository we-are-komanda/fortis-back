package infrastructure

import (
	"context"
	"errors"

	"github.com/fortis/backend/internal/modules/enterprise/domain"
	"github.com/fortis/backend/internal/rdbms"
	"gorm.io/gorm"
)

// userEnterpriseModel — GORM-модель для связи пользователя и предприятия.
type userEnterpriseModel struct {
	UserID       string `gorm:"primaryKey;type:uuid"`
	EnterpriseID string `gorm:"primaryKey;type:uuid"`
}

func (userEnterpriseModel) TableName() string {
	return "user_enterprises"
}

// EnterpriseRepository — реализация репозитория Enterprise.
type EnterpriseRepository struct {
	executor rdbms.Executor
}

// NewEnterpriseRepository создаёт новый репозиторий.
func NewEnterpriseRepository(executor rdbms.Executor) domain.EnterpriseRepositoryInterface {
	return &EnterpriseRepository{
		executor: executor,
	}
}

// Save сохраняет предприятие (создаёт или обновляет).
func (r *EnterpriseRepository) Save(ctx context.Context, enterprise *domain.Enterprise) error {
	model := ToModel(enterprise)
	db := r.executor.WithContext(ctx)

	// Upsert
	var existing EnterpriseModel
	result := db.Where("id = ?", model.ID).First(&existing)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return db.Create(model).Error
		}
		return result.Error
	}

	// Обновление
	return db.Model(&EnterpriseModel{}).Where("id = ?", model.ID).Updates(map[string]interface{}{
		"name":      model.Name,
		"address":   model.Address,
		"status":    model.Status,
		"latitude":  model.Latitude,
		"longitude": model.Longitude,
		"updated_at": model.UpdatedAt,
	}).Error
}

// FindByID ищет предприятие по ID.
func (r *EnterpriseRepository) FindByID(ctx context.Context, id string) (*domain.Enterprise, error) {
	var model EnterpriseModel
	result := r.executor.WithContext(ctx).Where("id = ?", id).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrEnterpriseNotFound
		}
		return nil, result.Error
	}

	return model.ToDomain(), nil
}

// FindAll возвращает список предприятий с пагинацией.
func (r *EnterpriseRepository) FindAll(ctx context.Context, limit, offset int) ([]*domain.Enterprise, int64, error) {
	db := r.executor.WithContext(ctx)

	var total int64
	if err := db.Model(&EnterpriseModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var models []EnterpriseModel
	if err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&models).Error; err != nil {
		return nil, 0, err
	}

	enterprises := make([]*domain.Enterprise, len(models))
	for i, m := range models {
		enterprises[i] = m.ToDomain()
	}

	return enterprises, total, nil
}

// Delete удаляет предприятие по ID.
func (r *EnterpriseRepository) Delete(ctx context.Context, id string) error {
	result := r.executor.WithContext(ctx).Where("id = ?", id).Delete(&EnterpriseModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrEnterpriseNotFound
	}
	return nil
}

// FindAllByUserID возвращает предприятия, доступные пользователю, с пагинацией.
func (r *EnterpriseRepository) FindAllByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Enterprise, int64, error) {
	db := r.executor.WithContext(ctx)

	var total int64
	if err := db.Model(&EnterpriseModel{}).
		Joins("JOIN user_enterprises ON enterprises.id = user_enterprises.enterprise_id").
		Where("user_enterprises.user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var models []EnterpriseModel
	if err := db.Joins("JOIN user_enterprises ON enterprises.id = user_enterprises.enterprise_id").
		Where("user_enterprises.user_id = ?", userID).
		Order("enterprises.created_at DESC").
		Limit(limit).Offset(offset).
		Find(&models).Error; err != nil {
		return nil, 0, err
	}

	enterprises := make([]*domain.Enterprise, len(models))
	for i, m := range models {
		enterprises[i] = m.ToDomain()
	}

	return enterprises, total, nil
}

// AddUserToEnterprise добавляет пользователя к предприятию.
func (r *EnterpriseRepository) AddUserToEnterprise(ctx context.Context, userID, enterpriseID string) error {
	link := userEnterpriseModel{
		UserID:       userID,
		EnterpriseID: enterpriseID,
	}
	return r.executor.WithContext(ctx).Create(&link).Error
}

// RemoveUserFromEnterprise удаляет пользователя из предприятия.
func (r *EnterpriseRepository) RemoveUserFromEnterprise(ctx context.Context, userID, enterpriseID string) error {
	result := r.executor.WithContext(ctx).
		Where("user_id = ? AND enterprise_id = ?", userID, enterpriseID).
		Delete(&userEnterpriseModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrEnterpriseNotFound
	}
	return nil
}

// CheckUserEnterpriseAccess проверяет, имеет ли пользователь доступ к предприятию.
func (r *EnterpriseRepository) CheckUserEnterpriseAccess(ctx context.Context, userID, enterpriseID string) (bool, error) {
	var count int64
	err := r.executor.WithContext(ctx).
		Model(&userEnterpriseModel{}).
		Where("user_id = ? AND enterprise_id = ?", userID, enterpriseID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
