package application

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/fortis/backend/internal/modules/enterprise/domain"
	"github.com/google/uuid"
)

var (
	ErrInvalidProvisionInput = errors.New("invalid provisioning input: require valid IDs, name, address and explicit valid latitude/longitude")
	ErrProvisionFailed       = errors.New("enterprise provisioning failed")
)

type ProvisionResult struct{ EnterpriseID, UserID string }

type ProvisionService struct {
	repo domain.EnterpriseProvisionRepository
}

func NewProvisionService(repo domain.EnterpriseProvisionRepository) *ProvisionService {
	return &ProvisionService{repo: repo}
}

func (s *ProvisionService) Provision(ctx context.Context, command ProvisionEnterpriseCommand) (ProvisionResult, error) {
	enterpriseID, enterpriseErr := uuid.Parse(command.EnterpriseID)
	userID, userErr := uuid.Parse(command.UserID)
	name, address := strings.TrimSpace(command.Name), strings.TrimSpace(command.Address)
	if enterpriseErr != nil || userErr != nil || enterpriseID == uuid.Nil || userID == uuid.Nil ||
		name == "" || utf8.RuneCountInString(name) > 512 || address == "" ||
		command.Latitude == nil || command.Longitude == nil {
		return ProvisionResult{}, ErrInvalidProvisionInput
	}
	now := time.Now().UTC()
	enterprise, err := domain.NewEnterprise(enterpriseID.String(), name, address, domain.EnterpriseStatusActive, *command.Latitude, *command.Longitude, now, now)
	if err != nil {
		return ProvisionResult{}, ErrInvalidProvisionInput
	}
	if err := s.repo.Provision(ctx, enterprise, userID.String()); err != nil {
		switch {
		case errors.Is(err, domain.ErrProvisionUserNotFound):
			return ProvisionResult{}, domain.ErrProvisionUserNotFound
		case errors.Is(err, domain.ErrProvisionConflict):
			return ProvisionResult{}, domain.ErrProvisionConflict
		default:
			return ProvisionResult{}, ErrProvisionFailed
		}
	}
	return ProvisionResult{EnterpriseID: enterprise.ID(), UserID: userID.String()}, nil
}
