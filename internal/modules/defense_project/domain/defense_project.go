package domain

import "time"

// SchemaVersion текущая версия схемы DefenseProject.
const SchemaVersion = 1

// Coordinates — value object для географических координат.
type Coordinates struct {
	lat float64
	lng float64
}

func NewCoordinates(lat, lng float64) Coordinates {
	return Coordinates{lat: lat, lng: lng}
}

func (c Coordinates) Lat() float64 { return c.lat }
func (c Coordinates) Lng() float64 { return c.lng }

// ProtectedObject — value object защищаемого объекта.
type ProtectedObject struct {
	id     string
	name   string
	center Coordinates
}

func NewProtectedObject(id, name string, center Coordinates) ProtectedObject {
	return ProtectedObject{id: id, name: name, center: center}
}

func (p ProtectedObject) ID() string        { return p.id }
func (p ProtectedObject) Name() string      { return p.name }
func (p ProtectedObject) Center() Coordinates { return p.center }

// LayerGeometryType — тип геометрии слоя.
type LayerGeometryType string

const (
	LayerGeometryCircle   LayerGeometryType = "circle"
	LayerGeometryRing     LayerGeometryType = "ring"
	LayerGeometryPolygon  LayerGeometryType = "polygon"
	LayerGeometryFreeform LayerGeometryType = "freeform"
)

// LayerGeometry — value object геометрии слоя (discriminated union).
type LayerGeometry struct {
	geometryType LayerGeometryType
	center       *Coordinates
	radiusM      *float64
	minRadiusM   *float64
	maxRadiusM   *float64
	points       []Coordinates
}

func NewCircleGeometry(center Coordinates, radiusM float64) LayerGeometry {
	c := center
	return LayerGeometry{
		geometryType: LayerGeometryCircle,
		center:       &c,
		radiusM:      &radiusM,
	}
}

func NewRingGeometry(center Coordinates, minRadiusM, maxRadiusM float64) LayerGeometry {
	c := center
	return LayerGeometry{
		geometryType: LayerGeometryRing,
		center:       &c,
		minRadiusM:   &minRadiusM,
		maxRadiusM:   &maxRadiusM,
	}
}

func NewPolygonGeometry(points []Coordinates) LayerGeometry {
	return LayerGeometry{
		geometryType: LayerGeometryPolygon,
		points:       points,
	}
}

func NewFreeformGeometry(points []Coordinates) LayerGeometry {
	return LayerGeometry{
		geometryType: LayerGeometryFreeform,
		points:       points,
	}
}

func (g LayerGeometry) Type() LayerGeometryType  { return g.geometryType }
func (g LayerGeometry) Center() *Coordinates      { return g.center }
func (g LayerGeometry) RadiusM() *float64         { return g.radiusM }
func (g LayerGeometry) MinRadiusM() *float64      { return g.minRadiusM }
func (g LayerGeometry) MaxRadiusM() *float64      { return g.maxRadiusM }
func (g LayerGeometry) Points() []Coordinates     { return g.points }

// EditableDefenseLayer — value object редактируемого слоя (эшелона) защиты.
type EditableDefenseLayer struct {
	id               string
	name             string
	code             string
	description      *string
	order            int
	distanceFromObjMin *float64
	distanceFromObjMax *float64
	geometryType     LayerGeometryType
	geometry         LayerGeometry
	color            *string
	opacity          *float64
	isActive         bool
	isVisible        bool
	isLocked         bool
}

func NewEditableDefenseLayer(
	id, name, code string,
	description *string,
	order int,
	distanceFromObjMin, distanceFromObjMax *float64,
	geometryType LayerGeometryType,
	geometry LayerGeometry,
	color *string,
	opacity *float64,
	isActive, isVisible, isLocked bool,
) EditableDefenseLayer {
	return EditableDefenseLayer{
		id: id, name: name, code: code,
		description:        description,
		order:              order,
		distanceFromObjMin: distanceFromObjMin,
		distanceFromObjMax: distanceFromObjMax,
		geometryType:       geometryType,
		geometry:           geometry,
		color:              color,
		opacity:            opacity,
		isActive:           isActive,
		isVisible:          isVisible,
		isLocked:           isLocked,
	}
}

