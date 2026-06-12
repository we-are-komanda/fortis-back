package application

import (
	"github.com/fortis/backend/internal/modules/platform/domain"
)

type StatusService struct {
	repository domain.StatusRepositoryInterface
}

func NewStatusService(
	repository domain.StatusRepositoryInterface,
) *StatusService {
	return &StatusService{
		repository: repository,
	}
}

func (service *StatusService) Get() (platform *domain.Platform, err error) {
	return service.repository.Get()
}
