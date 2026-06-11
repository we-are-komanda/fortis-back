package domain

type Platform struct {
	Status string `json:"status"`
}

type StatusRepositoryInterface interface {
	Get() (platform *Platform, err error)
}
