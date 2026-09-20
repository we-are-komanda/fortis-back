package infrastructure

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestPrivateDocumentStorage(t *testing.T) {
	root := filepath.Join(t.TempDir(), "private")
	s := NewLocalDocumentStorage(DocumentStorageConfig{RootDir: root, Enabled: true})
	require.True(t, s.Ready())
	key := uuid.NewString()
	body := []byte("synthetic text")
	path, err := s.Put(context.Background(), key, body)
	require.NoError(t, err)
	info, err := os.Stat(path)
	require.NoError(t, err)
	require.EqualValues(t, 0600, info.Mode().Perm())
	got, err := s.Read(context.Background(), key)
	require.NoError(t, err)
	require.Equal(t, body, got)
	_, err = s.Put(context.Background(), key, body)
	require.Error(t, err)
	_, err = s.Read(context.Background(), "../outside")
	require.Error(t, err)
	require.NoError(t, s.Remove(context.Background(), key))
	require.False(t, NewLocalDocumentStorage(DocumentStorageConfig{}).Ready())
}
