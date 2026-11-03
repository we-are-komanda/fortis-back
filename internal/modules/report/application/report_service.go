package application

import (
	"context"
	"encoding/json"
	"errors"
	budgetDocument "github.com/fortis/backend/internal/modules/budget/document"
	budget "github.com/fortis/backend/internal/modules/budget/domain"
	project "github.com/fortis/backend/internal/modules/defense_project/domain"
	"github.com/fortis/backend/internal/modules/report/domain"
	"github.com/google/uuid"
	"time"
)

type ProjectService interface {
	GetRevision(context.Context, string, string, *int) (*project.DefenseProject, error)
}
type BudgetServiceInterface interface {
	FrozenCost(context.Context, *project.DefenseProject) (*budget.CostProjection, error)
}

// PayloadRenderer is the transport serialization boundary; only allowlisted JSON is persisted.
type PayloadRenderer func(*domain.ReportDraft) (string, error)
type ReportService struct {
	projects   ProjectService
	finance    BudgetServiceInterface
	repository domain.ReportRepository
	render     PayloadRenderer
}

func NewReportService(projects ProjectService, finance BudgetServiceInterface, repository domain.ReportRepository, render PayloadRenderer) *ReportService {
	return &ReportService{projects, finance, repository, render}
}
func (s *ReportService) GetReport(ctx context.Context, actor string, q GetReportQuery) (*domain.IssuedReport, error) {
	if q.ReportID != "" {
		return s.repository.FindByID(ctx, actor, q.ReportID)
	}
	version := q.Key.ProjectVersion()
	p, err := s.projects.GetRevision(ctx, actor, q.Key.ProjectID(), &version)
	if err != nil {
		return nil, err
	}
	existing, err := s.repository.FindByKey(ctx, actor, q.Key)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, domain.ErrReportNotFound) {
		return nil, err
	}
	var cost *budget.CostProjection
	var projection *budget.BudgetProjection
	if !q.Key.HideCost() {
		cost, err = s.finance.FrozenCost(ctx, p)
		if err != nil {
			return nil, err
		}
		var snapshot struct {
			Budget json.RawMessage `json:"budget"`
		}
		if err = json.Unmarshal([]byte(p.Snapshot()), &snapshot); err != nil {
			return nil, domain.ErrInvalidSnapshot
		}
		config, err := budgetDocument.ReadBudget(p.ProjectID(), string(snapshot.Budget))
		if err != nil {
			return nil, err
		}
		projection = budget.NewBudgetProjection(config, cost)
	}
	draft, err := domain.NewReportDraft(uuid.NewString(), time.Now().UTC(), q.Key, p, cost, projection)
	if err != nil {
		return nil, err
	}
	payload, err := s.render(draft)
	if err != nil {
		return nil, err
	}
	return s.repository.Issue(ctx, actor, domain.NewIssuedReport(draft.ID(), q.Key, p.EnterpriseID(), draft.GeneratedAt(), payload))
}
