//go:build integration

package integration

import (
	"database/sql"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"

	"github.com/fortis/backend/internal/modules/platform/domain"
	"github.com/fortis/backend/internal/modules/platform/infrastructure"
	"github.com/fortis/backend/internal/metrics"
)

type PlatformSuite struct {
	suite.Suite
	db *sql.DB
}

func (s *PlatformSuite) SetupSuite() {
	// TODO: подключение к тестовой БД (testcontainers или внешний хост)
}

func (s *PlatformSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *PlatformSuite) TestStatusRepository_Get() {
	prom := metrics.NewPrometheusService("test", "localhost")
	repo := infrastructure.NewStatusRepository(prom)

	platform, err := repo.Get()
	s.Require().NoError(err)
	s.NotNil(platform)
	s.Equal("ok", platform.Status())
}

func TestPlatformIntegration(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(PlatformSuite))
}
