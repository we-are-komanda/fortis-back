package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/fortis/backend/internal/modules/enterprise/domain"
)

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

// Create создаёт новое предприятие.
func (s *EnterpriseService) Create(
	ctx context.Context,
	name, address string,
	status domain.EnterpriseStatus,
	latitude, longitude float64,
) (*domain.Enterprise, error) {
	now := time.Now().UTC()
	enterprise, err := domain.NewEnterprise(
		uuid.New().String(),
		name,
		address,
		status,
		latitude,
		longitude,
		now,
		now,
	)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, enterprise); err != nil {
		return nil, fmt.Errorf("save enterprise: %w", err)
	}

	return enterprise, nil
}

// Get возвращает предприятие по ID.
func (s *EnterpriseService) Get(ctx context.Context, id string) (*domain.Enterprise, error) {
	enterprise, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return enterprise, nil
}

// List возвращает список предприятий с пагинацией.
func (s *EnterpriseService) List(ctx context.Context, limit, offset int) ([]*domain.Enterprise, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	enterprises, total, err := s.repo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list enterprises: %w", err)
	}

	return enterprises, total, nil
}

// Update обновляет существующее предприятие.
func (s *EnterpriseService) Update(
	ctx context.Context,
	id, name, address string,
	status domain.EnterpriseStatus,
	latitude, longitude float64,
) (*domain.Enterprise, error) {
	enterprise, err := s.repo.FindByID(ctx, id)
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
func (s *EnterpriseService) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	return nil
}

// ListByUser возвращает предприятия, доступные пользователю, с пагинацией.
func (s *EnterpriseService) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Enterprise, int64, error) {
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
// Возвращает nil если доступ есть, или ErrEnterpriseAccessDenied.
func (s *EnterpriseService) CheckUserAccess(ctx context.Context, userID, enterpriseID string) error {
	if userID == "" {
		return domain.ErrEnterpriseAccessDenied
	}

	// Проверяем, что предприятие существует
	if _, err := s.repo.FindByID(ctx, enterpriseID); err != nil {
		return err
	}

	// Проверяем доступ через user_enterprises
	hasAccess, err := s.repo.CheckUserEnterpriseAccess(ctx, userID, enterpriseID)
	if err != nil {
		return fmt.Errorf("check access: %w", err)
	}

	if !hasAccess {
		return domain.ErrEnterpriseAccessDenied
	}

	return nil
}

// AddUser добавляет пользователя к предприятию.
func (s *EnterpriseService) AddUser(ctx context.Context, userID, enterpriseID string) error {
	// Проверяем, что предприятие существует
	_, err := s.repo.FindByID(ctx, enterpriseID)
	if err != nil {
		return err
	}

	return s.repo.AddUserToEnterprise(ctx, userID, enterpriseID)
}

// RemoveUser удаляет пользователя из предприятия.
func (s *EnterpriseService) RemoveUser(ctx context.Context, userID, enterpriseID string) error {
	return s.repo.RemoveUserFromEnterprise(ctx, userID, enterpriseID)
}
