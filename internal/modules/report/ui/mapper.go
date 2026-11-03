package ui

import (
	"encoding/json"
	budget "github.com/fortis/backend/internal/modules/budget/ui"
	"github.com/fortis/backend/internal/modules/report/domain"
	"math"
	"regexp"
	"time"
)

// frozenDocument is a field allowlist, not a generic project passthrough.
// Free text remains text; callers must render without HTML or automatic links.
type frozenDocument struct {
	BaseObject ReportBaseObject `json:"baseObject"`
	Layers     []struct {
		ReportLayer
		IsVisible *bool `json:"isVisible"`
	} `json:"layers"`
	AssetLibrary []struct {
		ID             string               `json:"id"`
		Name           string               `json:"name"`
		Category       string               `json:"category"`
		Description    *string              `json:"description"`
		UnitLabel      string               `json:"unitLabel"`
		ProtectionType string               `json:"protectionType"`
		Compound       *ReportCompoundCard  `json:"compoundProfile"`
		Weapon         *ReportWeaponCard    `json:"weaponSpec"`
		Detection      *ReportDetectionCard `json:"detectionSpec"`
		EW             *ReportEWCard        `json:"ewSpec"`
	} `json:"assetLibrary"`
	PlacedObjects []struct {
		ID             string     `json:"id"`
		AssetID        string     `json:"assetId"`
		LayerID        string     `json:"layerId"`
		Name           *string    `json:"name"`
		Quantity       int        `json:"quantity"`
		Coordinates    Coordinate `json:"coordinates"`
		Rotation       *float64   `json:"rotation"`
		Scale          *float64   `json:"scale"`
		IsVisibleOnMap *bool      `json:"isVisibleOnMap"`
		Notes          *string    `json:"notes"`
	} `json:"placedObjects"`
}

