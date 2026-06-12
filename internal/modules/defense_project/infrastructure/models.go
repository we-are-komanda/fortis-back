package infrastructure

import (
	"encoding/json"
	"time"

	"github.com/fortis/backend/internal/modules/defense_project/domain"
)

// DefenseProjectModel — GORM-модель для хранения DefenseProject.
type DefenseProjectModel struct {
	ID          string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ProjectData string    `gorm:"type:jsonb;not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func (DefenseProjectModel) TableName() string {
	return "defense_projects"
}

// projectDataJSON — промежуточная структура для сериализации/десериализации JSONB.
// Содержит все поля DefenseProject, включая schemaVersion.
type projectDataJSON struct {
	SchemaVersion    int                              `json:"schemaVersion"`
	ProjectID        string                           `json:"projectId"`
	ProjectName      string                           `json:"projectName"`
	BaseObject       protectedObjectJSON              `json:"baseObject"`
	Layers           []editableDefenseLayerJSON       `json:"layers"`
	AssetLibrary     []defenseAssetJSON               `json:"assetLibrary"`
	PlacedObjects    []placedDefenseObjectJSON        `json:"placedObjects"`
	ActiveLayerID    *string                          `json:"activeLayerId,omitempty"`
	SelectedAssetID  *string                          `json:"selectedAssetId,omitempty"`
	SelectedObjectID *string                          `json:"selectedObjectId,omitempty"`
	Mode             string                           `json:"mode"`
	Source           string                           `json:"source,omitempty"`
	BasePresetID     *string                          `json:"basePresetId,omitempty"`
	UpdatedAt        string                           `json:"updatedAt"`
}

type coordinatesJSON struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type protectedObjectJSON struct {
	ID     string          `json:"id"`
	Name   string          `json:"name"`
	Center coordinatesJSON `json:"center"`
}

type layerGeometryJSON struct {
	Type       string            `json:"type"`
	Center     *coordinatesJSON  `json:"center,omitempty"`
	RadiusM    *float64          `json:"radiusM,omitempty"`
	MinRadiusM *float64          `json:"minRadiusM,omitempty"`
	MaxRadiusM *float64          `json:"maxRadiusM,omitempty"`
	Points     []coordinatesJSON `json:"points,omitempty"`
}

type editableDefenseLayerJSON struct {
	ID                 string             `json:"id"`
	Name               string             `json:"name"`
	Code               string             `json:"code"`
	Description        *string            `json:"description,omitempty"`
	Order              int                `json:"order"`
	DistanceFromObjMin *float64           `json:"distanceFromObjectMin,omitempty"`
	DistanceFromObjMax *float64           `json:"distanceFromObjectMax,omitempty"`
	GeometryType       string             `json:"geometryType"`
	Geometry           layerGeometryJSON  `json:"geometry"`
	Color              *string            `json:"color,omitempty"`
	Opacity            *float64           `json:"opacity,omitempty"`
	IsActive           bool               `json:"isActive"`
	IsVisible          *bool              `json:"isVisible,omitempty"`
	IsLocked           *bool              `json:"isLocked,omitempty"`
}

type defenseAssetJSON struct {
	ID                    string   `json:"id"`
	Name                  string   `json:"name"`
	ShortName             *string  `json:"shortName,omitempty"`
	Description           *string  `json:"description,omitempty"`
	Category              string   `json:"category"`
	Roles                 []string `json:"roles"`
	PricePerUnitMln       *float64 `json:"pricePerUnitMln,omitempty"`
	Currency              string   `json:"currency"`
	UnitLabel             string   `json:"unitLabel"`
	CompatibleLayerTypes  []string `json:"compatibleLayerTypes,omitempty"`
	RecommendedLayerCodes []string `json:"recommendedLayerCodes,omitempty"`
	CompatibleLayerCodes  []string `json:"compatibleLayerCodes,omitempty"`
	IncompatibleLayerCodes []string `json:"incompatibleLayerCodes,omitempty"`
	MinEffectiveDistance  *float64 `json:"minEffectiveDistance,omitempty"`
	MaxEffectiveDistance  *float64 `json:"maxEffectiveDistance,omitempty"`
	CoverageType          string   `json:"coverageType"`
	CoverageRadius        *float64 `json:"coverageRadius,omitempty"`
	CoverageAngle         *float64 `json:"coverageAngle,omitempty"`
	DeploymentType        string   `json:"deploymentType"`
	PlacementType         string   `json:"placementType"`
	IconURL               *string  `json:"iconUrl,omitempty"`
	ModelURL              *string  `json:"modelUrl,omitempty"`
	Score                 *int     `json:"score,omitempty"`
	Priority              *int     `json:"priority,omitempty"`
	Tags                  []string `json:"tags,omitempty"`
	LegacyItemID          *string  `json:"legacyItemId,omitempty"`
	CalculatorAssetID     *string  `json:"calculatorAssetId,omitempty"`
	MapCatalogGroupIDs    []string `json:"mapCatalogGroupIds,omitempty"`
}

type placedDefenseObjectJSON struct {
	ID                    string           `json:"id"`
	AssetID               string           `json:"assetId"`
	LayerID               string           `json:"layerId"`
	Name                  *string          `json:"name,omitempty"`
	Coordinates           coordinatesJSON  `json:"coordinates"`
	Rotation              *float64         `json:"rotation,omitempty"`
	Scale                 *float64         `json:"scale,omitempty"`
	Quantity              int              `json:"quantity"`
	Status                string           `json:"status"`
	CustomPricePerUnitMln *float64         `json:"customPricePerUnitMln,omitempty"`
	CustomCoverageRadius  *float64         `json:"customCoverageRadius,omitempty"`
	CustomCoverageAngle   *float64         `json:"customCoverageAngle,omitempty"`
	HasGeometryConflict   bool             `json:"hasGeometryConflict"`
	HasCoverageConflict   bool             `json:"hasCoverageConflict"`
	HasTerrainConflict    bool             `json:"hasTerrainConflict"`
	Notes                 *string          `json:"notes,omitempty"`
	CreatedAt             string           `json:"createdAt"`
	UpdatedAt             string           `json:"updatedAt"`
}

// ToDomain преобразует GORM-модель в доменный агрегат.
func (m *DefenseProjectModel) ToDomain() (*domain.DefenseProject, error) {
	var data projectDataJSON
	if err := json.Unmarshal([]byte(m.ProjectData), &data); err != nil {
		return nil, err
	}

	project := jsonToDomain(data)
	project.SetProjectID(m.ID)
	return project, nil
}

// ToModel преобразует доменный агрегат в GORM-модель.
func ToModel(p *domain.DefenseProject) (*DefenseProjectModel, error) {
	data := domainToJSON(p)
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &DefenseProjectModel{
		ID:          p.ProjectID(),
		ProjectData: string(raw),
		UpdatedAt:   p.UpdatedAt(),
	}, nil
}

// domainToJSON преобразует доменный объект в JSON-структуру для сериализации.
func domainToJSON(p *domain.DefenseProject) projectDataJSON {
	layers := make([]editableDefenseLayerJSON, len(p.Layers()))
	for i, l := range p.Layers() {
		layers[i] = layerToJSON(l)
	}

	assets := make([]defenseAssetJSON, len(p.AssetLibrary()))
	for i, a := range p.AssetLibrary() {
		assets[i] = assetToJSON(a)
	}

	objects := make([]placedDefenseObjectJSON, len(p.PlacedObjects()))
	for i, o := range p.PlacedObjects() {
		objects[i] = placedObjectToJSON(o)
	}

	return projectDataJSON{
		SchemaVersion:    p.SchemaVersion(),
		ProjectID:        p.ProjectID(),
		ProjectName:      p.ProjectName(),
		BaseObject:       protectedObjectToJSON(p.BaseObject()),
		Layers:           layers,
		AssetLibrary:     assets,
		PlacedObjects:    objects,
		ActiveLayerID:    p.ActiveLayerID(),
		SelectedAssetID:  p.SelectedAssetID(),
		SelectedObjectID: p.SelectedObjectID(),
		Mode:             string(p.Mode()),
		Source:           string(p.Source()),
		BasePresetID:     p.BasePresetID(),
		UpdatedAt:        p.UpdatedAt().Format(time.RFC3339Nano),
	}
}

// jsonToDomain преобразует JSON-структуру в доменный объект.
func jsonToDomain(data projectDataJSON) *domain.DefenseProject {
	baseObj := domain.NewProtectedObject(
		data.BaseObject.ID,
		data.BaseObject.Name,
		domain.NewCoordinates(data.BaseObject.Center.Lat, data.BaseObject.Center.Lng),
	)

	layers := make([]domain.EditableDefenseLayer, len(data.Layers))
	for i, l := range data.Layers {
		layers[i] = jsonToLayer(l)
	}

	assets := make([]domain.DefenseAsset, len(data.AssetLibrary))
	for i, a := range data.AssetLibrary {
		assets[i] = jsonToAsset(a)
	}

	objects := make([]domain.PlacedDefenseObject, len(data.PlacedObjects))
	for i, o := range data.PlacedObjects {
		objects[i] = jsonToPlacedObject(o)
	}

	updatedAt, _ := time.Parse(time.RFC3339Nano, data.UpdatedAt)

	//nolint:errcheck // валидация уже пройдена на уровне сервиса
	project, _ := domain.NewDefenseProject(
		data.ProjectID, data.ProjectName,
		baseObj, layers, assets, objects,
		data.ActiveLayerID, data.SelectedAssetID, data.SelectedObjectID,
		domain.DefenseProjectMode(data.Mode),
		domain.DefenseProjectSource(data.Source),
		data.BasePresetID,
		updatedAt,
	)
	return project
}

func layerToJSON(l domain.EditableDefenseLayer) editableDefenseLayerJSON {
	geom := l.Geometry()
	geomJSON := layerGeometryJSON{Type: string(geom.Type())}

	switch geom.Type() {
	case domain.LayerGeometryCircle:
		if c := geom.Center(); c != nil {
			geomJSON.Center = &coordinatesJSON{Lat: c.Lat(), Lng: c.Lng()}
		}
		geomJSON.RadiusM = geom.RadiusM()
	case domain.LayerGeometryRing:
		if c := geom.Center(); c != nil {
			geomJSON.Center = &coordinatesJSON{Lat: c.Lat(), Lng: c.Lng()}
		}
		geomJSON.MinRadiusM = geom.MinRadiusM()
		geomJSON.MaxRadiusM = geom.MaxRadiusM()
	case domain.LayerGeometryPolygon, domain.LayerGeometryFreeform:
		pts := geom.Points()
		geomJSON.Points = make([]coordinatesJSON, len(pts))
		for i, p := range pts {
			geomJSON.Points[i] = coordinatesJSON{Lat: p.Lat(), Lng: p.Lng()}
		}
	}

	isVisible := l.IsVisible()
	isLocked := l.IsLocked()

	return editableDefenseLayerJSON{
		ID: l.ID(), Name: l.Name(), Code: l.Code(),
		Description:        l.Description(),
		Order:              l.Order(),
		DistanceFromObjMin: l.DistanceFromObjectMin(),
		DistanceFromObjMax: l.DistanceFromObjectMax(),
		GeometryType:       string(l.GeometryType()),
		Geometry:           geomJSON,
		Color:              l.Color(),
		Opacity:            l.Opacity(),
		IsActive:           l.IsActive(),
		IsVisible:          &isVisible,
		IsLocked:           &isLocked,
	}
}

func jsonToLayer(l editableDefenseLayerJSON) domain.EditableDefenseLayer {
	return domain.NewEditableDefenseLayer(
		l.ID, l.Name, l.Code, l.Description,
		l.Order,
		l.DistanceFromObjMin, l.DistanceFromObjMax,
		domain.LayerGeometryType(l.GeometryType),
		jsonToGeometry(l.Geometry),
		l.Color, l.Opacity,
		l.IsActive,
		boolOrTrue(l.IsVisible),
		boolOrFalse(l.IsLocked),
	)
}

func jsonToGeometry(g layerGeometryJSON) domain.LayerGeometry {
	switch domain.LayerGeometryType(g.Type) {
	case domain.LayerGeometryCircle:
		center := domain.NewCoordinates(g.Center.Lat, g.Center.Lng)
		return domain.NewCircleGeometry(center, *g.RadiusM)
	case domain.LayerGeometryRing:
		center := domain.NewCoordinates(g.Center.Lat, g.Center.Lng)
		return domain.NewRingGeometry(center, *g.MinRadiusM, *g.MaxRadiusM)
	case domain.LayerGeometryPolygon:
		pts := make([]domain.Coordinates, len(g.Points))
		for i, p := range g.Points {
			pts[i] = domain.NewCoordinates(p.Lat, p.Lng)
		}
		return domain.NewPolygonGeometry(pts)
	case domain.LayerGeometryFreeform:
		pts := make([]domain.Coordinates, len(g.Points))
		for i, p := range g.Points {
			pts[i] = domain.NewCoordinates(p.Lat, p.Lng)
		}
		return domain.NewFreeformGeometry(pts)
	default:
		return domain.NewCircleGeometry(domain.NewCoordinates(0, 0), 0)
	}
}

func assetToJSON(a domain.DefenseAsset) defenseAssetJSON {
	roles := make([]string, len(a.Roles()))
	for i, r := range a.Roles() {
		roles[i] = string(r)
	}

	compatTypes := make([]string, len(a.CompatibleLayerTypes()))
	for i, t := range a.CompatibleLayerTypes() {
		compatTypes[i] = string(t)
	}

	var priority *int
	if p := a.Priority(); p != nil {
		v := int(*p)
		priority = &v
	}

	return defenseAssetJSON{
		ID: a.ID(), Name: a.Name(),
		ShortName: a.ShortName(), Description: a.Description(),
		Category:              string(a.Category()),
		Roles:                 roles,
		PricePerUnitMln:       a.PricePerUnitMln(),
		Currency:              a.Currency(),
		UnitLabel:             a.UnitLabel(),
		CompatibleLayerTypes:  compatTypes,
		RecommendedLayerCodes: a.RecommendedLayerCodes(),
		CompatibleLayerCodes:  a.CompatibleLayerCodes(),
		IncompatibleLayerCodes: a.IncompatibleLayerCodes(),
		MinEffectiveDistance:  a.MinEffectiveDistance(),
		MaxEffectiveDistance:  a.MaxEffectiveDistance(),
		CoverageType:          string(a.CoverageType()),
		CoverageRadius:        a.CoverageRadius(),
		CoverageAngle:         a.CoverageAngle(),
		DeploymentType:        string(a.DeploymentType()),
		PlacementType:         string(a.PlacementType()),
		IconURL:               a.IconURL(),
		ModelURL:              a.ModelURL(),
		Score:                 a.Score(),
		Priority:              priority,
		Tags:                  a.Tags(),
		LegacyItemID:          a.LegacyItemID(),
		CalculatorAssetID:     a.CalculatorAssetID(),
		MapCatalogGroupIDs:    a.MapCatalogGroupIDs(),
	}
}

func jsonToAsset(a defenseAssetJSON) domain.DefenseAsset {
	roles := make([]domain.DefenseAssetRole, len(a.Roles))
	for i, r := range a.Roles {
		roles[i] = domain.DefenseAssetRole(r)
	}

	compatTypes := make([]domain.LayerGeometryType, len(a.CompatibleLayerTypes))
	for i, t := range a.CompatibleLayerTypes {
		compatTypes[i] = domain.LayerGeometryType(t)
	}

	var priority *domain.DefensePriority
	if a.Priority != nil {
		v := domain.DefensePriority(*a.Priority)
		priority = &v
	}

	return domain.NewDefenseAsset(
		a.ID, a.Name, a.ShortName, a.Description,
		domain.DefenseAssetCategory(a.Category),
		roles, a.PricePerUnitMln,
		a.Currency, a.UnitLabel,
		compatTypes,
		a.RecommendedLayerCodes, a.CompatibleLayerCodes, a.IncompatibleLayerCodes,
		a.MinEffectiveDistance, a.MaxEffectiveDistance,
		domain.DefenseAssetCoverageType(a.CoverageType),
		a.CoverageRadius, a.CoverageAngle,
		domain.DefenseAssetDeploymentType(a.DeploymentType),
		domain.DefenseAssetPlacementType(a.PlacementType),
		a.IconURL, a.ModelURL,
		a.Score, priority,
		a.Tags,
		a.LegacyItemID, a.CalculatorAssetID,
		a.MapCatalogGroupIDs,
	)
}

func placedObjectToJSON(o domain.PlacedDefenseObject) placedDefenseObjectJSON {
	return placedDefenseObjectJSON{
		ID:      o.ID(),
		AssetID: o.AssetID(),
		LayerID: o.LayerID(),
		Name:    o.Name(),
		Coordinates: coordinatesJSON{
			Lat: o.Coordinates().Lat(),
			Lng: o.Coordinates().Lng(),
		},
		Rotation:              o.Rotation(),
		Scale:                 o.Scale(),
		Quantity:              o.Quantity(),
		Status:                string(o.Status()),
		CustomPricePerUnitMln: o.CustomPricePerUnitMln(),
		CustomCoverageRadius:  o.CustomCoverageRadius(),
		CustomCoverageAngle:   o.CustomCoverageAngle(),
		HasGeometryConflict:   o.HasGeometryConflict(),
		HasCoverageConflict:   o.HasCoverageConflict(),
		HasTerrainConflict:    o.HasTerrainConflict(),
		Notes:                 o.Notes(),
		CreatedAt:             o.CreatedAt().Format(time.RFC3339Nano),
		UpdatedAt:             o.UpdatedAt().Format(time.RFC3339Nano),
	}
}

func jsonToPlacedObject(o placedDefenseObjectJSON) domain.PlacedDefenseObject {
	createdAt, _ := time.Parse(time.RFC3339Nano, o.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339Nano, o.UpdatedAt)

	return domain.NewPlacedDefenseObject(
		o.ID, o.AssetID, o.LayerID, o.Name,
		domain.NewCoordinates(o.Coordinates.Lat, o.Coordinates.Lng),
		o.Rotation, o.Scale,
		o.Quantity, domain.PlacedObjectStatus(o.Status),
		o.CustomPricePerUnitMln, o.CustomCoverageRadius, o.CustomCoverageAngle,
		o.HasGeometryConflict, o.HasCoverageConflict, o.HasTerrainConflict,
		o.Notes, createdAt, updatedAt,
	)
}

func protectedObjectToJSON(p domain.ProtectedObject) protectedObjectJSON {
	return protectedObjectJSON{
		ID:   p.ID(),
		Name: p.Name(),
		Center: coordinatesJSON{
			Lat: p.Center().Lat(),
			Lng: p.Center().Lng(),
		},
	}
}

func boolOrTrue(b *bool) bool {
	if b == nil {
		return true
	}
	return *b
}

func boolOrFalse(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}
