package ui

import (
	"encoding/json"
	projectUi "github.com/fortis/backend/internal/modules/defense_project/ui"
)

// ---- Budget Config DTOs ----

// BudgetConfigQuery DTO query-параметров для получения/обновления бюджета.
// swagger:parameters GetBudgetConfig UpdateBudgetConfig
type BudgetConfigQuery struct {
	// ID проекта; id — совместимый alias, конфликтующие значения отклоняются.
	// In: query
	ProjectID string `json:"projectId"`
	// In: query
	ID string `json:"id"`
	// Для GET: сохранённая версия; без параметра читается текущая.
	// In: query
	// Minimum: 1
	ProjectVersion *int `json:"projectVersion,omitempty"`
}

// UpdateBudgetRequest DTO запроса на обновление бюджета.
// swagger:parameters UpdateBudgetRequest
type UpdateBudgetRequest struct {
	// Ожидаемая положительная общая версия проекта.
	// Required: true
	Version int `json:"version"`
	// Режим бюджета.
	// Required: true
	// Enum: ["limited","unlimited"]
	BudgetMode string `json:"budgetMode"`
	// Точные копейки строкой для limited, явный null для unlimited; ноль допустим.
	// Required: true
	BudgetLimitMinor json.RawMessage `json:"budgetLimitMinor"`
}

// BudgetProjectionDTO — точный бюджет сохранённой ревизии, неизвестные суммы — null.
// swagger:response BudgetProjectionResponse
type BudgetProjectionDTO struct {
	Identity       CalculationIdentityDTO `json:"identity"`
	Config         BudgetSettingsDTO      `json:"config"`
	SpentMinor     *string                `json:"spentMinor"`
	RemainingMinor *string                `json:"remainingMinor"`
	Status         string                 `json:"status"`
}
type BudgetSettingsDTO struct {
	BudgetMode       string  `json:"budgetMode"`
	BudgetLimitMinor *string `json:"budgetLimitMinor"`
}

// BudgetUpdateResponse содержит новую каноническую ревизию и её бюджет.
// swagger:response BudgetUpdateResponse
type BudgetUpdateResponse struct {
	projectUi.ProjectResponse
	Budget BudgetProjectionDTO `json:"budget"`
}
type ProjectRefDTO struct {
	ProjectID      string `json:"projectId"`
	ProjectVersion int    `json:"projectVersion"`
}

// ---- Cost Calculation DTOs ----

// CostQuery DTO query-параметров для расчёта стоимости.
// swagger:parameters calculateCost
type CostQuery struct {
	// ID проекта (предпочтительный параметр).
	// In: query
	ProjectID string `json:"projectId"`
	// Совместимый alias projectId; конфликтующие значения отклоняются.
	// In: query
	ID string `json:"id"`
	// Версия сохранённой ревизии; без параметра читается текущая.
	// In: query
	// Minimum: 1
	ProjectVersion *int `json:"projectVersion,omitempty"`
}

type CalculationIdentityDTO struct {
	ProjectID          string            `json:"projectId"`
	ProjectVersion     int               `json:"projectVersion"`
	CalculationVersion string            `json:"calculationVersion"`
	InputDataVersions  map[string]string `json:"inputDataVersions"`
	SnapshotDigest     string            `json:"snapshotDigest"`
}
type FinancialLineDTO struct {
	ObjectID       string                 `json:"objectId"`
	AssetID        string                 `json:"assetId"`
	LayerID        string                 `json:"layerId"`
	Category       string                 `json:"category"`
	Name           string                 `json:"name"`
	Quantity       int                    `json:"quantity"`
	UnitPriceMinor *string                `json:"unitPriceMinor"`
	LineTotalMinor *string                `json:"lineTotalMinor"`
	PriceSource    string                 `json:"priceSource"`
	Provenance     *CostDataProvenanceDTO `json:"provenance"`
}
type CostDataProvenanceDTO struct {
	SourceLabel            string  `json:"sourceLabel"`
	SourceDocumentID       *string `json:"sourceDocumentId"`
	SourceURL              *string `json:"sourceUrl"`
	SourceDate             *string `json:"sourceDate"`
	RecordedAt             string  `json:"recordedAt"`
	RecordedBy             string  `json:"recordedBy"`
	Quality                string  `json:"quality"`
	Revision               string  `json:"revision"`
	SourceDocumentRevision *string `json:"sourceDocumentRevision,omitempty"`
	SourceDocumentChecksum *string `json:"sourceDocumentChecksum,omitempty"`
}
type CostGroupDTO struct {
	ID                 string  `json:"id"`
	Name               string  `json:"name"`
	ObjectCount        int     `json:"objectCount"`
	UnitCount          int     `json:"unitCount"`
	KnownSubtotalMinor string  `json:"knownSubtotalMinor"`
	TotalMinor         *string `json:"totalMinor"`
}
type CostIssueDTO struct {
	Code      string   `json:"code"`
	Severity  string   `json:"severity"`
	ObjectIDs []string `json:"objectIds"`
	Message   string   `json:"message"`
}

