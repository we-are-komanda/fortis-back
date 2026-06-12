package ui

import "github.com/fortis/backend/internal/modules/defense_project/domain"

func domainToImportResponse(p *domain.DefenseProject) ImportResponse {
	var resp ImportResponse
	resp.Body.ProjectID = p.ProjectID()
	resp.Body.ProjectName = p.ProjectName()
	resp.Body.UpdatedAt = p.UpdatedAt().UTC().Format("2006-01-02T15:04:05.000Z")
	return resp
}
