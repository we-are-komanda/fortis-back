package application

import "github.com/fortis/backend/internal/modules/budget/domain"

type UpdateBudgetCommand struct {
	Version    int
	Mode       domain.BudgetMode
	LimitMinor *string
}

type CheckBudgetCommand struct {
	Version     int
	ProjectJSON string
}
