package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	da "github.com/fortis/backend/internal/modules/demo_request/application"
	dd "github.com/fortis/backend/internal/modules/demo_request/domain"
	di "github.com/fortis/backend/internal/modules/demo_request/infrastructure"
	du "github.com/fortis/backend/internal/modules/demo_request/ui"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func frc12Database(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("FRC12_TEST_DSN")
	if dsn == "" {
		t.Skip("FRC12_TEST_DSN unset: real PostgreSQL not verified")
	}
	base, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	schema := "frc12_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	require.NoError(t, base.Exec("CREATE SCHEMA "+schema).Error)
	db, err := gorm.Open(postgres.Open(dsn+" search_path="+schema), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	t.Cleanup(func() {
		pool, _ := db.DB()
		pool.Close()
		base.Exec("DROP SCHEMA " + schema + " CASCADE")
		pool, _ = base.DB()
		pool.Close()
	})
	files, err := filepath.Glob("../../migrations/*_demo_requests.up.sql")
	require.NoError(t, err)
	require.Len(t, files, 1)
	sql, err := os.ReadFile(files[0])
	require.NoError(t, err)
	require.NoError(t, db.Exec(string(sql)).Error)
	return db
}

type frc12Notifier struct {
	ready, fail bool
	events      []string
}

func (n *frc12Notifier) Ready() bool { return n.ready }
func (n *frc12Notifier) Send(_ context.Context, r *dd.Request) error {
	n.events = append(n.events, r.EventID())
	if n.fail {
		return errors.New("synthetic provider failed with secret detail")
	}
	return nil
}
func frc12Command() da.SubmitCommand {
	return da.SubmitCommand{Name: " Test Person ", Organization: " Test Org ", Email: "Person@Example.TEST", Comment: "test only", Consent: true, ConsentVersion: "synthetic-test-v1", IdempotencyKey: uuid.NewString()}
}
func frc12Service(db *gorm.DB, n *frc12Notifier) *da.Service {
	return da.NewService(di.NewRepository(db), da.Config{Enabled: true, ConsentVersions: []string{"synthetic-test-v1"}}, n)
}

func TestFRC12DurableSubmitIdempotencyAndConcurrency(t *testing.T) {
	db := frc12Database(t)
	n := &frc12Notifier{ready: true}
	s := frc12Service(db, n)
	ctx := context.Background()
	now := time.Now().UTC()
	cmd := frc12Command()
	require.True(t, s.Ready())
	first, err := s.Submit(ctx, cmd, now)
	require.NoError(t, err)
	require.False(t, first.Replayed)
	require.NotEmpty(t, first.RequestID)
	cmd.Name = "Test Person"
	cmd.Organization = "Test Org"
	cmd.Email = "person@example.test"
	replay, err := s.Submit(ctx, cmd, now.Add(time.Minute))
	require.NoError(t, err)
	require.True(t, replay.Replayed)
	require.Equal(t, first.RequestID, replay.RequestID)
	changed := cmd
	changed.Comment = "different"
	_, err = s.Submit(ctx, changed, now)
	require.ErrorIs(t, err, dd.ErrIdempotencyConflict)
	rows, total, err := s.ListForOperator(ctx, 1, 0)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, rows, 1)
	require.Equal(t, "pending", rows[0].Status())
	require.Empty(t, n.events)
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	ids := make(chan string, 8)
	concurrent := frc12Command()
	concurrent.Email = "concurrent@example.test"
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := s.Submit(ctx, concurrent, now)
			ids <- result.RequestID
			errs <- err
		}()
	}
	wg.Wait()
	close(ids)
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
	unique := map[string]bool{}
	for id := range ids {
		unique[id] = true
	}
	require.Len(t, unique, 1)
	_, total, err = s.ListForOperator(ctx, 20, 0)
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	afterExpiry, err := s.Submit(ctx, cmd, now.Add(24*time.Hour))
	require.NoError(t, err)
	require.NotEqual(t, first.RequestID, afterExpiry.RequestID)
}

