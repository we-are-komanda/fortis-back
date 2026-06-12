package domain

import "context"

// DefenseProjectRepositoryInterface — интерфейс репозитория для DefenseProject.
type DefenseProjectRepositoryInterface interface {
	// Save сохраняет проект (создаёт или обновляет по ProjectID).
	Save(ctx context.Context, project *DefenseProject) error

	// FindByID находит проект по ID.
	FindByID(ctx context.Context, id string) (*DefenseProject, error)

	// FindAll возвращает список проектов с пагинацией и общим количеством.
	FindAll(ctx context.Context, limit, offset int) ([]*DefenseProject, int64, error)

	// FindAllByEnterprise возвращает проекты по enterprise ID с пагинацией.
	FindAllByEnterprise(ctx context.Context, enterpriseID string, limit, offset int) ([]*DefenseProject, int64, error)

	// Delete удаляет проект по ID.
	Delete(ctx context.Context, id string) error
}
