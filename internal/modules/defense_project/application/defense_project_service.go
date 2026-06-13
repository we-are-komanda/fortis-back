package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	defenseAssetDomain "github.com/fortis/backend/internal/modules/defense_asset/domain"
	"github.com/fortis/backend/internal/modules/defense_project/domain"
)

// importPayload — структура для парсинга JSON при импорте.
type importPayload struct {
	SchemaVersion    int                      `json:"schemaVersion"`
	ProjectID        string                   `json:"projectId"`
	Name             string                   `json:"name,omitempty"`
	EnterpriseID     string                   `json:"enterpriseId,omitempty"`
	ProjectName      string                   `json:"projectName"`
	BaseObject       importProtectedObject    `json:"baseObject"`
	Layers           []importEditableLayer    `json:"layers"`
	AssetLibrary     []importAsset            `json:"assetLibrary"`
	PlacedObjects    []importPlacedObject     `json:"placedObjects"`
	ActiveLayerID    *string                  `json:"activeLayerId,omitempty"`
	SelectedAssetID  *string                  `json:"selectedAssetId,omitempty"`
	SelectedObjectID *string                  `json:"selectedObjectId,omitempty"`
	Mode             string                   `json:"mode"`
	Source           string                   `json:"source,omitempty"`
	BasePresetID     *string                  `json:"basePresetId,omitempty"`
	UpdatedAt        string                   `json:"updatedAt"`
}

type importCoordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type importProtectedObject struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Center importCoordinates `json:"center"`
}

type importLayerGeometry struct {
	Type       string              `json:"type"`
	Center     *importCoordinates  `json:"center,omitempty"`
	RadiusM    *float64            `json:"radiusM,omitempty"`
	MinRadiusM *float64            `json:"minRadiusM,omitempty"`
	MaxRadiusM *float64            `json:"maxRadiusM,omitempty"`
	Points     []importCoordinates `json:"points,omitempty"`
}

type importEditableLayer struct {
	ID                 string              `json:"id"`
	Name               string              `json:"name"`
	Code               string              `json:"code"`
	Description        *string             `json:"description,omitempty"`
	Order              int                 `json:"order"`
	DistanceFromObjMin *float64            `json:"distanceFromObjectMin,omitempty"`
	DistanceFromObjMax *float64            `json:"distanceFromObjectMax,omitempty"`
	GeometryType       string              `json:"geometryType"`
	Geometry           importLayerGeometry `json:"geometry"`
	Color              *string             `json:"color,omitempty"`
	Opacity            *float64            `json:"opacity,omitempty"`
	IsActive           bool                `json:"isActive"`
	IsVisible          *bool               `json:"isVisible,omitempty"`
	IsLocked           *bool               `json:"isLocked,omitempty"`
}