func (l EditableDefenseLayer) ID() string                { return l.id }
func (l EditableDefenseLayer) Name() string              { return l.name }
func (l EditableDefenseLayer) Code() string              { return l.code }
func (l EditableDefenseLayer) Description() *string      { return l.description }
func (l EditableDefenseLayer) Order() int                { return l.order }
func (l EditableDefenseLayer) DistanceFromObjectMin() *float64 { return l.distanceFromObjMin }
func (l EditableDefenseLayer) DistanceFromObjectMax() *float64 { return l.distanceFromObjMax }
func (l EditableDefenseLayer) GeometryType() LayerGeometryType  { return l.geometryType }
func (l EditableDefenseLayer) Geometry() LayerGeometry          { return l.geometry }
func (l EditableDefenseLayer) Color() *string           { return l.color }
func (l EditableDefenseLayer) Opacity() *float64        { return l.opacity }
func (l EditableDefenseLayer) IsActive() bool           { return l.isActive }
func (l EditableDefenseLayer) IsVisible() bool          { return l.isVisible }
func (l EditableDefenseLayer) IsLocked() bool           { return l.isLocked }

// DefenseAssetCategory — категория средства защиты.
type DefenseAssetCategory string

const (
	DefenseAssetCategoryEarlyWarning        DefenseAssetCategory = "early-warning"
	DefenseAssetCategoryDetection           DefenseAssetCategory = "detection"
	DefenseAssetCategoryClassification      DefenseAssetCategory = "classification"
	DefenseAssetCategoryJamming             DefenseAssetCategory = "jamming"
	DefenseAssetCategorySpoofing            DefenseAssetCategory = "spoofing"
	DefenseAssetCategoryKinetic             DefenseAssetCategory = "kinetic"
	DefenseAssetCategoryInterceptor        DefenseAssetCategory = "interceptor"
	DefenseAssetCategoryPassiveProtection   DefenseAssetCategory = "passive-protection"
	DefenseAssetCategoryEngineeringProtection DefenseAssetCategory = "engineering-protection"
	DefenseAssetCategoryInfrastructure      DefenseAssetCategory = "infrastructure"
	DefenseAssetCategorySoftware            DefenseAssetCategory = "software"
	DefenseAssetCategoryCommandCenter       DefenseAssetCategory = "command-center"
	DefenseAssetCategoryExternalService     DefenseAssetCategory = "external-service"
)

// DefenseAssetRole — роль средства защиты.
type DefenseAssetRole string

const (
	DefenseAssetRoleDetect     DefenseAssetRole = "detect"
	DefenseAssetRoleTrack      DefenseAssetRole = "track"
	DefenseAssetRoleClassify   DefenseAssetRole = "classify"
	DefenseAssetRoleSuppress   DefenseAssetRole = "suppress"
	DefenseAssetRoleDestroy    DefenseAssetRole = "destroy"
	DefenseAssetRoleDelay      DefenseAssetRole = "delay"
	DefenseAssetRoleProtect    DefenseAssetRole = "protect"
	DefenseAssetRoleCoordinate DefenseAssetRole = "coordinate"
	DefenseAssetRoleMonitor    DefenseAssetRole = "monitor"
	DefenseAssetRoleAlert      DefenseAssetRole = "alert"
)

// DefenseAssetCoverageType — тип зоны покрытия средства.
type DefenseAssetCoverageType string

const (
	DefenseAssetCoverageCircle  DefenseAssetCoverageType = "circle"
	DefenseAssetCoverageSector  DefenseAssetCoverageType = "sector"
	DefenseAssetCoverageLine    DefenseAssetCoverageType = "line"
	DefenseAssetCoveragePolygon DefenseAssetCoverageType = "polygon"
	DefenseAssetCoverageNone    DefenseAssetCoverageType = "none"
)

// DefenseAssetDeploymentType — тип размещения средства.
type DefenseAssetDeploymentType string

const (
	DefenseAssetDeploymentStatic         DefenseAssetDeploymentType = "static"
	DefenseAssetDeploymentMobile         DefenseAssetDeploymentType = "mobile"
	DefenseAssetDeploymentInfrastructure DefenseAssetDeploymentType = "infrastructure"
	DefenseAssetDeploymentSoftware       DefenseAssetDeploymentType = "software"
	DefenseAssetDeploymentExternal       DefenseAssetDeploymentType = "external"
)

// DefenseAssetPlacementType — тип расположения средства на карте.
type DefenseAssetPlacementType string

const (
	DefenseAssetPlacementMapObject    DefenseAssetPlacementType = "map-object"
	DefenseAssetPlacementZoneObject   DefenseAssetPlacementType = "zone-object"
	DefenseAssetPlacementNonPhysical  DefenseAssetPlacementType = "non-physical"
)

// DefensePriority — приоритет средства защиты.
type DefensePriority int

