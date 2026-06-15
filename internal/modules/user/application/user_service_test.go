package application

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fortis/backend/internal/modules/user/domain"
)

// mockUserRepo — мок репозитория для тестирования сервиса.
type mockUserRepo struct {
	users map[string]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users: make(map[string]*domain.User),
	}
}

func (m *mockUserRepo) Save(ctx context.Context, user *domain.User) error {
	m.users[user.ID()] = user
	return nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	for _, u := range m.users {
		if u.Email() == email {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepo) Delete(ctx context.Context, id string) error {
	if _, ok := m.users[id]; !ok {
		return domain.ErrUserNotFound
	}
	delete(m.users, id)
	return nil
}

func TestRegister_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, "test-secret", 24)

	user, token, err := svc.Register(context.Background(), "user@example.com", "password123", "Иван Петров")

	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "user@example.com", user.Email())
	assert.Equal(t, "Иван Петров", user.Name())
	assert.NotEmpty(t, token)

	// Проверяем, что сохранён в репозитории
	saved, err := repo.FindByID(context.Background(), user.ID())
	require.NoError(t, err)
	assert.Equal(t, user.Email(), saved.Email())
}

func TestRegister_DuplicateEmail(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, "test-secret", 24)

	_, _, err := svc.Register(context.Background(), "user@example.com", "password123", "User")
	require.NoError(t, err)

	_, _, err = svc.Register(context.Background(), "user@example.com", "password456", "User 2")
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrEmailAlreadyExists)
}

func TestRegister_InvalidEmail(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, "test-secret", 24)

	_, _, err := svc.Register(context.Background(), "invalid", "password123", "User")
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrInvalidEmail)
}

func TestLogin_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, "test-secret", 24)

	// Сначала регистрируем
	registered, _, err := svc.Register(context.Background(), "user@example.com", "password123", "Иван")
	require.NoError(t, err)

	// Логинимся
	user, token, err := svc.Login(context.Background(), "user@example.com", "password123")
	require.NoError(t, err)
	assert.Equal(t, registered.ID(), user.ID())
	assert.NotEmpty(t, token)
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, "test-secret", 24)

	_, _, err := svc.Register(context.Background(), "user@example.com", "password123", "Иван")
	require.NoError(t, err)

	_, _, err = svc.Login(context.Background(), "user@example.com", "wrongpassword")
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, "test-secret", 24)

	_, _, err := svc.Login(context.Background(), "nonexistent@example.com", "password123")
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

func TestGetProfile_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, "test-secret", 24)

	registered, _, err := svc.Register(context.Background(), "user@example.com", "password123", "Иван")
	require.NoError(t, err)

	profile, err := svc.GetProfile(context.Background(), registered.ID())
	require.NoError(t, err)
	assert.Equal(t, registered.ID(), profile.ID())
	assert.Equal(t, "user@example.com", profile.Email())
}

func TestGetProfile_NotFound(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, "test-secret", 24)

	_, err := svc.GetProfile(context.Background(), "nonexistent-id")
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestValidateToken_Success(t *testing.T) {
	svc := NewUserService(newMockUserRepo(), "test-secret", 24)

	_, token, err := svc.Register(context.Background(), "user@example.com", "password123", "Иван")
	require.NoError(t, err)

	claims, err := svc.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user@example.com", claims.Email)
	assert.NotEmpty(t, claims.UserID)
}

func TestValidateToken_Invalid(t *testing.T) {
	svc := NewUserService(newMockUserRepo(), "test-secret", 24)

	_, err := svc.ValidateToken("invalid-token")
	require.Error(t, err)
}