type importAsset struct {
	ID                     string                                      `json:"id"`
	Name                   string                                      `json:"name"`
	ShortName              *string                                     `json:"shortName,omitempty"`
	Description            *string                                     `json:"description,omitempty"`
	Category               string                                      `json:"category"`
	Roles                  []string                                    `json:"roles"`
	PricePerUnitMln        *float64                                    `json:"pricePerUnitMln,omitempty"`
	Currency               string                                      `json:"currency"`
	UnitLabel              string                                      `json:"unitLabel"`
	CompatibleLayerTypes   []string                                    `json:"compatibleLayerTypes,omitempty"`
	RecommendedLayerCodes  []string                                    `json:"recommendedLayerCodes,omitempty"`
	CompatibleLayerCodes   []string                                    `json:"compatibleLayerCodes,omitempty"`
	IncompatibleLayerCodes []string                                    `json:"incompatibleLayerCodes,omitempty"`
	ProtectionType         string                                      `json:"protectionType,omitempty"`
	MinEffectiveDistance   *float64                                    `json:"minEffectiveDistance,omitempty"`
	MaxEffectiveDistance   *float64                                    `json:"maxEffectiveDistance,omitempty"`
	CoverageType           string                                      `json:"coverageType"`
	CoverageRadius         *float64                                    `json:"coverageRadius,omitempty"`
	CoverageAngle          *float64                                    `json:"coverageAngle,omitempty"`
	DeploymentType         string                                      `json:"deploymentType"`
	PlacementType          string                                      `json:"placementType"`
	IconURL                *string                                     `json:"iconUrl,omitempty"`
	ModelURL               *string                                     `json:"modelUrl,omitempty"`
	Score                  *int                                        `json:"score,omitempty"`
	Priority               *string                                     `json:"priority,omitempty"`
	CompoundProfile        *defenseAssetDomain.DefenseAssetCompoundProfile `json:"compoundProfile,omitempty"`
	Tags                   []string                                    `json:"tags,omitempty"`
	LegacyItemID           *string                                     `json:"legacyItemId,omitempty"`
	CalculatorAssetID      *string                                     `json:"calculatorAssetId,omitempty"`
	MapCatalogGroupIDs     []string                                    `json:"mapCatalogGroupIds,omitempty"`
}

type importPlacedObject struct {
	ID                    string            `json:"id"`
	AssetID               string            `json:"assetId"`
	LayerID               string            `json:"layerId"`
	Name                  *string           `json:"name,omitempty"`
	Coordinates           importCoordinates `json:"coordinates"`
	Rotation              *float64          `json:"rotation,omitempty"`
	Scale                 *float64          `json:"scale,omitempty"`
	Quantity              int               `json:"quantity"`
	Status                string            `json:"status"`
	CustomPricePerUnitMln *float64          `json:"customPricePerUnitMln,omitempty"`
	CustomCoverageRadius  *float64          `json:"customCoverageRadius,omitempty"`
	CustomCoverageAngle   *float64          `json:"customCoverageAngle,omitempty"`
	HasGeometryConflict   bool              `json:"hasGeometryConflict"`
	HasCoverageConflict   bool              `json:"hasCoverageConflict"`
	HasTerrainConflict    bool              `json:"hasTerrainConflict"`
	Notes                 *string           `json:"notes,omitempty"`
	CreatedAt             string            `json:"createdAt"`
	UpdatedAt             string            `json:"updatedAt"`
}

// DefenseProjectService — сервис для операций импорта/экспорта DefenseProject.
type DefenseProjectService struct {
	repo domain.DefenseProjectRepositoryInterface
}

func NewDefenseProjectService(repo domain.DefenseProjectRepositoryInterface) *DefenseProjectService {
	return &DefenseProjectService{
		repo: repo,
	}
}

// CreateFromJSON создаёт проект из JSON с заданным именем конфигурации и enterpriseID.
func (s *DefenseProjectService) CreateFromJSON(ctx context.Context, name, enterpriseID, rawJSON string) (*domain.DefenseProject, error) {
	if name == "" {
		return nil, domain.ErrInvalidConfigName
	}

	var payload importPayload
	if err := json.Unmarshal([]byte(rawJSON), &payload); err != nil {
		return nil, fmt.Errorf("unmarshal project json: %w", err)
	}

	if payload.SchemaVersion != domain.SchemaVersion {
		return nil, domain.ErrInvalidSchemaVersion
	}

	if payload.ProjectName == "" {
		return nil, fmt.Errorf("projectName: %w", domain.ErrInvalidProjectData)
	}

	projectID := uuid.New().String()

	mode := domain.DefenseProjectModeView
	if payload.Mode != "" {
		mode = domain.DefenseProjectMode(payload.Mode)
	}

	source := domain.DefenseProjectSourceCustom
	if payload.Source != "" {
		source = domain.DefenseProjectSource(payload.Source)
	}

	now := time.Now().UTC()

	project, err := domain.NewDefenseProject(
		projectID, name, enterpriseID, payload.ProjectName,
		domain.NewProtectedObject(
			payload.BaseObject.ID,
			payload.BaseObject.Name,
			domain.NewCoordinates(payload.BaseObject.Center.Lat, payload.BaseObject.Center.Lng),
		),
		mapImportLayers(payload.Layers),
		mapImportAssets(payload.AssetLibrary),
		mapImportPlacedObjects(payload.PlacedObjects),
		payload.ActiveLayerID,
		payload.SelectedAssetID,
		payload.SelectedObjectID,
		mode, source,
		payload.BasePresetID,
		now,
	)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, project); err != nil {
		return nil, fmt.Errorf("save project: %w", err)
	}

	return project, nil
}

