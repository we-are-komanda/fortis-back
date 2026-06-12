package domain

import "errors"

var (
	// ErrInvalidSchemaVersion возвращается при неверной версии схемы проекта.
	ErrInvalidSchemaVersion = errors.New("invalid schema version: expected 1")

	// ErrInvalidProjectData возвращается при некорректной структуре проекта.
	ErrInvalidProjectData = errors.New("invalid project data")

	// ErrProjectNotFound возвращается когда проект не найден.
	ErrProjectNotFound = errors.New("project not found")
)
