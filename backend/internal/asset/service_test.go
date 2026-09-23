package asset

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type fakeRepository struct {
	Repository
	parent  Asset
	created *Equipment
}

func (f *fakeRepository) FindAsset(context.Context, uuid.UUID) (Asset, error) {
	return f.parent, nil
}

func (f *fakeRepository) CreateEquipment(_ context.Context, equipment *Equipment) error {
	f.created = equipment
	return nil
}

func TestDepreciationCannotExceedCost(t *testing.T) {
	err := validateAsset(AssetRequest{AcquisitionValue: 1000, AccumulatedDepreciation: 1001})
	if !errors.Is(err, errAssetDepreciationTooHigh) {
		t.Fatalf("got %v, want errAssetDepreciationTooHigh", err)
	}
	if err := validateAsset(AssetRequest{AcquisitionValue: 1000, AccumulatedDepreciation: 1000}); err != nil {
		t.Fatalf("fully depreciated asset must be accepted, got %v", err)
	}
}

func TestBookValueAndMonthlyDepreciation(t *testing.T) {
	response := newAssetResponse(Asset{AcquisitionValue: 1200000, AccumulatedDepreciation: 300000, UsefulLifeMonths: 48})
	if response.BookValue != 900000 || response.MonthlyDepreciation != 25000 {
		t.Fatalf("got book %d monthly %d", response.BookValue, response.MonthlyDepreciation)
	}
	if monthlyDepreciation(1000, 0) != 0 {
		t.Fatal("asset without useful life must not depreciate")
	}
}

func TestOperatingEquipmentNeedsProject(t *testing.T) {
	repository := &fakeRepository{}

	_, err := NewService(repository).CreateEquipment(context.Background(), EquipmentRequest{Status: equipmentOperating})
	if !errors.Is(err, errEquipmentProjectRequired) || repository.created != nil {
		t.Fatalf("got %v created %v", err, repository.created)
	}
}

func TestEquipmentCannotUseDisposedAsset(t *testing.T) {
	repository := &fakeRepository{parent: Asset{Status: assetDisposed}}
	assetID := uuid.New()

	_, err := NewService(repository).CreateEquipment(context.Background(), EquipmentRequest{Status: "idle", AssetID: &assetID})
	if !errors.Is(err, errEquipmentAssetDisposed) {
		t.Fatalf("got %v, want errEquipmentAssetDisposed", err)
	}
}
