package main

import (
	"context"
	du "github.com/fortis/backend/internal/modules/demo_request/ui"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"net"
	"os"
	"testing"
	"time"
)

// Opt-in synthetic browser fixture; this server is never part of the app binary.
func TestFRC12BrowserServer(t *testing.T) {
	if os.Getenv("FRC12_BROWSER_SERVER") != "true" {
		t.Skip("browser fixture disabled")
	}
	db := frc12Database(t)
	service := frc12Service(db, &frc12Notifier{ready: true, fail: true})
	controller := du.NewController(service, du.TransportConfig{AllowedOrigins: []string{"http://127.0.0.1:3111"}, TrustedProxyCIDRs: []string{"127.0.0.0/8"}})
	listener, err := net.Listen("tcp", "127.0.0.1:8093")
	require.NoError(t, err)
	stopped := make(chan struct{})
	server := &fasthttp.Server{Handler: func(c *fasthttp.RequestCtx) {
		c.SetContentType("application/json")
		switch string(c.Path()) {
		case "/api/v1/demo-requests":
			if !c.IsPost() {
				c.SetStatusCode(405)
				return
			}
			controller.Submit(c)
		case "/__fixture":
			rows, total, err := service.ListForOperator(context.Background(), 100, 0)
			if err != nil {
				c.SetStatusCode(500)
				return
			}
			items := []map[string]any{}
			for _, row := range rows {
				items = append(items, map[string]any{"requestId": row.Request().ID(), "status": row.Status(), "attempts": row.Attempts()})
			}
			c.SetBodyString(js(map[string]any{"total": total, "items": items}))
		case "/__storage-failure":
			if !c.IsPost() {
				c.SetStatusCode(405)
				return
			}
			sql := "ALTER TABLE demo_request_outbox DROP CONSTRAINT IF EXISTS synthetic_commit_failure"
			if string(c.QueryArgs().Peek("enabled")) == "true" {
				sql = "ALTER TABLE demo_request_outbox ADD CONSTRAINT synthetic_commit_failure CHECK (attempts < 0) NOT VALID"
			}
			if db.Exec(sql).Error != nil {
				c.SetStatusCode(500)
			}
		case "/__deliver":
			if !c.IsPost() {
				c.SetStatusCode(405)
				return
			}
			if _, err := service.DeliverNext(context.Background(), time.Now().UTC()); err != nil {
				c.SetStatusCode(500)
			}
		case "/__stop":
			if !c.IsPost() {
				c.SetStatusCode(405)
				return
			}
			select {
			case <-stopped:
			default:
				close(stopped)
			}
		default:
			c.SetStatusCode(404)
		}
	}}
	t.Cleanup(func() { _ = server.Shutdown(); _ = listener.Close() })
	go func() { _ = server.Serve(listener) }()
	t.Log("FRC12 synthetic browser fixture ready on 127.0.0.1:8093")
	select {
	case <-stopped:
	case <-time.After(30 * time.Minute):
		t.Fatal("browser fixture timed out")
	}
}
