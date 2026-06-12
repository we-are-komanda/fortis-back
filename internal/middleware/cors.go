package middleware

import (
	"github.com/valyala/fasthttp"
	"github.com/fortis/backend/internal/config"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type Cors struct {
	cfg config.Cors
}

func NewCors(cfg config.Cors) *Cors {
	return &Cors{
		cfg: cfg,
	}
}

const (
	HeaderVary      = "Vary"
	HeaderXRealHost = "X-Real-Host"
)

var simpleHeaders = []string{
	"accept",
	"accept-language",
	"content-language",
	"origin",
}

//go:cover off
func (middleware *Cors) Process(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		if !middleware.cfg.Enable {
			next(ctx)
			return
		}

		origin := string(ctx.Request.Header.Peek(fasthttp.HeaderOrigin))

		if origin == "" || origin == getSchemeAndHost(ctx) {
			next(ctx)
			return
		}

		if middleware.checkOrigin(origin, middleware.cfg.AllowOrigins) {
			middleware.setCorsHeaders(ctx, origin)
		}

		if ctx.IsOptions() && string(ctx.Request.Header.Peek(fasthttp.HeaderAccessControlRequestMethod)) != "" {
			middleware.preparePreflight(ctx, origin)
			return
		}

		next(ctx)
	}

}

func (middleware *Cors) setCorsHeaders(ctx *fasthttp.RequestCtx, origin string) {
	if middleware.cfg.AllowCredentials {
		ctx.Response.Header.Set(fasthttp.HeaderAccessControlAllowCredentials, "true")
	}

	if len(middleware.cfg.ExposeHeaders) > 0 {
		ctx.Response.Header.Set(fasthttp.HeaderAccessControlExposeHeaders, strings.ToLower(strings.Join(middleware.cfg.ExposeHeaders, ", ")))
	}

	ctx.Response.Header.Set(fasthttp.HeaderAccessControlAllowOrigin, origin)
}

func (middleware *Cors) preparePreflight(ctx *fasthttp.RequestCtx, origin string) {
	middleware.setCorsHeaders(ctx, origin)
	ctx.Response.Header.Set(HeaderVary, fasthttp.HeaderOrigin)

	if !middleware.checkOrigin(origin, middleware.cfg.AllowOrigins) {
		ctx.Response.Header.Set(fasthttp.HeaderAccessControlAllowOrigin, "null")
	}

	if len(middleware.cfg.AllowMethods) > 0 {
		ctx.Response.Header.Set(fasthttp.HeaderAccessControlAllowMethods, strings.Join(middleware.cfg.AllowMethods, ", "))
	}
	if len(middleware.cfg.AllowHeaders) > 0 {
		ctx.Response.Header.Set(fasthttp.HeaderAccessControlAllowHeaders, strings.ToLower(strings.Join(middleware.cfg.AllowHeaders, ", ")))
	}
	if middleware.cfg.MaxAge > 0 {
		ctx.Response.Header.Set(fasthttp.HeaderAccessControlMaxAge, strconv.Itoa(middleware.cfg.MaxAge))
	}

	requestMethod := string(ctx.Request.Header.Peek(fasthttp.HeaderAccessControlRequestMethod))
	if !sliceContains(middleware.cfg.AllowMethods, strings.ToUpper(requestMethod)) {
		ctx.Response.SetStatusCode(http.StatusMethodNotAllowed)
		return
	}

	if !sliceContains(middleware.cfg.AllowMethods, requestMethod) {
		allowedMethods := middleware.cfg.AllowMethods
		allowedMethods = append(allowedMethods, requestMethod)
		ctx.Response.Header.Set(fasthttp.HeaderAccessControlAllowMethods, strings.Join(allowedMethods, ", "))
	}

	requestHeaders := string(ctx.Request.Header.Peek(fasthttp.HeaderAccessControlRequestHeaders))
	if len(middleware.cfg.AllowHeaders) > 0 && requestHeaders != "" {
		requestHeaders = strings.ToLower(requestHeaders)
		for _, requestHeader := range strings.Split(requestHeaders, ",") {
			requestHeader = strings.TrimSpace(requestHeader)
			if sliceContains(simpleHeaders, requestHeader) {
				continue
			}

			var allowHeaders []string
			for _, allowHeader := range middleware.cfg.AllowHeaders {
				allowHeaders = append(allowHeaders, strings.ToLower(allowHeader))
			}

			if !sliceContains(allowHeaders, requestHeader) {
				ctx.Response.SetStatusCode(http.StatusBadRequest)
				ctx.Response.SetBody([]byte("Unauthorized header " + requestHeader))
				return
			}
		}
	}

	ctx.Response.SetStatusCode(http.StatusNoContent)
}

func (middleware *Cors) checkOrigin(origin string, allowOrigins []string) bool {
	if len(allowOrigins) > 0 && allowOrigins[0] == "*" {
		return true
	}

	if sliceContains(allowOrigins, origin) {
		return true
	}

	if middleware.cfg.OriginRegex {
		for _, pattern := range allowOrigins {
			matched, err := regexp.MatchString(pattern, origin)
			if err != nil {
				slog.Error("Origin pattern is invalid: " + err.Error())
				continue
			}
			if matched {
				return true
			}
		}
	}

	return false
}

func getSchemeAndHost(ctx *fasthttp.RequestCtx) string {
	host := string(ctx.Host())
	xRealHost := string(ctx.Request.Header.Peek(HeaderXRealHost))
	if xRealHost != "" {
		host = xRealHost
	}
	return string(ctx.URI().Scheme()) + "://" + host
}

func sliceContains(s []string, e string) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
