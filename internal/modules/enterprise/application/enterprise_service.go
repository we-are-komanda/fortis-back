package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/fortis/backend/internal/auth"

	"github.com/fortis/backend/internal/modules/enterprise/domain"
)

// AccessChecker is shared by enterprise-owned application services.
type AccessChecker interface {
	CheckUserAccess(ctx context.Context, userID, enterpriseID string) error
}

// EnterpriseService — сервис для управления предприятиями.
type EnterpriseService struct {
	repo domain.EnterpriseRepositoryInterface
}

// NewEnterpriseService создаёт новый сервис предприятий.
func NewEnterpriseService(repo domain.EnterpriseRepositoryInterface) *EnterpriseService {
	return &EnterpriseService{
		repo: repo,
	}
}

// Create is closed until a separate trusted provisioning path is approved.
func (s *EnterpriseService) Create(ctx context.Context, userID, name, address string, status domain.EnterpriseStatus, latitude, longitude float64) (*domain.Enterprise, error) {
	if err := auth.RequireIdentity(userID); err != nil {
		return nil, err
	}
	// ponytail: enterprise provisioning is operator-only; add a separate trusted use case if self-service is approved.
	return nil, auth.ErrForbidden
}

// Get возвращает предприятие по ID.
func (s *EnterpriseService) Get(ctx context.Context, userID, id string) (*domain.Enterprise, error) {
	if err := s.CheckUserAccess(ctx, userID, id); err != nil {
		return nil, err
	}
	enterprise, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return enterprise, nil
}

// Update обновляет существующее предприятие.
func (s *EnterpriseService) Update(
	ctx context.Context,
	userID, id, name, address string,
	status domain.EnterpriseStatus,
	latitude, longitude float64,
) (*domain.Enterprise, error) {
	enterprise, err := s.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		if err := enterprise.UpdateName(name); err != nil {
			return nil, err
		}
	}

	enterprise.UpdateAddress(address)

	if status != "" {
		if err := enterprise.UpdateStatus(status); err != nil {
			return nil, err
		}
	}

	if err := enterprise.UpdateCoordinates(latitude, longitude); err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, enterprise); err != nil {
		return nil, fmt.Errorf("update enterprise: %w", err)
	}

	return enterprise, nil
}

// Delete удаляет предприятие по ID.
func (s *EnterpriseService) Delete(ctx context.Context, userID, id string) error {
	if err := s.CheckUserAccess(ctx, userID, id); err != nil {
		return err
	}
	return auth.ErrForbidden
}

// ListByUser возвращает предприятия, доступные пользователю, с пагинацией.
func (s *EnterpriseService) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Enterprise, int64, error) {
	if err := auth.RequireIdentity(userID); err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	enterprises, total, err := s.repo.FindAllByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list enterprises by user: %w", err)
	}

	return enterprises, total, nil
}

// CheckUserAccess проверяет, имеет ли пользователь доступ к предприятию.
// Missing enterprises and absent membership share ErrNotFound; storage failures remain errors.
func (s *EnterpriseService) CheckUserAccess(ctx context.Context, userID, enterpriseID string) error {
	if err := auth.RequireIdentity(userID); err != nil {
		return err
	}
	if enterpriseID == "" {
		return auth.ErrNotFound
	}

	// Проверяем, что предприятие существует
	if _, err := s.repo.FindByID(ctx, enterpriseID); err != nil {
		if errors.Is(err, domain.ErrEnterpriseNotFound) {
			return auth.ErrNotFound
		}
		return err
	}

	// Проверяем доступ через user_enterprises
	hasAccess, err := s.repo.CheckUserEnterpriseAccess(ctx, userID, enterpriseID)
	if err != nil {
		return fmt.Errorf("check access: %w", err)
	}

	if !hasAccess {
		return auth.ErrNotFound
	}

	return nil
}

// AddUser is operator-only; an ordinary authenticated actor cannot grant membership.
func (s *EnterpriseService) AddUser(ctx context.Context, actorID, userID, enterpriseID string) error {
	if err := auth.RequireIdentity(actorID); err != nil {
		return err
	}
	return auth.ErrForbidden
}

// RemoveUser is closed to ordinary user sessions, like AddUser.
func (s *EnterpriseService) RemoveUser(ctx context.Context, actorID, userID, enterpriseID string) error {
	if err := auth.RequireIdentity(actorID); err != nil {
		return err
	}
	return auth.ErrForbidden
}
