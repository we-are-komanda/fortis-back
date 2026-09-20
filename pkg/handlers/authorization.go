package handlers

import (
	"errors"

	"github.com/fortis/backend/internal/auth"
	"github.com/valyala/fasthttp"
)

// AuthorizationError uses the same public response for missing and inaccessible resources.
// Storage/checker failures do not match these sentinels and retain the controller's 500 path.
func AuthorizationError(ctx *fasthttp.RequestCtx, err error) bool {
	switch {
	case errors.Is(err, auth.ErrIdentityRequired):
		ErrorHandler(ctx, "unauthorized", "unauthorized", &ResponseBody{}, fasthttp.StatusUnauthorized)
	case errors.Is(err, auth.ErrNotFound):
		ErrorHandler(ctx, "not_found", "resource not found", &ResponseBody{}, fasthttp.StatusNotFound)
	case errors.Is(err, auth.ErrForbidden):
		ErrorHandler(ctx, "forbidden", "operation forbidden", &ResponseBody{}, fasthttp.StatusForbidden)
	default:
		return false
	}
	return true
}

func ActorID(ctx *fasthttp.RequestCtx) string {
	id, _ := ctx.UserValue("userID").(string)
	return id
}
