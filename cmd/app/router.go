package main

import (
	"github.com/fasthttp/router"
	defenseUi "github.com/fortis/backend/internal/modules/defense_project/ui"
	enterpriseUi "github.com/fortis/backend/internal/modules/enterprise/ui"
	"github.com/fortis/backend/internal/modules/platform/ui"
)

//go:cover off
func (app *Application) registerHandlers(r *router.Router) error {
	err := app.container.Invoke(
		func(
			exampleController *ui.ExampleController,
			defenseProjectController *defenseUi.DefenseProjectController,
			enterpriseController *enterpriseUi.EnterpriseController,
		) {
			r.GET("/api/v1/example", exampleController.Get)
			r.POST("/api/v1/projects/import", defenseProjectController.Import)
			r.GET("/api/v1/projects/export", defenseProjectController.Export)
			r.POST("/api/v1/enterprises", enterpriseController.Create)
			r.GET("/api/v1/enterprises", enterpriseController.GetOrList)
			r.PUT("/api/v1/enterprises", enterpriseController.Update)
			r.DELETE("/api/v1/enterprises", enterpriseController.Delete)
		})

	return err
}
