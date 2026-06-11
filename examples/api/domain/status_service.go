package domain

type StatusService struct {
	repository StatusRepositoryInterface
}

func NewStatusService(
	repository StatusRepositoryInterface,
) *StatusService {
	return &StatusService{
		repository: repository,
	}
}

func (service *StatusService) Get() (platform *Platform, err error) {
	return service.repository.Get()
}
