package infrastructure

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/fortis/backend/internal/modules/demo_request/domain"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

type SMTPConfig struct{ Address, Username, Password, From, To string }
type SMTPNotifier struct{ cfg SMTPConfig }

func NewSMTPNotifier(cfg SMTPConfig) *SMTPNotifier { return &SMTPNotifier{cfg} }
func (n *SMTPNotifier) Ready() bool {
	if n == nil {
		return false
	}
	host, port, err := net.SplitHostPort(n.cfg.Address)
	if err != nil || host == "" || strings.ContainsAny(host, "\r\n ") {
		return false
	}
	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > 65535 {
		return false
	}
	for _, value := range []string{n.cfg.From, n.cfg.To} {
		address, err := mail.ParseAddress(value)
		if err != nil || address.Address != value || strings.ContainsAny(value, "\r\n") {
			return false
		}
	}
	return (n.cfg.Username == "") == (n.cfg.Password == "")
}
func (n *SMTPNotifier) Send(ctx context.Context, request *domain.Request) error {
	if !n.Ready() || request == nil {
		return domain.ErrUnavailable
	}
	host, _, _ := net.SplitHostPort(n.cfg.Address)
	conn, err := (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, "tcp", n.cfg.Address)
	if err != nil {
		return domain.ErrDeliveryFailed
	}
	defer conn.Close()
	deadline := time.Now().Add(30 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	if err := conn.SetDeadline(deadline); err != nil {
		return domain.ErrDeliveryFailed
	}
	stop := context.AfterFunc(ctx, func() { _ = conn.SetDeadline(time.Now()) })
	defer stop()
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return domain.ErrDeliveryFailed
	}
	defer client.Close()
	// No plaintext fallback, even when the server omits STARTTLS.
	if ok, _ := client.Extension("STARTTLS"); !ok {
		return domain.ErrDeliveryFailed
	}
	if err := client.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}); err != nil {
		return domain.ErrDeliveryFailed
	}
	if n.cfg.Username != "" {
		if err := client.Auth(smtp.PlainAuth("", n.cfg.Username, n.cfg.Password, host)); err != nil {
			return domain.ErrDeliveryFailed
		}
	}
	if err := client.Mail(n.cfg.From); err != nil {
		return domain.ErrDeliveryFailed
	}
	if err := client.Rcpt(n.cfg.To); err != nil {
		return domain.ErrDeliveryFailed
	}
	writer, err := client.Data()
	if err != nil {
		return domain.ErrDeliveryFailed
	}
	if _, err = writer.Write(message(n.cfg, request)); err != nil {
		_ = writer.Close()
		return domain.ErrDeliveryFailed
	}
	if err = writer.Close(); err != nil {
		return domain.ErrDeliveryFailed
	}
	if err = client.Quit(); err != nil {
		return domain.ErrDeliveryFailed
	}
	return nil
}
func message(cfg SMTPConfig, request *domain.Request) []byte {
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "From: %s\r\nTo: %s\r\nSubject: Fortis demo request\r\nMessage-ID: <%s@fortis.invalid>\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n", cfg.From, cfg.To, request.EventID())
	writer := quotedprintable.NewWriter(&buffer)
	fmt.Fprintf(writer, "Request: %s\nEvent: %s\nName: %s\nOrganization: %s\nEmail: %s\nComment:\n%s\n", request.ID(), request.EventID(), request.Name(), request.Organization(), request.Email(), request.Comment())
	_ = writer.Close()
	return buffer.Bytes()
}
