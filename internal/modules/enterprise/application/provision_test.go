package application

import (
	"context"
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/fortis/backend/internal/modules/enterprise/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type provisionRecorder struct {
	enterprise *domain.Enterprise
	userID     string
	err        error
}

func (r *provisionRecorder) Provision(_ context.Context, enterprise *domain.Enterprise, userID string) error {
	r.enterprise, r.userID = enterprise, userID
	return r.err
}

func validProvisionCommand() ProvisionEnterpriseCommand {
	zero := 0.0
	return ProvisionEnterpriseCommand{EnterpriseID: uuid.NewString(), UserID: uuid.NewString(), Name: "  Synthetic facility  ", Address: "  Synthetic address  ", Latitude: &zero, Longitude: &zero}
}

func TestProvisionValidatesBeforePersistence(t *testing.T) {
	for name, change := range map[string]func(*ProvisionEnterpriseCommand){
		"missing enterprise ID": func(c *ProvisionEnterpriseCommand) { c.EnterpriseID = "" },
		"invalid user ID":       func(c *ProvisionEnterpriseCommand) { c.UserID = "not-a-uuid" },
		"nil UUID":              func(c *ProvisionEnterpriseCommand) { c.UserID = uuid.Nil.String() },
		"blank name":            func(c *ProvisionEnterpriseCommand) { c.Name = " \n\t" },
		"oversized name":        func(c *ProvisionEnterpriseCommand) { c.Name = strings.Repeat("я", 513) },
		"blank address":         func(c *ProvisionEnterpriseCommand) { c.Address = " " },
		"missing latitude":      func(c *ProvisionEnterpriseCommand) { c.Latitude = nil },
		"missing longitude":     func(c *ProvisionEnterpriseCommand) { c.Longitude = nil },
		"latitude range":        func(c *ProvisionEnterpriseCommand) { v := 90.01; c.Latitude = &v },
		"longitude range":       func(c *ProvisionEnterpriseCommand) { v := -180.01; c.Longitude = &v },
		"NaN":                   func(c *ProvisionEnterpriseCommand) { v := math.NaN(); c.Latitude = &v },
		"infinity":              func(c *ProvisionEnterpriseCommand) { v := math.Inf(1); c.Longitude = &v },
	} {
		t.Run(name, func(t *testing.T) {
			command := validProvisionCommand()
			change(&command)
			repo := &provisionRecorder{}
			_, err := NewProvisionService(repo).Provision(context.Background(), command)
			require.ErrorIs(t, err, ErrInvalidProvisionInput)
			require.Nil(t, repo.enterprise)
		})
	}
}

func TestProvisionExplicitZeroAndBoundaries(t *testing.T) {
	for _, coordinates := range [][2]float64{{0, 0}, {-90, -180}, {90, 180}} {
		command := validProvisionCommand()
		command.Latitude, command.Longitude = &coordinates[0], &coordinates[1]
		repo := &provisionRecorder{}
		result, err := NewProvisionService(repo).Provision(context.Background(), command)
		require.NoError(t, err)
		require.Equal(t, command.EnterpriseID, result.EnterpriseID)
		require.Equal(t, command.UserID, result.UserID)
		require.Equal(t, "Synthetic facility", repo.enterprise.Name())
		require.Equal(t, "Synthetic address", repo.enterprise.Address())
		require.Equal(t, domain.EnterpriseStatusActive, repo.enterprise.Status())
		require.Equal(t, coordinates[0], repo.enterprise.Latitude())
		require.Equal(t, coordinates[1], repo.enterprise.Longitude())
	}
}

func TestProvisionFailureReturnsNoResultOrSensitiveDetails(t *testing.T) {
	repo := &provisionRecorder{err: errors.New("database error includes private address")}
	result, err := NewProvisionService(repo).Provision(context.Background(), validProvisionCommand())
	require.ErrorIs(t, err, ErrProvisionFailed)
	require.NotContains(t, err.Error(), "private address")
	require.Zero(t, result)
}
