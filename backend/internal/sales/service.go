package sales

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/duedate"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

const (
	unitAvailable = "available"
	unitReserved  = "reserved"
	unitSold      = "sold"

	contractDraft     = "draft"
	contractActive    = "active"
	contractPaid      = "paid"
	contractCancelled = "cancelled"
)

var contractTransitions = map[string][]string{
	contractDraft:  {contractActive, contractCancelled},
	contractActive: {contractPaid, contractCancelled},
}

var unitStatusAfterContract = map[string]string{
	contractPaid:      unitSold,
	contractCancelled: unitAvailable,
}

var (
	errCustomerNotFound     = apperror.NotFound("customer_not_found", "Pelanggan tidak ditemukan")
	errCustomerIdentityUsed = apperror.Conflict("customer_identity_number_used", "Nomor identitas pelanggan sudah dipakai")
	errCustomerInUse        = apperror.Conflict("customer_in_use", "Pelanggan masih dipakai oleh kontrak atau piutang")

	errUnitNotFound     = apperror.NotFound("unit_not_found", "Unit tidak ditemukan")
	errUnitCodeUsed     = apperror.Conflict("unit_code_used", "Kode unit sudah dipakai")
	errUnitInUse        = apperror.Conflict("unit_in_use", "Unit masih dipakai oleh kontrak")
	errUnitLocked       = apperror.Conflict("unit_locked", "Unit sedang terikat kontrak sehingga tidak bisa diubah manual")
	errUnitNotAvailable = apperror.Conflict("unit_not_available", "Unit sudah dipesan atau terjual")
	errUnitUnknown      = apperror.Unprocessable("unit_not_found", "Unit tidak ditemukan")
	errProjectNotFound  = apperror.Unprocessable("project_not_found", "Proyek tidak ditemukan")

	errLeadNotFound = apperror.NotFound("lead_not_found", "Prospek tidak ditemukan")

	errContractNotFound   = apperror.NotFound("contract_not_found", "Kontrak tidak ditemukan")
	errContractNumberUsed = apperror.Conflict("contract_number_used", "Nomor kontrak sudah dipakai")
	errContractCustomer   = apperror.Unprocessable("customer_not_found", "Pelanggan tidak ditemukan")
	errContractLocked     = apperror.Conflict("contract_locked", "Kontrak hanya bisa diubah atau dihapus selama masih draf")
	errContractUnitChange = apperror.Unprocessable("contract_unit_change_invalid", "Unit kontrak tidak bisa diganti. Hapus draf lalu buat kontrak baru")
	errContractTransition = apperror.Unprocessable("contract_status_transition_invalid", "Perubahan status kontrak tidak diizinkan")
	errContractUnpaid     = apperror.Unprocessable("contract_unpaid", "Kontrak belum bisa lunas karena masih ada cicilan yang belum dibayar")
	errContractClosed     = apperror.Conflict("contract_closed", "Kontrak sudah lunas atau dibatalkan sehingga cicilannya tidak bisa diubah")
	errContractNotActive  = apperror.Conflict("contract_not_active", "Cicilan hanya bisa dibayar pada kontrak yang aktif")

	errInstallmentNotFound   = apperror.NotFound("installment_not_found", "Cicilan tidak ditemukan")
	errInstallmentNumberUsed = apperror.Conflict("installment_number_used", "Nomor angsuran sudah dipakai di kontrak ini")
	errInstallmentPaid       = apperror.Conflict("installment_paid", "Cicilan yang sudah lunas tidak bisa diubah")
	errInstallmentExceeds    = apperror.Unprocessable("installment_exceeds_contract", "Total cicilan tidak boleh melebihi nilai kontrak")
)

