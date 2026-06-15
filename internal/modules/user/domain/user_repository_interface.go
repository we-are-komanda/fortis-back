package domain

import "context"

// UserRepositoryInterface — интерфейс репозитория для User.
type UserRepositoryInterface interface {
	// Save создаёт или обновляет пользователя (upsert).
	Save(ctx context.Context, user *User) error

	// FindByID возвращает пользователя по ID.
	FindByID(ctx context.Context, id string) (*User, error)

	// FindByEmail возвращает пользователя по email.
	FindByEmail(ctx context.Context, email string) (*User, error)

	// Delete удаляет пользователя по ID.
	Delete(ctx context.Context, id string) error
}
