package domain

// BudgetCheckInput — входные данные для проверки добавления средства в бюджет.
type BudgetCheckInput struct {
	assetID   string
	quantity  int
	echelonID string
}

// NewBudgetCheckInput создаёт новые входные данные для проверки бюджета.
func NewBudgetCheckInput(assetID string, quantity int, echelonID string) BudgetCheckInput {
	return BudgetCheckInput{
		assetID:   assetID,
		quantity:  quantity,
		echelonID: echelonID,
	}
}

// AssetID возвращает ID средства защиты.
func (i BudgetCheckInput) AssetID() string { return i.assetID }

// Quantity возвращает количество единиц.
func (i BudgetCheckInput) Quantity() int { return i.quantity }

// EchelonID возвращает ID эшелона (слоя).
func (i BudgetCheckInput) EchelonID() string { return i.echelonID }

// BudgetCheckResult — результат проверки добавления средства в бюджет.
type BudgetCheckResult struct {
	fits           bool
	remainingMln   float64
	requiredMln    float64
	budgetMode     BudgetMode
}

// NewBudgetCheckResult создаёт новый результат проверки бюджета.
func NewBudgetCheckResult(fits bool, remainingMln, requiredMln float64, mode BudgetMode) BudgetCheckResult {
	return BudgetCheckResult{
		fits:         fits,
		remainingMln: remainingMln,
		requiredMln:  requiredMln,
		budgetMode:   mode,
	}
}

// Fits возвращает true, если добавление помещается в остаток бюджета.
func (r BudgetCheckResult) Fits() bool { return r.fits }

// RemainingMln возвращает остаток бюджета в млн ₽.
func (r BudgetCheckResult) RemainingMln() float64 { return r.remainingMln }

// RequiredMln возвращает требуемую сумму для добавления в млн ₽.
func (r BudgetCheckResult) RequiredMln() float64 { return r.requiredMln }

// BudgetMode возвращает режим бюджета.
func (r BudgetCheckResult) BudgetMode() BudgetMode { return r.budgetMode }