var (
	customerReadErrors   = database.ErrorMap{NotFound: errCustomerNotFound}
	customerWriteErrors  = database.ErrorMap{NotFound: errCustomerNotFound, Duplicate: errCustomerIdentityUsed}
	customerDeleteErrors = database.ErrorMap{NotFound: errCustomerNotFound, Referenced: errCustomerInUse}

	unitReadErrors   = database.ErrorMap{NotFound: errUnitNotFound}
	unitWriteErrors  = database.ErrorMap{NotFound: errUnitNotFound, Duplicate: errUnitCodeUsed, Referenced: errProjectNotFound}
	unitDeleteErrors = database.ErrorMap{NotFound: errUnitNotFound, Referenced: errUnitInUse}
	contractUnitErrs = database.ErrorMap{NotFound: errUnitUnknown}

	leadReadErrors  = database.ErrorMap{NotFound: errLeadNotFound}
	leadWriteErrors = database.ErrorMap{NotFound: errLeadNotFound, Referenced: errProjectNotFound}

	contractReadErrors  = database.ErrorMap{NotFound: errContractNotFound}
	contractWriteErrors = database.ErrorMap{NotFound: errContractNotFound, Duplicate: errContractNumberUsed, Referenced: errContractCustomer}

	installmentReadErrors  = database.ErrorMap{NotFound: errInstallmentNotFound}
	installmentWriteErrors = database.ErrorMap{NotFound: errInstallmentNotFound, Duplicate: errInstallmentNumberUsed}
)

type Service interface {
	ListCustomers(ctx context.Context, query CustomerQuery) (pagination.Page[CustomerResponse], error)
	GetCustomer(ctx context.Context, id uuid.UUID) (CustomerResponse, error)
	CreateCustomer(ctx context.Context, request CustomerRequest) (CustomerResponse, error)
	UpdateCustomer(ctx context.Context, id uuid.UUID, request CustomerRequest) (CustomerResponse, error)
	DeleteCustomer(ctx context.Context, id uuid.UUID) error

	ListUnits(ctx context.Context, query UnitQuery) (pagination.Page[UnitResponse], error)
	SummarizeUnits(ctx context.Context, query UnitSummaryQuery) ([]UnitStatusCountResponse, error)
	GetUnit(ctx context.Context, id uuid.UUID) (UnitResponse, error)
	CreateUnit(ctx context.Context, request UnitRequest) (UnitResponse, error)
	UpdateUnit(ctx context.Context, id uuid.UUID, request UnitRequest) (UnitResponse, error)
	DeleteUnit(ctx context.Context, id uuid.UUID) error

	ListLeads(ctx context.Context, query LeadQuery) (pagination.Page[LeadResponse], error)
	GetLead(ctx context.Context, id uuid.UUID) (LeadResponse, error)
	CreateLead(ctx context.Context, request LeadRequest) (LeadResponse, error)
	UpdateLead(ctx context.Context, id uuid.UUID, request LeadRequest) (LeadResponse, error)
	DeleteLead(ctx context.Context, id uuid.UUID) error

	ListContracts(ctx context.Context, query ContractQuery) (pagination.Page[ContractResponse], error)
	GetContract(ctx context.Context, id uuid.UUID) (ContractResponse, error)
	CreateContract(ctx context.Context, request ContractRequest) (ContractResponse, error)
	UpdateContract(ctx context.Context, id uuid.UUID, request ContractRequest) (ContractResponse, error)
	UpdateContractStatus(ctx context.Context, id uuid.UUID, request ContractStatusRequest) (ContractResponse, error)
	DeleteContract(ctx context.Context, id uuid.UUID) error

	ListInstallments(ctx context.Context, query InstallmentQuery) (pagination.Page[InstallmentResponse], error)
	CreateInstallment(ctx context.Context, contractID uuid.UUID, request InstallmentRequest) (InstallmentResponse, error)
	UpdateInstallment(ctx context.Context, contractID, installmentID uuid.UUID, request InstallmentRequest) (InstallmentResponse, error)
	DeleteInstallment(ctx context.Context, contractID, installmentID uuid.UUID) error
	PayInstallment(ctx context.Context, contractID, installmentID uuid.UUID, request InstallmentPaymentRequest) (InstallmentResponse, error)
}

