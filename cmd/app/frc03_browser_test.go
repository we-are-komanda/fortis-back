package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fortis/backend/internal/config"
	"github.com/fortis/backend/internal/db"
	ua "github.com/fortis/backend/internal/modules/user/application"
	ui "github.com/fortis/backend/internal/modules/user/infrastructure"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/dig"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Opt-in, test-only server. A dedicated database lets the real operator binary
// use the normal runtime config; no schema/search_path injection is necessary.
func TestFRC03BrowserServer(t *testing.T) {
	if os.Getenv("FRC03_BROWSER_SERVER") != "true" {
		t.Skip("FRC03 browser fixture disabled")
	}
	// Intentionally pinned to the disposable local cluster, never the developer DB.
	admin, err := gorm.Open(postgres.Open("host=127.0.0.1 port=55434 user=fortis_test dbname=postgres sslmode=disable"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	adminPool, err := admin.DB()
	require.NoError(t, err)
	databaseName := "frc03_browser_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	require.NoError(t, admin.Exec("CREATE DATABASE "+databaseName).Error)
	t.Cleanup(func() {
		// The runtime migration driver can retain a leased connection. This name
		// is generated above for this test alone, so only its sessions are closed.
		require.NoError(t, admin.Exec("DROP DATABASE "+databaseName+" WITH (FORCE)").Error)
		require.NoError(t, adminPool.Close())
		t.Log("removed disposable database", databaseName)
	})
	t.Chdir("../..")
	previousConfig := cnf
	cnf = &config.Config{Environment: "test", Postgres: config.Postgres{Host: "127.0.0.1", Port: 55434, User: "fortis_test", Password: "fixture-password-only", DbName: databaseName}, Auth: config.Auth{JWTSecret: "frc03-synthetic-only-secret", JWTExpiry: 24}}
	t.Cleanup(func() { cnf = previousConfig })
	app := NewApplication(dig.New())
	app.registerCoreDependencies()
	app.provideDependencies()
	var database *db.DataBase
	require.NoError(t, app.container.Invoke(func(value *db.DataBase) { database = value }))
	database.GormORM.DB = database.GormORM.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})
	pool, err := database.GormORM.DB.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pool.Close()) })
	users := ua.NewUserService(ui.NewUserRepository(database.GormORM), cnf.Auth.JWTSecret, cnf.Auth.JWTExpiry)
	fixture := map[string]any{"database": map[string]any{"host": "127.0.0.1", "port": 55434, "user": "fortis_test", "dbname": databaseName}}
	for _, name := range []string{"a", "b"} {
		email := "frc03-" + name + "@example.test"
		user, _, err := users.Register(context.Background(), email, "fixture-password-only", "FRC03 "+name)
		require.NoError(t, err)
		fixture[name] = map[string]string{"userId": user.ID(), "email": email}
	}
	for _, table := range []string{"enterprises", "user_enterprises", "defense_projects"} {
		var count int64
		require.NoError(t, database.GormORM.Table(table).Count(&count).Error)
		require.Zero(t, count, table)
	}
	envFile := os.Getenv("FRC03_OPERATOR_ENV_FILE")
	if envFile != "" {
		encoded, err := json.Marshal(map[string]string{"APP_DB_HOST": "127.0.0.1", "APP_DB_PORT": "55434", "APP_DB_USER": "fortis_test", "APP_DB_PASSWORD": "fixture-password-only", "APP_DB_DBNAME": databaseName})
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(envFile, encoded, 0600))
	}
	handler, err := app.getWepApp()
	require.NoError(t, err)
	listener, err := net.Listen("tcp", "127.0.0.1:8092")
	require.NoError(t, err)
	stopped, once := make(chan struct{}), sync.Once{}
	server := &fasthttp.Server{MaxRequestBodySize: 11 * 1024 * 1024, Handler: func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/__fixture":
			ctx.SetContentType("application/json")
			ctx.SetBodyString(js(fixture))
		case "/__stop":
			if !ctx.IsPost() {
				ctx.SetStatusCode(405)
				return
			}
			once.Do(func() { close(stopped) })
		default:
			handler(ctx)
		}
	}}
	t.Cleanup(func() { require.NoError(t, server.Shutdown()); _ = listener.Close() })
	go func() { _ = server.Serve(listener) }()
	t.Log(fmt.Sprintf("FRC03 browser fixture ready on 127.0.0.1:8092, database=%s, initial enterprises/memberships/projects=0", databaseName))
	select {
	case <-stopped:
	case <-time.After(2 * time.Hour):
		t.Fatal("FRC03 browser fixture timed out")
	}
}
