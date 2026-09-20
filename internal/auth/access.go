package auth

import "errors"

var (
	ErrIdentityRequired = errors.New("authenticated identity required")
	ErrNotFound         = errors.New("resource not found")
	ErrForbidden        = errors.New("operation forbidden")
)

func RequireIdentity(userID string) error {
	if userID == "" {
		return ErrIdentityRequired
	}
	return nil
}