type service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) Service {
	return &service{repository: repository, now: time.Now}
}

func (s *service) ListCustomers(ctx context.Context, query CustomerQuery) (pagination.Page[CustomerResponse], error) {
	customers, total, err := s.repository.ListCustomers(ctx, CustomerFilter{
		Search: query.Search,
		Offset: query.Offset(),
		Limit:  query.Size(),
	})
	if err != nil {
		return pagination.Page[CustomerResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(customers, newCustomerResponse), query.Query, total), nil
}

func (s *service) GetCustomer(ctx context.Context, id uuid.UUID) (CustomerResponse, error) {
	customer, err := s.repository.FindCustomer(ctx, id)
	if err != nil {
		return CustomerResponse{}, customerReadErrors.Resolve(err)
	}
	return newCustomerResponse(customer), nil
}

func (s *service) CreateCustomer(ctx context.Context, request CustomerRequest) (CustomerResponse, error) {
	var customer Customer
	applyCustomerRequest(&customer, request)
	if err := s.repository.CreateCustomer(ctx, &customer); err != nil {
		return CustomerResponse{}, customerWriteErrors.Resolve(err)
	}
	return newCustomerResponse(customer), nil
}

func (s *service) UpdateCustomer(ctx context.Context, id uuid.UUID, request CustomerRequest) (CustomerResponse, error) {
	customer, err := s.repository.FindCustomer(ctx, id)
	if err != nil {
		return CustomerResponse{}, customerReadErrors.Resolve(err)
	}

	applyCustomerRequest(&customer, request)
	if err := s.repository.SaveCustomer(ctx, &customer); err != nil {
		return CustomerResponse{}, customerWriteErrors.Resolve(err)
	}
	return newCustomerResponse(customer), nil
}

func (s *service) DeleteCustomer(ctx context.Context, id uuid.UUID) error {
	return customerDeleteErrors.Resolve(s.repository.DeleteCustomer(ctx, id))
}

func (s *service) ListUnits(ctx context.Context, query UnitQuery) (pagination.Page[UnitResponse], error) {
	units, total, err := s.repository.ListUnits(ctx, UnitFilter{
		Search:    query.Search,
		ProjectID: query.ProjectID,
		Status:    query.Status,
		Offset:    query.Offset(),
		Limit:     query.Size(),
	})
	if err != nil {
		return pagination.Page[UnitResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(units, newUnitResponse), query.Query, total), nil
}

func (s *service) SummarizeUnits(ctx context.Context, query UnitSummaryQuery) ([]UnitStatusCountResponse, error) {
	counts, err := s.repository.CountUnitsByStatus(ctx, query.ProjectID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return pagination.Map(counts, newUnitStatusCountResponse), nil
}

func (s *service) GetUnit(ctx context.Context, id uuid.UUID) (UnitResponse, error) {
	unit, err := s.repository.FindUnit(ctx, id)
	if err != nil {
		return UnitResponse{}, unitReadErrors.Resolve(err)
	}
	return newUnitResponse(unit), nil
}

func (s *service) CreateUnit(ctx context.Context, request UnitRequest) (UnitResponse, error) {
	var unit Unit
	applyUnitRequest(&unit, request)
	if err := s.repository.CreateUnit(ctx, &unit); err != nil {
		return UnitResponse{}, unitWriteErrors.Resolve(err)
	}
	return newUnitResponse(unit), nil
}

func (s *service) UpdateUnit(ctx context.Context, id uuid.UUID, request UnitRequest) (UnitResponse, error) {
	unit, err := s.repository.FindUnit(ctx, id)
	if err != nil {
		return UnitResponse{}, unitReadErrors.Resolve(err)
	}
	if unitBoundToContract(unit) {
		return UnitResponse{}, errUnitLocked
	}

	applyUnitRequest(&unit, request)
	if err := s.repository.SaveUnit(ctx, &unit); err != nil {
		return UnitResponse{}, unitWriteErrors.Resolve(err)
	}
	return newUnitResponse(unit), nil
}

func (s *service) DeleteUnit(ctx context.Context, id uuid.UUID) error {
	return unitDeleteErrors.Resolve(s.repository.DeleteUnit(ctx, id))
}

func (s *service) ListLeads(ctx context.Context, query LeadQuery) (pagination.Page[LeadResponse], error) {
	leads, total, err := s.repository.ListLeads(ctx, LeadFilter{
		Search:    query.Search,
		ProjectID: query.ProjectID,
		Stage:     query.Stage,
		Offset:    query.Offset(),
		Limit:     query.Size(),
	})
	if err != nil {
		return pagination.Page[LeadResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(leads, newLeadResponse), query.Query, total), nil
}

func (s *service) GetLead(ctx context.Context, id uuid.UUID) (LeadResponse, error) {
	lead, err := s.repository.FindLead(ctx, id)
	if err != nil {
		return LeadResponse{}, leadReadErrors.Resolve(err)
	}
	return newLeadResponse(lead), nil
}

func (s *service) CreateLead(ctx context.Context, request LeadRequest) (LeadResponse, error) {
	var lead Lead
	applyLeadRequest(&lead, request)
	if err := s.repository.CreateLead(ctx, &lead); err != nil {
		return LeadResponse{}, leadWriteErrors.Resolve(err)
	}
	return newLeadResponse(lead), nil
}

func (s *service) UpdateLead(ctx context.Context, id uuid.UUID, request LeadRequest) (LeadResponse, error) {
	lead, err := s.repository.FindLead(ctx, id)
	if err != nil {
		return LeadResponse{}, leadReadErrors.Resolve(err)
	}

	applyLeadRequest(&lead, request)
	if err := s.repository.SaveLead(ctx, &lead); err != nil {
		return LeadResponse{}, leadWriteErrors.Resolve(err)
	}
	return newLeadResponse(lead), nil
}

func (s *service) DeleteLead(ctx context.Context, id uuid.UUID) error {
	return leadReadErrors.Resolve(s.repository.DeleteLead(ctx, id))
}

func (s *service) ListContracts(ctx context.Context, query ContractQuery) (pagination.Page[ContractResponse], error) {
	rows, total, err := s.repository.ListContracts(ctx, ContractFilter{
		Search:     query.Search,
		CustomerID: query.CustomerID,
		UnitID:     query.UnitID,
		Status:     query.Status,
		Type:       query.Type,
		Offset:     query.Offset(),
		Limit:      query.Size(),
	})
	if err != nil {
		return pagination.Page[ContractResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(rows, newContractResponse), query.Query, total), nil
}

func (s *service) GetContract(ctx context.Context, id uuid.UUID) (ContractResponse, error) {
	row, err := s.repository.FindContractRow(ctx, id)
	if err != nil {
		return ContractResponse{}, contractReadErrors.Resolve(err)
	}
	return newContractResponse(row), nil
}

func (s *service) CreateContract(ctx context.Context, request ContractRequest) (ContractResponse, error) {
	contract := Contract{Status: contractDraft}
	applyContractRequest(&contract, request)
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		return createContract(ctx, repository, &contract)
	})
	if err != nil {
		return ContractResponse{}, apperror.From(err)
	}
	return s.GetContract(ctx, contract.ID)
}

func (s *service) UpdateContract(ctx context.Context, id uuid.UUID, request ContractRequest) (ContractResponse, error) {
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		return updateContract(ctx, repository, id, request)
	})
	if err != nil {
		return ContractResponse{}, apperror.From(err)
	}
	return s.GetContract(ctx, id)
}

func (s *service) UpdateContractStatus(ctx context.Context, id uuid.UUID, request ContractStatusRequest) (ContractResponse, error) {
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		return changeContractStatus(ctx, repository, id, request.Status)
	})
	if err != nil {
		return ContractResponse{}, apperror.From(err)
	}
	return s.GetContract(ctx, id)
}

