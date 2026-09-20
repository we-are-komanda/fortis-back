package infrastructure

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fortis/backend/internal/modules/demo_request/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func testDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("FRC12_TEST_DSN")
	if dsn == "" {
		t.Skip("FRC12_TEST_DSN unset: real PostgreSQL not verified")
	}
	base, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	schema := "frc12_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	require.NoError(t, base.Exec("CREATE SCHEMA "+schema).Error)
	db, err := gorm.Open(postgres.Open(dsn+" search_path="+schema), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	t.Cleanup(func() {
		pool, _ := db.DB()
		pool.Close()
		base.Exec("DROP SCHEMA " + schema + " CASCADE")
		pool, _ = base.DB()
		pool.Close()
	})
	files, err := filepath.Glob("../../../../migrations/*_demo_requests.up.sql")
	require.NoError(t, err)
	require.Len(t, files, 1)
	sql, err := os.ReadFile(files[0])
	require.NoError(t, err)
	require.NoError(t, db.Exec(string(sql)).Error)
	return db
}
func fixture(t *testing.T, now time.Time) *domain.Request {
	t.Helper()
	r, err := domain.NewRequest(uuid.NewString(), uuid.NewString(), "Test Person", "Test Org", "person@example.test", "synthetic", "synthetic-test-v1", true, []string{"synthetic-test-v1"}, now)
	require.NoError(t, err)
	return r
}
func TestRepositoryAtomicRollbackAndReceipt(t *testing.T) {
	db := testDatabase(t)
	repo := NewRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()
	lead := fixture(t, now)
	rollback := errors.New("synthetic commit abort")
	err := repo.Transaction(ctx, func(tx domain.Repository) error {
		_, err := tx.LockReceipt(ctx, "key", now.Add(24*time.Hour))
		require.NoError(t, err)
		require.NoError(t, tx.SaveRequest(ctx, lead))
		require.NoError(t, tx.SaveOutbox(ctx, lead))
		require.NoError(t, tx.SaveReceipt(ctx, domain.NewReceipt("key", lead.Digest(), lead.ID(), now.Add(24*time.Hour))))
		return rollback
	})
	require.ErrorIs(t, err, rollback)
	rows, total, err := repo.List(ctx, 20, 0)
	require.NoError(t, err)
	require.Empty(t, rows)
	require.Zero(t, total)
	require.NoError(t, repo.Transaction(ctx, func(tx domain.Repository) error {
		receipt, err := tx.LockReceipt(ctx, "key", now.Add(24*time.Hour))
		require.NoError(t, err)
		require.Empty(t, receipt.RequestID())
		require.NoError(t, tx.SaveRequest(ctx, lead))
		require.NoError(t, tx.SaveOutbox(ctx, lead))
		return tx.SaveReceipt(ctx, domain.NewReceipt("key", lead.Digest(), lead.ID(), now.Add(24*time.Hour)))
	}))
	rows, total, err = repo.List(ctx, 20, 0)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Equal(t, lead.ID(), rows[0].Request().ID())
	require.Equal(t, "pending", rows[0].Status())
}
func TestRepositoryRateLockAndWorkerLeaseAcrossConnections(t *testing.T) {
	db := testDatabase(t)
	repo := NewRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()
	lead := fixture(t, now)
	var wg sync.WaitGroup
	errCh := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errCh <- repo.Transaction(ctx, func(tx domain.Repository) error {
				times, err := tx.LockRate(ctx, "ip:key")
				if err != nil {
					return err
				}
				return tx.SaveRate(ctx, "ip:key", append(times, now))
			})
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}
	require.NoError(t, repo.Transaction(ctx, func(tx domain.Repository) error {
		times, err := tx.LockRate(ctx, "ip:key")
		require.Len(t, times, 8)
		return err
	}))
	require.NoError(t, repo.Transaction(ctx, func(tx domain.Repository) error {
		if err := tx.SaveRequest(ctx, lead); err != nil {
			return err
		}
		return tx.SaveOutbox(ctx, lead)
	}))
	claims := make(chan *domain.Delivery, 8)
	errCh = make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			d, err := repo.Claim(ctx, now.Add(time.Second), uuid.NewString(), now.Add(time.Minute))
			claims <- d
			errCh <- err
		}()
	}
	wg.Wait()
	close(claims)
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}
	var first *domain.Delivery
	count := 0
	for d := range claims {
		if d != nil {
			first = d
			count++
		}
	}
	require.Equal(t, 1, count)
	require.Equal(t, lead.EventID(), first.Request().EventID())
	second, err := repo.Claim(ctx, now.Add(2*time.Minute), uuid.NewString(), now.Add(3*time.Minute))
	require.NoError(t, err)
	require.NotNil(t, second)
	require.Equal(t, 2, second.Attempts())
	first.Complete()
	require.NoError(t, repo.Finish(ctx, first))
	rows, _, err := repo.List(ctx, 20, 0)
	require.NoError(t, err)
	require.Equal(t, "pending", rows[0].Status(), "expired worker must not acknowledge a newer lease")
	second.Complete()
	require.NoError(t, repo.Finish(ctx, second))
	rows, _, err = repo.List(ctx, 20, 0)
	require.NoError(t, err)
	require.Equal(t, "sent", rows[0].Status())
}
