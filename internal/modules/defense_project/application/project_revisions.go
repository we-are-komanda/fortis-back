package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/fortis/backend/internal/modules/defense_project/document"
	"github.com/fortis/backend/internal/modules/defense_project/domain"
)

func (s *DefenseProjectService) saveProject(ctx context.Context, p *domain.DefenseProject, actorID string, create bool, operation, key string, budget *string) error {
	if !create {
		p.SetUpdatedAt(time.Now().UTC())
	}
	write := domain.ProjectWrite{ActorID: actorID, Create: create, Operation: operation, IdempotencyKey: key, BudgetJSON: budget}
	if key != "" {
		if len(key) > 128 || strings.TrimSpace(key) != key || strings.IndexFunc(key, func(r rune) bool { return r < 33 || r > 126 }) >= 0 {
			return domain.ErrInvalidIdempotencyKey
		}
		body, err := json.Marshal(map[string]any{"name": p.Name(), "enterpriseId": p.EnterpriseID(), "project": json.RawMessage(p.Document())})
		if err != nil {
			return err
		}
		body, err = document.Canonical(body)
		if err != nil {
			return err
		}
		digest := sha256.Sum256(body)
		write.PayloadDigest = hex.EncodeToString(digest[:])
	}
	if repository, ok := s.repo.(domain.ProjectRevisionRepository); ok {
		saved, err := repository.Commit(ctx, p, write)
		if err == nil {
			*p = *saved
		}
		return err
	}
	// Existing unit-test/in-memory ports have no revision store. New operations
	// fail closed there; production DI always supplies ProjectRevisionRepository.
	if key != "" || budget != nil {
		return domain.ErrInvalidProjectData
	}
	return s.repo.Save(ctx, p)
}

// GetRevision always authorizes the current resource before exposing history.
func (s *DefenseProjectService) GetRevision(ctx context.Context, actorID, id string, version *int) (*domain.DefenseProject, error) {
	current, err := s.GetProject(ctx, actorID, id)
	if err != nil {
		return nil, err
	}
	wanted := current.Version()
	if version != nil {
		if *version <= 0 {
			return nil, domain.ErrInvalidProjectData
		}
		wanted = *version
	}
	if repository, ok := s.repo.(domain.ProjectRevisionRepository); ok {
		return repository.FindRevision(ctx, id, wanted)
	}
	if version != nil {
		return nil, domain.ErrRevisionNotFound
	}
	return current, nil
}

func (s *DefenseProjectService) ExportRevision(ctx context.Context, actorID, id string, version *int) (string, error) {
	p, err := s.GetRevision(ctx, actorID, id, version)
	if err != nil {
		return "", err
	}
	return serializeProject(p)
}

// UpdateBudget is the same CAS/revision transaction used by project saves.
// FRC-07 owns monetary validation; this boundary accepts only a JSON budget object.
func (s *DefenseProjectService) UpdateBudget(ctx context.Context, actorID, id string, expectedVersion int, budgetJSON string) (*domain.DefenseProject, error) {
	p, err := s.GetProject(ctx, actorID, id)
	if err != nil {
		return nil, err
	}
	if expectedVersion <= 0 {
		return nil, domain.ErrVersionRequired
	}
	if p.Version() != expectedVersion {
		return nil, domain.ErrVersionConflict
	}
	var config map[string]json.RawMessage
	if err := json.Unmarshal([]byte(budgetJSON), &config); err != nil || config == nil {
		return nil, domain.ErrInvalidProjectData
	}
	if err := s.saveProject(ctx, p, actorID, false, "budget", "", &budgetJSON); err != nil {
		return nil, err
	}
	return p, nil
}
