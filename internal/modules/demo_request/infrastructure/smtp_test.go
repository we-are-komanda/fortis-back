package infrastructure

import (
	"context"
	"github.com/fortis/backend/internal/modules/demo_request/domain"
	"github.com/stretchr/testify/require"
	"io"
	"mime/quotedprintable"
	"net/mail"
	"strings"
	"testing"
	"time"
)

func TestSMTPConfigurationAndMessageDoNotAllowHeaderInjection(t *testing.T) {
	cfg := SMTPConfig{Address: "smtp.example.test:587", From: "fortis@example.test", To: "operator@example.test"}
	require.True(t, NewSMTPNotifier(cfg).Ready())
	for _, bad := range []SMTPConfig{{}, {Address: "smtp.example.test", From: cfg.From, To: cfg.To}, {Address: cfg.Address, From: cfg.From, To: "operator@example.test\r\nBcc: victim@example.test"}, {Address: cfg.Address, From: cfg.From, To: cfg.To, Username: "user"}} {
		n := NewSMTPNotifier(bad)
		require.False(t, n.Ready())
		require.ErrorIs(t, n.Send(context.Background(), nil), domain.ErrUnavailable)
	}
	request, err := domain.NewRequest("lead", "event-id", "Тестовое имя", "Org", "person@example.test", "hello\nBcc: victim@example.test\n<html>", "synthetic-test-v1", true, []string{"synthetic-test-v1"}, time.Now())
	require.NoError(t, err)
	raw := message(cfg, request)
	parsed, err := mail.ReadMessage(strings.NewReader(string(raw)))
	require.NoError(t, err)
	require.Equal(t, "<event-id@fortis.invalid>", parsed.Header.Get("Message-ID"))
	require.Empty(t, parsed.Header.Get("Bcc"))
	require.Equal(t, cfg.To, parsed.Header.Get("To"))
	require.Equal(t, "text/plain; charset=UTF-8", parsed.Header.Get("Content-Type"))
	body, err := io.ReadAll(quotedprintable.NewReader(parsed.Body))
	require.NoError(t, err)
	require.Contains(t, string(body), "Bcc: victim@example.test")
	require.Contains(t, string(body), "Тестовое имя")
}