func TestFRC12RollingRatesCommitFailureAndDisabled(t *testing.T) {
	db := frc12Database(t)
	s := frc12Service(db, &frc12Notifier{ready: true})
	ctx := context.Background()
	now := time.Now().UTC()
	for i := 0; i < 5; i++ {
		require.NoError(t, s.AllowAttempt(ctx, "192.0.2.1", now.Add(time.Duration(i)*time.Second)))
	}
	err := s.AllowAttempt(ctx, "192.0.2.1", now.Add(5*time.Second))
	var limited *dd.RateLimitError
	require.ErrorAs(t, err, &limited)
	require.Equal(t, 595*time.Second, limited.RetryAfter)
	require.NoError(t, s.AllowAttempt(ctx, "192.0.2.1", now.Add(10*time.Minute)))
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); cmd := frc12Command(); _, err := s.Submit(ctx, cmd, now); errs <- err }()
	}
	wg.Wait()
	close(errs)
	accepted, rejected := 0, 0
	for err := range errs {
		if err == nil {
			accepted++
		} else {
			require.ErrorAs(t, err, &limited)
			rejected++
		}
	}
	require.Equal(t, 3, accepted)
	require.Equal(t, 5, rejected)
	// A real DB constraint fails after the lead insert; no orphan lead/receipt can commit.
	require.NoError(t, db.Exec("ALTER TABLE demo_request_outbox ADD CONSTRAINT synthetic_fail CHECK (attempts < 0) NOT VALID").Error)
	cmd := frc12Command()
	cmd.Email = "rollback@example.test"
	_, err = s.Submit(ctx, cmd, now)
	require.ErrorIs(t, err, dd.ErrUnavailable)
	var count int64
	require.NoError(t, db.Model(&di.RequestModel{}).Where("email = ?", cmd.Email).Count(&count).Error)
	require.Zero(t, count)
	require.NoError(t, db.Exec("ALTER TABLE demo_request_outbox DROP CONSTRAINT synthetic_fail").Error)
	_, err = s.Submit(ctx, cmd, now)
	require.NoError(t, err)
	for _, cfg := range []da.Config{{}, {Enabled: true}, {Enabled: true, ConsentVersions: []string{""}}} {
		disabled := da.NewService(di.NewRepository(db), cfg, &frc12Notifier{ready: true})
		require.False(t, disabled.Ready())
		_, err := disabled.Submit(ctx, frc12Command(), now)
		require.ErrorIs(t, err, dd.ErrUnavailable)
	}
	disabled := frc12Service(db, &frc12Notifier{})
	require.False(t, disabled.Ready())
	require.ErrorIs(t, disabled.AllowAttempt(ctx, "192.0.2.9", now), dd.ErrUnavailable)
}

func TestFRC12OutboxRetryScheduleStableEventAndExpiry(t *testing.T) {
	db := frc12Database(t)
	n := &frc12Notifier{ready: true, fail: true}
	s := frc12Service(db, n)
	ctx := context.Background()
	now := time.Now().UTC()
	_, err := s.Submit(ctx, frc12Command(), now)
	require.NoError(t, err)
	for i, offset := range []time.Duration{0, time.Minute, 6 * time.Minute, 36 * time.Minute} {
		worked, err := s.DeliverNext(ctx, now.Add(offset))
		require.NoError(t, err)
		require.True(t, worked)
		rows, _, err := s.ListForOperator(ctx, 20, 0)
		require.NoError(t, err)
		require.Equal(t, i+1, rows[0].Attempts())
		require.Equal(t, "pending", rows[0].Status())
		require.NotContains(t, rows[0].LastError(), "secret")
		worked, err = s.DeliverNext(ctx, now.Add(offset+time.Second))
		require.NoError(t, err)
		require.False(t, worked)
	}
	require.Len(t, n.events, 4)
	for _, id := range n.events {
		require.Equal(t, n.events[0], id)
	}
	worked, err := s.DeliverNext(ctx, now.Add(24*time.Hour))
	require.NoError(t, err)
	require.True(t, worked)
	require.Len(t, n.events, 4, "expired notification must not be sent")
	rows, _, err := s.ListForOperator(ctx, 20, 0)
	require.NoError(t, err)
	require.Equal(t, "failed", rows[0].Status())
	n.fail = false
	cmd := frc12Command()
	cmd.Email = fmt.Sprintf("next-%s@example.test", uuid.NewString())
	_, err = s.Submit(ctx, cmd, now.Add(24*time.Hour))
	require.NoError(t, err)
	worked, err = s.DeliverNext(ctx, now.Add(24*time.Hour))
	require.NoError(t, err)
	require.True(t, worked)
	rows, _, err = s.ListForOperator(ctx, 20, 0)
	require.NoError(t, err)
	require.Equal(t, "sent", rows[0].Status())
}

