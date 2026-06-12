package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEnterprise_Success(t *testing.T) {
	now := time.Now().UTC()
	e, err := NewEnterprise(
		"550e8400-e29b-41d4-a716-446655440000",
		"Нефтебаза Альфа",
		"г. Москва, ул. Ленина, д. 1",
		EnterpriseStatusActive,
		55.75,
		37.62,
		now,
		now,
	)

	require.NoError(t, err)
	assert.NotNil(t, e)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", e.ID())
	assert.Equal(t, "Нефтебаза Альфа", e.Name())
	assert.Equal(t, "г. Москва, ул. Ленина, д. 1", e.Address())
	assert.Equal(t, EnterpriseStatusActive, e.Status())
	assert.InDelta(t, 55.75, e.Latitude(), 0.0001)
	assert.InDelta(t, 37.62, e.Longitude(), 0.0001)
}

func TestNewEnterprise_EmptyName(t *testing.T) {
	now := time.Now().UTC()
	e, err := NewEnterprise(
		"test-id",
		"",
		"address",
		EnterpriseStatusActive,
		55.75,
		37.62,
		now,
		now,
	)

	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidEnterpriseName)
	assert.Nil(t, e)
}

func TestNewEnterprise_InvalidStatus(t *testing.T) {
	now := time.Now().UTC()
	e, err := NewEnterprise(
		"test-id",
		"Test",
		"address",
		EnterpriseStatus("invalid_status"),
		55.75,
		37.62,
		now,
		now,
	)

	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidEnterpriseStatus)
	assert.Nil(t, e)
}

func TestNewEnterprise_InvalidLatitude(t *testing.T) {
	now := time.Now().UTC()
	e, err := NewEnterprise(
		"test-id",
		"Test",
		"address",
		EnterpriseStatusActive,
		100.0,
		37.62,
		now,
		now,
	)

	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidCoordinates)
	assert.Nil(t, e)
}

func TestNewEnterprise_InvalidLongitude(t *testing.T) {
	now := time.Now().UTC()
	e, err := NewEnterprise(
		"test-id",
		"Test",
		"address",
		EnterpriseStatusActive,
		55.75,
		200.0,
		now,
		now,
	)

	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidCoordinates)
	assert.Nil(t, e)
}

func TestEnterprise_UpdateName(t *testing.T) {
	e := validEnterprise()
	err := e.UpdateName("Новое название")
	require.NoError(t, err)
	assert.Equal(t, "Новое название", e.Name())
}

func TestEnterprise_UpdateName_Empty(t *testing.T) {
	e := validEnterprise()
	err := e.UpdateName("")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidEnterpriseName)
}

func TestEnterprise_UpdateAddress(t *testing.T) {
	e := validEnterprise()
	e.UpdateAddress("Новый адрес")
	assert.Equal(t, "Новый адрес", e.Address())
}

func TestEnterprise_UpdateStatus(t *testing.T) {
	e := validEnterprise()
	err := e.UpdateStatus(EnterpriseStatusOffline)
	require.NoError(t, err)
	assert.Equal(t, EnterpriseStatusOffline, e.Status())
}

func TestEnterprise_UpdateStatus_Invalid(t *testing.T) {
	e := validEnterprise()
	err := e.UpdateStatus(EnterpriseStatus("invalid"))
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidEnterpriseStatus)
}

func TestEnterprise_UpdateCoordinates(t *testing.T) {
	e := validEnterprise()
	err := e.UpdateCoordinates(60.0, 30.0)
	require.NoError(t, err)
	assert.InDelta(t, 60.0, e.Latitude(), 0.0001)
	assert.InDelta(t, 30.0, e.Longitude(), 0.0001)
}

func TestEnterprise_UpdateCoordinates_Invalid(t *testing.T) {
	e := validEnterprise()
	err := e.UpdateCoordinates(91.0, 0.0)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCoordinates)
}

func validEnterprise() *Enterprise {
	now := time.Now().UTC()
	e, _ := NewEnterprise(
		"test-id",
		"Тестовое предприятие",
		"Адрес",
		EnterpriseStatusActive,
		55.75,
		37.62,
		now,
		now,
	)
	return e
}
