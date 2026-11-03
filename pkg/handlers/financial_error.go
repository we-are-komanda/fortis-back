package handlers

import (
	"encoding/json"
	"errors"

	"github.com/fortis/backend/internal/audit"
	"github.com/fortis/backend/internal/modules/budget/domain"
	"github.com/valyala/fasthttp"
)

// FinancialError keeps the same sanitized contract on every project/budget write.
func FinancialError(ctx *fasthttp.RequestCtx, err error) bool {
	var violation *domain.BudgetViolation
	var invalid *domain.CostValidationError
	status, code, message := 400, "validation_error", "Invalid budget configuration"
	var details map[string]any
	switch {
	case errors.As(err, &violation):
		status, code, message = 422, violation.Code(), "Candidate does not satisfy the project budget"
		projection := violation.Projection()
		field := "budgetLimitMinor"
		if code == "price_unknown" {
			field = "placedObjects"
		}
		details = map[string]any{"field": field, "objectIds": violation.ObjectIDs(), "budgetLimitMinor": projection.Config().BudgetLimitMinor(), "projectedSpentMinor": projection.SpentMinor(), "projectedRemainingMinor": projection.RemainingMinor(), "currentSpentMinor": violation.CurrentSpentMinor()}
	case errors.As(err, &invalid):
		code, message = invalid.Code, invalid.Message
		details = map[string]any{"field": invalid.Field, "objectIds": invalid.ObjectIDs}
	case errors.Is(err, domain.ErrInvalidBudgetAmount), errors.Is(err, domain.ErrInvalidBudgetMode):
	default:
		return false
	}
	requestID := audit.RequestID(ctx)
	ctx.Response.Header.Set("X-Request-ID", requestID)
	body, _ := json.Marshal(map[string]any{"error": map[string]any{"code": code, "message": message, "requestId": requestID, "retryable": false, "details": details}})
	ctx.SetStatusCode(status)
	ctx.SetBody(body)
	return true
}
