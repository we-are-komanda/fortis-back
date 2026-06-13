package ui

// DefenseAssetDTO DTO ответа с данными средства защиты.
// swagger:model DefenseAssetDTO
type DefenseAssetDTO struct {
	// ID средства защиты
	// Example: 550e8400-e29b-41d4-a716-446655440000
	ID string `json:"id"`
	// Название средства защиты
	// Example: РЛС 55Ж6
	Name string `json:"name"`
	// Краткое название
	// Example: 55Ж6
	ShortName string `json:"shortName,omitempty"`
	// Описание
	Description string `json:"description,omitempty"`
	// Категория (radiotechnical, radar, electronic-warfare, missile, anti-missile, shturmovaya, artillery, aircraft, helicopter, uav, ship, fortification, infrastructure)
	// Enum: ["radiotechnical","radar","electronic-warfare","missile","anti-missile","shturmovaya","artillery","aircraft","helicopter","uav","ship","fortification","infrastructure"]
	// Example: radar
	Category string `json:"category"`
	// Роли (detection, destruction, ew, c2, cover, deception, supply, engineering, recon, special)
	// items.enum: ["detection","destruction","ew","c2","cover","deception","supply","engineering","recon","special"]
	Roles []string `json:"roles,omitempty"`
	// Цена за единицу в млн руб
	PricePerUnitMln *float64 `json:"pricePerUnitMln,omitempty"`
	// Валюта
	// Example: RUB
	Currency string `json:"currency,omitempty"`
	// Название единицы измерения
	// Example: шт.
	UnitLabel string `json:"unitLabel,omitempty"`
	// Совместимые типы слоёв
	CompatibleLayerTypes []string `json:"compatibleLayerTypes,omitempty"`
	// Рекомендуемые коды слоёв
	RecommendedLayerCodes []string `json:"recommendedLayerCodes,omitempty"`
	// Совместимые коды слоёв
	CompatibleLayerCodes []string `json:"compatibleLayerCodes,omitempty"`
	// Несовместимые коды слоёв
	IncompatibleLayerCodes []string `json:"incompatibleLayerCodes,omitempty"`
	// Тип защиты
	ProtectionType string `json:"protectionType,omitempty"`
	// Минимальная эффективная дальность (км)
	MinEffectiveDistance *float64 `json:"minEffectiveDistance,omitempty"`
	// Максимальная эффективная дальность (км)
	MaxEffectiveDistance *float64 `json:"maxEffectiveDistance,omitempty"`
	// Тип покрытия (circle, sector, line, polygon, none)
	// Enum: ["circle","sector","line","polygon","none"]
	// Example: circle
	CoverageType string `json:"coverageType"`
	// Радиус покрытия (км)
	CoverageRadius *float64 `json:"coverageRadius,omitempty"`
	// Угол покрытия (градусы)
	CoverageAngle *float64 `json:"coverageAngle,omitempty"`
	// Тип развёртывания (static, mobile, infrastructure, software, external)
	// Enum: ["static","mobile","infrastructure","software","external"]
	// Example: static
	DeploymentType string `json:"deploymentType,omitempty"`
	// Тип размещения (map-object, zone-object, non-physical)
	// Enum: ["map-object","zone-object","non-physical"]
	// Example: map-object
	PlacementType string `json:"placementType,omitempty"`
	// URL иконки
	IconURL string `json:"iconUrl,omitempty"`
	// URL 3D-модели
	ModelURL string `json:"modelUrl,omitempty"`
	// Оценка/рейтинг
	Score *int `json:"score,omitempty"`
	// Приоритет (critical, high, medium, low)
	// Enum: ["critical","high","medium","low"]
	Priority *string `json:"priority,omitempty"`
	// Профиль составной установки
	CompoundProfile *CompoundProfileDTO `json:"compoundProfile,omitempty"`
	// ТТХ оружия/установок
	WeaponSpec *WeaponSpecDTO `json:"weaponSpec,omitempty"`
	// ТТХ средств обнаружения
	DetectionSpec *DetectionSpecDTO `json:"detectionSpec,omitempty"`
	// ТТХ средств РЭБ/спуферов
	EWSpec *EWSpecDTO `json:"ewSpec,omitempty"`
	// Теги
	Tags []string `json:"tags,omitempty"`
	// ID устаревшего элемента
	LegacyItemID string `json:"legacyItemId,omitempty"`
	// ID калькуляторного средства
	CalculatorAssetID *string `json:"calculatorAssetId,omitempty"`
	// ID групп каталога карты
	MapCatalogGroupIDs []string `json:"mapCatalogGroupIds,omitempty"`
	// ID предприятия (null для общего каталога)
	EnterpriseID *string `json:"enterpriseId,omitempty"`
	// Флаг общего доступа
	IsPublic bool `json:"isPublic"`
	// Дата создания (RFC3339)
	// Example: 2026-06-12T14:00:00Z
	CreatedAt string `json:"createdAt"`
	// Дата обновления (RFC3339)
	// Example: 2026-06-12T14:00:00Z
	UpdatedAt string `json:"updatedAt"`
}

