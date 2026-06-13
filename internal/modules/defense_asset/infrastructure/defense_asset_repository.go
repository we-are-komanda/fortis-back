package infrastructure

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/fortis/backend/internal/modules/defense_asset/domain"
	"github.com/fortis/backend/internal/rdbms"
)

// DefenseAssetRepository — реализация репозитория DefenseAsset.
type DefenseAssetRepository struct {
	executor rdbms.Executor
}

// NewDefenseAssetRepository создаёт новый репозиторий.
func NewDefenseAssetRepository(executor rdbms.Executor) domain.DefenseAssetRepositoryInterface {
	return &DefenseAssetRepository{
		executor: executor,
	}
}

// Save создаёт новое средство защиты.
func (r *DefenseAssetRepository) Save(ctx context.Context, asset *domain.DefenseAsset) error {
	model, err := ToModel(asset)
	if err != nil {
		return err
	}

	return r.executor.WithContext(ctx).Create(model).Error
}

// FindByID ищет средство защиты по ID.
func (r *DefenseAssetRepository) FindByID(ctx context.Context, id string) (*domain.DefenseAsset, error) {
	var model DefenseAssetModel
	result := r.executor.WithContext(ctx).Where("id = ?", id).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrDefenseAssetNotFound
		}
		return nil, result.Error
	}

	return model.ToDomain()
}

// FindAll возвращает список средств защиты с пагинацией и фильтрацией.
func (r *DefenseAssetRepository) FindAll(ctx context.Context, filter domain.DefenseAssetFilter) ([]*domain.DefenseAsset, int64, error) {
	db := r.executor.WithContext(ctx)

	// Базовый запрос
	query := db.Model(&DefenseAssetModel{})

	// Фильтры
	if filter.EnterpriseID != nil {
		eid, err := uuid.Parse(*filter.EnterpriseID)
		if err == nil {
			query = query.Where("enterprise_id = ?", eid)
		}
	}
	if filter.IsPublic != nil {
		query = query.Where("is_public = ?", *filter.IsPublic)
	}
	if filter.Category != nil {
		query = query.Where("asset_data->>'category' = ?", string(*filter.Category))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	var models []DefenseAssetModel
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&models).Error; err != nil {
		return nil, 0, err
	}

	assets := make([]*domain.DefenseAsset, len(models))
	for i, m := range models {
		asset, err := m.ToDomain()
		if err != nil {
			return nil, 0, err
		}
		assets[i] = asset
	}

	return assets, total, nil
}

// Update обновляет существующее средство защиты.
func (r *DefenseAssetRepository) Update(ctx context.Context, asset *domain.DefenseAsset) error {
	model, err := ToModel(asset)
	if err != nil {
		return err
	}

	result := r.executor.WithContext(ctx).Model(&DefenseAssetModel{}).Where("id = ?", model.ID).Updates(map[string]interface{}{
		"asset_data":  model.AssetData,
		"is_public":   model.IsPublic,
		"updated_at":  model.UpdatedAt,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrDefenseAssetNotFound
	}
	return nil
}

// Delete удаляет средство защиты по ID.
func (r *DefenseAssetRepository) Delete(ctx context.Context, id string) error {
	result := r.executor.WithContext(ctx).Where("id = ?", id).Delete(&DefenseAssetModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrDefenseAssetNotFound
	}
	return nil
}
