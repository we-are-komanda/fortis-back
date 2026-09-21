package domain

import (
	"context"
	"errors"
)

var (
	ErrProvisionUserNotFound = errors.New("provisioning user not found")
	ErrProvisionConflict     = errors.New("enterprise provisioning conflicts with existing data")
)

// EnterpriseProvisionRepository commits creation and the initial membership together.
// An existing ID succeeds only for an unchanged enterprise and an existing membership.
// It must never update an existing enterprise or restore a removed membership.
type EnterpriseProvisionRepository interface {
	Provision(context.Context, *Enterprise, string) error
}

func (e *Enterprise) MatchesProvision(other *Enterprise) bool {
	return other != nil && e.id == other.id && e.name == other.name &&
		e.address == other.address && e.status == other.status &&
		e.latitude == other.latitude && e.longitude == other.longitude
}
