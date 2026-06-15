package domain

import "errors"

var (
	// ErrProjectNotFound возвращается, когда проект не найден.
	ErrProjectNotFound = errors.New("project not found")

	// ErrInvalidProjectID возвращается при некорректном ID проекта.
	ErrInvalidProjectID = errors.New("invalid project ID: id is required")
)
