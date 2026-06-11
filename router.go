package main

import (
	"github.com/fasthttp/router"
	"github.com/fortis/backend/examples/api/ui"
	uiApp "github.com/fortis/backend/examples/web_app/ui"
)

//go:cover off
func (app *Application) registerHandlers(r *router.Router) error {
	err := app.container.Invoke(
		func(
			exampleController *ui.ExampleController,
			webController *uiApp.WebController,
		) {
			// API app
			r.GET("/api/v1/example", exampleController.Get)

			// Web UI app
			r.GET("/", webController.Main)
			r.ServeFiles("/static/{filepath:*}", "html/static")
		})

	return err
}
