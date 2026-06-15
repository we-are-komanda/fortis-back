package main

import (
	"github.com/fortis/backend/internal/config"
	"github.com/fortis/backend/internal/db"
	"github.com/fortis/backend/internal/middleware"
	budgetApp "github.com/fortis/backend/internal/modules/budget/application"
	budgetDomain "github.com/fortis/backend/internal/modules/budget/domain"
	budgetInfra "github.com/fortis/backend/internal/modules/budget/infrastructure"
	budgetUi "github.com/fortis/backend/internal/modules/budget/ui"
	defenseAssetApp "github.com/fortis/backend/internal/modules/defense_asset/application"
	defenseAssetInfra "github.com/fortis/backend/internal/modules/defense_asset/infrastructure"
	defenseAssetUi "github.com/fortis/backend/internal/modules/defense_asset/ui"
	defenseApp "github.com/fortis/backend/internal/modules/defense_project/application"
	defenseInfra "github.com/fortis/backend/internal/modules/defense_project/infrastructure"
	defenseUi "github.com/fortis/backend/internal/modules/defense_project/ui"
	enterpriseApp "github.com/fortis/backend/internal/modules/enterprise/application"
	enterpriseInfra "github.com/fortis/backend/internal/modules/enterprise/infrastructure"
	enterpriseUi "github.com/fortis/backend/internal/modules/enterprise/ui"
	platformApp "github.com/fortis/backend/internal/modules/platform/application"
	"github.com/fortis/backend/internal/modules/platform/infrastructure"
	platformUi "github.com/fortis/backend/internal/modules/platform/ui"
	reportApp "github.com/fortis/backend/internal/modules/report/application"
	reportUi "github.com/fortis/backend/internal/modules/report/ui"
	userApp "github.com/fortis/backend/internal/modules/user/application"
	userInfra "github.com/fortis/backend/internal/modules/user/infrastructure"
	userUi "github.com/fortis/backend/internal/modules/user/ui"
	"github.com/fortis/backend/internal/probe"
	"github.com/fortis/backend/internal/rdbms"
	"go.uber.org/dig"
)

//go:cover off
func (app *Application) provideDependencies() {
	// БД — провайдер *db.DataBase (config.Postgres регистрируется в registerCoreDependencies).
	err := app.container.Provide(db.NewDataBase)
	processError(err)

	err = app.container.Provide(platformUi.NewExampleController)
	processError(err)
	err = app.container.Provide(platformApp.NewStatusService)
	processError(err)
	err = app.container.Provide(infrastructure.NewStatusRepository)
	processError(err)

	err = app.container.Provide(func() *probe.Controller {
		return probe.NewProbeController(
			*probe.NewCompositeCheckService(),
			*probe.NewCompositeCheckService(),
			*probe.NewCompositeCheckService(),
		)
	})
	processError(err)

	// DefenseProject module
	err = app.container.Provide(defenseUi.NewDefenseProjectController)
	processError(err)
	err = app.container.Provide(defenseApp.NewDefenseProjectService,
		dig.As(new(defenseUi.DefenseProjectServiceInterface)))
	processError(err)
	err = app.container.Provide(defenseInfra.NewDefenseProjectRepository)
	processError(err)
	err = app.container.Provide(func(database *db.DataBase) rdbms.Executor {
		return database.GormORM
	})
	processError(err)

	// Enterprise module
	err = app.container.Provide(enterpriseUi.NewEnterpriseController)
	processError(err)
	err = app.container.Provide(enterpriseApp.NewEnterpriseService,
		dig.As(new(enterpriseUi.EnterpriseServiceInterface)))
	processError(err)
	err = app.container.Provide(enterpriseInfra.NewEnterpriseRepository)
	processError(err)

	// DefenseAsset module
	err = app.container.Provide(defenseAssetUi.NewDefenseAssetController)
	processError(err)
	err = app.container.Provide(defenseAssetApp.NewDefenseAssetService,
		dig.As(new(defenseAssetUi.DefenseAssetServiceInterface)))
	processError(err)
	err = app.container.Provide(defenseAssetInfra.NewDefenseAssetRepository)
	processError(err)

	// DefenseAsset Document module
	err = app.container.Provide(defenseAssetUi.NewDocumentController)
	processError(err)
	err = app.container.Provide(defenseAssetApp.NewDocumentService,
		dig.As(new(defenseAssetUi.DocumentServiceInterface)))
	processError(err)
	err = app.container.Provide(defenseAssetInfra.NewDocumentRepository)
	processError(err)

	// Budget module
	err = app.container.Provide(budgetUi.NewBudgetController)
	processError(err)
	err = app.container.Provide(budgetApp.NewBudgetService,
		dig.As(new(budgetUi.BudgetServiceInterface), new(reportApp.BudgetServiceInterface)))
	processError(err)
	err = app.container.Provide(budgetInfra.NewBudgetConfigRepository,
		dig.As(new(budgetDomain.BudgetConfigRepositoryInterface)))
	processError(err)

	// Report module
	err = app.container.Provide(reportUi.NewReportController)
	processError(err)
	err = app.container.Provide(reportApp.NewReportService,
		dig.As(new(reportUi.ReportServiceInterface)))
	processError(err)

	// User module
	err = app.container.Provide(userUi.NewUserController)
	processError(err)
	err = app.container.Provide(userInfra.NewUserRepository)
	processError(err)
	err = app.container.Provide(userApp.NewUserService,
		dig.As(new(userUi.UserServiceInterface)))
	processError(err)

	// Auth middleware
	err = app.container.Provide(func(cfg config.Auth) *middleware.AuthRequired {
		return middleware.NewAuthRequired(cfg.JWTSecret, []string{
			"^/_/[a-z]+$",
			"^/api/v1/auth/register$",
			"^/api/v1/auth/login$",
			"^/api/v1/token_validate$",
		})
	})
	processError(err)
}
