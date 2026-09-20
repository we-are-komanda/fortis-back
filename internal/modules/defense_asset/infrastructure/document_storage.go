package infrastructure

import (
	"context"
	"github.com/fortis/backend/internal/modules/defense_asset/domain"
	"github.com/google/uuid"
	"io"
	"os"
	"path/filepath"
)

type DocumentStorageConfig struct {
	RootDir string
	Enabled bool
}
type LocalDocumentStorage struct{ root string }

func NewLocalDocumentStorage(c DocumentStorageConfig) *LocalDocumentStorage {
	s := &LocalDocumentStorage{}
	if !c.Enabled || c.RootDir == "" || !filepath.IsAbs(c.RootDir) {
		return s
	}
	if err := os.MkdirAll(c.RootDir, 0700); err != nil {
		return s
	}
	info, err := os.Lstat(c.RootDir)
	if err != nil || !info.IsDir() || info.Mode().Perm()&0077 != 0 {
		return s
	}
	s.root = c.RootDir
	return s
}
func (s *LocalDocumentStorage) Ready() bool { return s != nil && s.root != "" }
func (s *LocalDocumentStorage) path(key string) (string, error) {
	if !s.Ready() {
		return "", domain.ErrDocumentUnavailable
	}
	id, err := uuid.Parse(key)
	if err != nil || id.String() != key {
		return "", domain.ErrDocumentUnavailable
	}
	return filepath.Join(s.root, key), nil
}
func (s *LocalDocumentStorage) Put(ctx context.Context, key string, body []byte) (string, error) {
	if ctx.Err() != nil {
		return "", domain.ErrDocumentUnavailable
	}
	if len(body) > domain.MaxDocumentBytes {
		return "", domain.ErrDocumentTooLarge
	}
	path, err := s.path(key)
	if err != nil {
		return "", err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return "", domain.ErrDocumentUnavailable
	}
	_, err = f.Write(body)
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		os.Remove(path)
		return "", domain.ErrDocumentUnavailable
	}
	return path, nil
}
func (s *LocalDocumentStorage) Read(ctx context.Context, key string) ([]byte, error) {
	if ctx.Err() != nil {
		return nil, domain.ErrDocumentUnavailable
	}
	path, err := s.path(key)
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 || info.Size() > domain.MaxDocumentBytes {
		return nil, domain.ErrDocumentUnavailable
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, domain.ErrDocumentUnavailable
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, domain.MaxDocumentBytes+1))
	if err != nil || len(b) > domain.MaxDocumentBytes {
		return nil, domain.ErrDocumentUnavailable
	}
	return b, nil
}
func (s *LocalDocumentStorage) Remove(ctx context.Context, key string) error {
	path, err := s.path(key)
	if err != nil {
		return err
	}
	if err = os.Remove(path); err != nil && !os.IsNotExist(err) {
		return domain.ErrDocumentUnavailable
	}
	return nil
}