// DefenseAsset — value object средства защиты.
type DefenseAsset struct {
	id                   string
	name                 string
	shortName            *string
	description          *string
	category             DefenseAssetCategory
	roles                []DefenseAssetRole
	pricePerUnitMln      *float64
	currency             string
	unitLabel            string
	compatibleLayerTypes []LayerGeometryType
	recommendedLayerCodes []string
	compatibleLayerCodes  []string
	incompatibleLayerCodes []string
	minEffectiveDistance  *float64
	maxEffectiveDistance  *float64
	coverageType          DefenseAssetCoverageType
	coverageRadius        *float64
	coverageAngle         *float64
	deploymentType        DefenseAssetDeploymentType
	placementType         DefenseAssetPlacementType
	iconURL               *string
	modelURL              *string
	score                 *int
	priority              *DefensePriority
	tags                  []string
	legacyItemID          *string
	calculatorAssetID     *string
	mapCatalogGroupIDs    []string
}

func NewDefenseAsset(
	id, name string,
	shortName, description *string,
	category DefenseAssetCategory,
	roles []DefenseAssetRole,
	pricePerUnitMln *float64,
	currency, unitLabel string,
	compatibleLayerTypes []LayerGeometryType,
	recommendedLayerCodes, compatibleLayerCodes, incompatibleLayerCodes []string,
	minEffectiveDistance, maxEffectiveDistance *float64,
	coverageType DefenseAssetCoverageType,
	coverageRadius, coverageAngle *float64,
	deploymentType DefenseAssetDeploymentType,
	placementType DefenseAssetPlacementType,
	iconURL, modelURL *string,
	score *int,
	priority *DefensePriority,
	tags []string,
	legacyItemID, calculatorAssetID *string,
	mapCatalogGroupIDs []string,
) DefenseAsset {
	return DefenseAsset{
		id: id, name: name,
		shortName: shortName, description: description,
		category: category, roles: roles,
		pricePerUnitMln: pricePerUnitMln,
		currency: currency, unitLabel: unitLabel,
		compatibleLayerTypes:   compatibleLayerTypes,
		recommendedLayerCodes:  recommendedLayerCodes,
		compatibleLayerCodes:   compatibleLayerCodes,
		incompatibleLayerCodes: incompatibleLayerCodes,
		minEffectiveDistance:   minEffectiveDistance,
		maxEffectiveDistance:   maxEffectiveDistance,
		coverageType:           coverageType,
		coverageRadius:         coverageRadius,
		coverageAngle:          coverageAngle,
		deploymentType:         deploymentType,
		placementType:          placementType,
		iconURL: iconURL, modelURL: modelURL,
		score: score, priority: priority,
		tags: tags,
		legacyItemID: legacyItemID, calculatorAssetID: calculatorAssetID,
		mapCatalogGroupIDs: mapCatalogGroupIDs,
	}
}

// Getters for DefenseAsset.
func (a DefenseAsset) ID() string                        { return a.id }
func (a DefenseAsset) Name() string                      { return a.name }
func (a DefenseAsset) ShortName() *string                { return a.shortName }
func (a DefenseAsset) Description() *string              { return a.description }
func (a DefenseAsset) Category() DefenseAssetCategory     { return a.category }
func (a DefenseAsset) Roles() []DefenseAssetRole          { return a.roles }
func (a DefenseAsset) PricePerUnitMln() *float64         { return a.pricePerUnitMln }
func (a DefenseAsset) Currency() string                  { return a.currency }
func (a DefenseAsset) UnitLabel() string                 { return a.unitLabel }
func (a DefenseAsset) CompatibleLayerTypes() []LayerGeometryType { return a.compatibleLayerTypes }
func (a DefenseAsset) RecommendedLayerCodes() []string   { return a.recommendedLayerCodes }
func (a DefenseAsset) CompatibleLayerCodes() []string    { return a.compatibleLayerCodes }
func (a DefenseAsset) IncompatibleLayerCodes() []string  { return a.incompatibleLayerCodes }
func (a DefenseAsset) MinEffectiveDistance() *float64    { return a.minEffectiveDistance }
func (a DefenseAsset) MaxEffectiveDistance() *float64    { return a.maxEffectiveDistance }
func (a DefenseAsset) CoverageType() DefenseAssetCoverageType  { return a.coverageType }
func (a DefenseAsset) CoverageRadius() *float64          { return a.coverageRadius }
func (a DefenseAsset) CoverageAngle() *float64           { return a.coverageAngle }
func (a DefenseAsset) DeploymentType() DefenseAssetDeploymentType { return a.deploymentType }
func (a DefenseAsset) PlacementType() DefenseAssetPlacementType  { return a.placementType }
func (a DefenseAsset) IconURL() *string                  { return a.iconURL }
func (a DefenseAsset) ModelURL() *string                 { return a.modelURL }
func (a DefenseAsset) Score() *int                       { return a.score }
func (a DefenseAsset) Priority() *DefensePriority        { return a.priority }
func (a DefenseAsset) Tags() []string                    { return a.tags }
func (a DefenseAsset) LegacyItemID() *string             { return a.legacyItemID }
func (a DefenseAsset) CalculatorAssetID() *string        { return a.calculatorAssetID }
func (a DefenseAsset) MapCatalogGroupIDs() []string      { return a.mapCatalogGroupIDs }

