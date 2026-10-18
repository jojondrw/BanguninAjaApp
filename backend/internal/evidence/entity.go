package evidence

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

type Record struct {
	entity.Base
	Kind      string     `gorm:"type:varchar(30);not null;uniqueIndex:uq_evidence_record_version,priority:1"`
	SubjectID uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:uq_evidence_record_version,priority:2;index:idx_evidence_record_subject"`
	Digest    string     `gorm:"type:char(64);not null;uniqueIndex:uq_evidence_record_version,priority:3"`
	ProjectID *uuid.UUID `gorm:"type:uuid;index:idx_evidence_record_project"`
	BatchID   *uuid.UUID `gorm:"type:uuid;index:idx_evidence_record_batch"`
}

func (Record) TableName() string {
	return "evidence_record"
}

type Batch struct {
	entity.Base
	Root         string     `gorm:"type:char(66);not null"`
	RecordCount  int        `gorm:"not null"`
	ChainID      int64      `gorm:"not null"`
	CreatedBy    *uuid.UUID `gorm:"type:uuid;index:idx_evidence_batch_created_by"`
	TxHash       *string    `gorm:"type:char(66);uniqueIndex:uq_evidence_batch_tx"`
	AnchoredFrom string     `gorm:"type:varchar(42);not null;default:''"`
	BlockNumber  *int64
	AnchoredAt   *time.Time
}

func (Batch) TableName() string {
	return "evidence_batch"
}

type Approver struct {
	entity.Base
	ProjectID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uq_evidence_approver_project"`
	Address   string    `gorm:"type:varchar(42);not null"`
}

func (Approver) TableName() string {
	return "evidence_approver"
}

type ClaimLink struct {
	entity.Base
	ProjectID    uuid.UUID `gorm:"type:uuid;not null;index:idx_evidence_claim_project"`
	TermID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uq_evidence_claim_term"`
	Token        string    `gorm:"type:varchar(40);not null;uniqueIndex:uq_evidence_claim_token"`
	Signature    string    `gorm:"type:varchar(140);not null;default:''"`
	Signer       string    `gorm:"type:varchar(42);not null;default:''"`
	EvidenceRoot string    `gorm:"type:varchar(66);not null;default:''"`
	ApprovedAt   *time.Time
}

func (ClaimLink) TableName() string {
	return "evidence_claim"
}

func Entities() []any {
	return []any{&Batch{}, &Record{}, &Approver{}, &ClaimLink{}}
}

func Indexes() []string {
	return nil
}

func Constraints() []string {
	return []string{
		database.ForeignKey("evidence_record", "batch_id", "evidence_batch", database.DeleteSetNull),
		database.ForeignKey("evidence_batch", "created_by", "users", database.DeleteSetNull),
		database.ForeignKey("evidence_approver", "project_id", "project", database.DeleteCascade),
		database.ForeignKey("evidence_claim", "project_id", "project", database.DeleteCascade),
		database.ForeignKey("evidence_claim", "term_id", "owner_term", database.DeleteCascade),
	}
}