func frc12HTTP(c *du.Controller, ip, origin, key, body, forwarded string) *fasthttp.RequestCtx {
	var request fasthttp.Request
	request.SetRequestURI("https://backend.example.test/api/v1/demo-requests")
	request.Header.SetMethod("POST")
	request.Header.SetContentType("application/json")
	request.Header.Set("Origin", origin)
	request.Header.Set("Idempotency-Key", key)
	if forwarded != "" {
		request.Header.Set("X-Forwarded-For", forwarded)
	}
	request.SetBodyString(body)
	ctx := &fasthttp.RequestCtx{}
	ctx.Init(&request, &net.TCPAddr{IP: net.ParseIP(ip), Port: 10000}, nil)
	c.Submit(ctx)
	return ctx
}
func TestFRC12HTTPBoundaryAndRegistry(t *testing.T) {
	db := frc12Database(t)
	s := frc12Service(db, &frc12Notifier{ready: true})
	c := du.NewController(s, du.TransportConfig{AllowedOrigins: []string{"https://workspace.example.test"}, TrustedProxyCIDRs: []string{"127.0.0.0/8"}})
	valid := `{"name":"Test Person","organization":"Test Org","email":"person@example.test","comment":"<b>hello</b>","consent":true,"consentVersion":"synthetic-test-v1"}`
	key := uuid.NewString()
	ctx := frc12HTTP(c, "127.0.0.1", "https://workspace.example.test", key, valid, "192.0.2.1")
	require.Equal(t, 201, ctx.Response.StatusCode(), string(ctx.Response.Body()))
	require.NotContains(t, string(ctx.Response.Body()), "person@example.test")
	var accepted du.DemoRequestAccepted
	require.NoError(t, json.Unmarshal(ctx.Response.Body(), &accepted))
	require.Equal(t, "received", accepted.Status)
	require.NotEmpty(t, accepted.RequestID)
	ctx = frc12HTTP(c, "127.0.0.1", "https://workspace.example.test", key, valid, "192.0.2.1")
	require.Equal(t, 200, ctx.Response.StatusCode())
	require.Contains(t, string(ctx.Response.Body()), accepted.RequestID)
	for i, tc := range []struct {
		body, origin, xff string
		status            int
	}{
		{valid, "https://attacker.example.test", "192.0.2.2", 403},
		{valid, "", "192.0.2.2", 403},
		{valid, "https://workspace.example.test", "", 400},
		{valid, "https://workspace.example.test", "invalid", 400},
		{strings.Repeat("x", 16385), "https://workspace.example.test", "192.0.2.2", 413},
		{"{", "https://workspace.example.test", "192.0.2.2", 400},
		{valid + " {}", "https://workspace.example.test", "192.0.2.2", 400},
		{strings.Replace(valid, `"consent":true`, `"consent":false`, 1), "https://workspace.example.test", "192.0.2.2", 400},
		{strings.Replace(valid, `"name":"Test Person"`, `"name":"Test Person","notificationTo":"victim@example.test"`, 1), "https://workspace.example.test", "192.0.2.2", 400},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			ctx := frc12HTTP(c, "127.0.0.1", tc.origin, uuid.NewString(), tc.body, strings.Replace(tc.xff, "192.0.2.2", fmt.Sprintf("192.0.2.%d", i+10), 1))
			require.Equal(t, tc.status, ctx.Response.StatusCode(), string(ctx.Response.Body()))
			require.NotEmpty(t, ctx.Response.Header.Peek("X-Request-ID"))
		})
	}
	// A direct untrusted peer cannot rotate the IP limit using a forged proxy header.
	for i := 0; i < 5; i++ {
		ctx = frc12HTTP(c, "192.0.2.200", "https://workspace.example.test", uuid.NewString(), "{", fmt.Sprintf("198.51.100.%d", i))
		require.Equal(t, 400, ctx.Response.StatusCode())
	}
	ctx = frc12HTTP(c, "192.0.2.200", "https://workspace.example.test", uuid.NewString(), valid, "198.51.100.99")
	require.Equal(t, 429, ctx.Response.StatusCode())
	require.NotEmpty(t, ctx.Response.Header.Peek("Retry-After"))
	var output bytes.Buffer
	require.NoError(t, du.WriteRegistry(context.Background(), s, &output, 1, 0))
	var page du.RegistryPage
	require.NoError(t, json.Unmarshal(output.Bytes(), &page))
	require.EqualValues(t, 1, page.TotalItems)
	require.Len(t, page.Items, 1)
	require.Equal(t, "person@example.test", page.Items[0].Email)
	closed := du.NewController(s, du.TransportConfig{})
	ctx = frc12HTTP(closed, "192.0.2.1", "https://workspace.example.test", uuid.NewString(), valid, "")
	require.Equal(t, 503, ctx.Response.StatusCode())
}
