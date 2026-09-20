package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/fortis/backend/internal/modules/demo_request/domain"
	"github.com/fortis/backend/internal/rdbms"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"time"
)

type Repository struct{ executor rdbms.Executor }

func NewRepository(executor rdbms.Executor) domain.Repository { return &Repository{executor} }
func (r *Repository) db(ctx context.Context) *gorm.DB {
	// Lead fields must never enter the application's SQL/access log.
	return r.executor.WithContext(ctx).Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})
}
func (r *Repository) Transaction(ctx context.Context, fn func(domain.Repository) error) error {
	return r.db(ctx).Transaction(func(tx *gorm.DB) error { return fn(NewRepository(tx)) })
}
func (r *Repository) LockRate(ctx context.Context, key string) ([]time.Time, error) {
	db := r.db(ctx)
	m := RateModel{Key: key, Timestamps: "[]"}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&m).Error; err != nil {
		return nil, err
	}
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("key = ?", key).First(&m).Error; err != nil {
		return nil, err
	}
	var times []time.Time
	err := json.Unmarshal([]byte(m.Timestamps), &times)
	return times, err
}
func (r *Repository) SaveRate(ctx context.Context, key string, times []time.Time) error {
	b, err := json.Marshal(times)
	if err != nil {
		return err
	}
	return r.db(ctx).Model(&RateModel{}).Where("key = ?", key).Update("timestamps", string(b)).Error
}
func (r *Repository) LockReceipt(ctx context.Context, key string, expires time.Time) (*domain.Receipt, error) {
	db := r.db(ctx)
	m := ReceiptModel{Key: key, ExpiresAt: expires}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&m).Error; err != nil {
		return nil, err
	}
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("key = ?", key).First(&m).Error; err != nil {
		return nil, err
	}
	id := ""
	if m.RequestID != nil {
		id = *m.RequestID
	}
	return domain.NewReceipt(m.Key, m.Digest, id, m.ExpiresAt), nil
}
func (r *Repository) SaveReceipt(ctx context.Context, receipt *domain.Receipt) error {
	return r.db(ctx).Model(&ReceiptModel{}).Where("key = ?", receipt.Key()).Updates(map[string]any{"digest": receipt.Digest(), "request_id": receipt.RequestID(), "expires_at": receipt.ExpiresAt()}).Error
}
func (r *Repository) SaveRequest(ctx context.Context, request *domain.Request) error {
	return r.db(ctx).Create(&RequestModel{ID: request.ID(), EventID: request.EventID(), Name: request.Name(), Organization: request.Organization(), Email: request.Email(), Comment: request.Comment(), ConsentVersion: request.ConsentVersion(), ConsentAt: request.ConsentAt()}).Error
}
func (r *Repository) SaveOutbox(ctx context.Context, request *domain.Request) error {
	return r.db(ctx).Omit("Request").Create(&OutboxModel{ID: request.EventID(), RequestID: request.ID(), Status: "pending", NextAttemptAt: request.ConsentAt()}).Error
}
func (r *Repository) Claim(ctx context.Context, now time.Time, token string, until time.Time) (delivery *domain.Delivery, err error) {
	err = r.db(ctx).Transaction(func(tx *gorm.DB) error {
		var m OutboxModel
		err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Where("status = ? AND next_attempt_at <= ? AND (lease_until IS NULL OR lease_until <= ?)", "pending", now, now).Order("next_attempt_at, id").First(&m).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		m.Attempts++
		m.LeaseToken = token
		m.LeaseUntil = &until
		if err := tx.Model(&OutboxModel{}).Where("id = ?", m.ID).Updates(map[string]any{"attempts": m.Attempts, "lease_token": token, "lease_until": until}).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", m.RequestID).First(&m.Request).Error; err != nil {
			return err
		}
		delivery, err = m.toDomain()
		return err
	})
	return
}
func (r *Repository) Finish(ctx context.Context, delivery *domain.Delivery) error {
	return r.db(ctx).Model(&OutboxModel{}).Where("id = ? AND lease_token = ?", delivery.Request().EventID(), delivery.LeaseToken()).Updates(map[string]any{"status": delivery.Status(), "next_attempt_at": delivery.DueAt(), "last_error": delivery.LastError(), "lease_token": "", "lease_until": nil}).Error
}
func (r *Repository) List(ctx context.Context, limit, offset int) ([]*domain.Delivery, int64, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	db := r.db(ctx)
	var total int64
	if err := db.Model(&RequestModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var models []OutboxModel
	if err := db.Preload("Request").Joins("JOIN demo_requests ON demo_requests.id = demo_request_outbox.request_id").Order("demo_requests.consent_at DESC, demo_request_outbox.id").Limit(limit).Offset(offset).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	rows := make([]*domain.Delivery, 0, len(models))
	for _, m := range models {
		d, err := m.toDomain()
		if err != nil {
			return nil, 0, err
		}
		rows = append(rows, d)
	}
	return rows, total, nil
}
