package infrastructure

import (
	"encoding/json"
	"time"

	"github.com/fortis/backend/internal/modules/budget/domain"
)

// BudgetConfigModel — GORM-модель для хранения BudgetConfig в таблице budget_configs.
type BudgetConfigModel struct {
	ProjectID  string    `gorm:"primaryKey;type:uuid"`
	ConfigData string    `gorm:"type:jsonb;not null"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}

func (BudgetConfigModel) TableName() string {
	return "budget_configs"
}

// budgetConfigDataJSON — промежуточная структура для сериализации/десериализации JSONB.
type budgetConfigDataJSON struct {
	BudgetMode      string  `json:"budgetMode"`
	BudgetAmountMln float64 `json:"budgetAmountMln"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

// ToDomain преобразует GORM-модель в доменный агрегат.
func (m *BudgetConfigModel) ToDomain() (*domain.BudgetConfig, error) {
	var data budgetConfigDataJSON
	if err := json.Unmarshal([]byte(m.ConfigData), &data); err != nil {
		return nil, err
	}

	createdAt, _ := time.Parse(time.RFC3339Nano, data.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339Nano, data.UpdatedAt)

	return domain.RestoreBudgetConfig(
		m.ProjectID,
		domain.BudgetMode(data.BudgetMode),
		data.BudgetAmountMln,
		createdAt,
		updatedAt,
	), nil
}

// ToModel преобразует доменный агрегат в GORM-модель.
func ToModel(config *domain.BudgetConfig) (*BudgetConfigModel, error) {
	data := budgetConfigDataJSON{
		BudgetMode:      string(config.BudgetMode()),
		BudgetAmountMln: config.BudgetAmountMln(),
		CreatedAt:       config.CreatedAt().Format(time.RFC3339Nano),
		UpdatedAt:       config.UpdatedAt().Format(time.RFC3339Nano),
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &BudgetConfigModel{
		ProjectID:  config.ProjectID(),
		ConfigData: string(raw),
	}, nil
}

type CostProjectionModel struct {
	ProjectID          string  `gorm:"primaryKey;type:uuid"`
	ProjectVersion     int     `gorm:"primaryKey"`
	CalculationVersion string  `gorm:"primaryKey"`
	ProjectionJSON     string  `gorm:"type:jsonb;not null"`
	KnownSubtotalMinor string  `gorm:"type:numeric;not null"`
	TotalMinor         *string `gorm:"type:numeric"`
	CreatedAt          time.Time
}

func (CostProjectionModel) TableName() string { return "project_cost_projections" }

type costIdentityJSON struct {
	ProjectID          string            `json:"projectId"`
	ProjectVersion     int               `json:"projectVersion"`
	CalculationVersion string            `json:"calculationVersion"`
	InputDataVersions  map[string]string `json:"inputDataVersions"`
	SnapshotDigest     string            `json:"snapshotDigest"`
}
type costLineJSON struct {
	ObjectID       string          `json:"objectId"`
	AssetID        string          `json:"assetId"`
	LayerID        string          `json:"layerId"`
	Category       string          `json:"category"`
	Name           string          `json:"name"`
	Quantity       int             `json:"quantity"`
	UnitPriceMinor *string         `json:"unitPriceMinor"`
	LineTotalMinor *string         `json:"lineTotalMinor"`
	PriceSource    string          `json:"priceSource"`
	Provenance     json.RawMessage `json:"provenance"`
}
type costGroupJSON struct {
	ID                 string  `json:"id"`
	Name               string  `json:"name"`
	ObjectCount        int     `json:"objectCount"`
	UnitCount          int     `json:"unitCount"`
	KnownSubtotalMinor string  `json:"knownSubtotalMinor"`
	TotalMinor         *string `json:"totalMinor"`
}
type costIssueJSON struct {
	Code      string   `json:"code"`
	Severity  string   `json:"severity"`
	ObjectIDs []string `json:"objectIds"`
	Message   string   `json:"message"`
}
type costProjectionJSON struct {
	Identity              costIdentityJSON `json:"identity"`
	Currency              string           `json:"currency"`
	Lines                 []costLineJSON   `json:"lines"`
	ByLayer               []costGroupJSON  `json:"byLayer"`
	ByType                []costGroupJSON  `json:"byType"`
	KnownSubtotalMinor    string           `json:"knownSubtotalMinor"`
	TotalMinor            *string          `json:"totalMinor"`
	IsComplete            bool             `json:"isComplete"`
	UnknownPriceObjectIDs []string         `json:"unknownPriceObjectIds"`
	Warnings              []costIssueJSON  `json:"warnings"`
}

func costProjectionToModel(p *domain.CostProjection) (*CostProjectionModel, error) {
	i := p.Identity()
	data := costProjectionJSON{Identity: costIdentityJSON{i.ProjectID(), i.ProjectVersion(), i.CalculationVersion(), i.InputDataVersions(), i.SnapshotDigest()}, Currency: "RUB", Lines: []costLineJSON{}, ByLayer: []costGroupJSON{}, ByType: []costGroupJSON{}, KnownSubtotalMinor: p.KnownSubtotalMinor(), TotalMinor: p.TotalMinor(), IsComplete: p.IsComplete(), UnknownPriceObjectIDs: p.UnknownPriceObjectIDs(), Warnings: []costIssueJSON{}}
	for _, l := range p.Lines() {
		var provenance json.RawMessage
		if l.Provenance() != "" {
			provenance = json.RawMessage(l.Provenance())
		}
		data.Lines = append(data.Lines, costLineJSON{l.ObjectID(), l.AssetID(), l.LayerID(), l.Category(), l.Name(), l.Quantity(), l.UnitPriceMinor(), l.LineTotalMinor(), l.PriceSource(), provenance})
	}
	groups := func(source []domain.CostGroup) []costGroupJSON {
		result := []costGroupJSON{}
		for _, g := range source {
			result = append(result, costGroupJSON{g.ID(), g.Name(), g.ObjectCount(), g.UnitCount(), g.KnownSubtotalMinor(), g.TotalMinor()})
		}
		return result
	}
	data.ByLayer, data.ByType = groups(p.ByLayer()), groups(p.ByType())
	for _, w := range p.Warnings() {
		data.Warnings = append(data.Warnings, costIssueJSON{w.Code(), w.Severity(), w.ObjectIDs(), w.Message()})
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return &CostProjectionModel{ProjectID: i.ProjectID(), ProjectVersion: i.ProjectVersion(), CalculationVersion: i.CalculationVersion(), ProjectionJSON: string(raw), KnownSubtotalMinor: p.KnownSubtotalMinor(), TotalMinor: p.TotalMinor()}, nil
}

func (m *CostProjectionModel) ToDomain() (*domain.CostProjection, error) {
	var data costProjectionJSON
	if err := json.Unmarshal([]byte(m.ProjectionJSON), &data); err != nil {
		return nil, err
	}
	i := data.Identity
	identity := domain.NewCalculationIdentity(i.ProjectID, i.ProjectVersion, i.CalculationVersion, i.SnapshotDigest, i.InputDataVersions)
	lines := []domain.FinancialLine{}
	for _, l := range data.Lines {
		provenance := string(l.Provenance)
		if provenance == "null" {
			provenance = ""
		}
		lines = append(lines, domain.RestoreFinancialLine(l.ObjectID, l.AssetID, l.LayerID, l.Category, l.Name, l.Quantity, l.UnitPriceMinor, l.LineTotalMinor, l.PriceSource, provenance))
	}
	groups := func(source []costGroupJSON) []domain.CostGroup {
		result := []domain.CostGroup{}
		for _, g := range source {
			result = append(result, domain.RestoreCostGroup(g.ID, g.Name, g.ObjectCount, g.UnitCount, g.KnownSubtotalMinor, g.TotalMinor))
		}
		return result
	}
	warnings := []domain.CostIssue{}
	for _, w := range data.Warnings {
		warnings = append(warnings, domain.RestoreCostIssue(w.Code, w.Severity, w.Message, w.ObjectIDs))
	}
	return domain.RestoreCostProjection(identity, lines, groups(data.ByLayer), groups(data.ByType), data.KnownSubtotalMinor, data.TotalMinor, data.UnknownPriceObjectIDs, warnings), nil
}
