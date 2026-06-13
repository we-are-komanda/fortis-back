package domain

import "context"

// DefenseAssetFilter — фильтр для списка средств защиты.
type DefenseAssetFilter struct {
	EnterpriseID *string
	IsPublic     *bool
	Category     *DefenseAssetCategory
	Limit        int
	Offset       int
}

// DefenseAssetRepositoryInterface — интерфейс репозитория для DefenseAsset.
type DefenseAssetRepositoryInterface interface {
	// Save создаёт новое средство защиты.
	Save(ctx context.Context, asset *DefenseAsset) error

	// FindByID возвращает средство защиты по ID.
	FindByID(ctx context.Context, id string) (*DefenseAsset, error)

	// FindAll возвращает список средств защиты с пагинацией и фильтрацией.
	FindAll(ctx context.Context, filter DefenseAssetFilter) ([]*DefenseAsset, int64, error)

	// Update обновляет существующее средство защиты.
	Update(ctx context.Context, asset *DefenseAsset) error

	// Delete удаляет средство защиты по ID.
	Delete(ctx context.Context, id string) error
}
