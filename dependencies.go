package main

import (
	"github.com/fortis/backend/examples/api/domain"
	"github.com/fortis/backend/examples/api/infrastructure"
	api "github.com/fortis/backend/examples/api/ui"
	uiApp "github.com/fortis/backend/examples/web_app/ui"
	"github.com/fortis/backend/probe"
)

//go:cover off
func (app *Application) provideDependencies() {
	// start - API examples
	err := app.container.Provide(api.NewExampleController)
	processError(err)
	err = app.container.Provide(domain.NewStatusService)
	processError(err)
	err = app.container.Provide(infrastructure.NewStatusRepository)
	processError(err)
	// end - API examples

	// start - UP app examples
	err = app.container.Provide(uiApp.NewWebController)
	processError(err)
	// end - UP app examples

	// start - App using DB examples
	// err = app.container.Provide(db.NewDataBase)
	// processError(err)
	// err = app.container.Provide(func(dataBase *db.DataBase) *db.GormORM {
	//	 return dataBase.GormORM
	// })
	// processError(err)
	// end - App using DB examples

	// ВАЖНО - нужно написать и зарегистрировать функциональности которые будут выполнять
	// проверку доступности компонентов сервиса, а так же его готовность.
	// Функциональности с проверками лучше размещать в пакете probe
	err = app.container.Provide(func() *probe.Controller {
		return probe.NewProbeController(
			// startupComposite
			*probe.NewCompositeCheckService(),
			// livenessComposite
			*probe.NewCompositeCheckService(),
			// readinessComposite
			*probe.NewCompositeCheckService(),
		)
	})
	processError(err)
}
