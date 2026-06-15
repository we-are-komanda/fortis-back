package main

import (
	"github.com/fasthttp/router"
	budgetUi "github.com/fortis/backend/internal/modules/budget/ui"
	defenseAssetUi "github.com/fortis/backend/internal/modules/defense_asset/ui"
	defenseUi "github.com/fortis/backend/internal/modules/defense_project/ui"
	enterpriseUi "github.com/fortis/backend/internal/modules/enterprise/ui"
	"github.com/fortis/backend/internal/modules/platform/ui"
	reportUi "github.com/fortis/backend/internal/modules/report/ui"
	userUi "github.com/fortis/backend/internal/modules/user/ui"
)

//go:cover off
func (app *Application) registerHandlers(r *router.Router) error {
	err := app.container.Invoke(
		func(
			exampleController *ui.ExampleController,
			defenseProjectController *defenseUi.DefenseProjectController,
			enterpriseController *enterpriseUi.EnterpriseController,
			defenseAssetController *defenseAssetUi.DefenseAssetController,
			documentController *defenseAssetUi.DocumentController,
			budgetController *budgetUi.BudgetController,
			reportController *reportUi.ReportController,
			userController *userUi.UserController,
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
			r.POST("/api/v1/enterprises/members", enterpriseController.AddMember)
			r.DELETE("/api/v1/enterprises/members", enterpriseController.RemoveMember)
			r.GET("/api/v1/assets", defenseAssetController.List)
			r.GET("/api/v1/assets/get", defenseAssetController.Get)
			r.POST("/api/v1/assets", defenseAssetController.Create)
			r.PUT("/api/v1/assets/update", defenseAssetController.Update)
			r.DELETE("/api/v1/assets/delete", defenseAssetController.Delete)
			r.GET("/api/v1/assets/documents/list", documentController.List)
			r.GET("/api/v1/assets/documents/get", documentController.Get)
			r.GET("/api/v1/assets/documents/download", documentController.Download)
			r.POST("/api/v1/assets/documents", documentController.Create)
			r.DELETE("/api/v1/assets/documents/delete", documentController.Delete)

			// Budget routes
			r.GET("/api/v1/projects/budget", budgetController.GetBudgetConfig)
			r.PUT("/api/v1/projects/budget", budgetController.UpdateBudgetConfig)
			r.GET("/api/v1/projects/cost", budgetController.CalculateCost)
			r.POST("/api/v1/projects/budget/check", budgetController.CheckBudget)
			r.GET("/api/v1/projects/compare", budgetController.Compare)

			// Report routes
			r.GET("/api/v1/projects/report", reportController.Get)

			// Auth routes
			r.POST("/api/v1/auth/register", userController.Register)
			r.POST("/api/v1/auth/login", userController.Login)
			r.GET("/api/v1/auth/me", userController.Me)
			r.GET("/api/v1/token_validate", userController.ValidateToken)
		})

	return err
}
