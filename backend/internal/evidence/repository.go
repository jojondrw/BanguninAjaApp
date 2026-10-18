package evidence

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

const batchListLimit = 10

type Repository interface {
	Transaction(ctx context.Context, work func(Repository) error) error
	RecordExists(ctx context.Context, kind string, subjectID uuid.UUID, digest string) (bool, error)
	CreateRecord(ctx context.Context, record *Record) error
	RecordsFor(ctx context.Context, projectID *uuid.UUID) ([]Record, error)
	UnbatchedRecords(ctx context.Context) ([]Record, error)
	PendingBatch(ctx context.Context) (Batch, bool, error)
	ReleaseBatch(ctx context.Context, batchID uuid.UUID) error
	CreateBatch(ctx context.Context, batch *Batch, records []Record) error
	FindBatch(ctx context.Context, id uuid.UUID) (Batch, error)
	SaveBatch(ctx context.Context, batch *Batch) error
	AnchoredBatches(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]Batch, error)
	ListBatches(ctx context.Context) ([]Batch, error)
	FindApprover(ctx context.Context, projectID uuid.UUID) (Approver, bool, error)
	SaveApprover(ctx context.Context, approver *Approver) error
	FindClaimByTerm(ctx context.Context, termID uuid.UUID) (ClaimLink, bool, error)
	FindClaimByToken(ctx context.Context, token string) (ClaimLink, error)
	ListClaims(ctx context.Context, projectID uuid.UUID) ([]ClaimLink, error)
	SaveClaim(ctx context.Context, claim *ClaimLink) error
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Transaction(ctx context.Context, work func(Repository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return work(&gormRepository{db: tx})
	})
}

func (r *gormRepository) RecordExists(ctx context.Context, kind string, subjectID uuid.UUID, digest string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Record{}).
		Where("kind = ? AND subject_id = ? AND digest = ?", kind, subjectID, digest).Count(&count).Error
	return count > 0, database.Translate(err)
}

func (r *gormRepository) CreateRecord(ctx context.Context, record *Record) error {
	return database.Translate(r.db.WithContext(ctx).Create(record).Error)
}

func (r *gormRepository) RecordsFor(ctx context.Context, projectID *uuid.UUID) ([]Record, error) {
	query := r.db.WithContext(ctx).Order("created_at, id")
	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}
	var records []Record
	return records, database.Translate(query.Find(&records).Error)
}

func (r *gormRepository) UnbatchedRecords(ctx context.Context) ([]Record, error) {
	var records []Record
	err := r.db.WithContext(ctx).Where("batch_id IS NULL").Order("created_at, id").Find(&records).Error
	return records, database.Translate(err)
}

func (r *gormRepository) PendingBatch(ctx context.Context) (Batch, bool, error) {
	var batch Batch
	err := r.db.WithContext(ctx).Where("tx_hash IS NULL").Order("created_at DESC").Take(&batch).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Batch{}, false, nil
	}
	return batch, err == nil, database.Translate(err)
}

func (r *gormRepository) ReleaseBatch(ctx context.Context, batchID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&Record{}).Where("batch_id = ?", batchID).Update("batch_id", nil).Error; err != nil {
		return database.Translate(err)
	}
	return database.Translate(r.db.WithContext(ctx).Delete(&Batch{}, "id = ?", batchID).Error)
}

func (r *gormRepository) CreateBatch(ctx context.Context, batch *Batch, records []Record) error {
	if err := r.db.WithContext(ctx).Create(batch).Error; err != nil {
		return database.Translate(err)
	}
	ids := make([]uuid.UUID, len(records))
	for index, record := range records {
		ids[index] = record.ID
	}
	return database.Translate(r.db.WithContext(ctx).Model(&Record{}).Where("id IN ?", ids).Update("batch_id", batch.ID).Error)
}

func (r *gormRepository) FindBatch(ctx context.Context, id uuid.UUID) (Batch, error) {
	var batch Batch
	return batch, database.Translate(r.db.WithContext(ctx).Where("id = ?", id).Take(&batch).Error)
}

func (r *gormRepository) SaveBatch(ctx context.Context, batch *Batch) error {
	return database.Translate(r.db.WithContext(ctx).Save(batch).Error)
}

func (r *gormRepository) AnchoredBatches(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]Batch, error) {
	found := map[uuid.UUID]Batch{}
	if len(ids) == 0 {
		return found, nil
	}
	var batches []Batch
	if err := r.db.WithContext(ctx).Where("id IN ? AND tx_hash IS NOT NULL", ids).Find(&batches).Error; err != nil {
		return nil, database.Translate(err)
	}
	for _, batch := range batches {
		found[batch.ID] = batch
	}
	return found, nil
}

func (r *gormRepository) ListBatches(ctx context.Context) ([]Batch, error) {
	var batches []Batch
	err := r.db.WithContext(ctx).Order("created_at DESC").Limit(batchListLimit).Find(&batches).Error
	return batches, database.Translate(err)
}

func (r *gormRepository) FindApprover(ctx context.Context, projectID uuid.UUID) (Approver, bool, error) {
	var approver Approver
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Take(&approver).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Approver{}, false, nil
	}
	return approver, err == nil, database.Translate(err)
}

func (r *gormRepository) SaveApprover(ctx context.Context, approver *Approver) error {
	return database.Translate(r.db.WithContext(ctx).Save(approver).Error)
}

func (r *gormRepository) FindClaimByTerm(ctx context.Context, termID uuid.UUID) (ClaimLink, bool, error) {
	var claim ClaimLink
	err := r.db.WithContext(ctx).Where("term_id = ?", termID).Take(&claim).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ClaimLink{}, false, nil
	}
	return claim, err == nil, database.Translate(err)
}

func (r *gormRepository) FindClaimByToken(ctx context.Context, token string) (ClaimLink, error) {
	var claim ClaimLink
	return claim, database.Translate(r.db.WithContext(ctx).Where("token = ?", token).Take(&claim).Error)
}

func (r *gormRepository) ListClaims(ctx context.Context, projectID uuid.UUID) ([]ClaimLink, error) {
	var claims []ClaimLink
	return claims, database.Translate(r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&claims).Error)
}

func (r *gormRepository) SaveClaim(ctx context.Context, claim *ClaimLink) error {
	return database.Translate(r.db.WithContext(ctx).Save(claim).Error)
}