// Import выполняет импорт проекта из JSON.
func (s *DefenseProjectService) Import(ctx context.Context, rawJSON string) (*domain.DefenseProject, error) {
	var payload importPayload
	if err := json.Unmarshal([]byte(rawJSON), &payload); err != nil {
		return nil, fmt.Errorf("unmarshal project json: %w", err)
	}

	// Валидация schemaVersion
	if payload.SchemaVersion != domain.SchemaVersion {
		return nil, domain.ErrInvalidSchemaVersion
	}

	// Генерация нового UUID для проекта (игнорируем projectId из запроса)
	projectID := uuid.New().String()

	// Валидация обязательных полей
	if payload.ProjectName == "" {
		return nil, fmt.Errorf("projectName: %w", domain.ErrInvalidProjectData)
	}
	if payload.BaseObject.ID == "" || payload.BaseObject.Name == "" {
		return nil, fmt.Errorf("baseObject: %w", domain.ErrInvalidProjectData)
	}

	// Сборка доменного объекта
	mode := domain.DefenseProjectModeView
	if payload.Mode != "" {
		mode = domain.DefenseProjectMode(payload.Mode)
	}

	source := domain.DefenseProjectSourceCustom
	if payload.Source != "" {
		source = domain.DefenseProjectSource(payload.Source)
	}

	now := time.Now().UTC()

	project, err := domain.NewDefenseProject(
		projectID,
		payload.Name,
		payload.EnterpriseID,
		payload.ProjectName,
		domain.NewProtectedObject(
			payload.BaseObject.ID,
			payload.BaseObject.Name,
			domain.NewCoordinates(payload.BaseObject.Center.Lat, payload.BaseObject.Center.Lng),
		),
		mapImportLayers(payload.Layers),
		mapImportAssets(payload.AssetLibrary),
		mapImportPlacedObjects(payload.PlacedObjects),
		payload.ActiveLayerID,
		payload.SelectedAssetID,
		payload.SelectedObjectID,
		mode,
		source,
		payload.BasePresetID,
		now,
	)
	if err != nil {
		return nil, err
	}

	// Сохранение в БД
	if err := s.repo.Save(ctx, project); err != nil {
		return nil, fmt.Errorf("save project: %w", err)
	}

	return project, nil
}

// CreateProject создаёт новый проект с заданным именем и enterpriseID.
func (s *DefenseProjectService) CreateProject(ctx context.Context, name, enterpriseID, projectName string, baseObject domain.ProtectedObject, layers []domain.EditableDefenseLayer, assets []domain.DefenseAsset, placedObjects []domain.PlacedDefenseObject, mode domain.DefenseProjectMode, source domain.DefenseProjectSource, basePresetID *string) (*domain.DefenseProject, error) {
	if name == "" {
		return nil, domain.ErrInvalidConfigName
	}

	now := time.Now().UTC()
	projectID := uuid.New().String()

	project, err := domain.NewDefenseProject(
		projectID, name, enterpriseID, projectName,
		baseObject, layers, assets, placedObjects,
		nil, nil, nil,
		mode, source, basePresetID,
		now,
	)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, project); err != nil {
		return nil, fmt.Errorf("save project: %w", err)
	}

	return project, nil
}

