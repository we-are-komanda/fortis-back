package infrastructure

import (
	"context"
	"fmt"
	"github.com/fortis/backend/internal/modules/defense_asset/domain"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestScannerAdapterFailClosed(t *testing.T) {
	require.False(t, NewClamAVScanner(ClamAVConfig{}).Ready())
	require.ErrorIs(t, NewClamAVScanner(ClamAVConfig{}).Scan(context.Background(), "/synthetic"), domain.ErrDocumentUnavailable)
	for _, code := range []int{0, 1, 2} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			executable := filepath.Join(t.TempDir(), "synthetic-scanner")
			require.NoError(t, os.WriteFile(executable, []byte(fmt.Sprintf("#!/bin/sh\nexit %d\n", code)), 0700))
			scanner := NewClamAVScanner(ClamAVConfig{Executable: executable})
			require.True(t, scanner.Ready())
			err := scanner.Scan(context.Background(), "/synthetic-private-file")
			switch code {
			case 0:
				require.NoError(t, err)
			case 1:
				require.ErrorIs(t, err, domain.ErrDocumentRejected)
			default:
				require.ErrorIs(t, err, domain.ErrDocumentUnavailable)
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			require.ErrorIs(t, scanner.Scan(ctx, "/synthetic"), domain.ErrDocumentUnavailable)
		})
	}
}
