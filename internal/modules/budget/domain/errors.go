package domain

import "errors"

var (
	// ErrBudgetConfigNotFound возвращается, когда конфигурация бюджета не найдена.
	ErrBudgetConfigNotFound = errors.New("budget config not found")

	// ErrInvalidBudgetMode возвращается при некорректном режиме бюджета.
	ErrInvalidBudgetMode = errors.New("invalid budget mode: must be 'limited' or 'unlimited'")

	// ErrInvalidBudgetAmount возвращается при некорректной сумме бюджета.
	ErrInvalidBudgetAmount = errors.New("invalid budget amount: must be greater than 0 for limited mode")

	// ErrBudgetNotConfigured возвращается при попытке проверить остаток в unlimited режиме.
	ErrBudgetNotConfigured = errors.New("budget is not configured (unlimited mode)")

	// ErrProjectIDRequired возвращается, если ID проекта не указан.
	ErrProjectIDRequired = errors.New("project ID is required")

	// ErrComparisonFailed возвращается при ошибке сравнения конфигураций.
	ErrComparisonFailed = errors.New("comparison failed")

	// ErrBothIDsRequired возвращается, если не указаны оба ID проекта для сравнения.
	ErrBothIDsRequired = errors.New("both project ids are required for comparison")

	// ErrInvalidEchelonProfile возвращается при некорректных данных профиля эшелона.
	ErrInvalidEchelonProfile = errors.New("invalid echelon profile: layerID is required")
)
