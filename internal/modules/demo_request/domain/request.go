package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/mail"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

var (
	ErrUnavailable         = errors.New("demo requests unavailable")
	ErrIdempotencyConflict = errors.New("idempotency conflict")
	ErrNotFound            = errors.New("demo request not found")
	ErrDeliveryFailed      = errors.New("notification delivery failed")
)

type RateLimitError struct{ RetryAfter time.Duration }

func (e *RateLimitError) Error() string { return "demo request rate limit exceeded" }

type ValidationError struct{ Fields map[string]string }

func (e *ValidationError) Error() string { return "invalid demo request" }

type Request struct {
	id, eventID, name, organization, email, comment, consentVersion string
	consentAt                                                       time.Time
}

func (r *Request) ID() string             { return r.id }
func (r *Request) EventID() string        { return r.eventID }
func (r *Request) Name() string           { return r.name }
func (r *Request) Organization() string   { return r.organization }
func (r *Request) Email() string          { return r.email }
func (r *Request) Comment() string        { return r.comment }
func (r *Request) ConsentVersion() string { return r.consentVersion }
func (r *Request) ConsentAt() time.Time   { return r.consentAt }
func (r *Request) Digest() string {
	b, _ := json.Marshal([]string{r.name, r.organization, r.email, r.comment, r.consentVersion})
	return Hash(string(b))
}
func Hash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
func NewRequest(id, eventID, name, organization, email, comment, version string, consent bool, allowed []string, now time.Time) (*Request, error) {
	name = strings.TrimSpace(name)
	organization = strings.TrimSpace(organization)
	email = strings.ToLower(strings.TrimSpace(email))
	comment = strings.TrimSpace(strings.ReplaceAll(comment, "\r\n", "\n"))
	fields := map[string]string{}
	check := func(key, value string, min, max int, multiline bool) {
		if !utf8.ValidString(value) || utf8.RuneCountInString(value) < min || utf8.RuneCountInString(value) > max {
			fields[key] = "invalid length"
		}
		for _, r := range value {
			if unicode.IsControl(r) && !(multiline && (r == '\n' || r == '\t')) {
				fields[key] = "invalid characters"
			}
		}
	}
	check("name", name, 2, 120, false)
	check("organization", organization, 1, 200, false)
	check("comment", comment, 0, 2000, true)
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || len(email) > 254 || !strings.Contains(email, "@") || strings.ContainsAny(email, "\r\n") {
		fields["email"] = "invalid email"
	}
	if !consent {
		fields["consent"] = "consent required"
	}
	found := false
	for _, v := range allowed {
		if v != "" && v == version {
			found = true
		}
	}
	if !found {
		fields["consentVersion"] = "unknown consent version"
	}
	if len(fields) > 0 {
		return nil, &ValidationError{Fields: fields}
	}
	return &Request{id: id, eventID: eventID, name: name, organization: organization, email: email, comment: comment, consentVersion: version, consentAt: now}, nil
}
func AddAttempt(times []time.Time, now time.Time, window time.Duration, limit int) ([]time.Time, time.Duration) {
	active := make([]time.Time, 0, len(times)+1)
	for _, at := range times {
		if at.After(now.Add(-window)) {
			active = append(active, at)
		}
	}
	if len(active) >= limit {
		return active, active[0].Add(window).Sub(now)
	}
	return append(active, now), 0
}
func RetryAt(created, now time.Time, attempts int) (time.Time, bool) {
	deadline := created.Add(24 * time.Hour)
	if !now.Before(deadline) {
		return now, true
	}
	delay := time.Hour
	switch attempts {
	case 1:
		delay = time.Minute
	case 2:
		delay = 5 * time.Minute
	case 3:
		delay = 30 * time.Minute
	}
	next := now.Add(delay)
	if next.After(deadline) {
		next = deadline
	}
	return next, false
}
