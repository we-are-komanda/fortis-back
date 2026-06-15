package domain

import (
	budgetDomain "github.com/fortis/backend/internal/modules/budget/domain"
)

// ReportBaseObject — защищаемый объект для отчёта (value object).
type ReportBaseObject struct {
	id   string
	name string
	lat  float64
	lng  float64
}

// NewReportBaseObject создаёт новый ReportBaseObject.
func NewReportBaseObject(id, name string, lat, lng float64) ReportBaseObject {
	return ReportBaseObject{id: id, name: name, lat: lat, lng: lng}
}

// ID возвращает ID защищаемого объекта.
func (o ReportBaseObject) ID() string { return o.id }

// Name возвращает название защищаемого объекта.
func (o ReportBaseObject) Name() string { return o.name }

// Lat возвращает широту центра.
func (o ReportBaseObject) Lat() float64 { return o.lat }

// Lng возвращает долготу центра.
func (o ReportBaseObject) Lng() float64 { return o.lng }

// ReportLayer — слой (эшелон) защиты для отчёта (value object).
type ReportLayer struct {
	id           string
	name         string
	code         string
	description  *string
	geometryType string
	centerLat    *float64
	centerLng    *float64
	radiusM      *float64
	minRadiusM   *float64
	maxRadiusM   *float64
	color        *string
	opacity      *float64
}

// NewReportLayer создаёт новый ReportLayer.
func NewReportLayer(
	id, name, code string,
	description *string,
	geometryType string,
	centerLat, centerLng *float64,
	radiusM, minRadiusM, maxRadiusM *float64,
	color *string,
	opacity *float64,
) ReportLayer {
	return ReportLayer{
		id: id, name: name, code: code,
		description:  description,
		geometryType: geometryType,
		centerLat:    centerLat,
		centerLng:    centerLng,
		radiusM:      radiusM,
		minRadiusM:   minRadiusM,
		maxRadiusM:   maxRadiusM,
		color:        color,
		opacity:      opacity,
	}
}

// ID возвращает ID слоя.
func (l ReportLayer) ID() string { return l.id }

// Name возвращает название слоя.
func (l ReportLayer) Name() string { return l.name }

// Code возвращает код слоя.
func (l ReportLayer) Code() string { return l.code }

// Description возвращает описание слоя.
func (l ReportLayer) Description() *string { return l.description }

// GeometryType возвращает тип геометрии слоя.
func (l ReportLayer) GeometryType() string { return l.geometryType }

// CenterLat возвращает широту центра геометрии.
func (l ReportLayer) CenterLat() *float64 { return l.centerLat }

// CenterLng возвращает долготу центра геометрии.
func (l ReportLayer) CenterLng() *float64 { return l.centerLng }

// RadiusM возвращает радиус круга.
func (l ReportLayer) RadiusM() *float64 { return l.radiusM }

// MinRadiusM возвращает минимальный радиус кольца.
func (l ReportLayer) MinRadiusM() *float64 { return l.minRadiusM }

// MaxRadiusM возвращает максимальный радиус кольца.
func (l ReportLayer) MaxRadiusM() *float64 { return l.maxRadiusM }

// Color возвращает цвет слоя.
func (l ReportLayer) Color() *string { return l.color }

// Opacity возвращает прозрачность слоя.
func (l ReportLayer) Opacity() *float64 { return l.opacity }

// CompositionSummary — сводка по составной установке (value object).
type CompositionSummary struct {
	postType       string
	personnel      string
	accountability string
	armament       string
	weaponUnits    string
	sectorOrRange  string
	azimuth        float64
}

// NewCompositionSummary создаёт новый CompositionSummary.
func NewCompositionSummary(
	postType, personnel, accountability, armament, weaponUnits, sectorOrRange string,
	azimuth float64,
) CompositionSummary {
	return CompositionSummary{
		postType: postType, personnel: personnel,
		accountability: accountability, armament: armament,
		weaponUnits: weaponUnits, sectorOrRange: sectorOrRange,
		azimuth: azimuth,
	}
}

// PostType возвращает тип поста.
func (s CompositionSummary) PostType() string { return s.postType }

// Personnel возвращает личный состав.
func (s CompositionSummary) Personnel() string { return s.personnel }

// Accountability возвращает подотчётность.
func (s CompositionSummary) Accountability() string { return s.accountability }

