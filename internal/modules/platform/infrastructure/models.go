package infrastructure

import (
	"github.com/fortis/backend/internal/modules/platform/domain"
)

type PlatformModel struct {
	Status string `gorm:"column:status"`
}

func (PlatformModel) TableName() string {
	return "platform"
}

func ToDomain(m *PlatformModel) *domain.Platform {
	return domain.NewPlatform(m.Status)
}

func ToModel(p *domain.Platform) *PlatformModel {
	return &PlatformModel{Status: p.Status()}
}
