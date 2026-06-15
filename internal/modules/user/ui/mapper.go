package ui

import (
	"time"

	"github.com/fortis/backend/internal/modules/user/domain"
)

// userToDTO преобразует доменный User в UserDTO.
func userToDTO(u *domain.User) UserDTO {
	return UserDTO{
		ID:        u.ID(),
		Email:     u.Email(),
		Name:      u.Name(),
		CreatedAt: u.CreatedAt().UTC().Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt().UTC().Format(time.RFC3339),
	}
}

// authResponse создаёт AuthResponse из пользователя и токена.
func authResponse(u *domain.User, token string) AuthResponse {
	return AuthResponse{
		Token: token,
		User:  userToDTO(u),
	}
}