// CompoundProfileDTO DTO профиля составной установки.
// swagger:model CompoundProfileDTO
type CompoundProfileDTO struct {
	// Тип профиля (compound-post)
	// Example: compound-post
	Kind string `json:"kind,omitempty"`
	// Тип поста (МОГ, ПВН, ГОР, КПП)
	// Example: МОГ
	PostType string `json:"postType,omitempty"`
	// Численность личного состава
	// Example: 4
	PersonnelCount string `json:"personnelCount,omitempty"`
	// Подотчётность (МО, Росгвардия, ЧОП)
	// Example: МО
	Accountability string `json:"accountability,omitempty"`
	// Вооружение (одна строка)
	// Example: Автомат/пулемёт/ПБС
	Armament string `json:"armament,omitempty"`
	// Количество единиц вооружения
	// Example: 2
	WeaponUnits string `json:"weaponUnits,omitempty"`
	// Сектор/дальность
	// Example: до 4-8 км, сектор 90-360°
	SectorOrRange string `json:"sectorOrRange,omitempty"`
	// Азимут
	Azimuth float64 `json:"azimuth,omitempty"`
}

// WeaponSpecDTO DTO ТТХ оружия/установок.
// swagger:model WeaponSpecDTO
type WeaponSpecDTO struct {
	// Калибр
	Caliber *string `json:"caliber,omitempty"`
	// Тип боеприпаса
	AmmunitionType *string `json:"ammunitionType,omitempty"`
	// Режим работы (ручная/автоматическая)
	OperationMode *string `json:"operationMode,omitempty"`
	// Количество модулей
	ModuleCount *int `json:"moduleCount,omitempty"`
	// Ручной режим
	IsManual *bool `json:"isManual,omitempty"`
}

// DetectionSpecDTO DTO ТТХ средств обнаружения.
// swagger:model DetectionSpecDTO
type DetectionSpecDTO struct {
	// Диапазон частот обнаружения
	FrequencyRange *string `json:"frequencyRange,omitempty"`
	// Режим обнаружения (активный/пассивный)
	DetectionMode *string `json:"detectionMode,omitempty"`
	// Скорость оборота (для активной РЛС)
	RotationSpeed *float64 `json:"rotationSpeed,omitempty"`
	// Поле зрения (градусы)
	FieldOfView *float64 `json:"fieldOfView,omitempty"`
	// Наличие тепловизора
	HasThermalImager *bool `json:"hasThermalImager,omitempty"`
}

// EWSpecDTO DTO ТТХ средств РЭБ/спуферов.
// swagger:model EWSpecDTO
type EWSpecDTO struct {
	// Диапазон частот
	FrequencyRange *string `json:"frequencyRange,omitempty"`
	// Дальность/зона действия (км)
	ActionRange *float64 `json:"actionRange,omitempty"`
	// Азимут
	Azimuth *float64 `json:"azimuth,omitempty"`
}

