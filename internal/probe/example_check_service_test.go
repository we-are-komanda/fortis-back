package probe

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCalculate(t *testing.T) {
	check := NewEmptySuccessCheckService()
	errCheck := NewErrorExampleCheckService()

	service := NewCompositeCheckService(check, errCheck)
	errs := service.Check()
	assert.Len(t, errs, 2)
}
