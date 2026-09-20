// Package classification fortis-backend.
//
// Documentation for fortis-backend.
//
//	Schemes: http
//	BasePath: /
//	Version: 1.0.0
//
// swagger:meta
package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/fasthttp/router"
	"github.com/fortis/backend/internal/config"
	"github.com/fortis/backend/internal/metrics"
	"github.com/fortis/backend/internal/middleware"
	demoApp "github.com/fortis/backend/internal/modules/demo_request/application"
	demoUi "github.com/fortis/backend/internal/modules/demo_request/ui"
	"github.com/fortis/backend/internal/probe"
	"github.com/valyala/fasthttp"
	"go.uber.org/dig"
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
	err = app.container.Provide(func() config.Auth { return cnf.Auth })
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
			authRequired *middleware.AuthRequired,
			access *middleware.Access,
		) {
			handler = cors.Process(handler)
			handler = swagger.Process(handler)
			handler = prometheus.Process(handler)
			handler = httpResponse.Process(handler)
			handler = authRequired.Process(handler)
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
	case "demo-request-registry":
		// Local operator command requires the process's OS and database privileges.
		flags := flag.NewFlagSet("demo-request-registry", flag.ExitOnError)
		limit, offset := flags.Int("limit", 100, "maximum rows"), flags.Int("offset", 0, "row offset")
		processError(flags.Parse(os.Args[1:]))
		processError(app.container.Invoke(func(service *demoApp.Service) error {
			return demoUi.WriteRegistry(context.Background(), service, os.Stdout, *limit, *offset)
		}))
	case "webapp":
		var webApp fasthttp.RequestHandler

		webApp, err = app.getWepApp()
		if err != nil {
			processError(err)
		}

		server := &fasthttp.Server{
			Handler:            webApp,
			ReadBufferSize:     10485760,
			MaxRequestBodySize: 11 * 1024 * 1024,
			WriteBufferSize:    10485760,
		}

		workerCtx, stopWorker := context.WithCancel(context.Background())
		defer stopWorker()
		workerDone := make(chan struct{})
		processError(app.container.Invoke(func(service *demoApp.Service) {
			go func() { defer close(workerDone); service.Run(workerCtx) }()
		}))

		go func(server *fasthttp.Server) {
			<-shutdown
			slog.Info("signal for termination received")
			stopWorker()
			GracefulShutdown(server)
		}(server)

		slog.Info("webserver=started")
		err = server.ListenAndServe("0.0.0.0:8090")
		stopWorker()
		<-workerDone
		if err != nil {
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
