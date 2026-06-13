package ui

import "github.com/fortis/backend/internal/modules/defense_project/domain"

func domainToImportResponse(p *domain.DefenseProject) ImportResponse {
	var resp ImportResponse
	resp.Body.ProjectID = p.ProjectID()
	resp.Body.ProjectName = p.ProjectName()
	resp.Body.Version = p.Version()
	resp.Body.UpdatedAt = p.UpdatedAt().UTC().Format("2006-01-02T15:04:05.000Z")
	return resp
}

func domainToProjectResponse(p *domain.DefenseProject) ProjectResponse {
	return ProjectResponse{
		ProjectID:    p.ProjectID(),
		Name:         p.Name(),
		EnterpriseID: p.EnterpriseID(),
		ProjectName:  p.ProjectName(),
		Version:      p.Version(),
		UpdatedAt:    p.UpdatedAt().UTC().Format("2006-01-02T15:04:05.000Z"),
	}
}

func domainToProjectListResponse(projects []*domain.DefenseProject, total int64) ProjectListResponse {
	items := make([]ProjectResponse, len(projects))
	for i, p := range projects {
		items[i] = domainToProjectResponse(p)
	}
	return ProjectListResponse{
		Items:      items,
		TotalItems: total,
	}
}
