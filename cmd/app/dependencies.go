package main

import (
	"context"
	"github.com/fortis/backend/internal/config"
	"github.com/fortis/backend/internal/db"
	"github.com/fortis/backend/internal/middleware"
	budgetApp "github.com/fortis/backend/internal/modules/budget/application"
	budgetDomain "github.com/fortis/backend/internal/modules/budget/domain"
	budgetInfra "github.com/fortis/backend/internal/modules/budget/infrastructure"
	budgetUi "github.com/fortis/backend/internal/modules/budget/ui"
	defenseAssetApp "github.com/fortis/backend/internal/modules/defense_asset/application"
	defenseAssetDomain "github.com/fortis/backend/internal/modules/defense_asset/domain"
	defenseAssetInfra "github.com/fortis/backend/internal/modules/defense_asset/infrastructure"
	defenseAssetUi "github.com/fortis/backend/internal/modules/defense_asset/ui"
	defenseApp "github.com/fortis/backend/internal/modules/defense_project/application"
	defenseInfra "github.com/fortis/backend/internal/modules/defense_project/infrastructure"
	defenseUi "github.com/fortis/backend/internal/modules/defense_project/ui"
	demoApp "github.com/fortis/backend/internal/modules/demo_request/application"
	demoDomain "github.com/fortis/backend/internal/modules/demo_request/domain"
	demoInfra "github.com/fortis/backend/internal/modules/demo_request/infrastructure"
	demoUi "github.com/fortis/backend/internal/modules/demo_request/ui"
	enterpriseApp "github.com/fortis/backend/internal/modules/enterprise/application"
	enterpriseInfra "github.com/fortis/backend/internal/modules/enterprise/infrastructure"
	enterpriseUi "github.com/fortis/backend/internal/modules/enterprise/ui"
	platformApp "github.com/fortis/backend/internal/modules/platform/application"
	"github.com/fortis/backend/internal/modules/platform/infrastructure"
	platformUi "github.com/fortis/backend/internal/modules/platform/ui"
	reportApp "github.com/fortis/backend/internal/modules/report/application"
	reportUi "github.com/fortis/backend/internal/modules/report/ui"
	userApp "github.com/fortis/backend/internal/modules/user/application"
	userDomain "github.com/fortis/backend/internal/modules/user/domain"
	userInfra "github.com/fortis/backend/internal/modules/user/infrastructure"
	userUi "github.com/fortis/backend/internal/modules/user/ui"
	"github.com/fortis/backend/internal/probe"
	"github.com/fortis/backend/internal/rdbms"
	"go.uber.org/dig"
)

