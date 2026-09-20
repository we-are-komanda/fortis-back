package ui

import "time"

// swagger:model DemoRequestInput
type DemoRequestInput struct {
	Name           string `json:"name"`
	Organization   string `json:"organization"`
	Email          string `json:"email"`
	Comment        string `json:"comment"`
	Consent        bool   `json:"consent"`
	ConsentVersion string `json:"consentVersion"`
}

// swagger:model DemoRequestAccepted
type DemoRequestAccepted struct {
	RequestID string `json:"requestId"`
	Status    string `json:"status"`
}

// swagger:parameters submitDemoRequest
type DemoRequestParameters struct {
	// in: body
	// required: true
	Body DemoRequestInput
	// in: header
	// required: true
	IdempotencyKey string `json:"Idempotency-Key"`
}
type RegistryItem struct {
	RequestID      string    `json:"requestId"`
	EventID        string    `json:"eventId"`
	Name           string    `json:"name"`
	Organization   string    `json:"organization"`
	Email          string    `json:"email"`
	Comment        string    `json:"comment"`
	ConsentVersion string    `json:"consentVersion"`
	ConsentAt      time.Time `json:"consentAt"`
	Status         string    `json:"status"`
	Attempts       int       `json:"attempts"`
	LastError      string    `json:"lastError"`
	NextAttemptAt  time.Time `json:"nextAttemptAt"`
}
type RegistryPage struct {
	Items      []RegistryItem `json:"items"`
	TotalItems int64          `json:"totalItems"`
}
