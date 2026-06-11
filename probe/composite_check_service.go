package probe

import (
	"fmt"
	"log/slog"
)

type CheckInterface interface {
	Check() []error
}

type CompositeCheckService struct {
	checkServices []CheckInterface
}

//go:cover off
func NewCompositeCheckService(checkServices ...CheckInterface) *CompositeCheckService {
	return &CompositeCheckService{
		checkServices: checkServices,
	}
}

//go:cover off
func (service *CompositeCheckService) Check() (errs []error) {
	for _, checkService := range service.checkServices {
		err := checkService.Check()

		if err != nil {
			errs = append(errs, err...)
		}
	}

	for _, err := range errs {
		slog.Error(fmt.Sprintf("Health check error: %s", err.Error()))
	}

	return errs
}