func (s *service) DeleteContract(ctx context.Context, id uuid.UUID) error {
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		return deleteContract(ctx, repository, id)
	})
	if err != nil {
		return apperror.From(err)
	}
	return nil
}

func (s *service) ListInstallments(ctx context.Context, query InstallmentQuery) (pagination.Page[InstallmentResponse], error) {
	today := s.today()
	window := duedate.RangeOf(query.Status, today)
	rows, total, err := s.repository.ListInstallments(ctx, InstallmentFilter{
		ContractID: query.ContractID,
		CustomerID: query.CustomerID,
		Settled:    window.Settled,
		DueFrom:    window.From,
		DueBefore:  window.Before,
		Offset:     query.Offset(),
		Limit:      query.Size(),
	})
	if err != nil {
		return pagination.Page[InstallmentResponse]{}, apperror.Internal(err)
	}
	responses := pagination.Map(rows, func(row InstallmentRow) InstallmentResponse {
		return newInstallmentResponse(row, today)
	})
	return pagination.New(responses, query.Query, total), nil
}

func (s *service) CreateInstallment(ctx context.Context, contractID uuid.UUID, request InstallmentRequest) (InstallmentResponse, error) {
	today := s.today()
	installment := Installment{ContractID: contractID}
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		return createInstallment(ctx, repository, &installment, request, today)
	})
	if err != nil {
		return InstallmentResponse{}, apperror.From(err)
	}
	return s.installmentResponse(ctx, installment.ID, today)
}

