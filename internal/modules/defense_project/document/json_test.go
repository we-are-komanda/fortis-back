package document

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProjectDocumentKeepsExtensionsButExcludesTransport(t *testing.T) {
	data, err := Encode(`{"requestId":"request-a","generatedAt":"transport-time","customFields":{"integer":9007199254740993,"note":"<>&"}}`, map[string]any{}, map[string]any{"version": 2})
	require.NoError(t, err)
	require.NotContains(t, string(data), "requestId")
	require.NotContains(t, string(data), "generatedAt")
	require.Contains(t, string(data), "9007199254740993")
	canonical, err := Canonical(data)
	require.NoError(t, err)
	require.Equal(t, `{"customFields":{"integer":9007199254740993,"note":"<>&"},"version":2}`, string(canonical))
}