var publicAuthPaths = []string{
	"^/_/[a-z]+$", "^/api/v1/auth/register$", "^/api/v1/auth/login$",
	"^/api/v1/token_validate$", "^/api/v1/demo-requests$",
}

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

	err = app.container.Provide(func(database *db.DataBase) *probe.Controller {
		databaseCheck := probe.NewDatabaseCheckService(func(ctx context.Context) error {
			pool, err := database.GormORM.DB.DB()
			if err != nil {
				return err
			}
			return pool.PingContext(ctx)
		})
		return probe.NewProbeController(
			*probe.NewCompositeCheckService(),
			*probe.NewCompositeCheckService(),
			*probe.NewCompositeCheckService(databaseCheck),
		)
	})
	processError(err)

	// DefenseProject module
	err = app.container.Provide(defenseUi.NewDefenseProjectController)
	processError(err)
	err = app.container.Provide(defenseApp.NewDefenseProjectService)
	processError(err)
	err = app.container.Provide(func(service *defenseApp.DefenseProjectService) defenseUi.DefenseProjectServiceInterface {
		return service
	})
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
	err = app.container.Provide(enterpriseApp.NewEnterpriseService)
	processError(err)
	err = app.container.Provide(func(service *enterpriseApp.EnterpriseService) enterpriseUi.EnterpriseServiceInterface { return service })
	processError(err)
	err = app.container.Provide(enterpriseInfra.NewEnterpriseRepository)
	processError(err)

	err = app.container.Provide(func(service *enterpriseApp.EnterpriseService) enterpriseApp.AccessChecker { return service })
	processError(err)

	// DefenseAsset module
	err = app.container.Provide(defenseAssetUi.NewDefenseAssetController)
	processError(err)
	err = app.container.Provide(func(repo defenseAssetDomain.DefenseAssetRepositoryInterface, access enterpriseApp.AccessChecker, documents defenseAssetDomain.DocumentRepositoryInterface) *defenseAssetApp.DefenseAssetService {
		return defenseAssetApp.NewDefenseAssetService(repo, access, documents)
	})
	processError(err)
	err = app.container.Provide(func(service *defenseAssetApp.DefenseAssetService) defenseAssetUi.DefenseAssetServiceInterface {
		return service
	})
	processError(err)
	err = app.container.Provide(defenseAssetInfra.NewDefenseAssetRepository)
	processError(err)

	// DefenseAsset Document module
	err = app.container.Provide(defenseAssetUi.NewDocumentController)
	processError(err)
	err = app.container.Provide(func() config.Documents { return cnf.Documents })
	processError(err)
	err = app.container.Provide(func(repo defenseAssetDomain.DocumentRepositoryInterface, assets *defenseAssetApp.DefenseAssetService, cfg config.Documents) *defenseAssetApp.DocumentService {
		return defenseAssetApp.NewDocumentService(repo, assets, defenseAssetApp.DocumentPipeline{
			Storage: defenseAssetInfra.NewLocalDocumentStorage(defenseAssetInfra.DocumentStorageConfig{RootDir: cfg.RootDir, Enabled: cfg.Enabled}),
			Scanner: defenseAssetInfra.NewClamAVScanner(defenseAssetInfra.ClamAVConfig{Executable: cfg.ScannerExecutable}),
		})
	}, dig.As(new(defenseAssetUi.DocumentServiceInterface)))
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
		dig.As(new(budgetDomain.BudgetConfigRepositoryInterface), new(budgetDomain.CostProjectionRepositoryInterface)))
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
	err = app.container.Provide(func(repo userDomain.UserRepositoryInterface, authCfg config.Auth) *userApp.UserService {
		return userApp.NewUserService(repo, authCfg.JWTSecret, authCfg.JWTExpiry)
	}, dig.As(new(userUi.UserServiceInterface)))
	processError(err)

	// Demo requests: durable intake and outbox share the existing database.
	err = app.container.Provide(demoInfra.NewRepository)
	processError(err)
	err = app.container.Provide(func(repo demoDomain.Repository) *demoApp.Service {
		cfg := cnf.DemoRequests
		notifier := demoInfra.NewSMTPNotifier(demoInfra.SMTPConfig{Address: cfg.SMTPAddress, Username: cfg.SMTPUsername, Password: cfg.SMTPPassword, From: cfg.SMTPFrom, To: cfg.SMTPTo})
		return demoApp.NewService(repo, demoApp.Config{Enabled: cfg.Enabled, ConsentVersions: cfg.ConsentVersions}, notifier)
	})
	processError(err)
	err = app.container.Provide(func(service *demoApp.Service) *demoUi.Controller {
		return demoUi.NewController(service, demoUi.TransportConfig{AllowedOrigins: cnf.DemoRequests.AllowedOrigins, TrustedProxyCIDRs: cnf.DemoRequests.TrustedProxyCIDRs})
	})
	processError(err)

	// Auth middleware
	err = app.container.Provide(func(cfg config.Auth, users userUi.UserServiceInterface) *middleware.AuthRequired {
		return middleware.NewAuthRequired(cfg.JWTSecret, publicAuthPaths, users)
	})
	processError(err)
}
