package application

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fortis/backend/internal/modules/enterprise/domain"
)

// mockRepo — мок репозитория для тестирования сервиса.
type mockRepo struct {
	enterprises map[string]*domain.Enterprise
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		enterprises: make(map[string]*domain.Enterprise),
	}
}

func (m *mockRepo) Save(ctx context.Context, e *domain.Enterprise) error {
	m.enterprises[e.ID()] = e
	return nil
}

func (m *mockRepo) FindByID(ctx context.Context, id string) (*domain.Enterprise, error) {
	e, ok := m.enterprises[id]
	if !ok {
		return nil, domain.ErrEnterpriseNotFound
	}
	return e, nil
}

func (m *mockRepo) FindAll(ctx context.Context, limit, offset int) ([]*domain.Enterprise, int64, error) {
	var result []*domain.Enterprise
	for _, e := range m.enterprises {
		result = append(result, e)
	}
	total := int64(len(result))

	// Простая пагинация
	if offset >= len(result) {
		return []*domain.Enterprise{}, total, nil
	}
	if limit <= 0 {
		limit = 20
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], total, nil
}

func (m *mockRepo) Delete(ctx context.Context, id string) error {
	if _, ok := m.enterprises[id]; !ok {
		return domain.ErrEnterpriseNotFound
	}
	delete(m.enterprises, id)
	return nil
}

func TestCreate_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewEnterpriseService(repo)

	e, err := svc.Create(context.Background(), "Тестовое предприятие", "Адрес", domain.EnterpriseStatusActive, 55.75, 37.62)

	require.NoError(t, err)
	assert.NotNil(t, e)
	assert.Equal(t, "Тестовое предприятие", e.Name())
	assert.Equal(t, domain.EnterpriseStatusActive, e.Status())

	// Проверяем, что сохранено в репозитории
	saved, err := repo.FindByID(context.Background(), e.ID())
	require.NoError(t, err)
	assert.Equal(t, e.Name(), saved.Name())
}

func TestCreate_EmptyName(t *testing.T) {
	repo := newMockRepo()
	svc := NewEnterpriseService(repo)

	e, err := svc.Create(context.Background(), "", "Адрес", domain.EnterpriseStatusActive, 55.75, 37.62)

	require.Error(t, err)
	assert.Nil(t, e)
}

func TestCreate_InvalidStatus(t *testing.T) {
	repo := newMockRepo()
	svc := NewEnterpriseService(repo)

	e, err := svc.Create(context.Background(), "Test", "Адрес", domain.EnterpriseStatus("invalid"), 55.75, 37.62)

	require.Error(t, err)
	assert.Nil(t, e)
}

func TestGet_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewEnterpriseService(repo)

	created, _ := svc.Create(context.Background(), "Тестовое предприятие", "Адрес", domain.EnterpriseStatusActive, 55.75, 37.62)

	found, err := svc.Get(context.Background(), created.ID())
	require.NoError(t, err)
	assert.Equal(t, created.ID(), found.ID())
	assert.Equal(t, created.Name(), found.Name())
}

func TestGet_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewEnterpriseService(repo)

	e, err := svc.Get(context.Background(), "nonexistent-id")

	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrEnterpriseNotFound)
	assert.Nil(t, e)
}

func TestList_Empty(t *testing.T) {
	repo := newMockRepo()
	svc := NewEnterpriseService(repo)

	enterprises, total, err := svc.List(context.Background(), 20, 0)

	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, enterprises)
}

func TestList_WithData(t *testing.T) {
	repo := newMockRepo()
	svc := NewEnterpriseService(repo)

	_, _ = svc.Create(context.Background(), "Предприятие 1", "Адрес 1", domain.EnterpriseStatusActive, 55.75, 37.62)
	_, _ = svc.Create(context.Background(), "Предприятие 2", "Адрес 2", domain.EnterpriseStatusConfiguring, 56.0, 38.0)

	enterprises, total, err := svc.List(context.Background(), 20, 0)

	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, enterprises, 2)
}

