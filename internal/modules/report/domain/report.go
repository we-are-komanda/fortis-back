package domain

import (
	budget "github.com/fortis/backend/internal/modules/budget/domain"
	project "github.com/fortis/backend/internal/modules/defense_project/domain"
	"time"
)

const SchemaVersion = "report-v1"

// ReportDraft owns one immutable revision. Hidden-cost drafts never hold finance.
type ReportDraft struct {
	id          string
	generatedAt time.Time
	key         ReportKey
	project     *project.DefenseProject
	cost        *budget.CostProjection
	budget      *budget.BudgetProjection
}

func NewReportDraft(id string, at time.Time, key ReportKey, p *project.DefenseProject, cost *budget.CostProjection, config *budget.BudgetProjection) (*ReportDraft, error) {
	if p == nil || p.ProjectID() != key.ProjectID() || p.Version() != key.ProjectVersion() || (key.HideCost() && (cost != nil || config != nil)) || (!key.HideCost() && (cost == nil || config == nil)) {
		return nil, ErrInvalidReport
	}
	if cost != nil {
		i, b := cost.Identity(), config.Identity()
		if i.ProjectID() != p.ProjectID() || i.ProjectVersion() != p.Version() || i.SnapshotDigest() != p.SnapshotDigest() || i.CalculationVersion() != p.CostCalculationVersion() || b.ProjectID() != i.ProjectID() || b.ProjectVersion() != i.ProjectVersion() || b.SnapshotDigest() != i.SnapshotDigest() || b.CalculationVersion() != i.CalculationVersion() {
			return nil, ErrInvalidReport
		}
	}
	return &ReportDraft{id, at, key, p, cost, config}, nil
}
func (d *ReportDraft) ID() string                       { return d.id }
func (d *ReportDraft) GeneratedAt() time.Time           { return d.generatedAt }
func (d *ReportDraft) Key() ReportKey                   { return d.key }
func (d *ReportDraft) Project() *project.DefenseProject { return d.project }
func (d *ReportDraft) Cost() *budget.CostProjection     { return d.cost }
func (d *ReportDraft) Budget() *budget.BudgetProjection { return d.budget }

// IssuedReport contains the immutable render-ready serialization, never a live project.
type IssuedReport struct {
	id           string
	key          ReportKey
	enterpriseID string
	generatedAt  time.Time
	payload      string
}

func NewIssuedReport(id string, key ReportKey, enterpriseID string, at time.Time, payload string) *IssuedReport {
	return &IssuedReport{id, key, enterpriseID, at, payload}
}
func (r *IssuedReport) ID() string             { return r.id }
func (r *IssuedReport) Key() ReportKey         { return r.key }
func (r *IssuedReport) EnterpriseID() string   { return r.enterpriseID }
func (r *IssuedReport) GeneratedAt() time.Time { return r.generatedAt }
func (r *IssuedReport) Payload() string        { return r.payload }
