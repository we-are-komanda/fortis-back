package middleware

import (
	"github.com/fortis/backend/internal/audit"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"strings"
)

type HttpResponse struct {
}

func NewHttpResponse() *HttpResponse {
	return &HttpResponse{}
}

func (m *HttpResponse) Process(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(c *fasthttp.RequestCtx) {
		requestID := uuid.NewString()
		c.SetUserValue(audit.RequestIDKey, requestID)
		c.Response.Header.Set("X-Request-ID", requestID)
		next(c)

		if strings.Contains(c.URI().String(), "/api/v") {
			c.Response.Header.SetContentType("application/json")
		} else {
			c.SetContentType("text/html; charset=utf-8")
		}
	}
}