func TestUpdate_Success(t *testing.T) {
	now := time.Now().UTC()
	repo := newMockRepo()
	svc := NewEnterpriseService(repo)

	created, _ := domain.NewEnterprise("test-id", "Старое название", "Старый адрес", domain.EnterpriseStatusActive, 55.75, 37.62, now, now)
	_ = repo.Save(context.Background(), created)

	updated, err := svc.Update(context.Background(), "test-id", "Новое название", "Новый адрес", domain.EnterpriseStatusOffline, 60.0, 30.0)

	require.NoError(t, err)
	assert.Equal(t, "Новое название", updated.Name())
	assert.Equal(t, "Новый адрес", updated.Address())
	assert.Equal(t, domain.EnterpriseStatusOffline, updated.Status())
	assert.InDelta(t, 60.0, updated.Latitude(), 0.0001)
	assert.InDelta(t, 30.0, updated.Longitude(), 0.0001)
}

func TestUpdate_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewEnterpriseService(repo)

	e, err := svc.Update(context.Background(), "nonexistent", "Имя", "Адрес", domain.EnterpriseStatusActive, 55.75, 37.62)

	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrEnterpriseNotFound)
	assert.Nil(t, e)
}

func TestDelete_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewEnterpriseService(repo)

	created, _ := svc.Create(context.Background(), "Тестовое предприятие", "Адрес", domain.EnterpriseStatusActive, 55.75, 37.62)

	err := svc.Delete(context.Background(), created.ID())
	require.NoError(t, err)

	_, err = svc.Get(context.Background(), created.ID())
	assert.ErrorIs(t, err, domain.ErrEnterpriseNotFound)
}

func TestDelete_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewEnterpriseService(repo)

	err := svc.Delete(context.Background(), "nonexistent-id")
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrEnterpriseNotFound)
}

func TestList_LimitDefaults(t *testing.T) {
	_ = newMockRepo()
	svc := NewEnterpriseService(newMockRepo())

	// limit <= 0 defaults to 20
	_, _, err := svc.List(context.Background(), 0, 0)
	assert.NoError(t, err)
}

func TestCreate_InvalidCoordinates(t *testing.T) {
	repo := newMockRepo()
	svc := NewEnterpriseService(repo)

	e, err := svc.Create(context.Background(), "Test", "Address", domain.EnterpriseStatusActive, 100.0, 37.62)
	require.Error(t, err)
	assert.Nil(t, e)

	e, err = svc.Create(context.Background(), "Test", "Address", domain.EnterpriseStatusActive, 55.75, 200.0)
	require.Error(t, err)
	assert.Nil(t, e)
}

func TestUpdate_WithEmptyName(t *testing.T) {
	now := time.Now().UTC()
	repo := newMockRepo()
	svc := NewEnterpriseService(repo)

	created, _ := domain.NewEnterprise("test-id", "Название", "Адрес", domain.EnterpriseStatusActive, 55.75, 37.62, now, now)
	_ = repo.Save(context.Background(), created)

	// Пустое имя в Update не меняет (оставляет старое)
	updated, err := svc.Update(context.Background(), "test-id", "", "Новый адрес", domain.EnterpriseStatusOffline, 60.0, 30.0)

	require.NoError(t, err)
	assert.Equal(t, "Название", updated.Name())
	assert.Equal(t, "Новый адрес", updated.Address())
}

func TestUpdate_WithEmptyStatus(t *testing.T) {
	now := time.Now().UTC()
	repo := newMockRepo()
	svc := NewEnterpriseService(repo)

	created, _ := domain.NewEnterprise("test-id", "Название", "Адрес", domain.EnterpriseStatusActive, 55.75, 37.62, now, now)
	_ = repo.Save(context.Background(), created)

	// Пустой статус = не меняем статус
	updated, err := svc.Update(context.Background(), "test-id", "Новое название", "Адрес", "", 55.75, 37.62)

	require.NoError(t, err)
	assert.Equal(t, domain.EnterpriseStatusActive, updated.Status())
}
