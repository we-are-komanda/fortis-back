package ui

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/valyala/fasthttp"

	"github.com/fortis/backend/internal/auth"
	"github.com/fortis/backend/internal/modules/user/domain"
	"github.com/fortis/backend/pkg/handlers"
)

// UserServiceInterface — интерфейс сервиса пользователей.
type UserServiceInterface interface {
	Register(ctx context.Context, email, password, name string) (*domain.User, string, error)
	Login(ctx context.Context, email, password string) (*domain.User, string, error)
	GetProfile(ctx context.Context, userID string) (*domain.User, error)
	ValidateToken(tokenString string) (*auth.Claims, error)
}

// UserController — контроллер для аутентификации и управления пользователями.
type UserController struct {
	service UserServiceInterface
}

// NewUserController создаёт новый контроллер пользователей.
func NewUserController(service UserServiceInterface) *UserController {
	return &UserController{
		service: service,
	}
}

// swagger:route POST /api/v1/auth/register auth register
// Регистрация нового пользователя
//
// Создаёт нового пользователя с указанными email, паролем и именем.
// Возвращает JWT-токен и данные пользователя.
//
// Consumes:
//   - application/json
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: AuthResponse
//	400: description: Bad Request — неверные данные
//	409: description: Conflict — email уже занят
//	500: description: Internal Server Error
func (c *UserController) Register(ctx *fasthttp.RequestCtx) {
	var req RegisterRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		handlers.ErrorHandler(ctx, "parse_error", "invalid request body", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		handlers.ErrorHandler(ctx, "validation_error", "email, password and name are required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	user, token, err := c.service.Register(ctx, req.Email, req.Password, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidEmail):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		case errors.Is(err, domain.ErrInvalidUserName):
			handlers.ErrorHandler(ctx, "validation_error", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		case errors.Is(err, domain.ErrEmailAlreadyExists):
			handlers.ErrorHandler(ctx, "conflict", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusConflict)
		default:
			handlers.ErrorHandler(ctx, "internal_error", "failed to register", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		}
		return
	}

	resp := authResponse(user, token)
	respJSON, err := json.Marshal(resp)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to marshal response", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	// Устанавливаем cookie с токеном
	cookie := fasthttp.AcquireCookie()
	defer fasthttp.ReleaseCookie(cookie)
	cookie.SetKey("access-token")
	cookie.SetValue(token)
	cookie.SetPath("/")
	cookie.SetHTTPOnly(true)
	ctx.Response.Header.SetCookie(cookie)

	ctx.SetBody(respJSON)
}

// swagger:route POST /api/v1/auth/login auth login
// Вход в систему
//
// Аутентифицирует пользователя по email и паролю.
// Возвращает JWT-токен и данные пользователя.
//
// Consumes:
//   - application/json
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: AuthResponse
//	400: description: Bad Request — неверные данные
//	401: description: Unauthorized — неверные учётные данные
//	500: description: Internal Server Error
func (c *UserController) Login(ctx *fasthttp.RequestCtx) {
	var req LoginRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		handlers.ErrorHandler(ctx, "parse_error", "invalid request body", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		handlers.ErrorHandler(ctx, "validation_error", "email and password are required", &handlers.ResponseBody{}, fasthttp.StatusBadRequest)
		return
	}

	user, token, err := c.service.Login(ctx, req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCredentials):
			handlers.ErrorHandler(ctx, "unauthorized", err.Error(), &handlers.ResponseBody{}, fasthttp.StatusUnauthorized)
		default:
			handlers.ErrorHandler(ctx, "internal_error", "failed to login", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		}
		return
	}

	resp := authResponse(user, token)
	respJSON, err := json.Marshal(resp)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to marshal response", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	// Устанавливаем cookie с токеном
	cookie := fasthttp.AcquireCookie()
	defer fasthttp.ReleaseCookie(cookie)
	cookie.SetKey("access-token")
	cookie.SetValue(token)
	cookie.SetPath("/")
	cookie.SetHTTPOnly(true)
	ctx.Response.Header.SetCookie(cookie)

	ctx.SetBody(respJSON)
}

// swagger:route GET /api/v1/auth/me auth getProfile
// Получение профиля текущего пользователя
//
// Возвращает данные пользователя на основе JWT-токена из cookie.
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: UserDTO
//	401: description: Unauthorized — не аутентифицирован
//	500: description: Internal Server Error
func (c *UserController) Me(ctx *fasthttp.RequestCtx) {
	userID := getUserIDFromCtx(ctx)
	if userID == "" {
		handlers.ErrorHandler(ctx, "unauthorized", "not authenticated", &handlers.ResponseBody{}, fasthttp.StatusUnauthorized)
		return
	}

	user, err := c.service.GetProfile(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			handlers.ErrorHandler(ctx, "not_found", "user not found", &handlers.ResponseBody{}, fasthttp.StatusNotFound)
		default:
			handlers.ErrorHandler(ctx, "internal_error", "failed to get profile", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		}
		return
	}

	resp := userToDTO(user)
	respJSON, err := json.Marshal(resp)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to marshal response", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetBody(respJSON)
}

// swagger:route GET /api/v1/token_validate auth validateToken
// Валидация JWT-токена
//
// Проверяет валидность JWT-токена. Используется для backward-compat с Access middleware.
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: TokenValidateResponse
//	403: description: Forbidden — токен невалиден
//	500: description: Internal Server Error
func (c *UserController) ValidateToken(ctx *fasthttp.RequestCtx) {
	tokenString := string(ctx.QueryArgs().Peek("token"))
	if tokenString == "" {
		// Пробуем из cookie
		tokenString = string(ctx.Request.Header.Cookie("access-token"))
	}
	if tokenString == "" {
		handlers.ErrorHandler(ctx, "forbidden", "token required", &handlers.ResponseBody{}, fasthttp.StatusForbidden)
		return
	}

	_, err := c.service.ValidateToken(tokenString)
	if err != nil {
		handlers.ErrorHandler(ctx, "forbidden", "invalid token", &handlers.ResponseBody{}, fasthttp.StatusForbidden)
		return
	}

	resp := TokenValidateResponse{Valid: true}
	respJSON, err := json.Marshal(resp)
	if err != nil {
		handlers.ErrorHandler(ctx, "internal_error", "failed to marshal response", &handlers.ResponseBody{}, fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetBody(respJSON)
}

// getUserIDFromCtx извлекает userId из контекста запроса.
func getUserIDFromCtx(ctx *fasthttp.RequestCtx) string {
	if userID, ok := ctx.UserValue("userID").(string); ok {
		return userID
	}
	return ""
}
