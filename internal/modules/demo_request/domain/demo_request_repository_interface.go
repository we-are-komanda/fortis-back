package domain

import (
	"context"
	"time"
)

type Receipt struct {
	key, digest, requestID string
	expiresAt              time.Time
}

func NewReceipt(key, digest, requestID string, expiresAt time.Time) *Receipt {
	return &Receipt{key, digest, requestID, expiresAt}
}
func (r *Receipt) Key() string          { return r.key }
func (r *Receipt) Digest() string       { return r.digest }
func (r *Receipt) RequestID() string    { return r.requestID }
func (r *Receipt) ExpiresAt() time.Time { return r.expiresAt }

type Delivery struct {
	request                       *Request
	status, leaseToken, lastError string
	attempts                      int
	dueAt, leaseUntil             time.Time
}

func RestoreDelivery(request *Request, status, leaseToken, lastError string, attempts int, dueAt, leaseUntil time.Time) *Delivery {
	return &Delivery{request, status, leaseToken, lastError, attempts, dueAt, leaseUntil}
}
func (d *Delivery) Request() *Request     { return d.request }
func (d *Delivery) Status() string        { return d.status }
func (d *Delivery) LeaseToken() string    { return d.leaseToken }
func (d *Delivery) LastError() string     { return d.lastError }
func (d *Delivery) Attempts() int         { return d.attempts }
func (d *Delivery) DueAt() time.Time      { return d.dueAt }
func (d *Delivery) LeaseUntil() time.Time { return d.leaseUntil }
func (d *Delivery) Complete()             { d.status = "sent"; d.lastError = "" }
func (d *Delivery) Fail(now time.Time) {
	var expired bool
	d.dueAt, expired = RetryAt(d.request.ConsentAt(), now, d.attempts)
	d.status = "pending"
	if expired {
		d.status = "failed"
	}
	d.lastError = "notification_delivery_failed"
}

type Repository interface {
	Transaction(context.Context, func(Repository) error) error
	LockRate(context.Context, string) ([]time.Time, error)
	SaveRate(context.Context, string, []time.Time) error
	LockReceipt(context.Context, string, time.Time) (*Receipt, error)
	SaveReceipt(context.Context, *Receipt) error
	SaveRequest(context.Context, *Request) error
	SaveOutbox(context.Context, *Request) error
	Claim(context.Context, time.Time, string, time.Time) (*Delivery, error)
	Finish(context.Context, *Delivery) error
	List(context.Context, int, int) ([]*Delivery, int64, error)
}

type Notifier interface {
	Ready() bool
	Send(context.Context, *Request) error
}
