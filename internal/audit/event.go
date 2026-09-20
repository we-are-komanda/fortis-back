// Package audit stores the minimal append-only event shared by owning modules.
package audit

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RequestIDKey is populated by the HTTP middleware from a server-generated UUID.
const RequestIDKey = "fortis.request_id"

func RequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		if parsed, err := uuid.Parse(id); err == nil {
			return parsed.String()
		}
	}
	return uuid.NewString()
}

// Event deliberately contains identifiers and version transitions only.
// Actor/entity identifiers are historical references without live-user/project FKs.
type Event struct {
	EventID         string  `gorm:"primaryKey;type:uuid"`
	ActorID         string  `gorm:"type:uuid;not null"`
	EnterpriseID    *string `gorm:"type:uuid"`
	EntityType      string
	EntityID        string `gorm:"type:uuid;not null"`
	Action          string
	PreviousVersion *int
	NewVersion      *int
	OccurredAt      time.Time
	RequestID       string `gorm:"type:uuid;not null"`
}

func (Event) TableName() string { return "audit_events" }

// Append must receive the caller's active transaction so an event cannot outlive
// a failed write, and an audit failure cannot leave a successful mutation behind.
func Append(tx *gorm.DB, event Event) error {
	event.EventID = uuid.NewString()
	event.OccurredAt = time.Now().UTC()
	return tx.Create(&event).Error
}
