package domain

import (
	"strings"
	"time"
)

// User — aggregate root для пользователя.
type User struct {
	id            string
	email         string
	passwordHash  string
	name          string
	createdAt     time.Time
	updatedAt     time.Time
}

// NewUser создаёт нового пользователя с валидацией.
func NewUser(
	id, email, passwordHash, name string,
	createdAt, updatedAt time.Time,
) (*User, error) {
	if err := ValidateEmail(email); err != nil {
		return nil, err
	}
	if name == "" {
		return nil, ErrInvalidUserName
	}
	if passwordHash == "" {
		return nil, ErrInvalidPassword
	}

	return &User{
		id:           id,
		email:        email,
		passwordHash: passwordHash,
		name:         name,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}, nil
}

// Getters.
func (u *User) ID() string              { return u.id }
func (u *User) Email() string           { return u.email }
func (u *User) PasswordHash() string    { return u.passwordHash }
func (u *User) Name() string            { return u.name }
func (u *User) CreatedAt() time.Time    { return u.createdAt }
func (u *User) UpdatedAt() time.Time    { return u.updatedAt }

// Setters used by repository after loading from DB.
func (u *User) SetID(id string)            { u.id = id }
func (u *User) SetCreatedAt(t time.Time)   { u.createdAt = t }
func (u *User) SetUpdatedAt(t time.Time)   { u.updatedAt = t }

// UpdateName обновляет имя пользователя.
func (u *User) UpdateName(name string) error {
	if name == "" {
		return ErrInvalidUserName
	}
	u.name = name
	u.updatedAt = time.Now().UTC()
	return nil
}

// UpdateEmail обновляет email пользователя.
func (u *User) UpdateEmail(email string) error {
	if err := ValidateEmail(email); err != nil {
		return err
	}
	u.email = email
	u.updatedAt = time.Now().UTC()
	return nil
}

// ValidateEmail проверяет корректность email.
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return ErrInvalidEmail
	}
	if !strings.Contains(email, "@") {
		return ErrInvalidEmail
	}
	return nil
}
