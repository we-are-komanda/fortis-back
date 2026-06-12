// Package classification fortis-backend.
//
// Documentation for fortis-backend.
//
//	Schemes: http
//	BasePath: /api/v1
//	Version: 1.0.0
//
// swagger:meta
package main

import (
	"fmt"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"go.uber.org/dig"
	"github.com/fortis/backend/internal/config"
	"github.com/fortis/backend/internal/metrics"
	"github.com/fortis/backend/internal/middleware"
	"github.com/fortis/backend/internal/probe"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

var cnf *config.Config

type Application struct {
	container *dig.Container
}

func NewApplication(
	container *dig.Container,
) *Application {
	return &Application{
		container: container,
	}
}

func (app *Application) registerTechUrls(r *router.Router) error {
	err := app.container.Invoke(
		func(
			metrics *metrics.Controller,
			probe *probe.Controller,
		) {
			r.GET("/_/metrics", metrics.GetMetrics)
			r.GET("/_/liveness", probe.LivenessProbe)
			r.GET("/_/readiness", probe.ReadinessProbe)
			r.GET("/_/startup", probe.StartupProbe)
		})

	return err
}

func (app *Application) registerCoreDependencies() {
	err := app.container.Provide(func() config.Cors { return cnf.Cors })
	processError(err)
	err = app.container.Provide(func() config.Access { return cnf.Access })
	processError(err)
	err = app.container.Provide(func() config.Postgres { return cnf.Postgres })
	processError(err)

	err = app.container.Provide(middleware.NewCors)
	processError(err)
	err = app.container.Provide(middleware.NewAccess)
	processError(err)
	err = app.container.Provide(middleware.NewHttpResponse)
	processError(err)

	err = app.container.Provide(func() *metrics.PrometheusService {
		hostName, _ := os.Hostname()
		return metrics.NewPrometheusService("app", hostName)
	})
	processError(err)
	err = app.container.Provide(middleware.NewPrometheus)
	processError(err)
	err = app.container.Provide(metrics.NewController)
	processError(err)

	err = app.container.Provide(middleware.NewSwagger)
	processError(err)
}

func (app *Application) registerMiddleware(next fasthttp.RequestHandler) (handler fasthttp.RequestHandler, err error) {
	handler = next
	err = app.container.Invoke(
		func(
			cors *middleware.Cors,
			swagger *middleware.Swagger,
			prometheus *middleware.Prometheus,
			httpResponse *middleware.HttpResponse,
			access *middleware.Access,
		) {
			handler = cors.Process(handler)
			handler = swagger.Process(handler)
			handler = prometheus.Process(handler)
			handler = httpResponse.Process(handler)
		})

	if err != nil {
		return
	}

	return handler, nil
}

func (app *Application) getWepApp() (handler fasthttp.RequestHandler, err error) {
	r := router.New()

	err = app.registerTechUrls(r)
	if err != nil {
		return
	}
	err = app.registerHandlers(r)
	if err != nil {
		return
	}

	handler, err = app.registerMiddleware(r.Handler)
	if err != nil {
		return
	}
	return handler, err
}

func (app *Application) Run() {
	var err error

	app.registerCoreDependencies()
	app.provideDependencies()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	slog.Info(fmt.Sprintf("work_mode=%s", cnf.WorkMode))

	switch cnf.WorkMode {
	case "webapp":
		var webApp fasthttp.RequestHandler

		webApp, err = app.getWepApp()
		if err != nil {
			processError(err)
		}

		server := &fasthttp.Server{
			Handler:         webApp,
			ReadBufferSize:  10485760,
			WriteBufferSize: 10485760,
		}

		go func(server *fasthttp.Server) {
			<-shutdown
			slog.Info("signal for termination received")
			GracefulShutdown(server)
		}(server)

		slog.Info("webserver=started")
		err = server.ListenAndServe("0.0.0.0:8090")
		if err == nil {
			processError(err)
		}
		slog.Info("webserver=stopped")
	default:
		slog.Error(fmt.Sprintf("unknown job type '%s'", cnf.WorkMode))
		os.Exit(1)
	}
}

func GracefulShutdown(server *fasthttp.Server) {
	if err := server.Shutdown(); err != nil {
		slog.Info(fmt.Sprintf("uncatchable error: %s", err.Error()))
	}
}

func processError(err error) {
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func main() {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "local"
	}

	cnf = config.NewConfig(env)
	err := cnf.ReadConfig("./config", os.ReadFile)
	processError(err)

	application := NewApplication(dig.New())
	application.Run()
}
