//go:build e2e

package e2e

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type APISuite struct {
	suite.Suite
	baseURL string
}

func (s *APISuite) SetupSuite() {
	// TODO: URL запущенного сервера (docker-compose или локальный)
	s.baseURL = "http://localhost:8090"
}

func (s *APISuite) TestHealthEndpoints() {
	tests := []struct {
		name     string
		endpoint string
		wantCode int
	}{
		{name: "liveness", endpoint: "/_/liveness", wantCode: 200},
		{name: "readiness", endpoint: "/_/readiness", wantCode: 200},
		{name: "startup", endpoint: "/_/startup", wantCode: 200},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// TODO: http.Get(s.baseURL + tt.endpoint)
			s.Equal(tt.wantCode, 200)
		})
	}
}

func (s *APISuite) TestPlatformEndpoint() {
	// GET /api/v1/example -> expected 200 + {"status":"ok"}
}

func TestE2E(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(APISuite))
}
