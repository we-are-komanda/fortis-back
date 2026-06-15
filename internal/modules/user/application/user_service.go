package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/fortis/backend/internal/auth"
	"github.com/fortis/backend/internal/modules/user/domain"
)

// UserService — сервис для управления пользователями и аутентификацией.
type UserService struct {
	repo      domain.UserRepositoryInterface
	jwtSecret string
	jwtExpiry int // hours
}

// NewUserService создаёт новый сервис пользователей.
func NewUserService(repo domain.UserRepositoryInterface, jwtSecret string, jwtExpiry int) *UserService {
	return &UserService{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

// Register регистрирует нового пользователя.
func (s *UserService) Register(ctx context.Context, email, password, name string) (*domain.User, string, error) {
	// Проверяем, не занят ли email
	existing, err := s.repo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, "", fmt.Errorf("check email: %w", err)
	}
	if existing != nil {
		return nil, "", domain.ErrEmailAlreadyExists
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("hash password: %w", err)
	}

	now := time.Now().UTC()
	user, err := domain.NewUser(
		uuid.New().String(),
		email,
		string(hashedPassword),
		name,
		now,
		now,
	)
	if err != nil {
		return nil, "", err
	}

	if err := s.repo.Save(ctx, user); err != nil {
		return nil, "", fmt.Errorf("save user: %w", err)
	}

	// Генерируем JWT
	token, err := auth.GenerateToken(user.ID(), user.Email(), s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, "", fmt.Errorf("generate token: %w", err)
	}

	return user, token, nil
}

// Login аутентифицирует пользователя по email и паролю.
func (s *UserService) Login(ctx context.Context, email, password string) (*domain.User, string, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, "", domain.ErrInvalidCredentials
		}
		return nil, "", fmt.Errorf("find user: %w", err)
	}

	// Сравниваем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash()), []byte(password)); err != nil {
		return nil, "", domain.ErrInvalidCredentials
	}

	// Генерируем JWT
	token, err := auth.GenerateToken(user.ID(), user.Email(), s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, "", fmt.Errorf("generate token: %w", err)
	}

	return user, token, nil
}

// GetProfile возвращает профиль пользователя по ID.
func (s *UserService) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// ValidateToken проверяет JWT-токен и возвращает claims.
func (s *UserService) ValidateToken(tokenString string) (*auth.Claims, error) {
	return auth.ValidateToken(tokenString, s.jwtSecret)
}

// SetPassword устанавливает новый пароль для пользователя.
func (s *UserService) SetPassword(ctx context.Context, userID, newPassword string) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	// Создаём нового пользователя с обновлённым хешем пароля
	// (используем внутреннее обновление через сохранение)
	now := time.Now().UTC()
	updated, err := domain.NewUser(
		user.ID(),
		user.Email(),
		string(hashedPassword),
		user.Name(),
		user.CreatedAt(),
		now,
	)
	if err != nil {
		return err
	}

	return s.repo.Save(ctx, updated)
}
