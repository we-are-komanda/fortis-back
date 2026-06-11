package infrastructure

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/fortis/backend/examples/api/domain"
	"github.com/fortis/backend/metrics"
	"time"
)

type StatusRepository struct {
	prometheus *metrics.PrometheusService
}

func NewStatusRepository(
	prometheus *metrics.PrometheusService,
) domain.StatusRepositoryInterface {
	return &StatusRepository{
		prometheus: prometheus,
	}
}

func (repository *StatusRepository) Get() (platform *domain.Platform, err error) {
	timer := prometheus.NewTimer(repository.prometheus.PgRequestsDuration.WithLabelValues("/api/v1/example"))
	time.Sleep(time.Second)
	timer.ObserveDuration()
	return &domain.Platform{Status: "ok"}, nil
}
