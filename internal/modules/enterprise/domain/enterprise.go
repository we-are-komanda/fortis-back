package domain

import (
	"fmt"
	"time"
)

// EnterpriseStatus — статус предприятия.
type EnterpriseStatus string

const (
	EnterpriseStatusActive       EnterpriseStatus = "active"
	EnterpriseStatusConfiguring  EnterpriseStatus = "configuring"
	EnterpriseStatusOffline      EnterpriseStatus = "offline"
)

// ValidEnterpriseStatuses — список допустимых статусов предприятия.
var ValidEnterpriseStatuses = []EnterpriseStatus{
	EnterpriseStatusActive,
	EnterpriseStatusConfiguring,
	EnterpriseStatusOffline,
}

// Enterprise — aggregate корень для предприятия/защищаемого объекта.
type Enterprise struct {
	id        string
	name      string
	address   string
	status    EnterpriseStatus
	latitude  float64
	longitude float64
	createdAt time.Time
	updatedAt time.Time
}

// NewEnterprise создаёт новый Enterprise с валидацией.
func NewEnterprise(
	id, name, address string,
	status EnterpriseStatus,
	latitude, longitude float64,
	createdAt, updatedAt time.Time,
) (*Enterprise, error) {
	if name == "" {
		return nil, ErrInvalidEnterpriseName
	}
	if !isValidStatus(status) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidEnterpriseStatus, status)
	}
	if !isValidLatitude(latitude) {
		return nil, fmt.Errorf("%w: latitude %f out of range [-90, 90]", ErrInvalidCoordinates, latitude)
	}
	if !isValidLongitude(longitude) {
		return nil, fmt.Errorf("%w: longitude %f out of range [-180, 180]", ErrInvalidCoordinates, longitude)
	}

	// Если статус пустой — установить по умолчанию active
	if status == "" {
		status = EnterpriseStatusActive
	}

	return &Enterprise{
		id:        id,
		name:      name,
		address:   address,
		status:    status,
		latitude:  latitude,
		longitude: longitude,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}, nil
}

// Getters.
func (e *Enterprise) ID() string              { return e.id }
func (e *Enterprise) Name() string            { return e.name }
func (e *Enterprise) Address() string          { return e.address }
func (e *Enterprise) Status() EnterpriseStatus { return e.status }
func (e *Enterprise) Latitude() float64        { return e.latitude }
func (e *Enterprise) Longitude() float64       { return e.longitude }
func (e *Enterprise) CreatedAt() time.Time     { return e.createdAt }
func (e *Enterprise) UpdatedAt() time.Time     { return e.updatedAt }

// Setters used by repository after loading from DB.
func (e *Enterprise) SetID(id string)          { e.id = id }
func (e *Enterprise) SetCreatedAt(t time.Time)  { e.createdAt = t }
func (e *Enterprise) SetUpdatedAt(t time.Time)  { e.updatedAt = t }

// UpdateName обновляет название предприятия.
func (e *Enterprise) UpdateName(name string) error {
	if name == "" {
		return ErrInvalidEnterpriseName
	}
	e.name = name
	e.updatedAt = time.Now().UTC()
	return nil
}

// UpdateAddress обновляет адрес предприятия.
func (e *Enterprise) UpdateAddress(address string) {
	e.address = address
	e.updatedAt = time.Now().UTC()
}

// UpdateStatus обновляет статус предприятия.
func (e *Enterprise) UpdateStatus(status EnterpriseStatus) error {
	if !isValidStatus(status) {
		return fmt.Errorf("%w: %s", ErrInvalidEnterpriseStatus, status)
	}
	e.status = status
	e.updatedAt = time.Now().UTC()
	return nil
}

// UpdateCoordinates обновляет координаты предприятия.
func (e *Enterprise) UpdateCoordinates(latitude, longitude float64) error {
	if !isValidLatitude(latitude) {
		return fmt.Errorf("%w: latitude %f out of range [-90, 90]", ErrInvalidCoordinates, latitude)
	}
	if !isValidLongitude(longitude) {
		return fmt.Errorf("%w: longitude %f out of range [-180, 180]", ErrInvalidCoordinates, longitude)
	}
	e.latitude = latitude
	e.longitude = longitude
	e.updatedAt = time.Now().UTC()
	return nil
}

func isValidStatus(s EnterpriseStatus) bool {
	for _, v := range ValidEnterpriseStatuses {
		if s == v {
			return true
		}
	}
	return false
}

func isValidLatitude(lat float64) bool {
	return lat >= -90 && lat <= 90
}

func isValidLongitude(lng float64) bool {
	return lng >= -180 && lng <= 180
}
