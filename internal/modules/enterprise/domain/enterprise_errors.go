package domain

import "errors"

var (
	// ErrEnterpriseNotFound возвращается когда предприятие не найдено.
	ErrEnterpriseNotFound = errors.New("enterprise not found")

	// ErrEnterpriseAccessDenied возвращается при попытке доступа к предприятию без прав.
	ErrEnterpriseAccessDenied = errors.New("enterprise access denied")

	// ErrInvalidEnterpriseName возвращается при пустом названии предприятия.
	ErrInvalidEnterpriseName = errors.New("invalid enterprise name: must not be empty")

	// ErrInvalidEnterpriseStatus возвращается при неверном статусе предприятия.
	ErrInvalidEnterpriseStatus = errors.New("invalid enterprise status")

	// ErrInvalidCoordinates возвращается при некорректных координатах.
	ErrInvalidCoordinates = errors.New("invalid coordinates")
)
