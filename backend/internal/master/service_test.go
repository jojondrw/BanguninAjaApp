package master

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

type fakeRepository struct {
	Repository
	regions     map[uuid.UUID]Region
	accounts    map[uuid.UUID]Account
	descendants map[uuid.UUID]uuid.UUID
}

func (f *fakeRepository) FindRegion(_ context.Context, id uuid.UUID) (Region, error) {
	region, ok := f.regions[id]
	if !ok {
		return Region{}, database.ErrNotFound
	}
	return region, nil
}

func (f *fakeRepository) FindAccount(_ context.Context, id uuid.UUID) (Account, error) {
	account, ok := f.accounts[id]
	if !ok {
		return Account{}, database.ErrNotFound
	}
	return account, nil
}

func (f *fakeRepository) AccountDescendsFrom(_ context.Context, candidateID, ancestorID uuid.UUID) (bool, error) {
	return f.descendants[candidateID] == ancestorID, nil
}

func TestValidateRegionParent(t *testing.T) {
	province := Region{Type: "province"}
	province.ID = uuid.New()
	city := Region{Type: "city"}
	city.ID = uuid.New()
	missing := uuid.New()
	service := &service{repository: &fakeRepository{regions: map[uuid.UUID]Region{province.ID: province, city.ID: city}}}

	cases := []struct {
		name    string
		request RegionRequest
		want    error
	}{
		{"province without parent", RegionRequest{Type: "province"}, nil},
		{"province with parent", RegionRequest{Type: "province", ParentID: &province.ID}, errRegionParentInvalid},
		{"city without parent", RegionRequest{Type: "city"}, errRegionParentRequired},
		{"city under province", RegionRequest{Type: "city", ParentID: &province.ID}, nil},
		{"district under city", RegionRequest{Type: "district", ParentID: &city.ID}, nil},
		{"district under province", RegionRequest{Type: "district", ParentID: &province.ID}, errRegionParentInvalid},
		{"unknown parent", RegionRequest{Type: "city", ParentID: &missing}, errRegionParentNotFound},
	}
	for _, tc := range cases {
		err := service.validateRegionParent(context.Background(), tc.request)
		if !errors.Is(err, tc.want) && !(err == nil && tc.want == nil) {
			t.Errorf("%s: got %v, want %v", tc.name, err, tc.want)
		}
	}
}

func TestValidateAccountParent(t *testing.T) {
	root := Account{Type: "asset"}
	root.ID = uuid.New()
	child := Account{Type: "asset", ParentID: &root.ID}
	child.ID = uuid.New()
	repository := &fakeRepository{
		accounts:    map[uuid.UUID]Account{root.ID: root, child.ID: child},
		descendants: map[uuid.UUID]uuid.UUID{child.ID: root.ID},
	}
	service := &service{repository: repository}

	if err := service.validateAccountParent(context.Background(), uuid.Nil, AccountRequest{Type: "liability", ParentID: &root.ID}); !errors.Is(err, errAccountParentTypeMismatch) {
		t.Errorf("type mismatch: got %v", err)
	}
	if err := service.validateAccountParent(context.Background(), root.ID, AccountRequest{Type: "asset", ParentID: &root.ID}); !errors.Is(err, errAccountParentCycle) {
		t.Errorf("self parent: got %v", err)
	}
	if err := service.validateAccountParent(context.Background(), root.ID, AccountRequest{Type: "asset", ParentID: &child.ID}); !errors.Is(err, errAccountParentCycle) {
		t.Errorf("descendant parent: got %v", err)
	}
	if err := service.validateAccountParent(context.Background(), uuid.Nil, AccountRequest{Type: "asset", ParentID: &root.ID}); err != nil {
		t.Errorf("valid parent: got %v", err)
	}
}

func TestBlankCodeOrNameIsRefusedBeforeTouchingTheDatabase(t *testing.T) {
	service := &service{repository: &fakeRepository{}}
	ctx := context.Background()
	id := uuid.New()

	cases := []struct {
		name string
		err  error
		want error
	}{
		{"create account", second(service.CreateAccount(ctx, AccountRequest{Code: "   ", Name: "Kas", Type: "asset"})), errAccountBlank},
		{"update account", second(service.UpdateAccount(ctx, id, AccountRequest{Code: "1110", Name: " ", Type: "asset"})), errAccountBlank},
		{"create region", second(service.CreateRegion(ctx, RegionRequest{Code: "	", Name: "Bogor", Type: "province"})), errRegionBlank},
		{"update region", second(service.UpdateRegion(ctx, id, RegionRequest{Code: "32", Name: "  ", Type: "province"})), errRegionBlank},
		{"create unit", second(service.CreateUnitOfMeasure(ctx, UnitOfMeasureRequest{Code: " ", Name: "Meter"})), errUnitOfMeasureBlank},
		{"update unit", second(service.UpdateUnitOfMeasure(ctx, id, UnitOfMeasureRequest{Code: "m", Name: "  "})), errUnitOfMeasureBlank},
	}
	for _, tc := range cases {
		if !errors.Is(tc.err, tc.want) {
			t.Errorf("%s: got %v, want %v", tc.name, tc.err, tc.want)
		}
	}
}

func second[T any](_ T, err error) error {
	return err
}
