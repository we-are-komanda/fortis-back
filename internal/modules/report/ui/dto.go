package ui

import budgetUi "github.com/fortis/backend/internal/modules/budget/ui"

// ---- Query DTO ----

// ReportQuery DTO query-параметров для получения отчёта.
// swagger:parameters GetReport
type ReportQuery struct {
	// ID проекта
	// Required: true
	// In: query
	ID string `json:"id"`
	// Режим скрытия стоимости (true — все *Mln поля = 0)
	// In: query
	// Example: false
	HideCost bool `json:"hideCost"`
}

// ---- Response DTOs ----

// ReportResponse DTO ответа отчёта GIS MVP.
// swagger:response ReportResponse
type ReportResponse struct {
	// ID проекта
	// Required: true
	// Example: 550e8400-e29b-41d4-a716-446655440000
	ProjectID string `json:"projectId"`
	// Название проекта
	// Required: true
	// Example: Защита объекта "Северный"
	ProjectName string `json:"projectName"`
	// Защищаемый объект
	// Required: true
	BaseObject BaseObjectDTO `json:"baseObject"`
	// Слои (эшелоны) защиты
	// Required: true
	Layers []LayerDTO `json:"layers"`
	// Размещённые объекты с МОГ-строками
	// Required: true
	PlacedObjects []ReportObjectLineDTO `json:"placedObjects"`
	// Расчёт стоимости
	// Required: true
	Estimate budgetUi.CostCalculationDTO `json:"estimate"`
	// Структурный профиль
	// Required: true
	StructuralProfile budgetUi.StructuralProfileDTO `json:"structuralProfile"`
	// Флаг скрытия стоимости
	// Required: true
	// Example: false
	HideCost bool `json:"hideCost"`
}

// BaseObjectDTO DTO защищаемого объекта.
// swagger:model BaseObjectDTO
type BaseObjectDTO struct {
	// ID защищаемого объекта
	// Required: true
	ID string `json:"id"`
	// Название защищаемого объекта
	// Required: true
	// Example: Объект "Северный"
	Name string `json:"name"`
	// Координаты центра
	// Required: true
	Center CenterDTO `json:"center"`
}

// CenterDTO DTO координат центра.
// swagger:model CenterDTO
type CenterDTO struct {
	// Широта
	// Required: true
	// Example: 55.751244
	Lat float64 `json:"lat"`
	// Долгота
	// Required: true
	// Example: 37.618423
	Lng float64 `json:"lng"`
}

// LayerDTO DTO слоя защиты для отчёта.
// swagger:model LayerDTO
type LayerDTO struct {
	// ID слоя
	// Required: true
	ID string `json:"id"`
	// Название слоя
	// Required: true
	// Example: Дальнее обнаружение
	Name string `json:"name"`
	// Код слоя
	// Required: true
	// Example: early-warning
	Code string `json:"code"`
	// Описание слоя
	Description *string `json:"description,omitempty"`
	// Тип геометрии слоя: "circle", "ring", "polygon", "freeform"
	// Required: true
	// Enum: ["circle","ring","polygon","freeform"]
	GeometryType string `json:"geometryType"`
	// Широта центра геометрии
	CenterLat *float64 `json:"centerLat,omitempty"`
	// Долгота центра геометрии
	CenterLng *float64 `json:"centerLng,omitempty"`
	// Радиус круга в метрах
	RadiusM *float64 `json:"radiusM,omitempty"`
	// Минимальный радиус кольца в метрах
	MinRadiusM *float64 `json:"minRadiusM,omitempty"`
	// Максимальный радиус кольца в метрах
	MaxRadiusM *float64 `json:"maxRadiusM,omitempty"`
	// Цвет слоя (hex)
	Color *string `json:"color,omitempty"`
	// Прозрачность (0-1)
	Opacity *float64 `json:"opacity,omitempty"`
}

// ReportObjectLineDTO DTO строки размещённого объекта для отчёта.
// swagger:model ReportObjectLineDTO
type ReportObjectLineDTO struct {
	// ID размещённого объекта
	// Required: true
	ObjectID string `json:"objectId"`
	// ID средства защиты
	// Required: true
	AssetID string `json:"assetId"`
	// Название средства защиты
	// Required: true
	// Example: РЛС "Небо-М"
	AssetName string `json:"assetName"`
	// ID слоя
	// Required: true
	LayerID string `json:"layerId"`
	// Код слоя
	// Required: true
	LayerCode string `json:"layerCode"`
	// Название слоя
	// Required: true
	LayerName string `json:"layerName"`
	// Количество единиц
	// Required: true
	// Example: 2
	Quantity int `json:"quantity"`
	// Тип защиты
	// Required: true
	// Example: detection
	ProtectionType string `json:"protectionType"`
	// Цена за единицу в млн ₽
	// Example: 150.0
	UnitPriceMln float64 `json:"unitPriceMln"`
	// Итоговая стоимость строки в млн ₽
	// Example: 300.0
	LineTotalMln float64 `json:"lineTotalMln"`
	// Является ли составной установкой (МОГ/ПВН/ГОР/КПП)
	// Required: true
	// Example: false
	IsCompoundPost bool `json:"isCompoundPost"`
	// Сводка по составной установке
	CompositionSummary *CompositionSummaryDTO `json:"compositionSummary,omitempty"`
	// Сводка по оружию
	WeaponSummary *WeaponSummaryDTO `json:"weaponSummary,omitempty"`
	// Сводка по азимутальному сектору
	AzimuthSectorSummary *AzimuthSectorSummaryDTO `json:"azimuthSectorSummary,omitempty"`
}

// CompositionSummaryDTO DTO сводки по составной установке.
// swagger:model CompositionSummaryDTO
type CompositionSummaryDTO struct {
	// Тип поста
	// Example: МОГ
	PostType string `json:"postType"`
	// Личный состав
	// Example: 4 чел.
	Personnel string `json:"personnel"`
	// Подотчётность
	// Example: командир роты
	Accountability string `json:"accountability"`
	// Вооружение
	// Example: ПЗРК "Игла"
	Armament string `json:"armament"`
	// Количество орудий/единиц
	// Example: 2
	WeaponUnits string `json:"weaponUnits"`
	// Сектор/дальность
	// Example: 120°
	SectorOrRange string `json:"sectorOrRange"`
	// Азимут (градусы)
	// Example: 180
	Azimuth float64 `json:"azimuth"`
}

// WeaponSummaryDTO DTO сводки по оружию.
// swagger:model WeaponSummaryDTO
type WeaponSummaryDTO struct {
	// Калибр
	// Example: 30 мм
	Caliber string `json:"caliber"`
	// Тип боеприпаса
	// Example: осколочно-фугасный
	AmmunitionType string `json:"ammunitionType"`
	// Режим работы
	// Example: автоматический
	OperationMode string `json:"operationMode"`
	// Количество модулей
	// Example: 2
	ModuleCount string `json:"moduleCount"`
	// Признак ручного управления
	// Example: автоматическое
	IsManual string `json:"isManual"`
}

// AzimuthSectorSummaryDTO DTO сводки по азимутальному сектору.
// swagger:model AzimuthSectorSummaryDTO
type AzimuthSectorSummaryDTO struct {
	// Азимут (градусы)
	// Required: true
	// Example: 180
	Azimuth float64 `json:"azimuth"`
	// Тип зоны покрытия
	// Required: true
	// Example: sector
	CoverageType string `json:"coverageType"`
	// Угол сектора (градусы)
	Angle *float64 `json:"angle,omitempty"`
}
