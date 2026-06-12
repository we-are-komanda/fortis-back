package infrastructure

import (
	"time"

	"github.com/fortis/backend/internal/modules/enterprise/domain"
)

// EnterpriseModel — GORM-модель для хранения Enterprise.
type EnterpriseModel struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name      string    `gorm:"not null"`
	Address   string    `gorm:"not null;default:''"`
	Status    string    `gorm:"not null;default:active"`
	Latitude  float64   `gorm:"not null"`
	Longitude float64   `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (EnterpriseModel) TableName() string {
	return "enterprises"
}

// ToDomain преобразует GORM-модель в доменный агрегат.
func (m *EnterpriseModel) ToDomain() *domain.Enterprise {
	ent, err := domain.NewEnterprise(
		m.ID,
		m.Name,
		m.Address,
		domain.EnterpriseStatus(m.Status),
		m.Latitude,
		m.Longitude,
		m.CreatedAt,
		m.UpdatedAt,
	)
	if err != nil {
		// При чтении из БД данные уже валидны, ошибка невозможна
		panic("invalid enterprise data in database: " + err.Error())
	}
	return ent
}

// ToModel преобразует доменный агрегат в GORM-модель.
func ToModel(e *domain.Enterprise) *EnterpriseModel {
	return &EnterpriseModel{
		ID:        e.ID(),
		Name:      e.Name(),
		Address:   e.Address(),
		Status:    string(e.Status()),
		Latitude:  e.Latitude(),
		Longitude: e.Longitude(),
		CreatedAt: e.CreatedAt(),
		UpdatedAt: e.UpdatedAt(),
	}
}
