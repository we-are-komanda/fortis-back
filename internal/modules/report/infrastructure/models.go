package infrastructure

import "time"

type IssuedReportModel struct {
	ID                  string `gorm:"primaryKey;type:uuid"`
	ProjectID           string `gorm:"type:uuid"`
	ProjectVersion      int
	EnterpriseID        string `gorm:"type:uuid"`
	ReportSchemaVersion string
	HideCost            bool
	MapView             string
	GeneratedAt         time.Time
	CreatedBy           string `gorm:"type:uuid"`
	PayloadJSON         string `gorm:"type:jsonb"`
}

func (IssuedReportModel) TableName() string { return "issued_reports" }
