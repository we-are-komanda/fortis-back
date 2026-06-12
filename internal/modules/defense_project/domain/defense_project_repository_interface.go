package domain

import "context"

// DefenseProjectRepositoryInterface — интерфейс репозитория для DefenseProject.
type DefenseProjectRepositoryInterface interface {
	// Save сохраняет проект (создаёт или обновляет по ProjectID).
	Save(ctx context.Context, project *DefenseProject) error

	// FindByID находит проект по ID.
	FindByID(ctx context.Context, id string) (*DefenseProject, error)

	// Delete удаляет проект по ID.
	Delete(ctx context.Context, id string) error
}
