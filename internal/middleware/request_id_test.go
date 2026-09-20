package middleware

import (
	"github.com/fortis/backend/internal/audit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"testing"
)

func TestHTTPResponseAssignsServerRequestIDForAudit(t *testing.T) {
	var ctx fasthttp.RequestCtx
	ctx.Request.Header.Set("X-Request-ID", "untrusted-user-payload")
	var eventID string
	NewHttpResponse().Process(func(c *fasthttp.RequestCtx) { eventID = audit.RequestID(c) })(&ctx)
	_, err := uuid.Parse(eventID)
	require.NoError(t, err)
	require.Equal(t, eventID, string(ctx.Response.Header.Peek("X-Request-ID")))
}
