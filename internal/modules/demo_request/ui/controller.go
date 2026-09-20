package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/fortis/backend/internal/modules/demo_request/application"
	"github.com/fortis/backend/internal/modules/demo_request/domain"
	"github.com/fortis/backend/pkg/handlers"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"io"
	"math"
	"mime"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type TransportConfig struct{ AllowedOrigins, TrustedProxyCIDRs []string }
type Controller struct {
	service *application.Service
	cfg     TransportConfig
}

func NewController(service *application.Service, cfg TransportConfig) *Controller {
	return &Controller{service, cfg}
}
func (c *Controller) ready() bool {
	if c == nil || c.service == nil || !c.service.Ready() || len(c.cfg.AllowedOrigins) == 0 {
		return false
	}
	for _, origin := range c.cfg.AllowedOrigins {
		u, err := url.Parse(origin)
		if err != nil || u.Host == "" || u.User != nil || origin != u.Scheme+"://"+u.Host {
			return false
		}
		if u.Scheme != "https" {
			ip := net.ParseIP(u.Hostname())
			if u.Scheme != "http" || (u.Hostname() != "localhost" && (ip == nil || !ip.IsLoopback())) {
				return false
			}
		}
	}
	for _, cidr := range c.cfg.TrustedProxyCIDRs {
		if _, _, err := net.ParseCIDR(cidr); err != nil {
			return false
		}
	}
	return true
}
func (c *Controller) trusted(ip net.IP) bool {
	for _, cidr := range c.cfg.TrustedProxyCIDRs {
		_, network, err := net.ParseCIDR(cidr)
		if err == nil && network.Contains(ip) {
			return true
		}
	}
	return false
}
func (c *Controller) clientIP(ctx *fasthttp.RequestCtx) (string, error) {
	peer := ctx.RemoteIP()
	if peer == nil {
		return "", domain.ErrUnavailable
	}
	if !c.trusted(peer) {
		return peer.String(), nil
	}
	parts := strings.Split(string(ctx.Request.Header.Peek("X-Forwarded-For")), ",")
	if len(parts) > 8 {
		return "", &domain.ValidationError{Fields: map[string]string{"clientIP": "invalid trusted proxy client IP"}}
	}
	ips := make([]net.IP, len(parts))
	for i, part := range parts {
		ips[i] = net.ParseIP(strings.TrimSpace(part))
		if ips[i] == nil {
			return "", &domain.ValidationError{Fields: map[string]string{"clientIP": "invalid trusted proxy client IP"}}
		}
	}
	for i := len(ips) - 1; i >= 0; i-- {
		if !c.trusted(ips[i]) || i == 0 {
			return ips[i].String(), nil
		}
	}
	return "", domain.ErrUnavailable
}

// swagger:route POST /api/v1/demo-requests demoRequests submitDemoRequest
// Приём заявки на демонстрацию
//
// Сохраняет заявку и уведомление атомарно; повтор ключа возвращает прежний номер.
// Consumes:
// - application/json
// Produces:
// - application/json
// Responses:
// 201: DemoRequestAccepted
// 200: DemoRequestAccepted
// 400: description: Некорректные поля
// 403: description: Запрещённый источник запроса
// 409: description: Конфликт ключа идемпотентности
// 413: description: Превышен размер тела
// 429: description: Превышен лимит заявок
// 503: description: Приём заявок недоступен
func (c *Controller) Submit(ctx *fasthttp.RequestCtx) {
	requestID := string(ctx.Request.Header.Peek("X-Request-ID"))
	if _, err := uuid.Parse(requestID); err != nil {
		requestID = uuid.NewString()
	}
	ctx.Response.Header.Set("X-Request-ID", requestID)
	ctx.Response.Header.Set("Cache-Control", "no-store")
	ctx.SetContentType("application/json")
	if !c.ready() {
		fail(ctx, 503, "unavailable", "demo requests unavailable", nil)
		return
	}
	if !ctx.IsPost() {
		fail(ctx, 405, "method_not_allowed", "method not allowed", nil)
		return
	}
	origin := string(ctx.Request.Header.Peek("Origin"))
	allowed := false
	for _, value := range c.cfg.AllowedOrigins {
		if origin == value {
			allowed = true
		}
	}
	if !allowed {
		fail(ctx, 403, "origin_forbidden", "origin forbidden", nil)
		return
	}
	ip, err := c.clientIP(ctx)
	if err != nil {
		writeError(ctx, err)
		return
	}
	now := time.Now().UTC()
	if err := c.service.AllowAttempt(ctx, ip, now); err != nil {
		writeError(ctx, err)
		return
	}
	if ctx.Request.Header.ContentLength() > 16*1024 || len(ctx.PostBody()) > 16*1024 {
		fail(ctx, 413, "payload_too_large", "request body exceeds 16 KiB", nil)
		return
	}
	contentType, _, err := mime.ParseMediaType(string(ctx.Request.Header.ContentType()))
	if err != nil || contentType != "application/json" {
		fail(ctx, 415, "unsupported_media_type", "application/json required", nil)
		return
	}
	var input DemoRequestInput
	decoder := json.NewDecoder(bytes.NewReader(ctx.PostBody()))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		fail(ctx, 400, "invalid_json", "invalid JSON request", nil)
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		fail(ctx, 400, "invalid_json", "one JSON object required", nil)
		return
	}
	accepted, err := c.service.Submit(ctx, toCommand(input, string(ctx.Request.Header.Peek("Idempotency-Key"))), now)
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.SetStatusCode(201)
	if accepted.Replayed {
		ctx.SetStatusCode(200)
	}
	body, _ := json.Marshal(DemoRequestAccepted{RequestID: accepted.RequestID, Status: "received"})
	ctx.SetBody(body)
}
func writeError(ctx *fasthttp.RequestCtx, err error) {
	var invalid *domain.ValidationError
	var limited *domain.RateLimitError
	switch {
	case errors.As(err, &invalid):
		fail(ctx, 400, "validation_error", "invalid request fields", invalid.Fields)
	case errors.As(err, &limited):
		ctx.Response.Header.Set("Retry-After", strconv.Itoa(max(1, int(math.Ceil(limited.RetryAfter.Seconds())))))
		fail(ctx, 429, "rate_limited", "request limit exceeded", nil)
	case errors.Is(err, domain.ErrIdempotencyConflict):
		fail(ctx, 409, "idempotency_conflict", "idempotency key already used for different input", nil)
	default:
		fail(ctx, 503, "unavailable", "demo requests unavailable", nil)
	}
}
func fail(ctx *fasthttp.RequestCtx, status int, code, message string, fields map[string]string) {
	base := handlers.ResponseBody{}
	handlers.ErrorHandler(ctx, code, message, &base, status)
	detail := map[string]any{"code": code, "message": message, "requestId": string(ctx.Response.Header.Peek("X-Request-ID")), "retryable": status == 503 || status == 429}
	if fields != nil {
		detail["details"] = map[string]any{"fields": fields}
	}
	body, _ := json.Marshal(struct {
		handlers.ResponseBody
		Error map[string]any `json:"error"`
	}{base, detail})
	ctx.SetBody(body)
}

// WriteRegistry is for the local trusted operator process only; never register an HTTP route to it.
func WriteRegistry(ctx context.Context, service *application.Service, output io.Writer, limit, offset int) error {
	rows, total, err := service.ListForOperator(ctx, limit, offset)
	if err != nil {
		return err
	}
	return json.NewEncoder(output).Encode(toRegistryPage(rows, total))
}
