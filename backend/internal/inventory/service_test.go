package inventory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type stockKey struct {
	material  uuid.UUID
	warehouse uuid.UUID
}

type fakeRepository struct {
	Repository
	stock     map[stockKey]float64
	movements []StockMovement
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{stock: map[stockKey]float64{}}
}

func (f *fakeRepository) Transaction(_ context.Context, work func(Repository) error) error {
	return work(f)
}

func (f *fakeRepository) CreateStockMovement(_ context.Context, movement *StockMovement) error {
	f.movements = append(f.movements, *movement)
	return nil
}

func (f *fakeRepository) IncreaseStock(_ context.Context, materialID, warehouseID uuid.UUID, quantity float64) error {
	f.stock[stockKey{materialID, warehouseID}] += quantity
	return nil
}

func (f *fakeRepository) DecreaseStock(_ context.Context, materialID, warehouseID uuid.UUID, quantity float64) (bool, error) {
	key := stockKey{materialID, warehouseID}
	if f.stock[key] < quantity {
		return false, nil
	}
	f.stock[key] -= quantity
	return true, nil
}

func TestDirectionMatchesType(t *testing.T) {
	cases := []struct {
		movementType         string
		hasSource, hasTarget bool
		valid                bool
	}{
		{movementIn, false, true, true},
		{movementIn, true, true, false},
		{movementOut, true, false, true},
		{movementOut, false, true, false},
		{movementTransfer, true, true, true},
		{movementTransfer, true, false, false},
		{movementAdjustment, true, false, true},
		{movementAdjustment, false, true, true},
		{movementAdjustment, true, true, false},
		{movementAdjustment, false, false, false},
	}
	for _, tc := range cases {
		if got := directionMatchesType(tc.movementType, tc.hasSource, tc.hasTarget); got != tc.valid {
			t.Errorf("%s source=%v target=%v: got %v, want %v", tc.movementType, tc.hasSource, tc.hasTarget, got, tc.valid)
		}
	}
}

func TestTransferToSameWarehouseIsRejected(t *testing.T) {
	warehouse := uuid.New()
	service := NewService(newFakeRepository())

	_, err := service.RecordStockMovement(context.Background(), StockMovementRequest{
		Date: time.Now(), Type: movementTransfer, MaterialID: uuid.New(), Quantity: 1,
		SourceWarehouseID: &warehouse, TargetWarehouseID: &warehouse,
	})
	if !errors.Is(err, errMovementSameTarget) {
		t.Fatalf("got %v, want errMovementSameTarget", err)
	}
}

func TestOutMovementWithoutEnoughStockIsRejected(t *testing.T) {
	material, warehouse := uuid.New(), uuid.New()
	repository := newFakeRepository()
	repository.stock[stockKey{material, warehouse}] = 5
	service := NewService(repository)

	_, err := service.RecordStockMovement(context.Background(), StockMovementRequest{
		Date: time.Now(), Type: movementOut, MaterialID: material, Quantity: 8, SourceWarehouseID: &warehouse,
	})
	if !errors.Is(err, errInsufficientStock) {
		t.Fatalf("got %v, want errInsufficientStock", err)
	}
	if repository.stock[stockKey{material, warehouse}] != 5 {
		t.Fatal("stock must stay untouched")
	}
}

func TestTransferMovesStockBetweenWarehouses(t *testing.T) {
	material, source, target := uuid.New(), uuid.New(), uuid.New()
	repository := newFakeRepository()
	repository.stock[stockKey{material, source}] = 10
	service := NewService(repository)

	_, err := service.RecordStockMovement(context.Background(), StockMovementRequest{
		Date: time.Now(), Type: movementTransfer, MaterialID: material, Quantity: 4.255,
		SourceWarehouseID: &source, TargetWarehouseID: &target,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := repository.stock[stockKey{material, source}]; got != 10-4.26 {
		t.Fatalf("source stock got %v", got)
	}
	if got := repository.stock[stockKey{material, target}]; got != 4.26 {
		t.Fatalf("target stock got %v", got)
	}
	if len(repository.movements) != 1 {
		t.Fatalf("got %d movements, want 1", len(repository.movements))
	}
}
