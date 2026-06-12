package metrics

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type Controller struct {
}

//go:cover off
func NewController() *Controller {
	return &Controller{}
}

// swagger:route GET /_/metrics telemetry getMetrics
// Метрики в формате Prometheus
//
// _
// Responses:
//
//	200: description:Success
//	500: description:Internal Server Error
func (controller *Controller) GetMetrics(context *fasthttp.RequestCtx) {
	handler := fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())
	handler(context)
}
