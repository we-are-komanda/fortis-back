package ui

// RegisterRequest DTO запроса на регистрацию.
// swagger:parameters RegisterRequest
type RegisterRequest struct {
	// Email пользователя
	// Required: true
	// Example: user@example.com
	Email string `json:"email"`
	// Пароль
	// Required: true
	// Example: securePassword123
	Password string `json:"password"`
	// Имя пользователя
	// Required: true
	// Example: Иван Петров
	Name string `json:"name"`
}

// LoginRequest DTO запроса на вход.
// swagger:parameters LoginRequest
type LoginRequest struct {
	// Email пользователя
	// Required: true
	// Example: user@example.com
	Email string `json:"email"`
	// Пароль
	// Required: true
	// Example: securePassword123
	Password string `json:"password"`
}

// UserDTO DTO ответа с данными пользователя.
// swagger:response UserDTO
type UserDTO struct {
	// ID пользователя
	// Example: 550e8400-e29b-41d4-a716-446655440000
	ID string `json:"id"`
	// Email пользователя
	// Example: user@example.com
	Email string `json:"email"`
	// Имя пользователя
	// Example: Иван Петров
	Name string `json:"name"`
	// Дата создания (RFC3339)
	// Example: 2026-06-12T14:00:00Z
	CreatedAt string `json:"createdAt"`
	// Дата обновления (RFC3339)
	// Example: 2026-06-12T14:00:00Z
	UpdatedAt string `json:"updatedAt"`
}

// AuthResponse DTO ответа с токеном и данными пользователя.
// swagger:response AuthResponse
type AuthResponse struct {
	// JWT-токен
	// Example: eyJhbGciOiJIUzI1NiIs...
	Token string  `json:"token"`
	// Данные пользователя
	User  UserDTO `json:"user"`
}

// TokenValidateResponse DTO ответа валидации токена.
// swagger:response TokenValidateResponse
type TokenValidateResponse struct {
	// Результат валидации
	// Example: true
	Valid bool `json:"valid"`
}
