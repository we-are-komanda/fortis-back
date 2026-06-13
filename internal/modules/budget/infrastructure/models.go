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
