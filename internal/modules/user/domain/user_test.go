package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser_Success(t *testing.T) {
	now := time.Now().UTC()
	u, err := NewUser(
		"550e8400-e29b-41d4-a716-446655440000",
		"user@example.com",
		"$2a$10$hashedpassword",
		"Иван Петров",
		now,
		now,
	)

	require.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", u.ID())
	assert.Equal(t, "user@example.com", u.Email())
	assert.Equal(t, "$2a$10$hashedpassword", u.PasswordHash())
	assert.Equal(t, "Иван Петров", u.Name())
}

func TestNewUser_InvalidEmail_Empty(t *testing.T) {
	now := time.Now().UTC()
	u, err := NewUser(
		"test-id",
		"",
		"hash",
		"Name",
		now,
		now,
	)

	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidEmail)
	assert.Nil(t, u)
}

func TestNewUser_InvalidEmail_NoAt(t *testing.T) {
	now := time.Now().UTC()
	u, err := NewUser(
		"test-id",
		"notanemail",
		"hash",
		"Name",
		now,
		now,
	)

	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidEmail)
	assert.Nil(t, u)
}

func TestNewUser_EmptyName(t *testing.T) {
	now := time.Now().UTC()
	u, err := NewUser(
		"test-id",
		"user@example.com",
		"hash",
		"",
		now,
		now,
	)

	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidUserName)
	assert.Nil(t, u)
}

func TestNewUser_EmptyPasswordHash(t *testing.T) {
	now := time.Now().UTC()
	u, err := NewUser(
		"test-id",
		"user@example.com",
		"",
		"Name",
		now,
		now,
	)

	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidPassword)
	assert.Nil(t, u)
}

func TestUser_UpdateName(t *testing.T) {
	u := validUser()
	err := u.UpdateName("Новое имя")
	require.NoError(t, err)
	assert.Equal(t, "Новое имя", u.Name())
}

func TestUser_UpdateName_Empty(t *testing.T) {
	u := validUser()
	err := u.UpdateName("")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidUserName)
}

func TestUser_UpdateEmail(t *testing.T) {
	u := validUser()
	err := u.UpdateEmail("new@example.com")
	require.NoError(t, err)
	assert.Equal(t, "new@example.com", u.Email())
}

func TestUser_UpdateEmail_Invalid(t *testing.T) {
	u := validUser()
	err := u.UpdateEmail("invalid")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidEmail)
}

func validUser() *User {
	now := time.Now().UTC()
	u, _ := NewUser(
		"test-id",
		"user@example.com",
		"$2a$10$hash",
		"Тестовый пользователь",
		now,
		now,
	)
	return u
}
