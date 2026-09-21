package application

import (
	"context"
	"encoding/json"

	"github.com/fortis/backend/internal/modules/budget/document"
	"github.com/fortis/backend/internal/modules/budget/domain"
	projectDomain "github.com/fortis/backend/internal/modules/defense_project/domain"
)

func budgetFromSnapshot(project *projectDomain.DefenseProject) (*domain.BudgetConfig, domain.CalculationIdentity, error) {
	var snapshot struct {
		Budget json.RawMessage   `json:"budget"`
		Inputs map[string]string `json:"inputDataVersions"`
	}
	if project.Snapshot() != "" {
		if err := json.Unmarshal([]byte(project.Snapshot()), &snapshot); err != nil {
			return nil, domain.CalculationIdentity{}, err
		}
	}
	config, err := document.ReadBudget(project.ProjectID(), string(snapshot.Budget))
	identity := domain.NewCalculationIdentity(project.ProjectID(), project.Version(), project.CostCalculationVersion(), project.SnapshotDigest(), snapshot.Inputs)
	return config, identity, err
}

func projectBudget(project *projectDomain.DefenseProject) (*domain.BudgetProjection, error) {
	config, identity, err := budgetFromSnapshot(project)
	if err != nil {
		return nil, err
	}
	cost, err := document.Calculate(project.Document(), identity)
	if err != nil {
		return nil, err
	}
	return domain.NewBudgetProjection(config, cost), nil
}

func (s *BudgetService) ProjectBudget(ctx context.Context, actorID, projectID string, version *int) (*domain.BudgetProjection, error) {
	project, err := s.projects.GetRevision(ctx, actorID, projectID, version)
	if err != nil {
		return nil, err
	}
	return projectBudget(project)
}

func (s *BudgetService) UpdateBudget(ctx context.Context, actorID, projectID string, command UpdateBudgetCommand) (*projectDomain.DefenseProject, *domain.BudgetProjection, error) {
	if _, err := s.projects.GetProject(ctx, actorID, projectID); err != nil {
		return nil, nil, err
	}
	config, err := domain.NewBudgetConfigMinor(projectID, command.Mode, command.LimitMinor)
	if err != nil {
		return nil, nil, err
	}
	raw, err := document.WriteBudget(config)
	if err != nil {
		return nil, nil, err
	}
	project, err := s.projects.UpdateBudget(ctx, actorID, projectID, command.Version, raw)
	if err != nil {
		return nil, nil, err
	}
	projection, err := projectBudget(project)
	return project, projection, err
}

func (s *BudgetService) CheckCandidate(ctx context.Context, actorID, projectID string, command CheckBudgetCommand) (*domain.CandidateBudgetCheck, error) {
	current, err := s.projects.GetRevision(ctx, actorID, projectID, nil)
	if err != nil {
		return nil, err
	}
	if command.Version <= 0 {
		return nil, projectDomain.ErrVersionRequired
	}
	if current.Version() != command.Version {
		return nil, projectDomain.ErrVersionConflict
	}
	var candidate struct {
		SchemaVersion int     `json:"schemaVersion"`
		EnterpriseID  *string `json:"enterpriseId"`
	}
	if err := json.Unmarshal([]byte(command.ProjectJSON), &candidate); err != nil || candidate.SchemaVersion != projectDomain.SchemaVersion {
		return nil, projectDomain.ErrInvalidProjectData
	}
	if candidate.EnterpriseID != nil && *candidate.EnterpriseID != current.EnterpriseID() {
		return nil, projectDomain.ErrProjectOwnershipImmutable
	}
	config, identity, err := budgetFromSnapshot(current)
	if err != nil {
		return nil, err
	}
	projection, err := document.Calculate(command.ProjectJSON, identity)
	if err != nil {
		return nil, err
	}
	before, _ := document.Calculate(current.Document(), identity)
	return domain.CheckCandidateBudget(config, projection, before), nil
}