func (s *service) UpdateInstallment(ctx context.Context, contractID, installmentID uuid.UUID, request InstallmentRequest) (InstallmentResponse, error) {
	today := s.today()
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		return updateInstallment(ctx, repository, contractID, installmentID, request, today)
	})
	if err != nil {
		return InstallmentResponse{}, apperror.From(err)
	}
	return s.installmentResponse(ctx, installmentID, today)
}

func (s *service) DeleteInstallment(ctx context.Context, contractID, installmentID uuid.UUID) error {
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		return deleteInstallment(ctx, repository, contractID, installmentID)
	})
	if err != nil {
		return apperror.From(err)
	}
	return nil
}

func (s *service) PayInstallment(ctx context.Context, contractID, installmentID uuid.UUID, request InstallmentPaymentRequest) (InstallmentResponse, error) {
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		return payInstallment(ctx, repository, contractID, installmentID, request.PaidDate)
	})
	if err != nil {
		return InstallmentResponse{}, apperror.From(err)
	}
	return s.installmentResponse(ctx, installmentID, s.today())
}

func (s *service) installmentResponse(ctx context.Context, id uuid.UUID, today time.Time) (InstallmentResponse, error) {
	row, err := s.repository.FindInstallmentRow(ctx, id)
	if err != nil {
		return InstallmentResponse{}, installmentReadErrors.Resolve(err)
	}
	return newInstallmentResponse(row, today), nil
}

func (s *service) today() time.Time {
	return duedate.Today(s.now())
}

func createContract(ctx context.Context, repository Repository, contract *Contract) error {
	unit, err := repository.LockUnit(ctx, contract.UnitID)
	if err != nil {
		return contractUnitErrs.Resolve(err)
	}
	if unit.Status != unitAvailable {
		return errUnitNotAvailable
	}
	if err := repository.CreateContract(ctx, contract); err != nil {
		return contractWriteErrors.Resolve(err)
	}

	unit.Status = unitReserved
	return unitWriteErrors.Resolve(repository.SaveUnit(ctx, &unit))
}

func updateContract(ctx context.Context, repository Repository, id uuid.UUID, request ContractRequest) error {
	contract, err := lockDraftContract(ctx, repository, id)
	if err != nil {
		return err
	}
	if request.UnitID != contract.UnitID {
		return errContractUnitChange
	}

	applyContractRequest(&contract, request)
	return contractWriteErrors.Resolve(repository.SaveContract(ctx, &contract))
}

