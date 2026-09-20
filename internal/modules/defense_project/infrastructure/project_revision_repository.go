package infrastructure

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/fortis/backend/internal/audit"
	"github.com/fortis/backend/internal/auth"
	budgetDomain "github.com/fortis/backend/internal/modules/budget/domain"
	budgetInfra "github.com/fortis/backend/internal/modules/budget/infrastructure"
	"github.com/fortis/backend/internal/modules/defense_project/document"
	"github.com/fortis/backend/internal/modules/defense_project/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Commit is the single database boundary for current state, budget and revision.
func (r *DefenseProjectRepository) Commit(ctx context.Context, project *domain.DefenseProject, write domain.ProjectWrite) (*domain.DefenseProject, error) {
	var saved *domain.DefenseProject
	err := r.executor.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if write.IdempotencyKey != "" {
			replay, err := reserveProjectRequest(tx, write)
			if err != nil {
				return err
			}
			if replay != nil {
				if replay.ProjectID == nil {
					return auth.ErrNotFound
				}
				var current DefenseProjectModel
				if err := tx.Where("id = ?", *replay.ProjectID).First(&current).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return auth.ErrNotFound
					}
					return err
				}
				if err := lockProjectMembership(tx, write.ActorID, current.EnterpriseID); err != nil {
					return err
				}
				if replay.PayloadDigest != write.PayloadDigest {
					return domain.ErrIdempotencyConflict
				}
				saved, err = readRevision(tx, *replay.ProjectID, replay.ProjectVersion)
				return err
			}
		}
		candidate := *project
		if !write.Create {
			var current DefenseProjectModel
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", project.ProjectID()).First(&current).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return auth.ErrNotFound
				}
				return err
			}
			if err := lockProjectMembership(tx, write.ActorID, current.EnterpriseID); err != nil {
				return err
			}
			if current.EnterpriseID != project.EnterpriseID() {
				return domain.ErrProjectOwnershipImmutable
			}
			if current.Version != project.Version() {
				return domain.ErrVersionConflict
			}
			candidate.SetVersion(current.Version + 1)
		} else {
			if err := lockProjectMembership(tx, write.ActorID, project.EnterpriseID()); err != nil {
				return err
			}
			candidate.SetVersion(1)
		}
		model, err := ToModel(&candidate)
		if err != nil {
			return err
		}
		if write.Create {
			if err := tx.Create(model).Error; err != nil {
				return err
			}
		} else {
			result := tx.Model(&DefenseProjectModel{}).Where("id = ? AND version = ?", model.ID, project.Version()).Updates(map[string]any{
				"name": model.Name, "project_data": model.ProjectData, "version": model.Version, "updated_at": model.UpdatedAt,
			})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return domain.ErrVersionConflict
			}
		}
		if write.BudgetJSON != nil {
			budget := budgetInfra.BudgetConfigModel{ProjectID: model.ID, ConfigData: *write.BudgetJSON}
			if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "project_id"}}, DoUpdates: clause.AssignmentColumns([]string{"config_data", "updated_at"})}).Create(&budget).Error; err != nil {
				return err
			}
		}
		// Re-read JSONB so canonical numbers/digest match the actual persisted snapshot.
		if err := tx.Where("id = ?", model.ID).First(model).Error; err != nil {
			return err
		}
		var budget budgetInfra.BudgetConfigModel
		budgetJSON := json.RawMessage("null")
		err = tx.Where("project_id = ?", model.ID).First(&budget).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil {
			budgetJSON = json.RawMessage(budget.ConfigData)
		}
		snapshot, err := json.Marshal(map[string]any{"project": json.RawMessage(model.ProjectData), "budget": budgetJSON, "inputDataVersions": map[string]string{}})
		if err != nil {
			return err
		}
		snapshot, err = document.Canonical(snapshot)
		if err != nil {
			return err
		}
		digest := sha256.Sum256(snapshot)
		revision := ProjectRevisionModel{ProjectID: model.ID, Version: model.Version, SnapshotJSON: string(snapshot), SnapshotDigest: hex.EncodeToString(digest[:]), CreatedAt: time.Now().UTC(), CostCalculationVersion: budgetDomain.CostCalculationV1}
		if write.ActorID != "" {
			revision.CreatedBy = &write.ActorID
		}
		if err := tx.Create(&revision).Error; err != nil {
			return err
		}
		if write.ActorID != "" {
			var previous *int
			if !write.Create {
				version := project.Version()
				previous = &version
			}
			next := model.Version
			entityType, action := "project", "project."+write.Operation
			if write.Operation == "budget" {
				entityType, action = "budget", "budget.update"
			}
			if err := audit.Append(tx, audit.Event{ActorID: write.ActorID, EnterpriseID: &model.EnterpriseID, EntityType: entityType, EntityID: model.ID, Action: action, PreviousVersion: previous, NewVersion: &next, RequestID: audit.RequestID(ctx)}); err != nil {
				return err
			}
		}
		if write.IdempotencyKey != "" {
			if err := tx.Model(&ProjectIdempotencyModel{}).Where("actor_id = ? AND operation = ? AND idempotency_key = ?", write.ActorID, write.Operation, write.IdempotencyKey).Updates(map[string]any{"project_id": model.ID, "project_version": model.Version}).Error; err != nil {
				return err
			}
		}
		saved, err = model.ToDomain()
		if err == nil {
			saved.SetSnapshot(revision.SnapshotJSON, revision.SnapshotDigest)
			saved.SetCostCalculationVersion(revision.CostCalculationVersion)
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return saved, nil
}

