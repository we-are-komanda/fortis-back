package middleware

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/fortis/backend/internal/config"
	"log/slog"
	"net/http"
	"regexp"
)

type Access struct {
	cfg config.Access
}

func NewAccess(cfg config.Access) *Access {
	return &Access{
		cfg: cfg,
	}
}

func (middleware *Access) Process(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(c *fasthttp.RequestCtx) {
		var inWhiteList bool
		for _, expr := range middleware.cfg.WhiteList {
			if matched, _ := regexp.MatchString(expr, string(c.URI().Path())); matched {
				inWhiteList = true
				next(c)
				break
			}
		}

		if inWhiteList {
			return
		}

		token := c.Request.Header.Cookie("access-token")
		if len(token) == 0 || token == nil {
			slog.Warn("Пользователь пытается получить доступ без access-token")
			c.Redirect(fmt.Sprintf("%s?redirect_url=%s", middleware.cfg.AuthUrl, c.URI().String()),
				fasthttp.StatusFound)
			return
		}

		for i := 0; i < middleware.cfg.ReTry; i++ {
			response, err := http.Get(fmt.Sprintf("%s?token=%s", middleware.cfg.ValidityEndpoint, token))
			if err != nil {
				slog.Error(fmt.Sprintf("Ошибка выполнения запроса валидации токена: %v", err))
				continue
			}
			defer response.Body.Close()

			if response.StatusCode == http.StatusOK {
				next(c)
				break
			}

			if response.StatusCode == http.StatusForbidden {
				c.Redirect(fmt.Sprintf("%s?redirect_url=%s", middleware.cfg.AuthUrl, c.URI().String()),
					fasthttp.StatusFound)
				break
			}

			if response.StatusCode != http.StatusOK {
				slog.Error(fmt.Sprintf("Ошибка при валидации токена: статус %d", response.StatusCode))
				continue
			}
		}
		c.SetStatusCode(fasthttp.StatusInternalServerError)
	}
}
