package application

import "github.com/fortis/backend/internal/modules/report/domain"

type GetReportQuery struct {
	ReportID string
	Key      domain.ReportKey
}
