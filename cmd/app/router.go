package main

import (
	"github.com/fasthttp/router"
	defenseUi "github.com/fortis/backend/internal/modules/defense_project/ui"
	"github.com/fortis/backend/internal/modules/platform/ui"
)

//go:cover off
func (app *Application) registerHandlers(r *router.Router) error {
	err := app.container.Invoke(
		func(
			exampleController *ui.ExampleController,
			defenseProjectController *defenseUi.DefenseProjectController,
		) {
			r.GET("/api/v1/example", exampleController.Get)
			r.POST("/api/v1/projects/import", defenseProjectController.Import)
			r.GET("/api/v1/projects/export", defenseProjectController.Export)
		})

	return err
}
