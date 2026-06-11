package probe

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
	"log/slog"
	"os"
)

type Controller struct {
	livenessComposite  CompositeCheckService
	readinessComposite CompositeCheckService
	startupComposite   CompositeCheckService
}

type SuccessProbeResponseBody struct {
	// Названия продукта
	// Example: cabinet
	ProductName string `json:"product_name"`

	// Названия сервиса
	// Example: orders
	ServiceName string `json:"service_name"`

	// Окружение в котором работает приложение
	// Example: prod
	Env string `json:"environment"`

	// Ошибка (при наличии)
	// Example: Нет подключения к базе данных
	Error []string `json:"error"`
}

// SuccessProbeResponse Успешный ответ
// swagger:response SuccessProbeResponse
type SuccessProbeResponse struct {
	// In: body
	Body SuccessProbeResponseBody
}

func NewProbeController(
	startupComposite CompositeCheckService,
	livenessComposite CompositeCheckService,
	readinessComposite CompositeCheckService,
) *Controller {
	return &Controller{
		livenessComposite:  livenessComposite,
		readinessComposite: readinessComposite,
		startupComposite:   startupComposite,
	}
}

// swagger:route GET /_/liveness k8s LivenessProbe
// Liveness Probe
//
// _
// Produces:
//   - application/json
//
// Responses:
//
//	200: SuccessProbeResponse
//	500: description:Internal Server Error
func (controller *Controller) LivenessProbe(context *fasthttp.RequestCtx) {
	probeErr := controller.livenessComposite.Check()
	controller.prepareProbeResponse(probeErr, context)
}

// swagger:route GET /_/readiness k8s ReadinessProbe
// Readiness Probe
//
// _
// Produces:
//   - application/json
//
// Responses:
//
//	200: SuccessProbeResponse
//	500: description:Internal Server Error
func (controller *Controller) ReadinessProbe(context *fasthttp.RequestCtx) {
	probeErr := controller.readinessComposite.Check()
	controller.prepareProbeResponse(probeErr, context)
}

// swagger:route GET /_/startup k8s StartupProbe
// Startup Probe
//
// _
// Produces:
//   - application/json
//
// Responses:
//
//	200: SuccessProbeResponse
//	500: description:Internal Server Error
func (controller *Controller) StartupProbe(context *fasthttp.RequestCtx) {
	probeErr := controller.startupComposite.Check()
	controller.prepareProbeResponse(probeErr, context)
}

//go:cover off
func (controller *Controller) prepareProbeResponse(errs []error, context *fasthttp.RequestCtx) {
	var errStrings []string
	response := SuccessProbeResponseBody{
		ProductName: os.Getenv("PRODUCT_NAME"),
		ServiceName: os.Getenv("SERVICE_NAME"),
		Env:         os.Getenv("ENVIRONMENT"),
		Error:       errStrings,
	}

	if len(errs) > 0 {
		context.SetStatusCode(fasthttp.StatusInternalServerError)

		for _, e := range errs {
			response.Error = append(response.Error, e.Error())
		}
	} else {
		context.SetStatusCode(fasthttp.StatusOK)
	}

	jsonResponse, _ := json.Marshal(response)
	_, err := context.Write(jsonResponse)
	if err != nil {
		slog.Error(err.Error())
	}
	context.SetContentType("application/json; charset=utf-8")
}