func (c *Coordinate) UnmarshalJSON(data []byte) error {
	var v struct {
		Lat      *float64 `json:"lat"`
		Lng      *float64 `json:"lng"`
		Altitude *float64 `json:"altitude"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if v.Lat == nil || v.Lng == nil {
		return domain.ErrInvalidSnapshot
	}
	*c = Coordinate{present: true, Lat: *v.Lat, Lng: *v.Lng, Altitude: v.Altitude}
	if !validCoordinate(*c) {
		return domain.ErrInvalidSnapshot
	}
	return nil
}

var reportColor = regexp.MustCompile(`(?i)^(#[0-9a-f]{3}|#[0-9a-f]{4}|#[0-9a-f]{6}|#[0-9a-f]{8}|black|white|red|green|blue|gray|grey|orange|purple|yellow|cyan|magenta|transparent|rgba?\([0-9.,% ]+\))$`)

func validCoordinate(c Coordinate) bool {
	return c.present && !math.IsNaN(c.Lat) && !math.IsNaN(c.Lng) && !math.IsInf(c.Lat, 0) && !math.IsInf(c.Lng, 0) && c.Lat >= -90 && c.Lat <= 90 && c.Lng >= -180 && c.Lng <= 180
}
func safeGeometry(g ReportGeometry) (ReportGeometry, error) {
	validRadius := func(n *float64) bool { return n != nil && *n >= 0 && !math.IsNaN(*n) && !math.IsInf(*n, 0) }
	switch g.Type {
	case "circle":
		if g.Center == nil || !validCoordinate(*g.Center) || !validRadius(g.RadiusM) {
			return g, domain.ErrInvalidSnapshot
		}
		return ReportGeometry{Type: g.Type, Center: g.Center, RadiusM: g.RadiusM}, nil
	case "ring":
		if g.Center == nil || !validCoordinate(*g.Center) || !validRadius(g.MinRadiusM) || !validRadius(g.MaxRadiusM) || *g.MinRadiusM > *g.MaxRadiusM {
			return g, domain.ErrInvalidSnapshot
		}
		return ReportGeometry{Type: g.Type, Center: g.Center, MinRadiusM: g.MinRadiusM, MaxRadiusM: g.MaxRadiusM}, nil
	case "polygon":
		points := g.Coordinates
		if points == nil {
			points = g.Points
		}
		if points == nil {
			return g, domain.ErrInvalidSnapshot
		}
		for _, p := range *points {
			if !validCoordinate(p) {
				return g, domain.ErrInvalidSnapshot
			}
		}
		return ReportGeometry{Type: g.Type, Coordinates: points, IsClosed: g.IsClosed}, nil
	case "freeform":
		if g.Points == nil {
			return g, domain.ErrInvalidSnapshot
		}
		for _, p := range *g.Points {
			if !validCoordinate(p) {
				return g, domain.ErrInvalidSnapshot
			}
		}
		return ReportGeometry{Type: g.Type, Points: g.Points}, nil
	default:
		return g, domain.ErrInvalidSnapshot
	}
}
func RenderReport(draft *domain.ReportDraft) (string, error) {
	p := draft.Project()
	var doc frozenDocument
	if err := json.Unmarshal([]byte(p.Document()), &doc); err != nil {
		return "", domain.ErrInvalidSnapshot
	}
	if !validCoordinate(doc.BaseObject.Center) {
		return "", domain.ErrInvalidSnapshot
	}
	var snapshot struct {
		Inputs map[string]string `json:"inputDataVersions"`
	}
	if err := json.Unmarshal([]byte(p.Snapshot()), &snapshot); err != nil {
		return "", domain.ErrInvalidSnapshot
	}
	if snapshot.Inputs == nil {
		snapshot.Inputs = map[string]string{}
	}
	common := ReportCommon{ReportID: draft.ID(), ReportSchemaVersion: domain.SchemaVersion,
		Identity:    budget.CalculationIdentityDTO{ProjectID: p.ProjectID(), ProjectVersion: p.Version(), CalculationVersion: p.CostCalculationVersion(), InputDataVersions: snapshot.Inputs, SnapshotDigest: p.SnapshotDigest()},
		GeneratedAt: draft.GeneratedAt().UTC().Format(time.RFC3339Nano), Title: p.Name(), Enterprise: ReportEnterprise{ID: p.EnterpriseID()},
		MapSnapshotData: ReportMap{BaseObject: doc.BaseObject, Layers: []ReportLayer{}, Objects: []ReportMapObject{}},
		MapVisibility:   ReportVisibility{Mode: draft.Key().MapView(), LayerIDs: []string{}, ObjectIDs: []string{}}, Objects: []ReportObject{}, Warnings: []budget.CostIssueDTO{},
		Limitations: []string{"Автоматическая проверка не выполнена.", "Схема использует сохранённые координаты без внешней картографической подложки; эффективность защиты не оценивалась."}}
	warning := func(code, message string) {
		common.Warnings = append(common.Warnings, budget.CostIssueDTO{Code: code, Severity: "warning", ObjectIDs: []string{}, Message: message})
	}
	warning("enterprise_name_unavailable", "Название предприятия не зафиксировано в этой версии.")
	warning("manual_text_review_required", "Проверьте свободный текст перед передачей отчёта: он может содержать суммы и другую чувствительную информацию.")
	if draft.Key().MapView() == "saved" {
		warning("selective_map", "На схеме показана сохранённая выборка; перечень объектов остаётся полным.")
	}
	visibleLayers := map[string]bool{}
	for _, l := range doc.Layers {
		layer := l.ReportLayer
		layer.IsVisible = l.IsVisible == nil || *l.IsVisible
		if layer.Color != nil && !reportColor.MatchString(*layer.Color) {
			return "", domain.ErrInvalidSnapshot
		}
		var err error
		layer.Geometry, err = safeGeometry(layer.Geometry)
		if err != nil {
			return "", err
		}
		common.MapSnapshotData.Layers = append(common.MapSnapshotData.Layers, layer)
		if draft.Key().MapView() == "all" || layer.IsVisible {
			visibleLayers[layer.ID] = true
			common.MapVisibility.LayerIDs = append(common.MapVisibility.LayerIDs, layer.ID)
		}
	}
	for _, o := range doc.PlacedObjects {
		if !validCoordinate(o.Coordinates) {
			return "", domain.ErrInvalidSnapshot
		}
		obj := ReportObject{ReportMapObject: ReportMapObject{ObjectID: o.ID, AssetID: o.AssetID, LayerID: o.LayerID, Name: o.AssetID, Quantity: o.Quantity, Coordinates: o.Coordinates, Rotation: o.Rotation, Scale: o.Scale, IsVisibleOnMap: o.IsVisibleOnMap == nil || *o.IsVisibleOnMap}, Notes: o.Notes}
		for _, a := range doc.AssetLibrary {
			if a.ID == o.AssetID {
				obj.Name = a.Name
				obj.Category = a.Category
				obj.Description = a.Description
				obj.UnitLabel = a.UnitLabel
				obj.ProtectionType = a.ProtectionType
				if a.Compound != nil || a.Weapon != nil || a.Detection != nil || a.EW != nil {
					obj.Card = &ReportCard{Compound: a.Compound, Weapon: a.Weapon, Detection: a.Detection, EW: a.EW}
				}
				break
			}
		}
		if o.Name != nil && *o.Name != "" {
			obj.Name = *o.Name
		}
		common.Objects = append(common.Objects, obj)
		common.MapSnapshotData.Objects = append(common.MapSnapshotData.Objects, obj.ReportMapObject)
		if draft.Key().MapView() == "all" || (obj.IsVisibleOnMap && visibleLayers[obj.LayerID]) {
			common.MapVisibility.ObjectIDs = append(common.MapVisibility.ObjectIDs, obj.ObjectID)
		}
	}
	var body []byte
	var err error
	if draft.Key().HideCost() {
		body, err = json.Marshal(UnpricedReportResponse{ReportCommon: common, HideCost: true})
	} else {
		cost := budget.CostProjectionToDTO(draft.Cost())
		projection := budget.BudgetProjectionToDTO(draft.Budget())
		sources := []budget.CostDataProvenanceDTO{}
		seen := map[string]bool{}
		preliminary := !cost.IsComplete
		for _, line := range cost.Lines {
			if line.Provenance == nil {
				preliminary = true
				continue
			}
			source := *line.Provenance
			raw, _ := json.Marshal(source)
			if !seen[string(raw)] {
				sources = append(sources, source)
				seen[string(raw)] = true
			}
			if source.Quality != "confirmed" {
				preliminary = true
			}
		}
		common.Warnings = append(common.Warnings, cost.Warnings...)
		if preliminary {
			warning("preliminary_report", "Предварительный отчёт: имеются неизвестные цены или непроверенные исходные данные.")
		}
		body, err = json.Marshal(PricedReportResponse{ReportCommon: common, HideCost: false, Cost: cost, Budget: projection, Sources: sources})
	}
	if err != nil {
		return "", domain.ErrInvalidSnapshot
	}
	return string(body), nil
}
