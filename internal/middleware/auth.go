package middleware

import (
	"regexp"

	"github.com/valyala/fasthttp"

	"github.com/fortis/backend/internal/auth"
)

// AuthRequired — middleware для проверки JWT-аутентификации.
type AuthRequired struct {
	jwtSecret string
	whitelist []string
}

// NewAuthRequired создаёт новый middleware аутентификации.
func NewAuthRequired(jwtSecret string, whitelist []string) *AuthRequired {
	return &AuthRequired{
		jwtSecret: jwtSecret,
		whitelist: whitelist,
	}
}

// Process возвращает fasthttp.RequestHandler с проверкой JWT.
func (m *AuthRequired) Process(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(c *fasthttp.RequestCtx) {
		// Проверяем whitelist
		path := string(c.URI().Path())
		for _, expr := range m.whitelist {
			if matched, _ := regexp.MatchString(expr, path); matched {
				next(c)
				return
			}
		}

		// Извлекаем токен: сначала из Authorization header, потом из cookie
		tokenString := string(c.Request.Header.Peek("Authorization"))
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		} else {
			tokenString = string(c.Request.Header.Cookie("access-token"))
		}

		if tokenString == "" {
			c.SetStatusCode(fasthttp.StatusUnauthorized)
			c.SetBodyString(`{"status":"error","errors":"unauthorized"}`)
			return
		}

		claims, err := auth.ValidateToken(tokenString, m.jwtSecret)
		if err != nil {
			c.SetStatusCode(fasthttp.StatusUnauthorized)
			c.SetBodyString(`{"status":"error","errors":"invalid token"}`)
			return
		}

		// Сохраняем userID в контексте запроса
		c.SetUserValue("userID", claims.UserID)

		next(c)
	}
}
