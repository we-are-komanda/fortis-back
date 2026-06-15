package ui

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/fortis/backend/internal/auth"
	"github.com/fortis/backend/internal/modules/user/domain"
)

// mockUserService — мок сервиса для тестирования контроллера.
type mockUserService struct {
	registerUser  *domain.User
	registerToken string
	registerErr   error
	loginUser     *domain.User
	loginToken    string
	loginErr      error
	profileUser   *domain.User
	profileErr    error
	validateErr   error
}

func (m *mockUserService) Register(ctx context.Context, email, password, name string) (*domain.User, string, error) {
	return m.registerUser, m.registerToken, m.registerErr
}

func (m *mockUserService) Login(ctx context.Context, email, password string) (*domain.User, string, error) {
	return m.loginUser, m.loginToken, m.loginErr
}

func (m *mockUserService) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	return m.profileUser, m.profileErr
}

func (m *mockUserService) ValidateToken(tokenString string) (*auth.Claims, error) {
	if m.validateErr != nil {
		return nil, m.validateErr
	}
	return &auth.Claims{UserID: "test-user-id", Email: "test@example.com"}, nil
}

func validUser() *domain.User {
	now := time.Now().UTC()
	u, _ := domain.NewUser(
		"550e8400-e29b-41d4-a716-446655440000",
		"user@example.com",
		"$2a$10$hash",
		"Тестовый пользователь",
		now,
		now,
	)
	return u
}

func TestRegister_Success(t *testing.T) {
	svc := &mockUserService{
		registerUser:  validUser(),
		registerToken: "test-jwt-token",
	}
	ctrl := NewUserController(svc)

	reqBody := `{"email":"user@example.com","password":"password123","name":"Тестовый пользователь"}`
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetBody([]byte(reqBody))

	ctrl.Register(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected status 200, got %d", ctx.Response.StatusCode())
	}

	var resp AuthResponse
	if err := json.Unmarshal(ctx.Response.Body(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Token != "test-jwt-token" {
		t.Errorf("expected token 'test-jwt-token', got %q", resp.Token)
	}
	if resp.User.Email != "user@example.com" {
		t.Errorf("expected email 'user@example.com', got %q", resp.User.Email)
	}
}

func TestRegister_EmptyBody(t *testing.T) {
	ctrl := NewUserController(&mockUserService{})

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetBody([]byte(`{}`))

	ctrl.Register(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status 400, got %d", ctx.Response.StatusCode())
	}
}

func TestRegister_BadJSON(t *testing.T) {
	ctrl := NewUserController(&mockUserService{})

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetBody([]byte(`not json`))

	ctrl.Register(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status 400, got %d", ctx.Response.StatusCode())
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc := &mockUserService{
		registerErr: domain.ErrEmailAlreadyExists,
	}
	ctrl := NewUserController(svc)

	reqBody := `{"email":"existing@example.com","password":"password123","name":"User"}`
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetBody([]byte(reqBody))

	ctrl.Register(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusConflict {
		t.Errorf("expected status 409, got %d", ctx.Response.StatusCode())
	}
}

func TestLogin_Success(t *testing.T) {
	svc := &mockUserService{
		loginUser:  validUser(),
		loginToken: "test-jwt-token",
	}
	ctrl := NewUserController(svc)

	reqBody := `{"email":"user@example.com","password":"password123"}`
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetBody([]byte(reqBody))

	ctrl.Login(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected status 200, got %d", ctx.Response.StatusCode())
	}

	var resp AuthResponse
	if err := json.Unmarshal(ctx.Response.Body(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Token != "test-jwt-token" {
		t.Errorf("expected token 'test-jwt-token', got %q", resp.Token)
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	svc := &mockUserService{
		loginErr: domain.ErrInvalidCredentials,
	}
	ctrl := NewUserController(svc)

	reqBody := `{"email":"user@example.com","password":"wrong"}`
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetBody([]byte(reqBody))

	ctrl.Login(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", ctx.Response.StatusCode())
	}
}

func TestMe_Success(t *testing.T) {
	svc := &mockUserService{
		profileUser: validUser(),
	}
	ctrl := NewUserController(svc)

	ctx := &fasthttp.RequestCtx{}
	ctx.SetUserValue("userID", "test-user-id")

	ctrl.Me(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected status 200, got %d", ctx.Response.StatusCode())
	}

	var resp UserDTO
	if err := json.Unmarshal(ctx.Response.Body(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.ID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected user ID, got %q", resp.ID)
	}
}

func TestMe_Unauthenticated(t *testing.T) {
	ctrl := NewUserController(&mockUserService{})

	ctx := &fasthttp.RequestCtx{}

	ctrl.Me(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", ctx.Response.StatusCode())
	}
}

func TestMe_NotFound(t *testing.T) {
	svc := &mockUserService{
		profileErr: domain.ErrUserNotFound,
	}
	ctrl := NewUserController(svc)

	ctx := &fasthttp.RequestCtx{}
	ctx.SetUserValue("userID", "nonexistent-id")

	ctrl.Me(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusNotFound {
		t.Errorf("expected status 404, got %d", ctx.Response.StatusCode())
	}
}

func TestValidateToken_Success(t *testing.T) {
	ctrl := NewUserController(&mockUserService{})

	ctx := &fasthttp.RequestCtx{}
	ctx.QueryArgs().Set("token", "valid-token")

	ctrl.ValidateToken(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected status 200, got %d", ctx.Response.StatusCode())
	}

	var resp TokenValidateResponse
	if err := json.Unmarshal(ctx.Response.Body(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !resp.Valid {
		t.Errorf("expected valid=true")
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	svc := &mockUserService{
		validateErr: errors.New("invalid token"),
	}
	ctrl := NewUserController(svc)

	ctx := &fasthttp.RequestCtx{}
	ctx.QueryArgs().Set("token", "invalid-token")

	ctrl.ValidateToken(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusForbidden {
		t.Errorf("expected status 403, got %d", ctx.Response.StatusCode())
	}
}

func TestValidateToken_NoToken(t *testing.T) {
	ctrl := NewUserController(&mockUserService{})

	ctx := &fasthttp.RequestCtx{}

	ctrl.ValidateToken(ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusForbidden {
		t.Errorf("expected status 403, got %d", ctx.Response.StatusCode())
	}
}
