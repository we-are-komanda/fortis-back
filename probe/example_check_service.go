package probe

import "fmt"

type EmptySuccessCheckService struct {
}

func NewEmptySuccessCheckService() *EmptySuccessCheckService {
	return &EmptySuccessCheckService{}
}

func (service *EmptySuccessCheckService) Check() []error {
	return nil
}

type ErrorExampleCheckService struct {
}

func NewErrorExampleCheckService() *ErrorExampleCheckService {
	return &ErrorExampleCheckService{}
}

func (service *ErrorExampleCheckService) Check() []error {
	var a []error

	a = append(a, fmt.Errorf("vse ploho"))
	a = append(a, fmt.Errorf("nado umirat"))

	return a
}
