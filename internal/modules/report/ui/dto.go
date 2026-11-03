package ui

import budget "github.com/fortis/backend/internal/modules/budget/ui"

// ReportQuery — выпуск фиксированной версии или открытие уже выпущенного отчёта.
// swagger:parameters getReport
type ReportQuery struct {
	// UUID выпущенного отчёта; исключает все другие параметры.
	// In: query
	ReportID string `json:"reportId"`
	// UUID проекта; обязателен при выпуске.
	// In: query
	ID string `json:"id"`
	// Положительная сохранённая версия; обязательна при выпуске.
	// In: query
	// Minimum: 1
	ProjectVersion int `json:"projectVersion"`
	// Обязателен при выпуске: только true или false.
	// In: query
	HideCost string `json:"hideCost"`
	// In: query
	// Enum: ["all","saved"]
	MapView string `json:"mapView"`
}
type Coordinate struct {
	present  bool
	Lat      float64  `json:"lat"`
	Lng      float64  `json:"lng"`
	Altitude *float64 `json:"altitude,omitempty"`
}
type ReportGeometry struct {
	Type        string        `json:"type"`
	Center      *Coordinate   `json:"center,omitempty"`
	RadiusM     *float64      `json:"radiusM,omitempty"`
	MinRadiusM  *float64      `json:"minRadiusM,omitempty"`
	MaxRadiusM  *float64      `json:"maxRadiusM,omitempty"`
	Coordinates *[]Coordinate `json:"coordinates,omitempty"`
	Points      *[]Coordinate `json:"points,omitempty"`
	IsClosed    *bool         `json:"isClosed,omitempty"`
}
type ReportBaseObject struct {
	ID     string     `json:"id"`
	Name   string     `json:"name"`
	Center Coordinate `json:"center"`
}
type ReportLayer struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Code        string         `json:"code"`
	Description *string        `json:"description"`
	Order       int            `json:"order"`
	Color       *string        `json:"color"`
	Opacity     *float64       `json:"opacity"`
	IsActive    bool           `json:"isActive"`
	IsVisible   bool           `json:"isVisible"`
	Geometry    ReportGeometry `json:"geometry"`
}
type ReportMapObject struct {
	ObjectID       string     `json:"objectId"`
	AssetID        string     `json:"assetId"`
	LayerID        string     `json:"layerId"`
	Name           string     `json:"name"`
	Category       string     `json:"category"`
	Quantity       int        `json:"quantity"`
	Coordinates    Coordinate `json:"coordinates"`
	Rotation       *float64   `json:"rotation"`
	Scale          *float64   `json:"scale"`
	IsVisibleOnMap bool       `json:"isVisibleOnMap"`
}
type ReportCompoundCard struct {
	Kind           string  `json:"kind"`
	PostType       string  `json:"postType"`
	PersonnelCount string  `json:"personnelCount"`
	Accountability string  `json:"accountability"`
	Armament       string  `json:"armament"`
	WeaponUnits    string  `json:"weaponUnits"`
	SectorOrRange  string  `json:"sectorOrRange"`
	Azimuth        float64 `json:"azimuth"`
}
type ReportWeaponCard struct {
	Caliber        *string `json:"caliber,omitempty"`
	AmmunitionType *string `json:"ammunitionType,omitempty"`
	OperationMode  *string `json:"operationMode,omitempty"`
	ModuleCount    *int    `json:"moduleCount,omitempty"`
	IsManual       *bool   `json:"isManual,omitempty"`
}
type ReportDetectionCard struct {
	FrequencyRange   *string  `json:"frequencyRange,omitempty"`
	DetectionMode    *string  `json:"detectionMode,omitempty"`
	RotationSpeed    *float64 `json:"rotationSpeed,omitempty"`
	FieldOfView      *float64 `json:"fieldOfView,omitempty"`
	HasThermalImager *bool    `json:"hasThermalImager,omitempty"`
}
type ReportEWCard struct {
	FrequencyRange *string  `json:"frequencyRange,omitempty"`
	ActionRange    *float64 `json:"actionRange,omitempty"`
	Azimuth        *float64 `json:"azimuth,omitempty"`
}
type ReportCard struct {
	Compound  *ReportCompoundCard  `json:"compound,omitempty"`
	Weapon    *ReportWeaponCard    `json:"weapon,omitempty"`
	Detection *ReportDetectionCard `json:"detection,omitempty"`
	EW        *ReportEWCard        `json:"ew,omitempty"`
}
type ReportObject struct {
	ReportMapObject
	Description    *string     `json:"description"`
	Notes          *string     `json:"notes"`
	UnitLabel      string      `json:"unitLabel"`
	ProtectionType string      `json:"protectionType"`
	Card           *ReportCard `json:"card"`
}
type ReportMap struct {
	BaseObject ReportBaseObject  `json:"baseObject"`
	Layers     []ReportLayer     `json:"layers"`
	Objects    []ReportMapObject `json:"objects"`
}
type ReportVisibility struct {
	Mode         string   `json:"mode"`
	LayerIDs     []string `json:"layerIds"`
	ObjectIDs    []string `json:"objectIds"`
	ShowCoverage bool     `json:"showCoverage"`
}
type ReportEnterprise struct {
	ID   string  `json:"id"`
	Name *string `json:"name"`
}
type ReportCommon struct {
	ReportID            string                        `json:"reportId"`
	ReportSchemaVersion string                        `json:"reportSchemaVersion"`
	Identity            budget.CalculationIdentityDTO `json:"identity"`
	GeneratedAt         string                        `json:"generatedAt"`
	Title               string                        `json:"title"`
	Enterprise          ReportEnterprise              `json:"enterprise"`
	MapSnapshotData     ReportMap                     `json:"mapSnapshotData"`
	MapVisibility       ReportVisibility              `json:"mapVisibility"`
	Objects             []ReportObject                `json:"objects"`
	Warnings            []budget.CostIssueDTO         `json:"warnings"`
	Limitations         []string                      `json:"limitations"`
}

// PricedReportResponse — сохранённый отчёт с точными суммами выбранной версии.
// swagger:response PricedReportResponse
type PricedReportResponse struct {
	ReportCommon
	HideCost bool                           `json:"hideCost"`
	Cost     budget.CostProjectionDTO       `json:"cost"`
	Budget   budget.BudgetProjectionDTO     `json:"budget"`
	Sources  []budget.CostDataProvenanceDTO `json:"sources"`
}

// UnpricedReportResponse — отдельный allowlist без финансовых полей и источников.
// swagger:response UnpricedReportResponse
type UnpricedReportResponse struct {
	ReportCommon
	HideCost bool `json:"hideCost"`
}
