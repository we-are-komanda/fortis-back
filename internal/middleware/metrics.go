package middleware

import (
	"github.com/valyala/fasthttp"
	"github.com/fortis/backend/internal/metrics"
	"strconv"
	"strings"
	"time"
)

type Prometheus struct {
	metrics *metrics.PrometheusService
}

func NewPrometheus(metrics *metrics.PrometheusService) *Prometheus {
	return &Prometheus{metrics: metrics}
}

//go:cover off
func (m *Prometheus) Process(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(c *fasthttp.RequestCtx) {
		path := strings.ToValidUTF8(string(c.Request.URI().Path()), "?")
		if path == "/_/metrics" {
			next(c)
			return
		}

		status := strconv.Itoa(c.Response.StatusCode())
		method := string(c.Method())
		host := string(c.Request.Host())

		m.metrics.RequestsTotal.WithLabelValues(status, method, host, path).Inc()

		start := time.Now()
		next(c)
		duration := time.Since(start)

		m.metrics.RequestsDuration.WithLabelValues(status, method, host, path).Observe(duration.Seconds())
	}
}
