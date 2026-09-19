package project

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

type Project struct {
	entity.Base
	Code          string     `gorm:"type:varchar(20);not null;uniqueIndex:uq_project_code"`
	Name          string     `gorm:"type:varchar(160);not null"`
	RegionID      *uuid.UUID `gorm:"type:uuid;index:idx_project_region"`
	Type          string     `gorm:"type:varchar(40);not null"`
	Status        string     `gorm:"type:varchar(20);not null;index:idx_project_status"`
	Progress      int        `gorm:"not null;default:0"`
	StartDate     *time.Time `gorm:"type:date"`
	TargetEndDate *time.Time `gorm:"type:date"`
	ContractValue int64      `gorm:"not null;default:0"`
}

func (Project) TableName() string {
	return "project"
}

type ProjectPhase struct {
	entity.Base
	ProjectID     uuid.UUID  `gorm:"type:uuid;not null;index:idx_stage_project"`
	Name          string     `gorm:"type:varchar(160);not null"`
	SortOrder     int        `gorm:"not null;default:1"`
	StartDate     *time.Time `gorm:"type:date"`
	TargetEndDate *time.Time `gorm:"type:date"`
	Progress      int        `gorm:"not null;default:0"`
	Status        string     `gorm:"type:varchar(20);not null"`
}

func (ProjectPhase) TableName() string {
	return "project_phase"
}

type BudgetItem struct {
	entity.Base
	ProjectID       uuid.UUID `gorm:"type:uuid;not null;index:idx_budget_item_project"`
	Code            string    `gorm:"type:varchar(20);not null"`
	Description     string    `gorm:"type:varchar(200);not null"`
	Volume          float64   `gorm:"type:numeric(14,2);not null;default:0"`
	UnitOfMeasureID uuid.UUID `gorm:"type:uuid;not null;index:idx_budget_item_unit_of_measure"`
	UnitPrice       int64     `gorm:"not null;default:0"`
	Total           int64     `gorm:"not null;default:0"`
}

func (BudgetItem) TableName() string {
	return "budget_item"
}

type Permit struct {
	entity.Base
	ProjectID  uuid.UUID  `gorm:"type:uuid;not null;index:idx_permit_project"`
	Type       string     `gorm:"type:varchar(60);not null"`
	Number     string     `gorm:"type:varchar(80);not null"`
	IssuedDate *time.Time `gorm:"type:date"`
	ValidUntil *time.Time `gorm:"type:date;index:idx_permit_valid_until"`
	Status     string     `gorm:"type:varchar(20);not null;index:idx_permit_status"`
}

func (Permit) TableName() string {
	return "permit"
}

func Entities() []any {
	return []any{&Project{}, &ProjectPhase{}, &BudgetItem{}, &Permit{}}
}

func Indexes() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_project_name_trgm ON project USING gin (name gin_trgm_ops)`,
		`CREATE INDEX IF NOT EXISTS idx_project_code_pattern ON project (code text_pattern_ops)`,
	}
}

func Constraints() []string {
	return []string{
		database.ForeignKey("project", "region_id", "region", database.DeleteRestrict),
		database.ForeignKey("project_phase", "project_id", "project", database.DeleteCascade),
		database.ForeignKey("budget_item", "project_id", "project", database.DeleteCascade),
		database.ForeignKey("budget_item", "unit_of_measure_id", "unit_of_measure", database.DeleteRestrict),
		database.ForeignKey("permit", "project_id", "project", database.DeleteCascade),
		database.Check("project", "progress", "progress BETWEEN 0 AND 100"),
		database.Check("project", "status", "status IN ('planning','ongoing','on_hold','completed','cancelled')"),
		database.Check("project", "contract_value", "contract_value >= 0"),
		database.Check("project", "date", "target_end_date IS NULL OR start_date IS NULL OR target_end_date >= start_date"),
		database.Check("project_phase", "progress", "progress BETWEEN 0 AND 100"),
		database.Check("budget_item", "volume", "volume >= 0"),
		database.Check("budget_item", "unit_price", "unit_price >= 0"),
		database.Check("permit", "status", "status IN ('submitted','issued','expired','rejected')"),
		database.Unique("budget_item", "project_code", "project_id, code"),
	}
}
