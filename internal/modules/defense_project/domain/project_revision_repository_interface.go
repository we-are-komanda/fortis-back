package domain

import "context"

// ProjectWrite is the persistence contract for one atomic saved state.
type ProjectWrite struct {
	ActorID        string
	Create         bool
	Operation      string
	IdempotencyKey string
	PayloadDigest  string
	BudgetJSON     *string
	BudgetSource   *ProjectRef
}

// ProjectRef identifies an immutable server revision, not embedded import metadata.
type ProjectRef struct {
	ProjectID      string
	ProjectVersion int
}

type ProjectRevisionRepository interface {
	Commit(context.Context, *DefenseProject, ProjectWrite) (*DefenseProject, error)
	FindRevision(context.Context, string, int) (*DefenseProject, error)
	DeleteAuthorized(context.Context, string, string) error
}
