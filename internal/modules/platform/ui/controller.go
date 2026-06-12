package ui

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
	"log/slog"

	"github.com/fortis/backend/internal/modules/platform/application"
)

type ExampleController struct {
	service *application.StatusService
}

func NewExampleController(
	service *application.StatusService,
) *ExampleController {
	return &ExampleController{
		service: service,
	}
}

// swagger:route GET /api/v1/example api Get
// Получение статуса платформы
//
// _
// Produces:
//   - application/json
//
// Responses:
//
//	200: SuccessGetResponse
//	500: description:Internal Server Error
func (ui *ExampleController) Get(ctx *fasthttp.RequestCtx) {
	result, err := ui.service.Get()
	if err != nil {
		slog.Error(err.Error())
	}

	dto := mapPlatformToDTO(result)
	resultJson, err := json.Marshal(dto)
	if err != nil {
		slog.Error(err.Error())
	}

	ctx.SetBody(resultJson)
}
