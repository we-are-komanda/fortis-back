package ui

import "github.com/fortis/backend/internal/modules/platform/domain"

func mapPlatformToDTO(p *domain.Platform) PlatformDTO {
	return PlatformDTO{Status: p.Status()}
}