// Armament возвращает вооружение.
func (s CompositionSummary) Armament() string { return s.armament }

// WeaponUnits возвращает количество орудий/единиц.
func (s CompositionSummary) WeaponUnits() string { return s.weaponUnits }

// SectorOrRange возвращает сектор/дальность.
func (s CompositionSummary) SectorOrRange() string { return s.sectorOrRange }

// Azimuth возвращает азимут.
func (s CompositionSummary) Azimuth() float64 { return s.azimuth }

// WeaponSummary — сводка по оружию (value object).
type WeaponSummary struct {
	caliber        string
	ammunitionType string
	operationMode  string
	moduleCount    string
	isManual       string
}

// NewWeaponSummary создаёт новый WeaponSummary.
func NewWeaponSummary(caliber, ammunitionType, operationMode, moduleCount, isManual string) WeaponSummary {
	return WeaponSummary{
		caliber: caliber, ammunitionType: ammunitionType,
		operationMode: operationMode, moduleCount: moduleCount,
		isManual: isManual,
	}
}

// Caliber возвращает калибр.
func (s WeaponSummary) Caliber() string { return s.caliber }

// AmmunitionType возвращает тип боеприпаса.
func (s WeaponSummary) AmmunitionType() string { return s.ammunitionType }

// OperationMode возвращает режим работы.
func (s WeaponSummary) OperationMode() string { return s.operationMode }

// ModuleCount возвращает количество модулей.
func (s WeaponSummary) ModuleCount() string { return s.moduleCount }

// IsManual возвращает признак ручного управления.
func (s WeaponSummary) IsManual() string { return s.isManual }

// AzimuthSectorSummary — сводка по азимутальному сектору (value object).
type AzimuthSectorSummary struct {
	azimuth      float64
	coverageType string
	angle        *float64
}

// NewAzimuthSectorSummary создаёт новый AzimuthSectorSummary.
func NewAzimuthSectorSummary(azimuth float64, coverageType string, angle *float64) AzimuthSectorSummary {
	return AzimuthSectorSummary{
		azimuth: azimuth, coverageType: coverageType, angle: angle,
	}
}

// Azimuth возвращает азимут.
func (s AzimuthSectorSummary) Azimuth() float64 { return s.azimuth }

// CoverageType возвращает тип зоны покрытия.
func (s AzimuthSectorSummary) CoverageType() string { return s.coverageType }

// Angle возвращает угол сектора.
func (s AzimuthSectorSummary) Angle() *float64 { return s.angle }

// ReportPlacedObject — размещённый объект для отчёта (value object).
type ReportPlacedObject struct {
	id                    string
	assetID               string
	assetName             string
	layerID               string
	layerCode             string
	layerName             string
	lat                   float64
	lng                   float64
	quantity              int
	protectionType        string
	unitPriceMln          float64
	lineTotalMln          float64
	isCompoundPost        bool
	compositionSummary    *CompositionSummary
	weaponSummary         *WeaponSummary
	azimuthSectorSummary  *AzimuthSectorSummary
}

// NewReportPlacedObject создаёт новый ReportPlacedObject.
func NewReportPlacedObject(
	id, assetID, assetName, layerID, layerCode, layerName string,
	lat, lng float64,
	quantity int,
	protectionType string,
	unitPriceMln, lineTotalMln float64,
	isCompoundPost bool,
	compositionSummary *CompositionSummary,
	weaponSummary *WeaponSummary,
	azimuthSectorSummary *AzimuthSectorSummary,
) ReportPlacedObject {
	return ReportPlacedObject{
		id:                   id,
		assetID:              assetID,
		assetName:            assetName,
		layerID:              layerID,
		layerCode:            layerCode,
		layerName:            layerName,
		lat:                  lat,
		lng:                  lng,
		quantity:             quantity,
		protectionType:       protectionType,
		unitPriceMln:         unitPriceMln,
		lineTotalMln:         lineTotalMln,
		isCompoundPost:       isCompoundPost,
		compositionSummary:   compositionSummary,
		weaponSummary:        weaponSummary,
		azimuthSectorSummary: azimuthSectorSummary,
	}
}

// ID возвращает ID размещённого объекта.
func (o ReportPlacedObject) ID() string { return o.id }

