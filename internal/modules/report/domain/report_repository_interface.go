package domain

import "context"

type ReportRepository interface {
	FindByKey(context.Context, string, ReportKey) (*IssuedReport, error)
	FindByID(context.Context, string, string) (*IssuedReport, error)
	Issue(context.Context, string, *IssuedReport) (*IssuedReport, error)
}
