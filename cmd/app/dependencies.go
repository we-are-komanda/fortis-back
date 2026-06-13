package main

import (
	"github.com/fortis/backend/internal/db"
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
	"github.com/fortis/backend/internal/probe"
	"github.com/fortis/backend/internal/rdbms"
)

//go:cover off
func (app *Application) provideDependencies() {
	err := app.container.Provide(platformUi.NewExampleController)
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
	err = app.container.Provide(defenseApp.NewDefenseProjectService)
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
	err = app.container.Provide(enterpriseInfra.NewEnterpriseRepository)
	processError(err)

	// DefenseAsset module
	err = app.container.Provide(defenseAssetUi.NewDefenseAssetController)
	processError(err)
	err = app.container.Provide(defenseAssetApp.NewDefenseAssetService)
	processError(err)
	err = app.container.Provide(defenseAssetInfra.NewDefenseAssetRepository)
	processError(err)
}
