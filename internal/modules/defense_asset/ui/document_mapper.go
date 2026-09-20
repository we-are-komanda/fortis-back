package ui

import (
	"net/url"
	"time"

	"github.com/fortis/backend/internal/modules/defense_asset/application"
	"github.com/fortis/backend/internal/modules/defense_asset/domain"
)

// documentToDTO преобразует доменный Document в DTO ответа.
func documentToDTO(doc *domain.Document) AssetDocumentDTO {
	download := ""
	if doc.Status() == "ready" {
		download = "/api/v1/assets/documents/download?id=" + url.QueryEscape(doc.ID()) + "&assetId=" + url.QueryEscape(doc.AssetID())
	}
	return AssetDocumentDTO{
		Revision: doc.Revision(), Checksum: nullable(doc.Checksum()), Status: doc.Status(), Commercial: doc.Commercial(),
		ID:          doc.ID(),
		AssetID:     doc.AssetID(),
		Name:        doc.Name(),
		MimeType:    doc.MimeType(),
		SizeBytes:   doc.SizeBytes(),
		DownloadURL: download,
		OwnerID:     doc.OwnerID(),
		CreatedAt:   doc.CreatedAt().UTC().Format(time.RFC3339),
		UpdatedAt:   doc.UpdatedAt().UTC().Format(time.RFC3339),
	}
}

// mapCreateDocumentRequestToServiceInput преобразует DTO создания в входные данные сервиса.
func mapCreateDocumentRequestToServiceInput(req CreateDocumentRequest) application.CreateDocumentInput {
	return application.CreateDocumentInput{
		AssetID:     req.AssetID,
		Name:        req.Name,
		MimeType:    req.MimeType,
		SizeBytes:   req.SizeBytes,
		StorageKey:  req.StorageKey,
		DownloadURL: req.DownloadURL,
		OwnerID:     req.OwnerID,
	}
}
