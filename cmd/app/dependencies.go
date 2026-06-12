package main

import (
	"github.com/fortis/backend/internal/modules/platform/application"
	"github.com/fortis/backend/internal/modules/platform/infrastructure"
	platformUi "github.com/fortis/backend/internal/modules/platform/ui"
	"github.com/fortis/backend/internal/probe"
)

//go:cover off
func (app *Application) provideDependencies() {
	err := app.container.Provide(platformUi.NewExampleController)
	processError(err)
	err = app.container.Provide(application.NewStatusService)
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
}