// PlacedObjectStatus — статус размещённого объекта.
type PlacedObjectStatus string

const (
	PlacedObjectStatusPlanned     PlacedObjectStatus = "planned"
	PlacedObjectStatusActive      PlacedObjectStatus = "active"
	PlacedObjectStatusInactive    PlacedObjectStatus = "inactive"
	PlacedObjectStatusMaintenance PlacedObjectStatus = "maintenance"
)

// PlacedDefenseObject — value object размещённого объекта средства защиты.
type PlacedDefenseObject struct {
	id                  string
	assetID             string
	layerID             string
	name                *string
	coordinates         Coordinates
	rotation            *float64
	scale               *float64
	quantity            int
	status              PlacedObjectStatus
	customPricePerUnitMln *float64
	customCoverageRadius  *float64
	customCoverageAngle   *float64
	hasGeometryConflict   bool
	hasCoverageConflict   bool
	hasTerrainConflict    bool
	notes               *string
	createdAt           time.Time
	updatedAt           time.Time
}

func NewPlacedDefenseObject(
	id, assetID, layerID string,
	name *string,
	coordinates Coordinates,
	rotation, scale *float64,
	quantity int,
	status PlacedObjectStatus,
	customPricePerUnitMln, customCoverageRadius, customCoverageAngle *float64,
	hasGeometryConflict, hasCoverageConflict, hasTerrainConflict bool,
	notes *string,
	createdAt, updatedAt time.Time,
) PlacedDefenseObject {
	return PlacedDefenseObject{
		id: id, assetID: assetID, layerID: layerID,
		name: name, coordinates: coordinates,
		rotation: rotation, scale: scale,
		quantity: quantity, status: status,
		customPricePerUnitMln: customPricePerUnitMln,
		customCoverageRadius:  customCoverageRadius,
		customCoverageAngle:   customCoverageAngle,
		hasGeometryConflict:   hasGeometryConflict,
		hasCoverageConflict:   hasCoverageConflict,
		hasTerrainConflict:    hasTerrainConflict,
		notes: notes,
		createdAt: createdAt, updatedAt: updatedAt,
	}
}

// Getters for PlacedDefenseObject.
func (o PlacedDefenseObject) ID() string                     { return o.id }
func (o PlacedDefenseObject) AssetID() string                { return o.assetID }
func (o PlacedDefenseObject) LayerID() string                { return o.layerID }
func (o PlacedDefenseObject) Name() *string                  { return o.name }
func (o PlacedDefenseObject) Coordinates() Coordinates        { return o.coordinates }
func (o PlacedDefenseObject) Rotation() *float64             { return o.rotation }
func (o PlacedDefenseObject) Scale() *float64                { return o.scale }
func (o PlacedDefenseObject) Quantity() int                  { return o.quantity }
func (o PlacedDefenseObject) Status() PlacedObjectStatus     { return o.status }
func (o PlacedDefenseObject) CustomPricePerUnitMln() *float64 { return o.customPricePerUnitMln }
func (o PlacedDefenseObject) CustomCoverageRadius() *float64 { return o.customCoverageRadius }
func (o PlacedDefenseObject) CustomCoverageAngle() *float64  { return o.customCoverageAngle }
func (o PlacedDefenseObject) HasGeometryConflict() bool      { return o.hasGeometryConflict }
func (o PlacedDefenseObject) HasCoverageConflict() bool      { return o.hasCoverageConflict }
func (o PlacedDefenseObject) HasTerrainConflict() bool       { return o.hasTerrainConflict }
func (o PlacedDefenseObject) Notes() *string                 { return o.notes }
func (o PlacedDefenseObject) CreatedAt() time.Time            { return o.createdAt }
func (o PlacedDefenseObject) UpdatedAt() time.Time            { return o.updatedAt }

