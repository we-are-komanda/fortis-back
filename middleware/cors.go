package middleware

import (
	"github.com/valyala/fasthttp"
	"github.com/fortis/backend/config"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type Cors struct {
	config *config.Config
}

func NewCors(config *config.Config) *Cors {
	return &Cors{
		config: config,
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
		if !middleware.config.Cors.Enable {
			next(ctx)
			return
		}

		origin := string(ctx.Request.Header.Peek(fasthttp.HeaderOrigin))

		if origin == "" || origin == getSchemeAndHost(ctx) {
			next(ctx)
			return
		}

		if middleware.checkOrigin(origin, middleware.config.Cors.AllowOrigins) {
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
	if middleware.config.Cors.AllowCredentials {
		ctx.Response.Header.Set(fasthttp.HeaderAccessControlAllowCredentials, "true")
	}

	if len(middleware.config.Cors.ExposeHeaders) > 0 {
		ctx.Response.Header.Set(fasthttp.HeaderAccessControlExposeHeaders, strings.ToLower(strings.Join(middleware.config.Cors.ExposeHeaders, ", ")))
	}

	ctx.Response.Header.Set(fasthttp.HeaderAccessControlAllowOrigin, origin)
}

func (middleware *Cors) preparePreflight(ctx *fasthttp.RequestCtx, origin string) {
	middleware.setCorsHeaders(ctx, origin)
	ctx.Response.Header.Set(HeaderVary, fasthttp.HeaderOrigin)

	if !middleware.checkOrigin(origin, middleware.config.Cors.AllowOrigins) {
		ctx.Response.Header.Set(fasthttp.HeaderAccessControlAllowOrigin, "null")
	}

	if len(middleware.config.Cors.AllowMethods) > 0 {
		ctx.Response.Header.Set(fasthttp.HeaderAccessControlAllowMethods, strings.Join(middleware.config.Cors.AllowMethods, ", "))
	}
	if len(middleware.config.Cors.AllowHeaders) > 0 {
		ctx.Response.Header.Set(fasthttp.HeaderAccessControlAllowHeaders, strings.ToLower(strings.Join(middleware.config.Cors.AllowHeaders, ", ")))
	}
	if middleware.config.Cors.MaxAge > 0 {
		ctx.Response.Header.Set(fasthttp.HeaderAccessControlMaxAge, strconv.Itoa(middleware.config.Cors.MaxAge))
	}

	requestMethod := string(ctx.Request.Header.Peek(fasthttp.HeaderAccessControlRequestMethod))
	if !sliceContains(middleware.config.Cors.AllowMethods, strings.ToUpper(requestMethod)) {
		ctx.Response.SetStatusCode(http.StatusMethodNotAllowed)
		return
	}

	// We have to allow the header in the case-set as we received it by the client.
	// Firefox f.e. sends the LINK method as "Link", and we have to allow it like this or the browser will deny the request.
	if !sliceContains(middleware.config.Cors.AllowMethods, requestMethod) {
		allowedMethods := middleware.config.Cors.AllowMethods
		allowedMethods = append(allowedMethods, requestMethod)
		ctx.Response.Header.Set(fasthttp.HeaderAccessControlAllowMethods, strings.Join(allowedMethods, ", "))
	}

	requestHeaders := string(ctx.Request.Header.Peek(fasthttp.HeaderAccessControlRequestHeaders))
	if len(middleware.config.Cors.AllowHeaders) > 0 && requestHeaders != "" {
		requestHeaders = strings.ToLower(requestHeaders)
		for _, requestHeader := range strings.Split(requestHeaders, ",") {
			requestHeader = strings.TrimSpace(requestHeader)
			if sliceContains(simpleHeaders, requestHeader) {
				continue
			}

			var allowHeaders []string
			for _, allowHeader := range middleware.config.Cors.AllowHeaders {
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

	if middleware.config.Cors.OriginRegex {
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
