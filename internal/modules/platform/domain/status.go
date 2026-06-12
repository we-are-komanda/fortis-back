package domain

type Platform struct {
	status string
}

func NewPlatform(status string) *Platform {
	return &Platform{status: status}
}

func (p *Platform) Status() string {
	return p.status
}

type StatusRepositoryInterface interface {
	Get() (platform *Platform, err error)
}