// DefenseProjectMode — режим работы с проектом.
type DefenseProjectMode string

const (
	DefenseProjectModeView        DefenseProjectMode = "view"
	DefenseProjectModeEditLayers  DefenseProjectMode = "edit-layers"
	DefenseProjectModePlaceObject DefenseProjectMode = "place-object"
	DefenseProjectModeMoveObject  DefenseProjectMode = "move-object"
	DefenseProjectModeMeasure     DefenseProjectMode = "measure"
)

// DefenseProjectSource — источник создания проекта.
type DefenseProjectSource string

const (
	DefenseProjectSourceCustom         DefenseProjectSource = "custom"
	DefenseProjectSourcePreset         DefenseProjectSource = "preset"
	DefenseProjectSourceLegacyMigration DefenseProjectSource = "legacy-migration"
)

// DefenseProject — aggregate корень для проекта защиты.
type DefenseProject struct {
	schemaVersion  int
	projectID      string
	projectName    string
	baseObject     ProtectedObject
	layers         []EditableDefenseLayer
	assetLibrary   []DefenseAsset
	placedObjects  []PlacedDefenseObject
	activeLayerID  *string
	selectedAssetID *string
	selectedObjectID *string
	mode           DefenseProjectMode
	source         DefenseProjectSource
	basePresetID   *string
	updatedAt      time.Time
}

// NewDefenseProject создаёт новый DefenseProject с валидацией.
func NewDefenseProject(
	projectID, projectName string,
	baseObject ProtectedObject,
	layers []EditableDefenseLayer,
	assetLibrary []DefenseAsset,
	placedObjects []PlacedDefenseObject,
	activeLayerID, selectedAssetID, selectedObjectID *string,
	mode DefenseProjectMode,
	source DefenseProjectSource,
	basePresetID *string,
	updatedAt time.Time,
) (*DefenseProject, error) {
	if projectID == "" {
		return nil, ErrInvalidProjectData
	}
	if projectName == "" {
		return nil, ErrInvalidProjectData
	}
	if layers == nil {
		layers = []EditableDefenseLayer{}
	}
	if assetLibrary == nil {
		assetLibrary = []DefenseAsset{}
	}
	if placedObjects == nil {
		placedObjects = []PlacedDefenseObject{}
	}

	return &DefenseProject{
		schemaVersion:    SchemaVersion,
		projectID:        projectID,
		projectName:      projectName,
		baseObject:       baseObject,
		layers:           layers,
		assetLibrary:     assetLibrary,
		placedObjects:    placedObjects,
		activeLayerID:    activeLayerID,
		selectedAssetID:  selectedAssetID,
		selectedObjectID: selectedObjectID,
		mode:             mode,
		source:           source,
		basePresetID:     basePresetID,
		updatedAt:        updatedAt,
	}, nil
}

// Getters for DefenseProject.
func (p *DefenseProject) SchemaVersion() int                       { return p.schemaVersion }
func (p *DefenseProject) ProjectID() string                       { return p.projectID }
func (p *DefenseProject) ProjectName() string                     { return p.projectName }
func (p *DefenseProject) BaseObject() ProtectedObject              { return p.baseObject }
func (p *DefenseProject) Layers() []EditableDefenseLayer           { return p.layers }
func (p *DefenseProject) AssetLibrary() []DefenseAsset            { return p.assetLibrary }
func (p *DefenseProject) PlacedObjects() []PlacedDefenseObject    { return p.placedObjects }
func (p *DefenseProject) ActiveLayerID() *string                  { return p.activeLayerID }
func (p *DefenseProject) SelectedAssetID() *string                { return p.selectedAssetID }
func (p *DefenseProject) SelectedObjectID() *string               { return p.selectedObjectID }
func (p *DefenseProject) Mode() DefenseProjectMode                { return p.mode }
func (p *DefenseProject) Source() DefenseProjectSource            { return p.source }
func (p *DefenseProject) BasePresetID() *string                   { return p.basePresetID }
func (p *DefenseProject) UpdatedAt() time.Time                    { return p.updatedAt }

// SetProjectID обновляет ID проекта (используется при сохранении).
func (p *DefenseProject) SetProjectID(id string) {
	p.projectID = id
}

// SetUpdatedAt обновляет время последнего изменения.
func (p *DefenseProject) SetUpdatedAt(t time.Time) {
	p.updatedAt = t
}
