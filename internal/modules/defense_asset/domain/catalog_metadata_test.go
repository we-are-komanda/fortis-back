package domain

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCatalogProvenanceAndExactPrice(t *testing.T) {
	for _, value := range []string{"-1", "1.5", "1000000000000001", "NaN", ""} {
		require.Error(t, ValidateMinor(&value), value)
	}
	for _, value := range []string{"0", "1000000000000000", "125"} {
		require.NoError(t, ValidateMinor(&value))
	}
	require.NoError(t, ValidateMinor(nil))
	_, err := NewProvenance("", "", "", "", "now", "actor", "confirmed", "revision", "", "")
	require.Error(t, err)
	_, err = NewProvenance("source", "", "https://example.test", "bad-date", "now", "actor", "confirmed", "revision", "", "")
	require.Error(t, err)
	p, err := NewProvenance("synthetic", "", "", "2026-09-20", "now", "actor", "demo", "revision", "", "")
	require.NoError(t, err)
	require.Equal(t, "demo", p.Quality())
}