// ListProjects возвращает список проектов с пагинацией.
func (s *DefenseProjectService) ListProjects(ctx context.Context, enterpriseID string, limit, offset int) ([]*domain.DefenseProject, int64, error) {
	if enterpriseID != "" {
		return s.repo.FindAllByEnterprise(ctx, enterpriseID, limit, offset)
	}
	return s.repo.FindAll(ctx, limit, offset)
}

// GetProject возвращает проект по ID.
func (s *DefenseProjectService) GetProject(ctx context.Context, id string) (*domain.DefenseProject, error) {
	return s.repo.FindByID(ctx, id)
}

// UpdateProject обновляет существующий проект.
//
// Если projectJSON непустой — содержимое карты полностью перезаписывается из
// переданного JSON, при этом сохраняется ID проекта и версия optimistic-lock.
// Если projectJSON пустой — обновляются только метаданные (имя, enterpriseID).
func (s *DefenseProjectService) UpdateProject(ctx context.Context, id, name, enterpriseID, projectJSON string) (*domain.DefenseProject, error) {
	project, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if projectJSON != "" {
		return s.overwriteProjectContent(ctx, project, name, enterpriseID, projectJSON)
	}

	if name != "" {
		project.SetName(name)
	}
	if enterpriseID != "" {
		project.SetEnterpriseID(enterpriseID)
	}

	if err := s.repo.Save(ctx, project); err != nil {
		return nil, fmt.Errorf("save project: %w", err)
	}

	return project, nil
}

// overwriteProjectContent перестраивает доменный объект из projectJSON, переиспользуя
// существующий ID проекта и сохраняя версию optimistic-lock.
func (s *DefenseProjectService) overwriteProjectContent(ctx context.Context, existing *domain.DefenseProject, name, enterpriseID, projectJSON string) (*domain.DefenseProject, error) {
	var payload importPayload
	if err := json.Unmarshal([]byte(projectJSON), &payload); err != nil {
		return nil, fmt.Errorf("unmarshal project json: %w", err)
	}

	if payload.SchemaVersion != domain.SchemaVersion {
		return nil, domain.ErrInvalidSchemaVersion
	}

	if payload.ProjectName == "" {
		return nil, fmt.Errorf("projectName: %w", domain.ErrInvalidProjectData)
	}
	if payload.BaseObject.ID == "" || payload.BaseObject.Name == "" {
		return nil, fmt.Errorf("baseObject: %w", domain.ErrInvalidProjectData)
	}

	mode := domain.DefenseProjectModeView
	if payload.Mode != "" {
		mode = domain.DefenseProjectMode(payload.Mode)
	}

	source := domain.DefenseProjectSourceCustom
	if payload.Source != "" {
		source = domain.DefenseProjectSource(payload.Source)
	}

	// Итоговое имя: явный аргумент name важнее, затем payload.Name, затем текущее имя.
	finalName := existing.Name()
	if payload.Name != "" {
		finalName = payload.Name
	}
	if name != "" {
		finalName = name
	}

	// Итоговый enterpriseID: тот же порядок предпочтения.
	finalEnterpriseID := existing.EnterpriseID()
	if payload.EnterpriseID != "" {
		finalEnterpriseID = payload.EnterpriseID
	}
	if enterpriseID != "" {
		finalEnterpriseID = enterpriseID
	}

	now := time.Now().UTC()

	rebuilt, err := domain.NewDefenseProject(
		existing.ProjectID(), finalName, finalEnterpriseID, payload.ProjectName,
		domain.NewProtectedObject(
			payload.BaseObject.ID,
			payload.BaseObject.Name,
			domain.NewCoordinates(payload.BaseObject.Center.Lat, payload.BaseObject.Center.Lng),
		),
		mapImportLayers(payload.Layers),
		mapImportAssets(payload.AssetLibrary),
		mapImportPlacedObjects(payload.PlacedObjects),
		payload.ActiveLayerID,
		payload.SelectedAssetID,
		payload.SelectedObjectID,
		mode, source,
		payload.BasePresetID,
		now,
	)
	if err != nil {
		return nil, err
	}

	// Сохраняем версию optimistic-lock с загруженного проекта, чтобы проверка
	// версии в repo.Save оставалась корректной (NewDefenseProject сбрасывает её на 1).
	rebuilt.SetVersion(existing.Version())

	if err := s.repo.Save(ctx, rebuilt); err != nil {
		return nil, fmt.Errorf("save project: %w", err)
	}

	return rebuilt, nil
}

