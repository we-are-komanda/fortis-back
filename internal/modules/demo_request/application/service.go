package application

import (
	"context"
	"errors"
	"github.com/fortis/backend/internal/modules/demo_request/domain"
	"github.com/google/uuid"
	"log/slog"
	"net"
	"strings"
	"time"
)

type Service struct {
	repo     domain.Repository
	cfg      Config
	notifier domain.Notifier
}

func NewService(repo domain.Repository, cfg Config, notifier domain.Notifier) *Service {
	return &Service{repo, cfg, notifier}
}
func (s *Service) Ready() bool {
	if s == nil || !s.cfg.Enabled || s.repo == nil || s.notifier == nil || !s.notifier.Ready() || len(s.cfg.ConsentVersions) == 0 {
		return false
	}
	for _, v := range s.cfg.ConsentVersions {
		if v == "" || len(v) > 100 || strings.ContainsAny(v, "\r\n\t ") {
			return false
		}
	}
	return true
}
func (s *Service) AllowAttempt(ctx context.Context, ip string, now time.Time) error {
	if !s.Ready() {
		return domain.ErrUnavailable
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return domain.ErrUnavailable
	}
	return s.rate(ctx, "ip:"+domain.Hash(parsed.String()), now, 10*time.Minute, 5)
}
func (s *Service) rate(ctx context.Context, key string, now time.Time, window time.Duration, limit int) error {
	err := s.repo.Transaction(ctx, func(tx domain.Repository) error { return rateInTransaction(ctx, tx, key, now, window, limit) })
	if err != nil {
		var limited *domain.RateLimitError
		if errors.As(err, &limited) {
			return err
		}
		return domain.ErrUnavailable
	}
	return nil
}
func rateInTransaction(ctx context.Context, tx domain.Repository, key string, now time.Time, window time.Duration, limit int) error {
	times, err := tx.LockRate(ctx, key)
	if err != nil {
		return err
	}
	times, retry := domain.AddAttempt(times, now, window, limit)
	if retry > 0 {
		return &domain.RateLimitError{RetryAfter: retry}
	}
	return tx.SaveRate(ctx, key, times)
}
func (s *Service) Submit(ctx context.Context, cmd SubmitCommand, now time.Time) (Accepted, error) {
	if !s.Ready() {
		return Accepted{}, domain.ErrUnavailable
	}
	if len(cmd.IdempotencyKey) < 1 || len(cmd.IdempotencyKey) > 128 || strings.IndexFunc(cmd.IdempotencyKey, func(r rune) bool { return r < 33 || r > 126 }) >= 0 {
		return Accepted{}, &domain.ValidationError{Fields: map[string]string{"Idempotency-Key": "invalid idempotency key"}}
	}
	request, err := domain.NewRequest(uuid.NewString(), uuid.NewString(), cmd.Name, cmd.Organization, cmd.Email, cmd.Comment, cmd.ConsentVersion, cmd.Consent, s.cfg.ConsentVersions, now)
	if err != nil {
		return Accepted{}, err
	}
	key := domain.Hash("demo-request:create:" + cmd.IdempotencyKey)
	expires := now.Add(24 * time.Hour)
	result := Accepted{RequestID: request.ID()}
	err = s.repo.Transaction(ctx, func(tx domain.Repository) error {
		receipt, err := tx.LockReceipt(ctx, key, expires)
		if err != nil {
			return err
		}
		if receipt.RequestID() != "" && receipt.ExpiresAt().After(now) {
			if receipt.Digest() != request.Digest() {
				return domain.ErrIdempotencyConflict
			}
			result = Accepted{RequestID: receipt.RequestID(), Replayed: true}
			return nil
		}
		if err := rateInTransaction(ctx, tx, "email:"+domain.Hash(request.Email()), now, time.Hour, 3); err != nil {
			return err
		}
		if err := tx.SaveRequest(ctx, request); err != nil {
			return err
		}
		if err := tx.SaveOutbox(ctx, request); err != nil {
			return err
		}
		return tx.SaveReceipt(ctx, domain.NewReceipt(key, request.Digest(), request.ID(), expires))
	})
	if err != nil {
		var limited *domain.RateLimitError
		if errors.Is(err, domain.ErrIdempotencyConflict) || errors.As(err, &limited) {
			return Accepted{}, err
		}
		slog.Error("demo request persistence failed", "requestId", request.ID())
		return Accepted{}, domain.ErrUnavailable
	}
	return result, nil
}
func (s *Service) DeliverNext(ctx context.Context, now time.Time) (bool, error) {
	if !s.Ready() {
		return false, domain.ErrUnavailable
	}
	delivery, err := s.repo.Claim(ctx, now, uuid.NewString(), now.Add(time.Minute))
	if err != nil {
		return false, domain.ErrUnavailable
	}
	if delivery == nil {
		return false, nil
	}
	if !now.Before(delivery.Request().ConsentAt().Add(24 * time.Hour)) {
		delivery.Fail(now)
	} else {
		sendCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		err = s.notifier.Send(sendCtx, delivery.Request())
		cancel()
		if err != nil {
			delivery.Fail(now)
		} else {
			delivery.Complete()
		}
	}
	if err := s.repo.Finish(ctx, delivery); err != nil {
		return true, domain.ErrUnavailable
	}
	if delivery.Status() == "failed" {
		slog.Error("demo request delivery exhausted", "eventId", delivery.Request().EventID(), "requestId", delivery.Request().ID(), "attempts", delivery.Attempts())
	}
	return true, nil
}
func (s *Service) Run(ctx context.Context) {
	if !s.Ready() {
		return
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		for ctx.Err() == nil {
			worked, err := s.DeliverNext(ctx, time.Now().UTC())
			if err != nil {
				slog.Error("demo request worker storage unavailable")
				break
			}
			if !worked {
				break
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
func (s *Service) ListForOperator(ctx context.Context, limit, offset int) ([]*domain.Delivery, int64, error) {
	return s.repo.List(ctx, limit, offset)
}