// CreateDefenseAssetRequest DTO запроса на создание средства защиты.
// swagger:parameters CreateDefenseAssetRequest
type CreateDefenseAssetRequest struct {
	// Название средства защиты
	// Required: true
	// Example: РЛС 55Ж6
	Name string `json:"name"`
	// Краткое название
	ShortName string `json:"shortName,omitempty"`
	// Описание
	Description string `json:"description,omitempty"`
	// Категория
	// Required: true
	// Enum: ["radiotechnical","radar","electronic-warfare","missile","anti-missile","shturmovaya","artillery","aircraft","helicopter","uav","ship","fortification","infrastructure"]
	// Example: radar
	Category string `json:"category"`
	// Роли
	// items.enum: ["detection","destruction","ew","c2","cover","deception","supply","engineering","recon","special"]
	Roles []string `json:"roles,omitempty"`
	// Цена за единицу в млн руб
	PricePerUnitMln *float64 `json:"pricePerUnitMln,omitempty"`
	// Валюта
	// Example: RUB
	Currency string `json:"currency,omitempty"`
	// Название единицы измерения
	UnitLabel string `json:"unitLabel,omitempty"`
	// Совместимые типы слоёв
	CompatibleLayerTypes []string `json:"compatibleLayerTypes,omitempty"`
	// Рекомендуемые коды слоёв
	RecommendedLayerCodes []string `json:"recommendedLayerCodes,omitempty"`
	// Совместимые коды слоёв
	CompatibleLayerCodes []string `json:"compatibleLayerCodes,omitempty"`
	// Несовместимые коды слоёв
	IncompatibleLayerCodes []string `json:"incompatibleLayerCodes,omitempty"`
	// Тип защиты
	ProtectionType string `json:"protectionType,omitempty"`
	// Минимальная эффективная дальность (км)
	MinEffectiveDistance *float64 `json:"minEffectiveDistance,omitempty"`
	// Максимальная эффективная дальность (км)
	MaxEffectiveDistance *float64 `json:"maxEffectiveDistance,omitempty"`
	// Тип покрытия
	// Required: true
	// Enum: ["circle","sector","line","polygon","none"]
	// Example: circle
	CoverageType string `json:"coverageType"`
	// Радиус покрытия (км)
	CoverageRadius *float64 `json:"coverageRadius,omitempty"`
	// Угол покрытия (градусы)
	CoverageAngle *float64 `json:"coverageAngle,omitempty"`
	// Тип развёртывания
	// Enum: ["static","mobile","infrastructure","software","external"]
	DeploymentType string `json:"deploymentType,omitempty"`
	// Тип размещения
	// Enum: ["map-object","zone-object","non-physical"]
	PlacementType string `json:"placementType,omitempty"`
	// URL иконки
	IconURL string `json:"iconUrl,omitempty"`
	// URL 3D-модели
	ModelURL string `json:"modelUrl,omitempty"`
	// Оценка
	Score *int `json:"score,omitempty"`
	// Приоритет
	// Enum: ["critical","high","medium","low"]
	Priority *string `json:"priority,omitempty"`
	// Профиль составной установки
	CompoundProfile *CompoundProfileDTO `json:"compoundProfile,omitempty"`
	// ТТХ оружия/установок
	WeaponSpec *WeaponSpecDTO `json:"weaponSpec,omitempty"`
	// ТТХ средств обнаружения
	DetectionSpec *DetectionSpecDTO `json:"detectionSpec,omitempty"`
	// ТТХ средств РЭБ/спуферов
	EWSpec *EWSpecDTO `json:"ewSpec,omitempty"`
	// Теги
	Tags []string `json:"tags,omitempty"`
	// ID устаревшего элемента
	LegacyItemID string `json:"legacyItemId,omitempty"`
	// ID калькуляторного средства
	CalculatorAssetID *string `json:"calculatorAssetId,omitempty"`
	// ID групп каталога карты
	MapCatalogGroupIDs []string `json:"mapCatalogGroupIds,omitempty"`
	// ID предприятия (null для общего каталога)
	EnterpriseID *string `json:"enterpriseId,omitempty"`
	// Флаг общего доступа
	IsPublic bool `json:"isPublic"`
}

