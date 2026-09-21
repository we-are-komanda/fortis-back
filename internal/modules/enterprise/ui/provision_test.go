package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	ea "github.com/fortis/backend/internal/modules/enterprise/application"
	ed "github.com/fortis/backend/internal/modules/enterprise/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type provisionRepository struct{ calls int }

func (r *provisionRepository) Provision(context.Context, *ed.Enterprise, string) error {
	r.calls++
	return nil
}

func TestProvisionCLIRequiresExplicitCoordinatesAndSafeResult(t *testing.T) {
	id, userID := uuid.NewString(), uuid.NewString()
	base := []string{"--enterprise-id", id, "--user-id", userID, "--name", "private name", "--address", "private address"}
	for _, extra := range [][]string{nil, {"--latitude", "0"}, {"--longitude", "0"}, {"--latitude", "NaN", "--longitude", "0"}, {"--latitude", "0", "--longitude", "Inf"}, {"--latitude", "secret-input", "--longitude", "0"}, {"--latitude", "0", "--longitude", "0", "unexpected"}} {
		repo, out := &provisionRepository{}, &bytes.Buffer{}
		err := WriteProvision(context.Background(), ea.NewProvisionService(repo), out, append(append([]string{}, base...), extra...))
		require.Error(t, err)
		require.Zero(t, repo.calls)
		require.Empty(t, out.String())
		require.NotContains(t, err.Error(), "secret-input")
	}
	repo, out := &provisionRepository{}, &bytes.Buffer{}
	require.NoError(t, WriteProvision(context.Background(), ea.NewProvisionService(repo), out, append(base, "--latitude", "0", "--longitude", "0")))
	var result map[string]string
	require.NoError(t, json.Unmarshal(out.Bytes(), &result))
	require.Equal(t, map[string]string{"enterpriseId": id, "userId": userID}, result)
	require.Equal(t, 1, repo.calls)
}
