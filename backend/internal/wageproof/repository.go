package wageproof

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

const circuitName = "wage-minimum-v1"

type Repository interface {
	Transaction(ctx context.Context, work func(Repository) error) error
	ReplaceProof(ctx context.Context, proof *Proof, receipts []Receipt) error
	ListProofs(ctx context.Context, projectID *uuid.UUID) ([]Proof, error)
	ReceiptsFor(ctx context.Context, proofIDs []uuid.UUID) ([]Receipt, error)
	FindProofByToken(ctx context.Context, token string) (Proof, error)
	FindProof(ctx context.Context, id uuid.UUID) (Proof, error)
	FindReceiptByToken(ctx context.Context, token string) (Receipt, error)
	LoadKeys(ctx context.Context) (Keys, bool, error)
	StoreKeys(ctx context.Context, keys Keys) error
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

func (r *gormRepository) ReplaceProof(ctx context.Context, proof *Proof, receipts []Receipt) error {
	db := r.db.WithContext(ctx)
	if err := db.Where("project_id = ? AND period = ?", proof.ProjectID, proof.Period).Delete(&Proof{}).Error; err != nil {
		return database.Translate(err)
	}
	if err := db.Create(proof).Error; err != nil {
		return database.Translate(err)
	}
	for index := range receipts {
		receipts[index].ProofID = proof.ID
	}
	if len(receipts) == 0 {
		return nil
	}
	return database.Translate(db.Create(&receipts).Error)
}

func (r *gormRepository) ListProofs(ctx context.Context, projectID *uuid.UUID) ([]Proof, error) {
	query := r.db.WithContext(ctx).Order("period DESC")
	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}
	var proofs []Proof
	return proofs, database.Translate(query.Find(&proofs).Error)
}

func (r *gormRepository) ReceiptsFor(ctx context.Context, proofIDs []uuid.UUID) ([]Receipt, error) {
	var receipts []Receipt
	if len(proofIDs) == 0 {
		return receipts, nil
	}
	err := r.db.WithContext(ctx).Where("proof_id IN ?", proofIDs).Order("leaf_index").Find(&receipts).Error
	return receipts, database.Translate(err)
}

func (r *gormRepository) FindProofByToken(ctx context.Context, token string) (Proof, error) {
	var proof Proof
	return proof, database.Translate(r.db.WithContext(ctx).Where("token = ?", token).Take(&proof).Error)
}

func (r *gormRepository) FindProof(ctx context.Context, id uuid.UUID) (Proof, error) {
	var proof Proof
	return proof, database.Translate(r.db.WithContext(ctx).Where("id = ?", id).Take(&proof).Error)
}

func (r *gormRepository) FindReceiptByToken(ctx context.Context, token string) (Receipt, error) {
	var receipt Receipt
	return receipt, database.Translate(r.db.WithContext(ctx).Where("token = ?", token).Take(&receipt).Error)
}

func (r *gormRepository) LoadKeys(ctx context.Context) (Keys, bool, error) {
	var key CircuitKey
	err := r.db.WithContext(ctx).Where("name = ?", circuitName).Take(&key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Keys{}, false, nil
	}
	if err != nil {
		return Keys{}, false, database.Translate(err)
	}
	return Keys{Proving: key.Proving, Verifying: key.Verifying}, true, nil
}

func (r *gormRepository) StoreKeys(ctx context.Context, keys Keys) error {
	key := CircuitKey{Name: circuitName, Proving: keys.Proving, Verifying: keys.Verifying}
	return database.Translate(r.db.WithContext(ctx).Create(&key).Error)
}
