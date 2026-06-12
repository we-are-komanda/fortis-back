package main

import (
	"github.com/fasthttp/router"
	"github.com/fortis/backend/internal/modules/platform/ui"
)

//go:cover off
func (app *Application) registerHandlers(r *router.Router) error {
	err := app.container.Invoke(
		func(
			exampleController *ui.ExampleController,
		) {
			r.GET("/api/v1/example", exampleController.Get)
		})

	return err
}
