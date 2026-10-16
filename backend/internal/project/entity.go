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
	ProjectID       uuid.UUID  `gorm:"type:uuid;not null;index:idx_budget_item_project"`
	PhaseID         *uuid.UUID `gorm:"type:uuid;index:idx_budget_item_phase"`
	Code            string     `gorm:"type:varchar(20);not null"`
	Description     string     `gorm:"type:varchar(200);not null"`
	Volume          float64    `gorm:"type:numeric(14,2);not null;default:0"`
	UnitOfMeasureID uuid.UUID  `gorm:"type:uuid;not null;index:idx_budget_item_unit_of_measure"`
	UnitPrice       int64      `gorm:"not null;default:0"`
	Total           int64      `gorm:"not null;default:0"`
}

func (BudgetItem) TableName() string {
	return "budget_item"
}

type ProgressSnapshot struct {
	entity.Base
	ProjectID  uuid.UUID `gorm:"type:uuid;not null;index:idx_project_progress_project"`
	RecordedOn time.Time `gorm:"type:date;not null"`
	Progress   int       `gorm:"not null"`
}

func (ProgressSnapshot) TableName() string {
	return "project_progress"
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

type ProjectMember struct {
	entity.Base
	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_project_member_user"`
	ProjectID uuid.UUID `gorm:"type:uuid;not null;index:idx_project_member_project"`
	RoleID    uuid.UUID `gorm:"type:uuid;not null;index:idx_project_member_role"`
	Position  *string   `gorm:"type:varchar(80)"`
}

func (ProjectMember) TableName() string {
	return "project_member"
}

func Entities() []any {
	return []any{&Project{}, &ProjectPhase{}, &BudgetItem{}, &ProgressSnapshot{}, &DailyReport{}, &ConstructionContract{}, &PaymentTerm{}, &ContractChange{}, &Permit{}, &ProjectMember{}}
}

func Indexes() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_project_name_trgm ON project USING gin (name gin_trgm_ops)`,
		`CREATE INDEX IF NOT EXISTS idx_project_code_pattern ON project (code text_pattern_ops)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_payment_term_retention ON payment_term (project_id) WHERE kind = 'retention'`,
	}
}

func Constraints() []string {
	return []string{
		database.ForeignKey("project", "region_id", "region", database.DeleteRestrict),
		database.ForeignKey("project_phase", "project_id", "project", database.DeleteCascade),
		database.ForeignKey("budget_item", "project_id", "project", database.DeleteCascade),
		database.ForeignKey("budget_item", "unit_of_measure_id", "unit_of_measure", database.DeleteRestrict),
		database.ForeignKey("budget_item", "phase_id", "project_phase", database.DeleteSetNull),
		database.ForeignKey("project_progress", "project_id", "project", database.DeleteCascade),
		database.ForeignKey("daily_report", "project_id", "project", database.DeleteCascade),
		database.ForeignKey("daily_report", "created_by", "users", database.DeleteSetNull),
		database.ForeignKey("construction_contract", "project_id", "project", database.DeleteCascade),
		database.ForeignKey("construction_contract", "vendor_id", "vendor", database.DeleteRestrict),
		database.ForeignKey("payment_term", "project_id", "project", database.DeleteCascade),
		database.ForeignKey("payment_term", "payable_id", "payable", database.DeleteSetNull),
		database.ForeignKey("payment_term", "claimed_by", "users", database.DeleteSetNull),
		database.ForeignKey("contract_change", "project_id", "project", database.DeleteCascade),
		database.ForeignKey("contract_change", "proposed_by", "users", database.DeleteSetNull),
		database.ForeignKey("contract_change", "decided_by", "users", database.DeleteSetNull),
		database.ForeignKey("permit", "project_id", "project", database.DeleteCascade),
		database.ForeignKey("project_member", "user_id", "users", database.DeleteCascade),
		database.ForeignKey("project_member", "project_id", "project", database.DeleteCascade),
		database.ForeignKey("project_member", "role_id", "role", database.DeleteRestrict),
		database.Check("project", "progress", "progress BETWEEN 0 AND 100"),
		database.Check("project", "status", "status IN ('planning','ongoing','on_hold','completed','cancelled')"),
		database.Check("project", "contract_value", "contract_value >= 0"),
		database.Check("project", "date", "target_end_date IS NULL OR start_date IS NULL OR target_end_date >= start_date"),
		database.Check("project_phase", "progress", "progress BETWEEN 0 AND 100"),
		database.Check("budget_item", "volume", "volume >= 0"),
		database.Check("budget_item", "unit_price", "unit_price >= 0"),
		database.Check("permit", "status", "status IN ('submitted','issued','expired','rejected')"),
		database.Check("project_progress", "progress", "progress BETWEEN 0 AND 100"),
		database.Check("daily_report", "weather", "weather_morning IN ('sunny','cloudy','drizzle','rain','heavy_rain') AND weather_afternoon IN ('sunny','cloudy','drizzle','rain','heavy_rain')"),
		database.Check("daily_report", "workers", "workers >= 0"),
		database.Check("construction_contract", "value", "value > 0"),
		database.Check("construction_contract", "retention", "retention_percent BETWEEN 0 AND 10"),
		database.Check("payment_term", "kind", "kind IN ('progress','retention')"),
		database.Check("payment_term", "required_progress", "required_progress BETWEEN 0 AND 100"),
		database.Check("payment_term", "percent", "percent BETWEEN 0 AND 100"),
		database.Check("payment_term", "amount", "gross_amount >= 0 AND retention_amount BETWEEN 0 AND gross_amount"),
		database.Check("contract_change", "status", "status IN ('proposed','approved','rejected')"),
		database.Check("contract_change", "not_empty", "value_change <> 0 OR days_change <> 0"),
		database.Check("contract_change", "days", "days_change >= 0"),
		database.Check("contract_change", "decided", "status = 'proposed' OR decided_on IS NOT NULL"),
		database.Unique("budget_item", "project_code", "project_id, code"),
		database.Unique("project_progress", "project_day", "project_id, recorded_on"),
		database.Unique("daily_report", "project_day", "project_id, report_date"),
		database.Unique("contract_change", "project_sequence", "project_id, sequence"),
		database.Unique("project_member", "user_project", "user_id, project_id"),
	}
}