// AssetID возвращает ID средства защиты.
func (o ReportPlacedObject) AssetID() string { return o.assetID }

// AssetName возвращает название средства защиты.
func (o ReportPlacedObject) AssetName() string { return o.assetName }

// LayerID возвращает ID слоя.
func (o ReportPlacedObject) LayerID() string { return o.layerID }

// LayerCode возвращает код слоя.
func (o ReportPlacedObject) LayerCode() string { return o.layerCode }

// LayerName возвращает название слоя.
func (o ReportPlacedObject) LayerName() string { return o.layerName }

// Lat возвращает широту.
func (o ReportPlacedObject) Lat() float64 { return o.lat }

// Lng возвращает долготу.
func (o ReportPlacedObject) Lng() float64 { return o.lng }

// Quantity возвращает количество единиц.
func (o ReportPlacedObject) Quantity() int { return o.quantity }

// ProtectionType возвращает тип защиты.
func (o ReportPlacedObject) ProtectionType() string { return o.protectionType }

// UnitPriceMln возвращает цену за единицу в млн ₽.
func (o ReportPlacedObject) UnitPriceMln() float64 { return o.unitPriceMln }

// LineTotalMln возвращает общую стоимость строки в млн ₽.
func (o ReportPlacedObject) LineTotalMln() float64 { return o.lineTotalMln }

// IsCompoundPost возвращает true, если объект является составной установкой (МОГ/ПВН/ГОР/КПП).
func (o ReportPlacedObject) IsCompoundPost() bool { return o.isCompoundPost }

// CompositionSummary возвращает сводку по составной установке.
func (o ReportPlacedObject) CompositionSummary() *CompositionSummary { return o.compositionSummary }

// WeaponSummary возвращает сводку по оружию.
func (o ReportPlacedObject) WeaponSummary() *WeaponSummary { return o.weaponSummary }

// AzimuthSectorSummary возвращает сводку по азимутальному сектору.
func (o ReportPlacedObject) AzimuthSectorSummary() *AzimuthSectorSummary { return o.azimuthSectorSummary }

// ReportPayload — агрегат данных для отчёта GIS MVP.
type ReportPayload struct {
	projectID         string
	projectName       string
	baseObject        ReportBaseObject
	layers            []ReportLayer
	placedObjects     []ReportPlacedObject
	estimate          budgetDomain.CostCalculation
	structuralProfile budgetDomain.StructuralProfile
	hideCost          bool
}

// NewReportPayload создаёт новый ReportPayload.
func NewReportPayload(
	projectID, projectName string,
	baseObject ReportBaseObject,
	layers []ReportLayer,
	placedObjects []ReportPlacedObject,
	estimate budgetDomain.CostCalculation,
	structuralProfile budgetDomain.StructuralProfile,
	hideCost bool,
) *ReportPayload {
	if layers == nil {
		layers = []ReportLayer{}
	}
	if placedObjects == nil {
		placedObjects = []ReportPlacedObject{}
	}
	return &ReportPayload{
		projectID:         projectID,
		projectName:       projectName,
		baseObject:        baseObject,
		layers:            layers,
		placedObjects:     placedObjects,
		estimate:          estimate,
		structuralProfile: structuralProfile,
		hideCost:          hideCost,
	}
}

// ProjectID возвращает ID проекта.
func (p *ReportPayload) ProjectID() string { return p.projectID }

// ProjectName возвращает название проекта.
func (p *ReportPayload) ProjectName() string { return p.projectName }

// BaseObject возвращает защищаемый объект.
func (p *ReportPayload) BaseObject() ReportBaseObject { return p.baseObject }

// Layers возвращает слои защиты.
func (p *ReportPayload) Layers() []ReportLayer { return p.layers }

// PlacedObjects возвращает размещённые объекты.
func (p *ReportPayload) PlacedObjects() []ReportPlacedObject { return p.placedObjects }

// Estimate возвращает расчёт стоимости.
func (p *ReportPayload) Estimate() budgetDomain.CostCalculation { return p.estimate }

// StructuralProfile возвращает структурный профиль.
func (p *ReportPayload) StructuralProfile() budgetDomain.StructuralProfile { return p.structuralProfile }

// HideCost возвращает флаг скрытия стоимости.
func (p *ReportPayload) HideCost() bool { return p.hideCost }
