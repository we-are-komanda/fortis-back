package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestRequestNormalizesAndRejectsUnsafeInput(t *testing.T) {
	now := time.Date(2026, 9, 20, 0, 0, 0, 0, time.UTC)
	create := func(name, org, email, comment, consentVersion string, consent bool) (*Request, error) {
		return NewRequest("lead", "event", name, org, email, comment, consentVersion, consent, []string{"synthetic-test-v1"}, now)
	}
	got, err := create("  Test Person  ", "  Test Org ", " Person@Example.TEST ", " hello\r\nworld ", "synthetic-test-v1", true)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name() != "Test Person" || got.Organization() != "Test Org" || got.Email() != "person@example.test" || got.Comment() != "hello\nworld" {
		t.Fatal("request not normalized")
	}
	same, err := create("Test Person", "Test Org", "person@example.test", "hello\nworld", "synthetic-test-v1", true)
	if err != nil || got.Digest() != same.Digest() {
		t.Fatal("normalized equivalent input must be idempotent")
	}
	for _, tc := range []struct {
		name, org, email, comment, version string
		consent                            bool
	}{
		{" ", "Org", "person@example.test", "", "synthetic-test-v1", true},
		{"Name\r\nBcc: injected", "Org", "person@example.test", "", "synthetic-test-v1", true},
		{"Name", "Org", "name@example.test\r\nBcc: victim@example.test", "", "synthetic-test-v1", true},
		{"Name", "Org", "Person <person@example.test>", "", "synthetic-test-v1", true},
		{"Name", "Org", "person@example.test", strings.Repeat("я", 2001), "synthetic-test-v1", true},
		{"Name", "Org", "person@example.test", "", "unknown", true},
		{"Name", "Org", "person@example.test", "", "synthetic-test-v1", false},
	} {
		_, err := create(tc.name, tc.org, tc.email, tc.comment, tc.version, tc.consent)
		var invalid *ValidationError
		if !errors.As(err, &invalid) {
			t.Errorf("expected field validation, got %v", err)
		}
	}
}

func TestRollingLimitAndRetryDeadline(t *testing.T) {
	now := time.Date(2026, 9, 20, 0, 0, 0, 0, time.UTC)
	times, retry := AddAttempt([]time.Time{now.Add(-10 * time.Minute), now.Add(-9 * time.Minute)}, now, 10*time.Minute, 2)
	if len(times) != 2 || retry != 0 {
		t.Fatal("expired boundary must free a slot")
	}
	_, retry = AddAttempt(times, now, 10*time.Minute, 2)
	if retry != time.Minute {
		t.Fatalf("rolling retry = %s", retry)
	}
	for attempts, want := range map[int]time.Duration{1: time.Minute, 2: 5 * time.Minute, 3: 30 * time.Minute, 4: time.Hour} {
		due, failed := RetryAt(now, now, attempts)
		if failed || due.Sub(now) != want {
			t.Fatalf("attempt %d: %v %v", attempts, due, failed)
		}
	}
	due, failed := RetryAt(now, now.Add(23*time.Hour+30*time.Minute), 10)
	if failed || !due.Equal(now.Add(24*time.Hour)) {
		t.Fatal("retry must be capped at 24 hours")
	}
	_, failed = RetryAt(now, now.Add(24*time.Hour), 11)
	if !failed {
		t.Fatal("delivery must expire after 24 hours")
	}
}
