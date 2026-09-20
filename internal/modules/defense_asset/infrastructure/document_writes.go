package infrastructure

import (
	"context"
	"errors"
	"github.com/fortis/backend/internal/audit"
	"github.com/fortis/backend/internal/auth"
	"github.com/fortis/backend/internal/modules/defense_asset/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"time"
)

func documentMutationAccess(tx *gorm.DB, actor, assetID string) (*DefenseAssetModel, error) {
	if err := auth.RequireIdentity(actor); err != nil {
		return nil, err
	}
	var asset DefenseAssetModel
	if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Where("id = ?", assetID).First(&asset).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, auth.ErrNotFound
		}
		return nil, err
	}
	if asset.IsPublic {
		return nil, auth.ErrForbidden
	}
	if asset.EnterpriseID == nil {
		return nil, auth.ErrNotFound
	}
	var membership struct{ UserID string }
	err := tx.Table("user_enterprises").Clauses(clause.Locking{Strength: "SHARE"}).Where("user_id = ? AND enterprise_id = ?", actor, *asset.EnterpriseID).Take(&membership).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, auth.ErrNotFound
	}
	return &asset, err
}
func (r *DocumentRepository) CommitDocument(ctx context.Context, doc *domain.Document, actor, action string) error {
	return r.executor.WithContext(ctx).Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Transaction(func(tx *gorm.DB) error {
		asset, err := documentMutationAccess(tx, actor, doc.AssetID())
		if err != nil {
			return err
		}
		model := toDocumentModel(doc)
		if action == "document.upload" {
			if err = tx.Create(model).Error; err != nil {
				return err
			}
		} else {
			result := tx.Model(&DefenseAssetDocumentModel{}).Where("id = ? AND asset_id = ? AND status = 'quarantined' AND deleted_at IS NULL", doc.ID(), doc.AssetID()).Updates(map[string]any{"status": doc.Status(), "updated_at": time.Now().UTC()})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return auth.ErrNotFound
			}
		}
		enterprise := asset.EnterpriseID.String()
		version := 1
		return audit.Append(tx, audit.Event{ActorID: actor, EnterpriseID: &enterprise, EntityType: "document", EntityID: doc.ID(), Action: action, NewVersion: &version, RequestID: audit.RequestID(ctx)})
	})
}
func (r *DocumentRepository) DeleteDocument(ctx context.Context, id, actor string) error {
	return r.executor.WithContext(ctx).Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Transaction(func(tx *gorm.DB) error {
		var doc DefenseAssetDocumentModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND deleted_at IS NULL", id).First(&doc).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return auth.ErrNotFound
			}
			return err
		}
		asset, err := documentMutationAccess(tx, actor, doc.AssetID.String())
		if err != nil {
			return err
		}
		if err = tx.Model(&doc).Updates(map[string]any{"deleted_at": time.Now().UTC(), "updated_at": time.Now().UTC()}).Error; err != nil {
			return err
		}
		enterprise := asset.EnterpriseID.String()
		version := 1
		return audit.Append(tx, audit.Event{ActorID: actor, EnterpriseID: &enterprise, EntityType: "document", EntityID: id, Action: "document.delete", PreviousVersion: &version, RequestID: audit.RequestID(ctx)})
	})
}
func (r *DocumentRepository) ListDocuments(ctx context.Context, id string, limit, offset int) ([]*domain.Document, int64, error) {
	query := r.executor.WithContext(ctx).Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Model(&DefenseAssetDocumentModel{}).Where("asset_id = ? AND deleted_at IS NULL", id)
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	var models []DefenseAssetDocumentModel
	if err := query.Order("created_at DESC, id").Limit(limit).Offset(offset).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	out := make([]*domain.Document, 0, len(models))
	for _, m := range models {
		d, err := m.toDomain()
		if err != nil {
			return nil, 0, err
		}
		out = append(out, d)
	}
	return out, count, nil
}