func deleteContract(ctx context.Context, repository Repository, id uuid.UUID) error {
	contract, err := lockDraftContract(ctx, repository, id)
	if err != nil {
		return err
	}
	if err := repository.DeleteContract(ctx, id); err != nil {
		return contractReadErrors.Resolve(err)
	}
	return releaseUnit(ctx, repository, contract.UnitID, unitAvailable)
}

func changeContractStatus(ctx context.Context, repository Repository, id uuid.UUID, next string) error {
	contract, err := repository.LockContract(ctx, id)
	if err != nil {
		return contractReadErrors.Resolve(err)
	}
	if contract.Status == next {
		return nil
	}
	if !canTransitionContract(contract.Status, next) {
		return errContractTransition
	}
	if err := ensureSettledForPaid(ctx, repository, id, next); err != nil {
		return err
	}

	contract.Status = next
	if err := repository.SaveContract(ctx, &contract); err != nil {
		return contractWriteErrors.Resolve(err)
	}
	return syncUnitWithContract(ctx, repository, contract.UnitID, next)
}

func ensureSettledForPaid(ctx context.Context, repository Repository, contractID uuid.UUID, next string) error {
	if next != contractPaid {
		return nil
	}

	unpaid, err := repository.CountUnpaidInstallments(ctx, contractID)
	if err != nil {
		return apperror.Internal(err)
	}
	if unpaid > 0 {
		return errContractUnpaid
	}
	return nil
}

func syncUnitWithContract(ctx context.Context, repository Repository, unitID uuid.UUID, contractStatus string) error {
	unitStatus, changes := unitStatusAfterContract[contractStatus]
	if !changes {
		return nil
	}
	return releaseUnit(ctx, repository, unitID, unitStatus)
}

func releaseUnit(ctx context.Context, repository Repository, unitID uuid.UUID, status string) error {
	unit, err := repository.LockUnit(ctx, unitID)
	if err != nil {
		return contractUnitErrs.Resolve(err)
	}
	unit.Status = status
	return unitWriteErrors.Resolve(repository.SaveUnit(ctx, &unit))
}

func lockDraftContract(ctx context.Context, repository Repository, id uuid.UUID) (Contract, error) {
	contract, err := repository.LockContract(ctx, id)
	if err != nil {
		return Contract{}, contractReadErrors.Resolve(err)
	}
	if contract.Status != contractDraft {
		return Contract{}, errContractLocked
	}
	return contract, nil
}

func lockOpenContract(ctx context.Context, repository Repository, id uuid.UUID) (Contract, error) {
	contract, err := repository.LockContract(ctx, id)
	if err != nil {
		return Contract{}, contractReadErrors.Resolve(err)
	}
	if contract.Status == contractPaid || contract.Status == contractCancelled {
		return Contract{}, errContractClosed
	}
	return contract, nil
}

func createInstallment(ctx context.Context, repository Repository, installment *Installment, request InstallmentRequest, today time.Time) error {
	contract, err := lockOpenContract(ctx, repository, installment.ContractID)
	if err != nil {
		return err
	}
	if err := ensureWithinContract(ctx, repository, contract, request.Amount, nil); err != nil {
		return err
	}

	applyInstallmentRequest(installment, request, today)
	return installmentWriteErrors.Resolve(repository.CreateInstallment(ctx, installment))
}

func updateInstallment(ctx context.Context, repository Repository, contractID, id uuid.UUID, request InstallmentRequest, today time.Time) error {
	contract, err := lockOpenContract(ctx, repository, contractID)
	if err != nil {
		return err
	}
	installment, err := unpaidInstallment(ctx, repository, contractID, id)
	if err != nil {
		return err
	}
	if err := ensureWithinContract(ctx, repository, contract, request.Amount, &id); err != nil {
		return err
	}

	applyInstallmentRequest(&installment, request, today)
	return installmentWriteErrors.Resolve(repository.SaveInstallment(ctx, &installment))
}

