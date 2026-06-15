package ui

// ---- Budget Config DTOs ----

// BudgetConfigQuery DTO query-параметров для получения/обновления бюджета.
// swagger:parameters GetBudgetConfig UpdateBudgetConfig
type BudgetConfigQuery struct {
	// ID проекта
	// Required: true
	// In: query
	ID string `json:"id"`
}

// BudgetConfigDTO DTO конфигурации бюджета проекта.
// swagger:response BudgetConfigResponse
type BudgetConfigDTO struct {
	// Режим бюджета: "limited" | "unlimited"
	// Required: true
	// Enum: ["limited","unlimited"]
	// Example: limited
	BudgetMode string `json:"budgetMode"`
	// Сумма бюджета в млн ₽ (только для limited режима)
	// Example: 5000.0
	BudgetAmountMln float64 `json:"budgetAmountMln"`
	// ID проекта
	// Example: 550e8400-e29b-41d4-a716-446655440000
	ProjectID string `json:"projectId"`
	// Дата создания
	// Example: 2026-06-12T14:00:00.000Z
	CreatedAt string `json:"createdAt"`
	// Дата обновления
	// Example: 2026-06-12T14:00:00.000Z
	UpdatedAt string `json:"updatedAt"`
}

// UpdateBudgetRequest DTO запроса на обновление бюджета.
// swagger:parameters UpdateBudgetRequest
type UpdateBudgetRequest struct {
	// Режим бюджета: "limited" | "unlimited"
	// Required: true
	// In: body
	// Enum: ["limited","unlimited"]
	BudgetMode string `json:"budgetMode"`
	// Сумма бюджета в млн ₽ (обязательна для limited)
	// In: body
	// Example: 5000.0
	BudgetAmountMln float64 `json:"budgetAmountMln"`
}

// ---- Cost Calculation DTOs ----

// CostQuery DTO query-параметров для расчёта стоимости.
// swagger:parameters CalculateCost
type CostQuery struct {
	// ID проекта
	// Required: true
	// In: query
	ID string `json:"id"`
}

// EstimateLineDTO DTO строки расчёта стоимости.
// swagger:model EstimateLineDTO
type EstimateLineDTO struct {
	// ID размещённого объекта
	ObjectID string `json:"objectId"`
	// ID средства защиты
	AssetID string `json:"assetId"`
	// Название средства защиты
	AssetName string `json:"assetName"`
	// ID эшелона (слоя)
	EchelonID string `json:"echelonId"`
	// Название эшелона
	EchelonName string `json:"echelonName"`
	// ID типа защиты
	TypeID string `json:"typeId"`
	// Название типа защиты
	TypeName string `json:"typeName"`
	// Количество единиц
	Quantity int `json:"quantity"`
	// Цена за единицу в млн ₽
	UnitPriceMln float64 `json:"unitPriceMln"`
	// Итоговая стоимость строки в млн ₽
	LineTotalMln float64 `json:"lineTotalMln"`
}

// EchelonEstimateDTO DTO стоимости по эшелону.
// swagger:model EchelonEstimateDTO
type EchelonEstimateDTO struct {
	// ID эшелона
	EchelonID string `json:"echelonId"`
	// Название эшелона
	EchelonName string `json:"echelonName"`
	// Строки расчёта в эшелоне
	Lines []EstimateLineDTO `json:"lines"`
	// Общая стоимость эшелона в млн ₽
	EchelonTotalMln float64 `json:"echelonTotalMln"`
}

// TypeEstimateDTO DTO стоимости по типу защиты.
// swagger:model TypeEstimateDTO
type TypeEstimateDTO struct {
	// ID типа защиты
	TypeID string `json:"typeId"`
	// Название типа защиты
	TypeName string `json:"typeName"`
	// Строки расчёта данного типа
	Lines []EstimateLineDTO `json:"lines"`
	// Общая стоимость по типу в млн ₽
	TypeTotalMln float64 `json:"typeTotalMln"`
}

// CostCalculationDTO DTO полного расчёта стоимости.
// swagger:response CostCalculationResponse
type CostCalculationDTO struct {
	// Общая стоимость конфигурации в млн ₽
	// Example: 2454.10
	TotalMln float64 `json:"totalMln"`
	// Стоимость по эшелонам
	ByEchelon []EchelonEstimateDTO `json:"byEchelon"`
	// Стоимость по типам защиты
	ByType []TypeEstimateDTO `json:"byType"`
	// Стоимость по объектам
	ByObject []EstimateLineDTO `json:"byObject"`
}

// ---- Budget Check DTOs ----

// BudgetCheckQuery DTO query-параметров для проверки бюджета.
// swagger:parameters CheckBudget
type BudgetCheckQuery struct {
	// ID проекта
	// Required: true
	// In: query
	ID string `json:"id"`
}

