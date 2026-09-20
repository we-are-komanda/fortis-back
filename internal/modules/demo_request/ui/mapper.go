package ui

import (
	"github.com/fortis/backend/internal/modules/demo_request/application"
	"github.com/fortis/backend/internal/modules/demo_request/domain"
)

func toCommand(input DemoRequestInput, key string) application.SubmitCommand {
	return application.SubmitCommand{Name: input.Name, Organization: input.Organization, Email: input.Email, Comment: input.Comment, Consent: input.Consent, ConsentVersion: input.ConsentVersion, IdempotencyKey: key}
}
func toRegistryPage(deliveries []*domain.Delivery, total int64) RegistryPage {
	page := RegistryPage{Items: make([]RegistryItem, 0, len(deliveries)), TotalItems: total}
	for _, delivery := range deliveries {
		r := delivery.Request()
		page.Items = append(page.Items, RegistryItem{RequestID: r.ID(), EventID: r.EventID(), Name: r.Name(), Organization: r.Organization(), Email: r.Email(), Comment: r.Comment(), ConsentVersion: r.ConsentVersion(), ConsentAt: r.ConsentAt(), Status: delivery.Status(), Attempts: delivery.Attempts(), LastError: delivery.LastError(), NextAttemptAt: delivery.DueAt()})
	}
	return page
}
