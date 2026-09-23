package master

import (
	"context"
	"slices"
	"strings"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

var (
	errRegionNotFound       = apperror.NotFound("region_not_found", "Wilayah tidak ditemukan")
	errRegionCodeUsed       = apperror.Conflict("region_code_used", "Kode wilayah sudah dipakai")
	errRegionInUse          = apperror.Conflict("region_in_use", "Wilayah masih dipakai oleh data lain")
	errRegionTypeLocked     = apperror.Conflict("region_type_locked", "Jenis wilayah tidak bisa diubah karena masih punya wilayah turunan")
	errRegionParentNotFound = apperror.Unprocessable("region_parent_not_found", "Wilayah induk tidak ditemukan")
	errRegionParentRequired = apperror.Unprocessable("region_parent_required", "Wilayah selain provinsi wajib punya wilayah induk")
	errRegionParentInvalid  = apperror.Unprocessable("region_parent_invalid", "Jenis wilayah induk tidak sesuai dengan jenis wilayah")

	errUnitOfMeasureNotFound = apperror.NotFound("unit_of_measure_not_found", "Satuan tidak ditemukan")
	errUnitOfMeasureCodeUsed = apperror.Conflict("unit_of_measure_code_used", "Kode satuan sudah dipakai")
	errUnitOfMeasureInUse    = apperror.Conflict("unit_of_measure_in_use", "Satuan masih dipakai oleh data lain")

	errAccountNotFound           = apperror.NotFound("account_not_found", "Akun tidak ditemukan")
	errAccountCodeUsed           = apperror.Conflict("account_code_used", "Kode akun sudah dipakai")
	errAccountInUse              = apperror.Conflict("account_in_use", "Akun masih dipakai oleh data lain")
	errAccountTypeLocked         = apperror.Conflict("account_type_locked", "Jenis akun tidak bisa diubah karena masih punya akun turunan")
	errAccountParentNotFound     = apperror.Unprocessable("account_parent_not_found", "Akun induk tidak ditemukan")
	errAccountParentTypeMismatch = apperror.Unprocessable("account_parent_type_mismatch", "Akun induk harus memiliki jenis yang sama")
	errAccountParentCycle        = apperror.Unprocessable("account_parent_cycle", "Akun induk tidak boleh akun itu sendiri atau turunannya")
)

var (
	regionReadErrors   = database.ErrorMap{NotFound: errRegionNotFound}
	regionWriteErrors  = database.ErrorMap{NotFound: errRegionNotFound, Duplicate: errRegionCodeUsed, Referenced: errRegionParentNotFound}
	regionDeleteErrors = database.ErrorMap{NotFound: errRegionNotFound, Referenced: errRegionInUse}
	regionParentErrors = database.ErrorMap{NotFound: errRegionParentNotFound}

	unitOfMeasureReadErrors   = database.ErrorMap{NotFound: errUnitOfMeasureNotFound}
	unitOfMeasureWriteErrors  = database.ErrorMap{NotFound: errUnitOfMeasureNotFound, Duplicate: errUnitOfMeasureCodeUsed}
	unitOfMeasureDeleteErrors = database.ErrorMap{NotFound: errUnitOfMeasureNotFound, Referenced: errUnitOfMeasureInUse}

	accountReadErrors   = database.ErrorMap{NotFound: errAccountNotFound}
	accountWriteErrors  = database.ErrorMap{NotFound: errAccountNotFound, Duplicate: errAccountCodeUsed, Referenced: errAccountParentNotFound}
	accountDeleteErrors = database.ErrorMap{NotFound: errAccountNotFound, Referenced: errAccountInUse}
	accountParentErrors = database.ErrorMap{NotFound: errAccountParentNotFound}
)

var regionParentTypes = map[string][]string{
	"province": nil,
	"city":     {"province"},
	"regency":  {"province"},
	"district": {"city", "regency"},
}

type Service interface {
	ListRegions(ctx context.Context, query RegionQuery) (pagination.Page[RegionResponse], error)
	GetRegion(ctx context.Context, id uuid.UUID) (RegionResponse, error)
	CreateRegion(ctx context.Context, request RegionRequest) (RegionResponse, error)
	UpdateRegion(ctx context.Context, id uuid.UUID, request RegionRequest) (RegionResponse, error)
	DeleteRegion(ctx context.Context, id uuid.UUID) error

	ListUnitsOfMeasure(ctx context.Context, query UnitOfMeasureQuery) (pagination.Page[UnitOfMeasureResponse], error)
	GetUnitOfMeasure(ctx context.Context, id uuid.UUID) (UnitOfMeasureResponse, error)
	CreateUnitOfMeasure(ctx context.Context, request UnitOfMeasureRequest) (UnitOfMeasureResponse, error)
	UpdateUnitOfMeasure(ctx context.Context, id uuid.UUID, request UnitOfMeasureRequest) (UnitOfMeasureResponse, error)
	DeleteUnitOfMeasure(ctx context.Context, id uuid.UUID) error

	ListAccounts(ctx context.Context, query AccountQuery) (pagination.Page[AccountResponse], error)
	GetAccount(ctx context.Context, id uuid.UUID) (AccountResponse, error)
	CreateAccount(ctx context.Context, request AccountRequest) (AccountResponse, error)
	UpdateAccount(ctx context.Context, id uuid.UUID, request AccountRequest) (AccountResponse, error)
	DeleteAccount(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) ListRegions(ctx context.Context, query RegionQuery) (pagination.Page[RegionResponse], error) {
	regions, total, err := s.repository.ListRegions(ctx, RegionFilter{
		Search:   query.Search,
		Type:     query.Type,
		ParentID: query.ParentID,
		Offset:   query.Offset(),
		Limit:    query.Size(),
	})
	if err != nil {
		return pagination.Page[RegionResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(regions, newRegionResponse), query.Query, total), nil
}

func (s *service) GetRegion(ctx context.Context, id uuid.UUID) (RegionResponse, error) {
	region, err := s.repository.FindRegion(ctx, id)
	if err != nil {
		return RegionResponse{}, regionReadErrors.Resolve(err)
	}
	return newRegionResponse(region), nil
}

func (s *service) CreateRegion(ctx context.Context, request RegionRequest) (RegionResponse, error) {
	if err := s.validateRegionParent(ctx, request); err != nil {
		return RegionResponse{}, err
	}

	var region Region
	applyRegionRequest(&region, request)
	if err := s.repository.CreateRegion(ctx, &region); err != nil {
		return RegionResponse{}, regionWriteErrors.Resolve(err)
	}
	return newRegionResponse(region), nil
}

func (s *service) UpdateRegion(ctx context.Context, id uuid.UUID, request RegionRequest) (RegionResponse, error) {
	region, err := s.repository.FindRegion(ctx, id)
	if err != nil {
		return RegionResponse{}, regionReadErrors.Resolve(err)
	}
	if err := s.ensureRegionTypeChangeable(ctx, region, request.Type); err != nil {
		return RegionResponse{}, err
	}
	if err := s.validateRegionParent(ctx, request); err != nil {
		return RegionResponse{}, err
	}

	applyRegionRequest(&region, request)
	if err := s.repository.SaveRegion(ctx, &region); err != nil {
		return RegionResponse{}, regionWriteErrors.Resolve(err)
	}
	return newRegionResponse(region), nil
}

func (s *service) DeleteRegion(ctx context.Context, id uuid.UUID) error {
	return regionDeleteErrors.Resolve(s.repository.DeleteRegion(ctx, id))
}

func (s *service) validateRegionParent(ctx context.Context, request RegionRequest) error {
	allowed := regionParentTypes[request.Type]
	if request.ParentID == nil && len(allowed) > 0 {
		return errRegionParentRequired
	}
	if request.ParentID == nil {
		return nil
	}
	if len(allowed) == 0 {
		return errRegionParentInvalid
	}

	parent, err := s.repository.FindRegion(ctx, *request.ParentID)
	if err != nil {
		return regionParentErrors.Resolve(err)
	}
	if !slices.Contains(allowed, parent.Type) {
		return errRegionParentInvalid
	}
	return nil
}

func (s *service) ensureRegionTypeChangeable(ctx context.Context, region Region, nextType string) error {
	if region.Type == nextType {
		return nil
	}

	children, err := s.repository.CountRegionChildren(ctx, region.ID)
	if err != nil {
		return apperror.Internal(err)
	}
	if children > 0 {
		return errRegionTypeLocked
	}
	return nil
}

func (s *service) ListUnitsOfMeasure(ctx context.Context, query UnitOfMeasureQuery) (pagination.Page[UnitOfMeasureResponse], error) {
	units, total, err := s.repository.ListUnitsOfMeasure(ctx, UnitOfMeasureFilter{
		Search: query.Search,
		Offset: query.Offset(),
		Limit:  query.Size(),
	})
	if err != nil {
		return pagination.Page[UnitOfMeasureResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(units, newUnitOfMeasureResponse), query.Query, total), nil
}

func (s *service) GetUnitOfMeasure(ctx context.Context, id uuid.UUID) (UnitOfMeasureResponse, error) {
	unit, err := s.repository.FindUnitOfMeasure(ctx, id)
	if err != nil {
		return UnitOfMeasureResponse{}, unitOfMeasureReadErrors.Resolve(err)
	}
	return newUnitOfMeasureResponse(unit), nil
}

func (s *service) CreateUnitOfMeasure(ctx context.Context, request UnitOfMeasureRequest) (UnitOfMeasureResponse, error) {
	unit := UnitOfMeasure{Code: strings.TrimSpace(request.Code), Name: strings.TrimSpace(request.Name)}
	if err := s.repository.CreateUnitOfMeasure(ctx, &unit); err != nil {
		return UnitOfMeasureResponse{}, unitOfMeasureWriteErrors.Resolve(err)
	}
	return newUnitOfMeasureResponse(unit), nil
}

func (s *service) UpdateUnitOfMeasure(ctx context.Context, id uuid.UUID, request UnitOfMeasureRequest) (UnitOfMeasureResponse, error) {
	unit, err := s.repository.FindUnitOfMeasure(ctx, id)
	if err != nil {
		return UnitOfMeasureResponse{}, unitOfMeasureReadErrors.Resolve(err)
	}

	unit.Code = strings.TrimSpace(request.Code)
	unit.Name = strings.TrimSpace(request.Name)
	if err := s.repository.SaveUnitOfMeasure(ctx, &unit); err != nil {
		return UnitOfMeasureResponse{}, unitOfMeasureWriteErrors.Resolve(err)
	}
	return newUnitOfMeasureResponse(unit), nil
}

func (s *service) DeleteUnitOfMeasure(ctx context.Context, id uuid.UUID) error {
	return unitOfMeasureDeleteErrors.Resolve(s.repository.DeleteUnitOfMeasure(ctx, id))
}

func (s *service) ListAccounts(ctx context.Context, query AccountQuery) (pagination.Page[AccountResponse], error) {
	accounts, total, err := s.repository.ListAccounts(ctx, AccountFilter{
		Search:   query.Search,
		Type:     query.Type,
		ParentID: query.ParentID,
		Offset:   query.Offset(),
		Limit:    query.Size(),
	})
	if err != nil {
		return pagination.Page[AccountResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(accounts, newAccountResponse), query.Query, total), nil
}

func (s *service) GetAccount(ctx context.Context, id uuid.UUID) (AccountResponse, error) {
	account, err := s.repository.FindAccount(ctx, id)
	if err != nil {
		return AccountResponse{}, accountReadErrors.Resolve(err)
	}
	return newAccountResponse(account), nil
}

func (s *service) CreateAccount(ctx context.Context, request AccountRequest) (AccountResponse, error) {
	if err := s.validateAccountParent(ctx, uuid.Nil, request); err != nil {
		return AccountResponse{}, err
	}

	var account Account
	applyAccountRequest(&account, request)
	if err := s.repository.CreateAccount(ctx, &account); err != nil {
		return AccountResponse{}, accountWriteErrors.Resolve(err)
	}
	return newAccountResponse(account), nil
}

func (s *service) UpdateAccount(ctx context.Context, id uuid.UUID, request AccountRequest) (AccountResponse, error) {
	account, err := s.repository.FindAccount(ctx, id)
	if err != nil {
		return AccountResponse{}, accountReadErrors.Resolve(err)
	}
	if err := s.ensureAccountTypeChangeable(ctx, account, request.Type); err != nil {
		return AccountResponse{}, err
	}
	if err := s.validateAccountParent(ctx, account.ID, request); err != nil {
		return AccountResponse{}, err
	}

	applyAccountRequest(&account, request)
	if err := s.repository.SaveAccount(ctx, &account); err != nil {
		return AccountResponse{}, accountWriteErrors.Resolve(err)
	}
	return newAccountResponse(account), nil
}

func (s *service) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	return accountDeleteErrors.Resolve(s.repository.DeleteAccount(ctx, id))
}

func (s *service) validateAccountParent(ctx context.Context, accountID uuid.UUID, request AccountRequest) error {
	if request.ParentID == nil {
		return nil
	}
	if *request.ParentID == accountID {
		return errAccountParentCycle
	}

	parent, err := s.repository.FindAccount(ctx, *request.ParentID)
	if err != nil {
		return accountParentErrors.Resolve(err)
	}
	if parent.Type != request.Type {
		return errAccountParentTypeMismatch
	}
	if accountID == uuid.Nil {
		return nil
	}

	descends, err := s.repository.AccountDescendsFrom(ctx, parent.ID, accountID)
	if err != nil {
		return apperror.Internal(err)
	}
	if descends {
		return errAccountParentCycle
	}
	return nil
}

func (s *service) ensureAccountTypeChangeable(ctx context.Context, account Account, nextType string) error {
	if account.Type == nextType {
		return nil
	}

	children, err := s.repository.CountAccountChildren(ctx, account.ID)
	if err != nil {
		return apperror.Internal(err)
	}
	if children > 0 {
		return errAccountTypeLocked
	}
	return nil
}

func applyRegionRequest(region *Region, request RegionRequest) {
	region.Code = strings.TrimSpace(request.Code)
	region.Name = strings.TrimSpace(request.Name)
	region.Type = request.Type
	region.ParentID = request.ParentID
}

func applyAccountRequest(account *Account, request AccountRequest) {
	account.Code = strings.TrimSpace(request.Code)
	account.Name = strings.TrimSpace(request.Name)
	account.Type = request.Type
	account.ParentID = request.ParentID
}