// DeleteProject удаляет проект по ID.
func (s *DefenseProjectService) DeleteProject(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// Export выполняет экспорт проекта в JSON.
func (s *DefenseProjectService) Export(ctx context.Context, projectID string) (string, error) {
	project, err := s.repo.FindByID(ctx, projectID)
	if err != nil {
		return "", err
	}

	// Сериализуем весь проект в JSON
	raw, err := serializeProject(project)
	if err != nil {
		return "", fmt.Errorf("serialize project: %w", err)
	}

	return raw, nil
}

// serializeProject сериализует проект в JSON.
func serializeProject(project *domain.DefenseProject) (string, error) {
	data := exportPayload{
		SchemaVersion: project.SchemaVersion(),
		ProjectID:     project.ProjectID(),
		Name:          project.Name(),
		EnterpriseID:  project.EnterpriseID(),
		ProjectName:   project.ProjectName(),
		BaseObject: exportProtectedObject{
			ID:   project.BaseObject().ID(),
			Name: project.BaseObject().Name(),
			Center: exportCoordinates{
				Lat: project.BaseObject().Center().Lat(),
				Lng: project.BaseObject().Center().Lng(),
			},
		},
		Layers:        exportLayers(project.Layers()),
		AssetLibrary:  exportAssets(project.AssetLibrary()),
		PlacedObjects: exportPlacedObjects(project.PlacedObjects()),
		ActiveLayerID:    project.ActiveLayerID(),
		SelectedAssetID:  project.SelectedAssetID(),
		SelectedObjectID: project.SelectedObjectID(),
		Mode:          string(project.Mode()),
		Source:        string(project.Source()),
		BasePresetID:  project.BasePresetID(),
		UpdatedAt:     project.UpdatedAt().Format(time.RFC3339Nano),
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

// exportPayload — структура для сериализации проекта в JSON.
type exportPayload struct {
	SchemaVersion    int                    `json:"schemaVersion"`
	ProjectID        string                 `json:"projectId"`
	Name             string                 `json:"name,omitempty"`
	EnterpriseID     string                 `json:"enterpriseId,omitempty"`
	ProjectName      string                 `json:"projectName"`
	BaseObject       exportProtectedObject  `json:"baseObject"`
	Layers           []exportLayer          `json:"layers"`
	AssetLibrary     []exportAsset          `json:"assetLibrary"`
	PlacedObjects    []exportPlacedObject   `json:"placedObjects"`
	ActiveLayerID    *string                `json:"activeLayerId,omitempty"`
	SelectedAssetID  *string                `json:"selectedAssetId,omitempty"`
	SelectedObjectID *string                `json:"selectedObjectId,omitempty"`
	Mode             string                 `json:"mode"`
	Source           string                 `json:"source,omitempty"`
	BasePresetID     *string                `json:"basePresetId,omitempty"`
	UpdatedAt        string                 `json:"updatedAt"`
}

type exportCoordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type exportProtectedObject struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Center exportCoordinates `json:"center"`
}

type exportGeometry struct {
	Type       string              `json:"type"`
	Center     *exportCoordinates  `json:"center,omitempty"`
	RadiusM    *float64            `json:"radiusM,omitempty"`
	MinRadiusM *float64            `json:"minRadiusM,omitempty"`
	MaxRadiusM *float64            `json:"maxRadiusM,omitempty"`
	Points     []exportCoordinates `json:"points,omitempty"`
}

type exportLayer struct {
	ID                 string          `json:"id"`
	Name               string          `json:"name"`
	Code               string          `json:"code"`
	Description        *string         `json:"description,omitempty"`
	Order              int             `json:"order"`
	DistanceFromObjMin *float64        `json:"distanceFromObjectMin,omitempty"`
	DistanceFromObjMax *float64        `json:"distanceFromObjectMax,omitempty"`
	GeometryType       string          `json:"geometryType"`
	Geometry           exportGeometry  `json:"geometry"`
	Color              *string         `json:"color,omitempty"`
	Opacity            *float64        `json:"opacity,omitempty"`
	IsActive           bool            `json:"isActive"`
	IsVisible          bool            `json:"isVisible"`
	IsLocked           bool            `json:"isLocked"`
}

type exportAsset struct {
	ID                     string   `json:"id"`
	Name                   string   `json:"name"`
	ShortName              *string  `json:"shortName,omitempty"`
	Description            *string  `json:"description,omitempty"`
	Category               string   `json:"category"`
	Roles                  []string `json:"roles"`
	PricePerUnitMln        *float64 `json:"pricePerUnitMln,omitempty"`
	Currency               string   `json:"currency"`
	UnitLabel              string   `json:"unitLabel"`
	CompatibleLayerTypes   []string `json:"compatibleLayerTypes,omitempty"`
	RecommendedLayerCodes  []string `json:"recommendedLayerCodes,omitempty"`
	CompatibleLayerCodes   []string `json:"compatibleLayerCodes,omitempty"`
	IncompatibleLayerCodes []string `json:"incompatibleLayerCodes,omitempty"`
	MinEffectiveDistance   *float64 `json:"minEffectiveDistance,omitempty"`
	MaxEffectiveDistance   *float64 `json:"maxEffectiveDistance,omitempty"`
	CoverageType           string   `json:"coverageType"`
	CoverageRadius         *float64 `json:"coverageRadius,omitempty"`
	CoverageAngle          *float64 `json:"coverageAngle,omitempty"`
	DeploymentType         string   `json:"deploymentType"`
	PlacementType          string   `json:"placementType"`
	IconURL                *string  `json:"iconUrl,omitempty"`
	ModelURL               *string  `json:"modelUrl,omitempty"`
	Score                  *int     `json:"score,omitempty"`
	Priority               *string  `json:"priority,omitempty"`
	Tags                   []string `json:"tags,omitempty"`
	LegacyItemID           *string  `json:"legacyItemId,omitempty"`
	CalculatorAssetID      *string  `json:"calculatorAssetId,omitempty"`
	MapCatalogGroupIDs     []string `json:"mapCatalogGroupIds,omitempty"`
}

type exportPlacedObject struct {
	ID                    string            `json:"id"`
	AssetID               string            `json:"assetId"`
	LayerID               string            `json:"layerId"`
	Name                  *string           `json:"name,omitempty"`
	Coordinates           exportCoordinates `json:"coordinates"`
	Rotation              *float64          `json:"rotation,omitempty"`
	Scale                 *float64          `json:"scale,omitempty"`
	Quantity              int               `json:"quantity"`
	Status                string            `json:"status"`
	CustomPricePerUnitMln *float64          `json:"customPricePerUnitMln,omitempty"`
	CustomCoverageRadius  *float64          `json:"customCoverageRadius,omitempty"`
	CustomCoverageAngle   *float64          `json:"customCoverageAngle,omitempty"`
	HasGeometryConflict   bool              `json:"hasGeometryConflict"`
	HasCoverageConflict   bool              `json:"hasCoverageConflict"`
	HasTerrainConflict    bool              `json:"hasTerrainConflict"`
	Notes                 *string           `json:"notes,omitempty"`
	CreatedAt             string            `json:"createdAt"`
	UpdatedAt             string            `json:"updatedAt"`
}

func mapImportLayers(layers []importEditableLayer) []domain.EditableDefenseLayer {
	if layers == nil {
		return nil
	}
	result := make([]domain.EditableDefenseLayer, len(layers))
	for i, l := range layers {
		geom := mapImportGeometry(l.Geometry)

		isVisible := true
		if l.IsVisible != nil {
			isVisible = *l.IsVisible
		}
		isLocked := false
		if l.IsLocked != nil {
			isLocked = *l.IsLocked
		}

		result[i] = domain.NewEditableDefenseLayer(
			l.ID, l.Name, l.Code, l.Description,
			l.Order,
			l.DistanceFromObjMin, l.DistanceFromObjMax,
			domain.LayerGeometryType(l.GeometryType),
			geom,
			l.Color, l.Opacity,
			l.IsActive, isVisible, isLocked,
		)
	}
	return result
}

func mapImportGeometry(g importLayerGeometry) domain.LayerGeometry {
	switch domain.LayerGeometryType(g.Type) {
	case domain.LayerGeometryCircle:
		var center domain.Coordinates
		if g.Center != nil {
			center = domain.NewCoordinates(g.Center.Lat, g.Center.Lng)
		}
		radius := 0.0
		if g.RadiusM != nil {
			radius = *g.RadiusM
		}
		return domain.NewCircleGeometry(center, radius)
	case domain.LayerGeometryRing:
		var center domain.Coordinates
		if g.Center != nil {
			center = domain.NewCoordinates(g.Center.Lat, g.Center.Lng)
		}
		minR, maxR := 0.0, 0.0
		if g.MinRadiusM != nil {
			minR = *g.MinRadiusM
		}
		if g.MaxRadiusM != nil {
			maxR = *g.MaxRadiusM
		}
		return domain.NewRingGeometry(center, minR, maxR)
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

func mapImportAssets(assets []importAsset) []domain.DefenseAsset {
	if assets == nil {
		return nil
	}
	result := make([]domain.DefenseAsset, len(assets))
	for i, a := range assets {
		roles := make([]domain.DefenseAssetRole, len(a.Roles))
		for j, r := range a.Roles {
			roles[j] = domain.DefenseAssetRole(r)
		}

		compatTypes := make([]domain.LayerGeometryType, len(a.CompatibleLayerTypes))
		for j, t := range a.CompatibleLayerTypes {
			compatTypes[j] = domain.LayerGeometryType(t)
		}

		var priority *domain.DefensePriority
		if a.Priority != nil {
			v := domain.DefensePriority(*a.Priority)
			priority = &v
		}

		result[i] = domain.NewDefenseAsset(
			a.ID, a.Name, a.ShortName, a.Description,
			domain.DefenseAssetCategory(a.Category),
			roles, a.PricePerUnitMln,
			a.Currency, a.UnitLabel,
			compatTypes,
			a.RecommendedLayerCodes, a.CompatibleLayerCodes, a.IncompatibleLayerCodes,
			a.ProtectionType,
			a.MinEffectiveDistance, a.MaxEffectiveDistance,
			domain.DefenseAssetCoverageType(a.CoverageType),
			a.CoverageRadius, a.CoverageAngle,
			domain.DefenseAssetDeploymentType(a.DeploymentType),
			domain.DefenseAssetPlacementType(a.PlacementType),
			a.IconURL, a.ModelURL,
			a.Score, priority,
			a.CompoundProfile,
			a.Tags,
			a.LegacyItemID, a.CalculatorAssetID,
			a.MapCatalogGroupIDs,
		)
	}
	return result
}

func mapImportPlacedObjects(objects []importPlacedObject) []domain.PlacedDefenseObject {
	if objects == nil {
		return nil
	}
	result := make([]domain.PlacedDefenseObject, len(objects))
	for i, o := range objects {
		createdAt, _ := time.Parse(time.RFC3339Nano, o.CreatedAt)
		updatedAt, _ := time.Parse(time.RFC3339Nano, o.UpdatedAt)

		result[i] = domain.NewPlacedDefenseObject(
			o.ID, o.AssetID, o.LayerID, o.Name,
			domain.NewCoordinates(o.Coordinates.Lat, o.Coordinates.Lng),
			o.Rotation, o.Scale,
			o.Quantity, domain.PlacedObjectStatus(o.Status),
			o.CustomPricePerUnitMln, o.CustomCoverageRadius, o.CustomCoverageAngle,
			o.HasGeometryConflict, o.HasCoverageConflict, o.HasTerrainConflict,
			o.Notes,
			createdAt, updatedAt,
		)
	}
	return result
}

func exportLayers(layers []domain.EditableDefenseLayer) []exportLayer {
	result := make([]exportLayer, len(layers))
	for i, l := range layers {
		geom := l.Geometry()
		eg := exportGeometry{Type: string(geom.Type())}

		switch geom.Type() {
		case domain.LayerGeometryCircle:
			if c := geom.Center(); c != nil {
				eg.Center = &exportCoordinates{Lat: c.Lat(), Lng: c.Lng()}
			}
			eg.RadiusM = geom.RadiusM()
		case domain.LayerGeometryRing:
			if c := geom.Center(); c != nil {
				eg.Center = &exportCoordinates{Lat: c.Lat(), Lng: c.Lng()}
			}
			eg.MinRadiusM = geom.MinRadiusM()
			eg.MaxRadiusM = geom.MaxRadiusM()
		case domain.LayerGeometryPolygon, domain.LayerGeometryFreeform:
			pts := geom.Points()
			eg.Points = make([]exportCoordinates, len(pts))
			for j, p := range pts {
				eg.Points[j] = exportCoordinates{Lat: p.Lat(), Lng: p.Lng()}
			}
		}

		result[i] = exportLayer{
			ID: l.ID(), Name: l.Name(), Code: l.Code(),
			Description:        l.Description(),
			Order:              l.Order(),
			DistanceFromObjMin: l.DistanceFromObjectMin(),
			DistanceFromObjMax: l.DistanceFromObjectMax(),
			GeometryType:       string(l.GeometryType()),
			Geometry:           eg,
			Color:              l.Color(),
			Opacity:            l.Opacity(),
			IsActive:           l.IsActive(),
			IsVisible:          l.IsVisible(),
			IsLocked:           l.IsLocked(),
		}
	}
	return result
}

func exportAssets(assets []domain.DefenseAsset) []exportAsset {
	result := make([]exportAsset, len(assets))
	for i, a := range assets {
		roles := make([]string, len(a.Roles()))
		for j, r := range a.Roles() {
			roles[j] = string(r)
		}

		compatTypes := make([]string, len(a.CompatibleLayerTypes()))
		for j, t := range a.CompatibleLayerTypes() {
			compatTypes[j] = string(t)
		}

		var priority *string
		if p := a.Priority(); p != nil {
			v := string(*p)
			priority = &v
		}

		result[i] = exportAsset{
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
	return result
}

func exportPlacedObjects(objects []domain.PlacedDefenseObject) []exportPlacedObject {
	result := make([]exportPlacedObject, len(objects))
	for i, o := range objects {
		result[i] = exportPlacedObject{
			ID:      o.ID(),
			AssetID: o.AssetID(),
			LayerID: o.LayerID(),
			Name:    o.Name(),
			Coordinates: exportCoordinates{
				Lat: o.Coordinates().Lat(),
				Lng: o.Coordinates().Lng(),
			},
			Rotation: o.Rotation(),
			Scale:    o.Scale(),
			Quantity: o.Quantity(),
			Status:   string(o.Status()),
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
	return result
}
