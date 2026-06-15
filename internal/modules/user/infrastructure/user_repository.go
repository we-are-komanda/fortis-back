package infrastructure

import (
	"context"
	"errors"

	"github.com/fortis/backend/internal/modules/user/domain"
	"github.com/fortis/backend/internal/rdbms"
	"gorm.io/gorm"
)

// UserRepository — реализация репозитория User.
type UserRepository struct {
	executor rdbms.Executor
}

// NewUserRepository создаёт новый репозиторий пользователей.
func NewUserRepository(executor rdbms.Executor) domain.UserRepositoryInterface {
	return &UserRepository{
		executor: executor,
	}
}

// Save сохраняет пользователя (создаёт или обновляет).
func (r *UserRepository) Save(ctx context.Context, user *domain.User) error {
	model := ToModel(user)
	db := r.executor.WithContext(ctx)

	var existing UserModel
	result := db.Where("id = ?", model.ID).First(&existing)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return db.Create(model).Error
		}
		return result.Error
	}

	// Обновление
	return db.Model(&UserModel{}).Where("id = ?", model.ID).Updates(map[string]interface{}{
		"email":        model.Email,
		"password_hash": model.PasswordHash,
		"name":         model.Name,
		"updated_at":   model.UpdatedAt,
	}).Error
}

// FindByID ищет пользователя по ID.
func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var model UserModel
	result := r.executor.WithContext(ctx).Where("id = ?", id).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, result.Error
	}

	return model.ToDomain(), nil
}

// FindByEmail ищет пользователя по email.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var model UserModel
	result := r.executor.WithContext(ctx).Where("email = ?", email).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, result.Error
	}

	return model.ToDomain(), nil
}

// Delete удаляет пользователя по ID.
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	result := r.executor.WithContext(ctx).Where("id = ?", id).Delete(&UserModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}