// UpdateDefenseAssetRequest DTO запроса на обновление средства защиты.
// swagger:parameters UpdateDefenseAssetRequest
type UpdateDefenseAssetRequest struct {
	// Название средства защиты
	Name *string `json:"name,omitempty"`
	// Краткое название
	ShortName *string `json:"shortName,omitempty"`
	// Описание
	Description *string `json:"description,omitempty"`
	// Категория
	// Enum: ["radiotechnical","radar","electronic-warfare","missile","anti-missile","shturmovaya","artillery","aircraft","helicopter","uav","ship","fortification","infrastructure"]
	Category *string `json:"category,omitempty"`
	// Роли
	// items.enum: ["detection","destruction","ew","c2","cover","deception","supply","engineering","recon","special"]
	Roles []string `json:"roles,omitempty"`
	// Цена за единицу
	PricePerUnitMln *float64 `json:"pricePerUnitMln,omitempty"`
	// Валюта
	Currency *string `json:"currency,omitempty"`
	// Название единицы измерения
	UnitLabel *string `json:"unitLabel,omitempty"`
	// Совместимые типы слоёв
	CompatibleLayerTypes []string `json:"compatibleLayerTypes,omitempty"`
	// Рекомендуемые коды слоёв
	RecommendedLayerCodes []string `json:"recommendedLayerCodes,omitempty"`
	// Совместимые коды слоёв
	CompatibleLayerCodes []string `json:"compatibleLayerCodes,omitempty"`
	// Несовместимые коды слоёв
	IncompatibleLayerCodes []string `json:"incompatibleLayerCodes,omitempty"`
	// Тип защиты
	ProtectionType *string `json:"protectionType,omitempty"`
	// Минимальная эффективная дальность
	MinEffectiveDistance *float64 `json:"minEffectiveDistance,omitempty"`
	// Максимальная эффективная дальность
	MaxEffectiveDistance *float64 `json:"maxEffectiveDistance,omitempty"`
	// Тип покрытия
	// Enum: ["circle","sector","line","polygon","none"]
	CoverageType *string `json:"coverageType,omitempty"`
	// Радиус покрытия
	CoverageRadius *float64 `json:"coverageRadius,omitempty"`
	// Угол покрытия
	CoverageAngle *float64 `json:"coverageAngle,omitempty"`
	// Тип развёртывания
	// Enum: ["static","mobile","infrastructure","software","external"]
	DeploymentType *string `json:"deploymentType,omitempty"`
	// Тип размещения
	// Enum: ["map-object","zone-object","non-physical"]
	PlacementType *string `json:"placementType,omitempty"`
	// URL иконки
	IconURL *string `json:"iconUrl,omitempty"`
	// URL 3D-модели
	ModelURL *string `json:"modelUrl,omitempty"`
	// Оценка
	Score *int `json:"score,omitempty"`
	// Приоритет
	// Enum: ["critical","high","medium","low"]
	Priority *string `json:"priority,omitempty"`
	// Профиль составной установки
	CompoundProfile *CompoundProfileDTO `json:"compoundProfile,omitempty"`
	// ТТХ оружия/установок
	WeaponSpec *WeaponSpecDTO `json:"weaponSpec,omitempty"`
	// ТТХ средств обнаружения
	DetectionSpec *DetectionSpecDTO `json:"detectionSpec,omitempty"`
	// ТТХ средств РЭБ/спуферов
	EWSpec *EWSpecDTO `json:"ewSpec,omitempty"`
	// Теги
	Tags []string `json:"tags,omitempty"`
	// ID устаревшего элемента
	LegacyItemID *string `json:"legacyItemId,omitempty"`
	// ID калькуляторного средства
	CalculatorAssetID *string `json:"calculatorAssetId,omitempty"`
	// ID групп каталога карты
	MapCatalogGroupIDs []string `json:"mapCatalogGroupIds,omitempty"`
	// Флаг общего доступа
	IsPublic *bool `json:"isPublic,omitempty"`
}

// DefenseAssetListResponse DTO ответа со списком средств защиты.
// swagger:response DefenseAssetListResponse
type DefenseAssetListResponse struct {
	// Список элементов
	Items []DefenseAssetDTO `json:"items"`
	// Общее количество элементов
	TotalItems int64 `json:"totalItems"`
}

// AssetListQuery DTO query-параметров для списка средств защиты.
// swagger:parameters AssetListQuery
type AssetListQuery struct {
	// ID предприятия для фильтрации
	// In: query
	EnterpriseID string `json:"enterpriseId"`
	// Фильтр по признаку общего доступа (true/false)
	// In: query
	IsPublic string `json:"isPublic"`
	// Фильтр по категории
	// In: query
	Category string `json:"category"`
	// Количество записей на странице (по умолчанию 20, макс 100)
	// In: query
	// Example: 20
	Limit int `json:"limit"`
	// Смещение от начала списка
	// In: query
	// Example: 0
	Offset int `json:"offset"`
}

// AssetGetQuery DTO query-параметров для получения по ID.
// swagger:parameters AssetGetQuery
type AssetGetQuery struct {
	// ID средства защиты
	// Required: true
	// In: query
	ID string `json:"id"`
}
