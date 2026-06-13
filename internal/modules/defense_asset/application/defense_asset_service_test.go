package application

import (
	"context"
	"errors"
	"testing"

	"github.com/fortis/backend/internal/modules/defense_asset/domain"
)

// mockDefenseAssetRepo — мок репозитория для тестирования сервиса.
type mockDefenseAssetRepo struct {
	assets  map[string]*domain.DefenseAsset
	saveErr error
}

func newMockRepo() *mockDefenseAssetRepo {
	return &mockDefenseAssetRepo{
		assets: make(map[string]*domain.DefenseAsset),
	}
}

func (m *mockDefenseAssetRepo) Save(ctx context.Context, asset *domain.DefenseAsset) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.assets[asset.ID()] = asset
	return nil
}

func (m *mockDefenseAssetRepo) FindByID(ctx context.Context, id string) (*domain.DefenseAsset, error) {
	asset, ok := m.assets[id]
	if !ok {
		return nil, domain.ErrDefenseAssetNotFound
	}
	return asset, nil
}

func (m *mockDefenseAssetRepo) FindAll(ctx context.Context, filter domain.DefenseAssetFilter) ([]*domain.DefenseAsset, int64, error) {
	var result []*domain.DefenseAsset
	for _, a := range m.assets {
		if filter.IsPublic != nil && a.IsPublic() != *filter.IsPublic {
			continue
		}
		if filter.Category != nil && a.Category() != *filter.Category {
			continue
		}
		result = append(result, a)
	}
	total := int64(len(result))

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	if offset >= len(result) {
		return []*domain.DefenseAsset{}, total, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}

	return result[offset:end], total, nil
}

func (m *mockDefenseAssetRepo) Update(ctx context.Context, asset *domain.DefenseAsset) error {
	if _, ok := m.assets[asset.ID()]; !ok {
		return domain.ErrDefenseAssetNotFound
	}
	m.assets[asset.ID()] = asset
	return nil
}

func (m *mockDefenseAssetRepo) Delete(ctx context.Context, id string) error {
	if _, ok := m.assets[id]; !ok {
		return domain.ErrDefenseAssetNotFound
	}
	delete(m.assets, id)
	return nil
}

func TestDefenseAssetService_Create_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewDefenseAssetService(repo)

	cat := domain.DefenseAssetCategoryRadar
	ct := domain.DefenseAssetCoverageCircle

	input := CreateInput{
		Name:         "РЛС 55Ж6",
		Category:     cat,
		CoverageType: ct,
	}

	asset, err := svc.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if asset.Name() != "РЛС 55Ж6" {
		t.Errorf("expected name 'РЛС 55Ж6', got %q", asset.Name())
	}
	if asset.Category() != cat {
		t.Errorf("expected category %v, got %v", cat, asset.Category())
	}
	if asset.CoverageType() != ct {
		t.Errorf("expected coverage type %v, got %v", ct, asset.CoverageType())
	}
	if asset.ID() == "" {
		t.Error("expected non-empty ID")
	}
}

func TestDefenseAssetService_Create_InvalidName(t *testing.T) {
	repo := newMockRepo()
	svc := NewDefenseAssetService(repo)

	input := CreateInput{
		Name:         "",
		Category:     domain.DefenseAssetCategoryRadar,
		CoverageType: domain.DefenseAssetCoverageCircle,
	}

	_, err := svc.Create(context.Background(), input)
	if !errors.Is(err, domain.ErrDefenseAssetInvalidName) {
		t.Errorf("expected ErrDefenseAssetInvalidName, got %v", err)
	}
}

func TestDefenseAssetService_Create_InvalidCategory(t *testing.T) {
	repo := newMockRepo()
	svc := NewDefenseAssetService(repo)

	input := CreateInput{
		Name:         "Test",
		Category:     "invalid-category",
		CoverageType: domain.DefenseAssetCoverageCircle,
	}

	_, err := svc.Create(context.Background(), input)
	if !errors.Is(err, domain.ErrDefenseAssetInvalidCategory) {
		t.Errorf("expected ErrDefenseAssetInvalidCategory, got %v", err)
	}
}

func TestDefenseAssetService_GetByID_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewDefenseAssetService(repo)

	cat := domain.DefenseAssetCategoryRadar
	ct := domain.DefenseAssetCoverageCircle

	created, err := svc.Create(context.Background(), CreateInput{
		Name:         "Test Asset",
		Category:     cat,
		CoverageType: ct,
	})
	if err != nil {
		t.Fatalf("failed to create asset: %v", err)
	}

	found, err := svc.GetByID(context.Background(), created.ID())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if found.Name() != "Test Asset" {
		t.Errorf("expected name 'Test Asset', got %q", found.Name())
	}
}

func TestDefenseAssetService_GetByID_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewDefenseAssetService(repo)

	_, err := svc.GetByID(context.Background(), "non-existent-id")
	if !errors.Is(err, domain.ErrDefenseAssetNotFound) {
		t.Errorf("expected ErrDefenseAssetNotFound, got %v", err)
	}
}

func TestDefenseAssetService_List_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewDefenseAssetService(repo)

	// Create test assets
	for range 5 {
		_, err := svc.Create(context.Background(), CreateInput{
			Name:         "Asset",
			Category:     domain.DefenseAssetCategoryRadar,
			CoverageType: domain.DefenseAssetCoverageCircle,
		})
		if err != nil {
			t.Fatalf("failed to create asset: %v", err)
		}
	}

	assets, total, err := svc.List(context.Background(), nil, nil, nil, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(assets) != 5 {
		t.Errorf("expected 5 assets, got %d", len(assets))
	}
}

func TestDefenseAssetService_List_WithPagination(t *testing.T) {
	repo := newMockRepo()
	svc := NewDefenseAssetService(repo)

	for range 5 {
		_, err := svc.Create(context.Background(), CreateInput{
			Name:         "Asset",
			Category:     domain.DefenseAssetCategoryRadar,
			CoverageType: domain.DefenseAssetCoverageCircle,
		})
		if err != nil {
			t.Fatalf("failed to create asset: %v", err)
		}
	}

	assets, total, err := svc.List(context.Background(), nil, nil, nil, 2, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(assets) != 2 {
		t.Errorf("expected 2 assets, got %d", len(assets))
	}
}

func TestDefenseAssetService_Delete_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewDefenseAssetService(repo)

	created, err := svc.Create(context.Background(), CreateInput{
		Name:         "Test Asset",
		Category:     domain.DefenseAssetCategoryRadar,
		CoverageType: domain.DefenseAssetCoverageCircle,
	})
	if err != nil {
		t.Fatalf("failed to create asset: %v", err)
	}

	err = svc.Delete(context.Background(), created.ID())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.GetByID(context.Background(), created.ID())
	if !errors.Is(err, domain.ErrDefenseAssetNotFound) {
		t.Errorf("expected ErrDefenseAssetNotFound after delete, got %v", err)
	}
}

func TestDefenseAssetService_Delete_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewDefenseAssetService(repo)

	err := svc.Delete(context.Background(), "non-existent-id")
	if !errors.Is(err, domain.ErrDefenseAssetNotFound) {
		t.Errorf("expected ErrDefenseAssetNotFound, got %v", err)
	}
}
