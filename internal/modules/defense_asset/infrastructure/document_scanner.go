package infrastructure

import (
	"context"
	"errors"
	"github.com/fortis/backend/internal/modules/defense_asset/domain"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type ClamAVConfig struct{ Executable string }
type ClamAVScanner struct{ executable string }

func NewClamAVScanner(c ClamAVConfig) *ClamAVScanner { return &ClamAVScanner{executable: c.Executable} }
func (s *ClamAVScanner) Ready() bool {
	if s == nil || !filepath.IsAbs(s.executable) {
		return false
	}
	i, e := os.Stat(s.executable)
	return e == nil && i.Mode().IsRegular() && i.Mode().Perm()&0111 != 0 && i.Mode().Perm()&0022 == 0
}
func (s *ClamAVScanner) Scan(ctx context.Context, path string) error {
	if !s.Ready() || !filepath.IsAbs(path) {
		return domain.ErrDocumentUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	// No shell, caller-supplied flags or output logging. Signature updates and the
	// executable are controlled by the deployment operator, never an HTTP request.
	c := exec.CommandContext(ctx, s.executable, "--no-summary", "--stdout", "--alert-exceeds-max=yes", "--alert-encrypted=yes", "--", path)
	c.Stdout = io.Discard
	c.Stderr = io.Discard
	err := c.Run()
	if ctx.Err() != nil {
		return domain.ErrDocumentUnavailable
	}
	if err == nil {
		return nil
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() == 1 {
		return domain.ErrDocumentRejected
	}
	return domain.ErrDocumentUnavailable
}
