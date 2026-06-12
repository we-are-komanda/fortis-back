package ui

import (
	"time"

	"github.com/fortis/backend/internal/modules/enterprise/domain"
)

// enterpriseToResponse преобразует доменный Enterprise в DTO ответа.
func enterpriseToResponse(e *domain.Enterprise) EnterpriseResponse {
	return EnterpriseResponse{
		ID:        e.ID(),
		Name:      e.Name(),
		Address:   e.Address(),
		Status:    string(e.Status()),
		Latitude:  e.Latitude(),
		Longitude: e.Longitude(),
		CreatedAt: e.CreatedAt().UTC().Format(time.RFC3339),
		UpdatedAt: e.UpdatedAt().UTC().Format(time.RFC3339),
	}
}
