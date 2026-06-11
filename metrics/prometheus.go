//go:cover off
package metrics

import "github.com/prometheus/client_golang/prometheus"

type PrometheusService struct {
	RequestsTotal        *prometheus.CounterVec
	RequestsDuration     *prometheus.HistogramVec
	PgRequestsErrorTotal *prometheus.CounterVec
	PgRequestsDuration   *prometheus.HistogramVec
}

func NewPrometheusService(subsystem string, hostName string) *PrometheusService {
	prom := &PrometheusService{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem:   subsystem,
				Name:        "requests_total",
				ConstLabels: map[string]string{"hostName": hostName},
			},
			[]string{"code", "method", "host", "url"},
		),
		RequestsDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Subsystem:   subsystem,
				Name:        "request_duration_seconds",
				Buckets:     []float64{0.0001, 0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 150, 200, 300, 400, 500, 600},
				ConstLabels: map[string]string{"hostName": hostName},
			},
			[]string{"code", "method", "host", "url"},
		),
		PgRequestsDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Subsystem:   subsystem,
				Name:        "pg_request_duration_seconds",
				Buckets:     []float64{0.0001, 0.0005, 0.001, 0.0015, 0.002, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 150, 200, 300, 400, 500, 600},
				ConstLabels: map[string]string{"hostName": hostName},
			},
			[]string{"method"},
		),
		PgRequestsErrorTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem:   subsystem,
				Name:        "pg_requests_error_total",
				ConstLabels: map[string]string{"hostName": hostName},
			},
			[]string{"method"},
		),
	}
	prometheus.MustRegister(
		prom.RequestsTotal,
		prom.RequestsDuration,
		prom.PgRequestsErrorTotal,
		prom.PgRequestsDuration,
	)
	return prom
}
