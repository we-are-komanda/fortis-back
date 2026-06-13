package main

import (
	"github.com/fasthttp/router"
	defenseAssetUi "github.com/fortis/backend/internal/modules/defense_asset/ui"
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
			defenseAssetController *defenseAssetUi.DefenseAssetController,
		) {
			r.GET("/api/v1/example", exampleController.Get)
			r.POST("/api/v1/projects/import", defenseProjectController.Import)
			r.GET("/api/v1/projects/export", defenseProjectController.Export)
			r.POST("/api/v1/projects", defenseProjectController.Create)
			r.GET("/api/v1/projects", defenseProjectController.List)
			r.GET("/api/v1/projects/get", defenseProjectController.Get)
			r.PUT("/api/v1/projects/update", defenseProjectController.Update)
			r.DELETE("/api/v1/projects/delete", defenseProjectController.Delete)
			r.POST("/api/v1/enterprises", enterpriseController.Create)
			r.GET("/api/v1/enterprises", enterpriseController.GetOrList)
			r.PUT("/api/v1/enterprises", enterpriseController.Update)
			r.DELETE("/api/v1/enterprises", enterpriseController.Delete)
			r.GET("/api/v1/assets", defenseAssetController.List)
			r.GET("/api/v1/assets/get", defenseAssetController.Get)
			r.POST("/api/v1/assets", defenseAssetController.Create)
			r.PUT("/api/v1/assets/update", defenseAssetController.Update)
			r.DELETE("/api/v1/assets/delete", defenseAssetController.Delete)
		})

	return err
}
