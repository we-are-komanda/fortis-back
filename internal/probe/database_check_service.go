package probe

import (
	"context"
	"errors"
	"time"
)

// DatabaseCheckService uses the existing pool; it never creates connections or runs migrations.
type DatabaseCheckService struct{ ping func(context.Context) error }

func NewDatabaseCheckService(ping func(context.Context) error) *DatabaseCheckService {
	return &DatabaseCheckService{ping: ping}
}

func (service *DatabaseCheckService) Check() []error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if service.ping(ctx) != nil {
		return []error{errors.New("database unavailable")}
	}
	return nil
}
