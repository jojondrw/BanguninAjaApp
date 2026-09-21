package reporting

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

type Report struct {
	entity.Base
	Name        string     `gorm:"type:varchar(160);not null"`
	Type        string     `gorm:"type:varchar(40);not null;index:idx_report_type"`
	Scope       string     `gorm:"type:varchar(120)"`
	ProjectID   *uuid.UUID `gorm:"type:uuid;index:idx_report_project"`
	PeriodStart *time.Time `gorm:"type:date"`
	PeriodEnd   *time.Time `gorm:"type:date"`
	Format      string     `gorm:"type:varchar(10);not null"`
	CreatedBy   uuid.UUID  `gorm:"type:uuid;not null;index:idx_report_created_by"`
}

func (Report) TableName() string {
	return "report"
}

func Entities() []any {
	return []any{&Report{}}
}

func Indexes() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_report_created ON report (created_at DESC)`,
	}
}

func Constraints() []string {
	return []string{
		database.ForeignKey("report", "project_id", "project", database.DeleteCascade),
		database.ForeignKey("report", "created_by", "users", database.DeleteRestrict),
		database.Check("report", "format", "format IN ('pdf','xlsx','csv')"),
		database.Check("report", "period", "period_end IS NULL OR period_start IS NULL OR period_end >= period_start"),
	}
}
