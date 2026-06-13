package domain

import "errors"

var (
	// ErrInvalidSchemaVersion возвращается при неверной версии схемы проекта.
	ErrInvalidSchemaVersion = errors.New("invalid schema version: expected 1")

	// ErrInvalidProjectData возвращается при некорректной структуре проекта.
	ErrInvalidProjectData = errors.New("invalid project data")

	// ErrProjectNotFound возвращается когда проект не найден.
	ErrProjectNotFound = errors.New("project not found")

	// ErrInvalidConfigName возвращается при пустом имени конфигурации.
	ErrInvalidConfigName = errors.New("invalid configuration name: name is required")

	// ErrVersionConflict возвращается при попытке сохранить устаревшую версию проекта.
	ErrVersionConflict = errors.New("version conflict: project has been modified by another user")

	// ErrInvalidDateFormat возвращается при некорректном формате даты.
	ErrInvalidDateFormat = errors.New("invalid date format: expected RFC3339Nano (e.g. 2026-06-12T14:00:00.000Z)")
)