func (r *DefenseProjectRepository) DeleteAuthorized(ctx context.Context, actorID, id string) error {
	return r.executor.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current DefenseProjectModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&current).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return auth.ErrNotFound
			}
			return err
		}
		if err := auth.RequireIdentity(actorID); err != nil {
			return err
		}
		if err := lockProjectMembership(tx, actorID, current.EnterpriseID); err != nil {
			return err
		}
		if err := tx.Delete(&current).Error; err != nil {
			return err
		}
		return audit.Append(tx, audit.Event{ActorID: actorID, EnterpriseID: &current.EnterpriseID, EntityType: "project", EntityID: id, Action: "project.delete", PreviousVersion: &current.Version, RequestID: audit.RequestID(ctx)})
	})
}

// The application already authorizes; this persisted guard closes revoke/write races.
func lockProjectMembership(tx *gorm.DB, actorID, enterpriseID string) error {
	if actorID == "" {
		return nil
	} // Trusted repository-only fixtures/provisioning have no browser actor.
	if enterpriseID == "" {
		return auth.ErrNotFound
	}
	var membership struct{ UserID string }
	err := tx.Table("user_enterprises").Clauses(clause.Locking{Strength: "SHARE"}).Where("user_id = ? AND enterprise_id = ?", actorID, enterpriseID).Take(&membership).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return auth.ErrNotFound
	}
	return err
}

func reserveProjectRequest(tx *gorm.DB, write domain.ProjectWrite) (*ProjectIdempotencyModel, error) {
	now := time.Now().UTC()
	record := ProjectIdempotencyModel{ActorID: write.ActorID, Operation: write.Operation, Key: write.IdempotencyKey, PayloadDigest: write.PayloadDigest, ProjectVersion: 1, ExpiresAt: now.Add(24 * time.Hour)}
	result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 1 {
		return nil, nil
	}
	var existing ProjectIdempotencyModel
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("actor_id = ? AND operation = ? AND idempotency_key = ?", write.ActorID, write.Operation, write.IdempotencyKey).First(&existing).Error; err != nil {
		return nil, err
	}
	if existing.ExpiresAt.After(now) {
		return &existing, nil
	}
	// Keep the locked row: DELETE/INSERT makes waiting READ COMMITTED readers
	// miss the replacement. Clear the old result before reserving a new request.
	err := tx.Model(&existing).Updates(map[string]any{
		"payload_digest":  record.PayloadDigest,
		"project_id":      nil,
		"project_version": record.ProjectVersion,
		"expires_at":      record.ExpiresAt,
	}).Error
	return nil, err
}

func (r *DefenseProjectRepository) FindRevision(ctx context.Context, id string, version int) (*domain.DefenseProject, error) {
	return readRevision(r.executor.WithContext(ctx), id, version)
}

func readRevision(tx *gorm.DB, id string, version int) (*domain.DefenseProject, error) {
	var revision ProjectRevisionModel
	if err := tx.Where("project_id = ? AND version = ?", id, version).First(&revision).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrRevisionNotFound
		}
		return nil, err
	}
	var envelope struct {
		Project json.RawMessage `json:"project"`
	}
	if err := json.Unmarshal([]byte(revision.SnapshotJSON), &envelope); err != nil {
		return nil, err
	}
	var metadata struct {
		Name         string    `json:"name"`
		EnterpriseID string    `json:"enterpriseId"`
		UpdatedAt    time.Time `json:"updatedAt"`
	}
	if err := json.Unmarshal(envelope.Project, &metadata); err != nil {
		return nil, err
	}
	model := DefenseProjectModel{ID: id, Name: metadata.Name, EnterpriseID: metadata.EnterpriseID, Version: version, ProjectData: string(envelope.Project), UpdatedAt: metadata.UpdatedAt}
	project, err := model.ToDomain()
	if err == nil {
		project.SetSnapshot(revision.SnapshotJSON, revision.SnapshotDigest)
		project.SetCostCalculationVersion(revision.CostCalculationVersion)
	}
	return project, err
}
