package infrastructure

import (
	"github.com/fortis/backend/internal/modules/demo_request/domain"
	"time"
)

type RequestModel struct {
	ID                                                 string `gorm:"type:uuid;primaryKey"`
	EventID                                            string `gorm:"type:uuid"`
	Name, Organization, Email, Comment, ConsentVersion string
	ConsentAt                                          time.Time
}

func (RequestModel) TableName() string { return "demo_requests" }
func (m RequestModel) toDomain() (*domain.Request, error) {
	return domain.NewRequest(m.ID, m.EventID, m.Name, m.Organization, m.Email, m.Comment, m.ConsentVersion, true, []string{m.ConsentVersion}, m.ConsentAt)
}

type OutboxModel struct {
	ID            string `gorm:"type:uuid;primaryKey"`
	RequestID     string `gorm:"type:uuid"`
	Status        string
	Attempts      int
	NextAttemptAt time.Time
	LeaseToken    string
	LeaseUntil    *time.Time
	LastError     string
	Request       RequestModel `gorm:"foreignKey:RequestID;references:ID"`
}

func (OutboxModel) TableName() string { return "demo_request_outbox" }
func (m OutboxModel) toDomain() (*domain.Delivery, error) {
	request, err := m.Request.toDomain()
	if err != nil {
		return nil, err
	}
	var lease time.Time
	if m.LeaseUntil != nil {
		lease = *m.LeaseUntil
	}
	return domain.RestoreDelivery(request, m.Status, m.LeaseToken, m.LastError, m.Attempts, m.NextAttemptAt, lease), nil
}

type ReceiptModel struct {
	Key       string `gorm:"primaryKey"`
	Digest    string
	RequestID *string `gorm:"type:uuid"`
	ExpiresAt time.Time
}

func (ReceiptModel) TableName() string { return "demo_request_receipts" }

type RateModel struct {
	Key        string `gorm:"primaryKey"`
	Timestamps string `gorm:"type:jsonb"`
}

func (RateModel) TableName() string { return "demo_request_rates" }