// BudgetCheckRequest DTO запроса проверки бюджета.
// swagger:parameters BudgetCheckRequest
type BudgetCheckRequest struct {
	// ID средства защиты
	// Required: true
	// In: body
	// Example: 550e8400-e29b-41d4-a716-446655440000
	AssetID string `json:"assetId"`
	// Количество единиц
	// Required: true
	// In: body
	// Example: 2
	Quantity int `json:"quantity"`
	// ID эшелона (слоя)
	// Required: true
	// In: body
	// Example: detection
	EchelonID string `json:"echelonId"`
}

// ---- Comparison DTOs ----

// CompareQuery — query-параметры для сравнения конфигураций.
// swagger:parameters CompareConfigs
type CompareQuery struct {
	// ID первого проекта
	// Required: true
	// In: query
	ProjectID1 string `json:"id1"`
	// ID второго проекта
	// Required: true
	// In: query
	ProjectID2 string `json:"id2"`
}

// ConfigComparisonResponse — ответ сравнения двух конфигураций.
// swagger:response ConfigComparisonResponse
type ConfigComparisonResponse struct {
	ProjectA ConfigSnapshotDTO `json:"projectA"`
	ProjectB ConfigSnapshotDTO `json:"projectB"`
	Diff     ConfigDiffDTO     `json:"diff"`
}

// ConfigSnapshotDTO — слепок одной конфигурации.
type ConfigSnapshotDTO struct {
	ProjectID         string               `json:"projectId"`
	ProjectName       string               `json:"projectName"`
	StructuralProfile StructuralProfileDTO `json:"structuralProfile"`
	CostCalculation   CostCalculationDTO   `json:"costCalculation"`
}

// StructuralProfileDTO — структурный профиль конфигурации.
type StructuralProfileDTO struct {
	ObjectCount     int                `json:"objectCount"`
	UnitCount       int                `json:"unitCount"`
	EchelonCount    int                `json:"echelonCount"`
	CategoryCount   int                `json:"categoryCount"`
	ConflictCount   int                `json:"conflictCount"`
	CoveredObjCount int                `json:"coveredObjCount"`
	TotalMln        float64            `json:"totalMln"`
	ByEchelon       []EchelonProfileDTO `json:"byEchelon"`
}

// EchelonProfileDTO — профиль одного эшелона.
type EchelonProfileDTO struct {
	LayerID        string `json:"layerId"`
	LayerCode      string `json:"layerCode"`
	LayerName      string `json:"layerName"`
	ObjectCount    int    `json:"objectCount"`
	UnitCount      int    `json:"unitCount"`
	CategoryCount  int    `json:"categoryCount"`
	ConflictCount  int    `json:"conflictCount"`
	CoveredObjCount int   `json:"coveredObjCount"`
}

// ConfigDiffDTO — разница между двумя конфигурациями.
type ConfigDiffDTO struct {
	ObjectCountDelta     int            `json:"objectCountDelta"`
	UnitCountDelta       int            `json:"unitCountDelta"`
	EchelonCountDelta    int            `json:"echelonCountDelta"`
	CategoryCountDelta   int            `json:"categoryCountDelta"`
	ConflictCountDelta   int            `json:"conflictCountDelta"`
	CoveredObjCountDelta int            `json:"coveredObjCountDelta"`
	CostDeltaMln         float64        `json:"costDeltaMln"`
	ByEchelon            []EchelonDiffDTO `json:"byEchelon"`
}

// EchelonDiffDTO — разница по одному эшелону.
type EchelonDiffDTO struct {
	LayerID           string `json:"layerId"`
	LayerCode         string `json:"layerCode"`
	LayerName         string `json:"layerName"`
	ObjectCountDelta  int    `json:"objectCountDelta"`
	UnitCountDelta    int    `json:"unitCountDelta"`
	CategoryCountDelta int   `json:"categoryCountDelta"`
	ConflictCountDelta int   `json:"conflictCountDelta"`
	CoveredObjDelta   int    `json:"coveredObjDelta"`
}

// BudgetCheckResponse DTO результата проверки бюджета.
// swagger:response BudgetCheckResponse
type BudgetCheckResponse struct {
	// Помещается ли в остаток бюджета
	Fits bool `json:"fits"`
	// Остаток бюджета в млн ₽
	// Example: 500.0
	RemainingMln float64 `json:"remainingMln"`
	// Требуемая сумма для добавления в млн ₽
	// Example: 700.0
	RequiredMln float64 `json:"requiredMln"`
	// Режим бюджета: "limited" | "unlimited"
	// Enum: ["limited","unlimited"]
	// Example: limited
	BudgetMode string `json:"budgetMode"`
}
