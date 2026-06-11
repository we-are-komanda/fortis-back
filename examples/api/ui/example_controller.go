package ui

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
	"log/slog"

	"github.com/fortis/backend/examples/api/domain"
)

type ExampleController struct {
	service *domain.StatusService
}

func NewExampleController(
	service *domain.StatusService,
) *ExampleController {
	return &ExampleController{
		service: service,
	}
}

type SuccessGetResponseBody struct {
	// Статус ответа
	// Example: ok
	Status string
}

// SuccessGetResponse Успешный ответ
// swagger:response SuccessGetResponse
type SuccessGetResponse struct {
	// In: body
	Body SuccessGetResponseBody
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

	resultJson, err := json.Marshal(result)
	if err != nil {
		slog.Error(err.Error())
	}

	ctx.SetBody(resultJson)
}
