package domain

import "context"

// EnterpriseRepositoryInterface — интерфейс репозитория для Enterprise.
type EnterpriseRepositoryInterface interface {
	// Save создаёт или обновляет предприятие (upsert).
	Save(ctx context.Context, enterprise *Enterprise) error

	// FindByID возвращает предприятие по ID.
	FindByID(ctx context.Context, id string) (*Enterprise, error)

	// FindAll возвращает список предприятий с пагинацией и общим количеством.
	FindAll(ctx context.Context, limit, offset int) ([]*Enterprise, int64, error)

	// FindAllByUserID возвращает предприятия, доступные пользователю, с пагинацией.
	FindAllByUserID(ctx context.Context, userID string, limit, offset int) ([]*Enterprise, int64, error)

	// CheckUserEnterpriseAccess проверяет, имеет ли пользователь доступ к предприятию.
	CheckUserEnterpriseAccess(ctx context.Context, userID, enterpriseID string) (bool, error)

	// AddUserToEnterprise добавляет пользователя к предприятию.
	AddUserToEnterprise(ctx context.Context, userID, enterpriseID string) error

	// RemoveUserFromEnterprise удаляет пользователя из предприятия.
	RemoveUserFromEnterprise(ctx context.Context, userID, enterpriseID string) error

	// Delete удаляет предприятие по ID.
	Delete(ctx context.Context, id string) error
}
