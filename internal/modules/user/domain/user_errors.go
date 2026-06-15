package domain

import "errors"

var (
	// ErrUserNotFound возвращается когда пользователь не найден.
	ErrUserNotFound = errors.New("user not found")

	// ErrInvalidEmail возвращается при некорректном email.
	ErrInvalidEmail = errors.New("invalid email")

	// ErrInvalidUserName возвращается при пустом имени пользователя.
	ErrInvalidUserName = errors.New("invalid user name: must not be empty")

	// ErrInvalidPassword возвращается при пустом хеше пароля.
	ErrInvalidPassword = errors.New("invalid password hash: must not be empty")

	// ErrEmailAlreadyExists возвращается если email уже занят.
	ErrEmailAlreadyExists = errors.New("email already exists")

	// ErrInvalidCredentials возвращается при неверных учётных данных.
	ErrInvalidCredentials = errors.New("invalid credentials")
)