// CostProjectionDTO — точная стоимость сохранённой ревизии в копейках RUB.
// swagger:model CostProjection
type CostProjectionDTO struct {
	Identity              CalculationIdentityDTO `json:"identity"`
	Currency              string                 `json:"currency"`
	Lines                 []FinancialLineDTO     `json:"lines"`
	ByLayer               []CostGroupDTO         `json:"byLayer"`
	ByType                []CostGroupDTO         `json:"byType"`
	KnownSubtotalMinor    string                 `json:"knownSubtotalMinor"`
	TotalMinor            *string                `json:"totalMinor"`
	IsComplete            bool                   `json:"isComplete"`
	UnknownPriceObjectIDs []string               `json:"unknownPriceObjectIds"`
	Warnings              []CostIssueDTO         `json:"warnings"`
}

// swagger:response CostProjectionResponse
type CostProjectionResponse struct {
	// In: body
	Body CostProjectionDTO
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
	// ID проекта; id — совместимый alias.
	// In: query
	ProjectID string `json:"projectId"`
	// In: query
	ID string `json:"id"`
}

// BudgetCheckRequest DTO запроса проверки бюджета.
// swagger:parameters BudgetCheckRequest
type BudgetCheckRequest struct {
	// Ожидаемая положительная версия проекта.
	// Required: true
	Version int `json:"version"`
	// Полный итоговый кандидат; проверка не записывает и не резервирует бюджет.
	// Required: true
	ProjectJSON string `json:"projectJson"`
}

// ---- Comparison DTOs ----

// CompareQuery — query-параметры для сравнения конфигураций.
// swagger:parameters CompareConfigs
type CompareQuery struct {
	// In: query
	ProjectIDA string `json:"projectIdA"`
	// In: query
	// Required: true
	// Minimum: 1
	ProjectVersionA int `json:"projectVersionA"`
	// In: query
	ProjectIDB string `json:"projectIdB"`
	// In: query
	// Required: true
	// Minimum: 1
	ProjectVersionB int `json:"projectVersionB"`
	// Совместимый alias projectIdA; версия всё равно обязательна.
	// In: query
	ID1 string `json:"id1"`
	// Совместимый alias projectIdB; версия всё равно обязательна.
	// In: query
	ID2 string `json:"id2"`
}

// ConfigComparisonResponse — фиксированные ревизии; без оценки эффективности.
// swagger:response ConfigComparisonResponse
type ConfigComparisonResponse struct {
	EnterpriseID      string              `json:"enterpriseId"`
	A                 ComparedRevisionDTO `json:"a"`
	B                 ComparedRevisionDTO `json:"b"`
	DifferenceMinor   *string             `json:"differenceMinor"`
	DifferencePercent *string             `json:"differencePercent"`
	Warnings          []CostIssueDTO      `json:"warnings"`
}
type ComparedRevisionDTO struct {
	Identity  CalculationIdentityDTO `json:"identity"`
	Name      string                 `json:"name"`
	UpdatedAt string                 `json:"updatedAt"`
	Cost      CostProjectionDTO      `json:"cost"`
	Budget    BudgetProjectionDTO    `json:"budget"`
}

// StructuralProfileDTO — структурный профиль конфигурации.
type StructuralProfileDTO struct {
	ObjectCount     int                 `json:"objectCount"`
	UnitCount       int                 `json:"unitCount"`
	EchelonCount    int                 `json:"echelonCount"`
	CategoryCount   int                 `json:"categoryCount"`
	ConflictCount   int                 `json:"conflictCount"`
	CoveredObjCount int                 `json:"coveredObjCount"`
	TotalMln        float64             `json:"totalMln"`
	ByEchelon       []EchelonProfileDTO `json:"byEchelon"`
}

// EchelonProfileDTO — профиль одного эшелона.
type EchelonProfileDTO struct {
	LayerID         string `json:"layerId"`
	LayerCode       string `json:"layerCode"`
	LayerName       string `json:"layerName"`
	ObjectCount     int    `json:"objectCount"`
	UnitCount       int    `json:"unitCount"`
	CategoryCount   int    `json:"categoryCount"`
	ConflictCount   int    `json:"conflictCount"`
	CoveredObjCount int    `json:"coveredObjCount"`
}

// BudgetCheckResponse DTO результата проверки бюджета.
// swagger:response BudgetCheckResponse
type BudgetCheckResponse struct {
	BasedOn                 ProjectRefDTO `json:"basedOn"`
	Status                  string        `json:"status"`
	ReasonCode              string        `json:"reasonCode"`
	ProjectedSpentMinor     *string       `json:"projectedSpentMinor"`
	ProjectedRemainingMinor *string       `json:"projectedRemainingMinor"`
}