func deleteInstallment(ctx context.Context, repository Repository, contractID, id uuid.UUID) error {
	if _, err := lockOpenContract(ctx, repository, contractID); err != nil {
		return err
	}
	if _, err := unpaidInstallment(ctx, repository, contractID, id); err != nil {
		return err
	}
	return installmentReadErrors.Resolve(repository.DeleteInstallment(ctx, contractID, id))
}

func payInstallment(ctx context.Context, repository Repository, contractID, id uuid.UUID, paidDate time.Time) error {
	contract, err := repository.LockContract(ctx, contractID)
	if err != nil {
		return contractReadErrors.Resolve(err)
	}
	if contract.Status != contractActive {
		return errContractNotActive
	}

	installment, err := unpaidInstallment(ctx, repository, contractID, id)
	if err != nil {
		return err
	}
	installment.PaidDate = &paidDate
	installment.Status = duedate.Paid
	return installmentWriteErrors.Resolve(repository.SaveInstallment(ctx, &installment))
}

func unpaidInstallment(ctx context.Context, repository Repository, contractID, id uuid.UUID) (Installment, error) {
	installment, err := repository.FindInstallment(ctx, contractID, id)
	if err != nil {
		return Installment{}, installmentReadErrors.Resolve(err)
	}
	if installment.Status == duedate.Paid {
		return Installment{}, errInstallmentPaid
	}
	return installment, nil
}

func ensureWithinContract(ctx context.Context, repository Repository, contract Contract, amount int64, excludeID *uuid.UUID) error {
	scheduled, err := repository.SumInstallments(ctx, contract.ID, excludeID)
	if err != nil {
		return apperror.Internal(err)
	}
	if scheduled+amount > contract.Value {
		return errInstallmentExceeds
	}
	return nil
}

func canTransitionContract(current, next string) bool {
	return slices.Contains(contractTransitions[current], next)
}

func unitBoundToContract(unit Unit) bool {
	return unit.Status == unitReserved || unit.Status == unitSold
}

func applyCustomerRequest(customer *Customer, request CustomerRequest) {
	customer.Name = strings.TrimSpace(request.Name)
	customer.Contact = strings.TrimSpace(request.Contact)
	customer.Email = strings.TrimSpace(request.Email)
	customer.IdentityNumber = strings.TrimSpace(request.IdentityNumber)
	customer.Address = strings.TrimSpace(request.Address)
}

func applyUnitRequest(unit *Unit, request UnitRequest) {
	unit.Code = strings.TrimSpace(request.Code)
	unit.ProjectID = request.ProjectID
	unit.UnitType = strings.TrimSpace(request.UnitType)
	unit.AreaSqm = request.AreaSqm
	unit.Price = request.Price
	unit.Status = request.Status
}

func applyLeadRequest(lead *Lead, request LeadRequest) {
	lead.Name = strings.TrimSpace(request.Name)
	lead.Contact = strings.TrimSpace(request.Contact)
	lead.ProjectID = request.ProjectID
	lead.Source = strings.TrimSpace(request.Source)
	lead.Stage = request.Stage
	lead.LastContactedAt = request.LastContactedAt
}

func applyContractRequest(contract *Contract, request ContractRequest) {
	contract.Number = strings.TrimSpace(request.Number)
	contract.CustomerID = request.CustomerID
	contract.UnitID = request.UnitID
	contract.Type = request.Type
	contract.Value = request.Value
	contract.Date = request.Date
}

func applyInstallmentRequest(installment *Installment, request InstallmentRequest, today time.Time) {
	installment.InstallmentNumber = request.InstallmentNumber
	installment.DueDate = request.DueDate
	installment.Amount = request.Amount
	installment.Status = duedate.Status(request.DueDate, false, today)
}
