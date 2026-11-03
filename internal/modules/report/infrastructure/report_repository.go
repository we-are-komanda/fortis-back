package infrastructure

import (
	"context"
	"errors"
	"github.com/fortis/backend/internal/audit"
	"github.com/fortis/backend/internal/auth"
	project "github.com/fortis/backend/internal/modules/defense_project/infrastructure"
	"github.com/fortis/backend/internal/modules/report/domain"
	"github.com/fortis/backend/internal/rdbms"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReportRepository struct{ db rdbms.Executor }

func NewReportRepository(db rdbms.Executor) *ReportRepository { return &ReportRepository{db} }
func keyQuery(tx *gorm.DB, key domain.ReportKey) *gorm.DB {
	return tx.Where("project_id=? AND project_version=? AND report_schema_version=? AND hide_cost=? AND map_view=?", key.ProjectID(), key.ProjectVersion(), domain.SchemaVersion, key.HideCost(), key.MapView())
}
func fromModel(m IssuedReportModel) *domain.IssuedReport {
	k, _ := domain.NewReportKey(m.ProjectID, m.ProjectVersion, m.HideCost, m.MapView)
	return domain.NewIssuedReport(m.ID, k, m.EnterpriseID, m.GeneratedAt, m.PayloadJSON)
}

// Keep the project stable while issuing an immutable report.
func lockAccess(tx *gorm.DB, actor, id string) (string, error) {
	var p project.DefenseProjectModel
	err := tx.Select("id", "enterprise_id").Clauses(clause.Locking{Strength: "KEY SHARE"}).Where("id=?", id).Take(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", auth.ErrNotFound
	}
	if err != nil {
		return "", err
	}
	if p.EnterpriseID == "" {
		return "", auth.ErrNotFound
	}
	return p.EnterpriseID, nil
}
func (r *ReportRepository) FindByKey(ctx context.Context, actor string, key domain.ReportKey) (*domain.IssuedReport, error) {
	var m IssuedReportModel
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		enterprise, err := lockAccess(tx, actor, key.ProjectID())
		if err != nil {
			return err
		}
		err = keyQuery(tx, key).Take(&m).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrReportNotFound
		}
		if err == nil && m.EnterpriseID != enterprise {
			return auth.ErrNotFound
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return fromModel(m), nil
}
func (r *ReportRepository) FindByID(ctx context.Context, actor, id string) (*domain.IssuedReport, error) {
	var m IssuedReportModel
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id=?", id).Take(&m).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return auth.ErrNotFound
			}
			return err
		}
		enterprise, err := lockAccess(tx, actor, m.ProjectID)
		if err != nil {
			return err
		}
		if enterprise != m.EnterpriseID {
			return auth.ErrNotFound
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return fromModel(m), nil
}
func (r *ReportRepository) Issue(ctx context.Context, actor string, report *domain.IssuedReport) (*domain.IssuedReport, error) {
	k := report.Key()
	m := IssuedReportModel{ID: report.ID(), ProjectID: k.ProjectID(), ProjectVersion: k.ProjectVersion(), EnterpriseID: report.EnterpriseID(), ReportSchemaVersion: domain.SchemaVersion, HideCost: k.HideCost(), MapView: k.MapView(), GeneratedAt: report.GeneratedAt(), CreatedBy: actor, PayloadJSON: report.Payload()}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		enterprise, err := lockAccess(tx, actor, k.ProjectID())
		if err != nil {
			return err
		}
		if enterprise != m.EnterpriseID {
			return auth.ErrNotFound
		}
		result := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "project_id"}, {Name: "project_version"}, {Name: "report_schema_version"}, {Name: "hide_cost"}, {Name: "map_view"}}, DoNothing: true}).Create(&m)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			m = IssuedReportModel{}
			return keyQuery(tx, k).Take(&m).Error
		}
		version := k.ProjectVersion()
		if err := audit.Append(tx, audit.Event{ActorID: actor, EnterpriseID: &enterprise, EntityType: "report", EntityID: m.ID, Action: "report.issue", NewVersion: &version, RequestID: audit.RequestID(ctx)}); err != nil {
			return err
		}
		// Return PostgreSQL's stored representation on first issuance too.
		m = IssuedReportModel{}
		return keyQuery(tx, k).Take(&m).Error
	})
	if err != nil {
		return nil, err
	}
	return fromModel(m), nil
}
