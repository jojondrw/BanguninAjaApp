package main

import (
	"context"
	"database/sql/driver"
	"fmt"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/project"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/sales"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
)

type fakeStore struct {
	rows  map[string]uuid.UUID
	calls map[string]int
}

func newFakeStore() *fakeStore {
	store := &fakeStore{rows: map[string]uuid.UUID{}, calls: map[string]int{}}
	for _, seed := range demoProjects() {
		store.rows[keyOf("region", map[string]any{"code": seed.region})] = uuid.New()
		for _, item := range seed.items {
			store.rows[keyOf("unit_of_measure", map[string]any{"code": item.unit})] = uuid.New()
		}
	}
	return store
}

func keyOf(table string, keys map[string]any) string {
	parts := make([]string, 0, len(keys))
	for column, value := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", column, value))
	}
	sort.Strings(parts)
	return table + "|" + strings.Join(parts, "|")
}

func (f *fakeStore) find(_ context.Context, table string, keys map[string]any) (uuid.UUID, bool, error) {
	id, found := f.rows[keyOf(table, keys)]
	return id, found, nil
}

func (f *fakeStore) first(context.Context, string, string) (uuid.UUID, bool, error) {
	return uuid.Nil, false, nil
}

func (f *fakeStore) put(table string, keys map[string]any) uuid.UUID {
	id := uuid.New()
	f.rows[keyOf(table, keys)] = id
	f.calls[table]++
	return id
}

type fakeProjects struct {
	project.Service
	store *fakeStore
}

func (f fakeProjects) CreateProject(_ context.Context, _ uuid.UUID, _ string, request project.ProjectRequest) (project.ProjectResponse, error) {
	return project.ProjectResponse{ID: f.store.put("project", map[string]any{"code": request.Code})}, nil
}

func (f fakeProjects) CreatePhase(_ context.Context, projectID uuid.UUID, request project.PhaseRequest) (project.PhaseResponse, error) {
	return project.PhaseResponse{ID: f.store.put("project_phase", map[string]any{"project_id": projectID, "name": request.Name})}, nil
}

func (f fakeProjects) CreateBudgetItem(_ context.Context, projectID uuid.UUID, request project.BudgetItemRequest) (project.BudgetItemResponse, error) {
	return project.BudgetItemResponse{ID: f.store.put("budget_item", map[string]any{"project_id": projectID, "code": request.Code})}, nil
}

func (f fakeProjects) CreatePermit(_ context.Context, projectID uuid.UUID, request project.PermitRequest) (project.PermitResponse, error) {
	return project.PermitResponse{ID: f.store.put("permit", map[string]any{"project_id": projectID, "number": request.Number})}, nil
}

func (f fakeProjects) UpdateProjectStatus(context.Context, uuid.UUID, project.ProjectStatusRequest) (project.ProjectResponse, error) {
	f.store.calls["project status"]++
	return project.ProjectResponse{}, nil
}

type fakeSales struct {
	sales.Service
	store *fakeStore
}

func (f fakeSales) CreateCustomer(_ context.Context, request sales.CustomerRequest) (sales.CustomerResponse, error) {
	return sales.CustomerResponse{ID: f.store.put("customer", map[string]any{"identity_number": request.IdentityNumber})}, nil
}

func (f fakeSales) CreateLead(_ context.Context, request sales.LeadRequest) (sales.LeadResponse, error) {
	return sales.LeadResponse{ID: f.store.put("lead", map[string]any{"name": request.Name, "project_id": *request.ProjectID})}, nil
}

func (f fakeSales) ConvertLead(_ context.Context, _ uuid.UUID, request sales.CustomerRequest) (sales.LeadConversionResponse, error) {
	customerID := f.store.put("customer", map[string]any{"identity_number": request.IdentityNumber})
	return sales.LeadConversionResponse{Customer: sales.CustomerResponse{ID: customerID}}, nil
}

func (f fakeSales) CreateUnit(_ context.Context, request sales.UnitRequest) (sales.UnitResponse, error) {
	return sales.UnitResponse{ID: f.store.put("unit", map[string]any{"code": request.Code})}, nil
}

func (f fakeSales) CreateContract(_ context.Context, request sales.ContractRequest) (sales.ContractResponse, error) {
	return sales.ContractResponse{ID: f.store.put("contract", map[string]any{"number": request.Number})}, nil
}

func (f fakeSales) CreateInstallment(context.Context, uuid.UUID, sales.InstallmentRequest) (sales.InstallmentResponse, error) {
	f.store.calls["installment"]++
	return sales.InstallmentResponse{ID: uuid.New()}, nil
}

func (f fakeSales) UpdateContractStatus(_ context.Context, _ uuid.UUID, request sales.ContractStatusRequest) (sales.ContractResponse, error) {
	f.store.calls["contract "+request.Status]++
	return sales.ContractResponse{}, nil
}

func (f fakeSales) PayInstallment(context.Context, uuid.UUID, uuid.UUID, sales.InstallmentPaymentRequest) (sales.InstallmentResponse, error) {
	f.store.calls["installment paid"]++
	return sales.InstallmentResponse{}, nil
}

func newTestSeeder(store *fakeStore) *seeder {
	return &seeder{
		ctx:      context.Background(),
		lookup:   store,
		services: services{projects: fakeProjects{store: store}, sales: fakeSales{store: store}},
		today:    time.Date(2026, time.October, 3, 0, 0, 0, 0, time.UTC),
		owner:    uuid.New(),
		ids:      map[string]uuid.UUID{},
		tally:    newTally(),
	}
}

func expectedSalesCalls() map[string]int {
	expected := map[string]int{
		"customer": len(demoCustomers()) + 1,
		"lead":     len(demoLeads()),
		"unit":     len(demoUnits()),
		"contract": len(demoContracts()),
	}
	for _, contract := range demoContracts() {
		expected["installment"] += len(contract.installments)
		for _, installment := range contract.installments {
			if installment.paid {
				expected["installment paid"]++
			}
		}
		for _, status := range contractStatusPath[contract.status] {
			expected["contract "+status]++
		}
	}
	return expected
}

func expectedProjectCalls() map[string]int {
	expected := map[string]int{"project": len(demoProjects())}
	for _, seed := range demoProjects() {
		expected["project_phase"] += len(seed.phases)
		expected["budget_item"] += len(seed.items)
		expected["permit"] += len(seed.permits)
		expected["project status"] += len(projectStatusPath[seed.status])
	}
	return expected
}

func TestSeedingProjectsAndSalesTwiceCreatesEverythingOnce(t *testing.T) {
	store := newFakeStore()
	for run := 1; run <= 2; run++ {
		s := newTestSeeder(store)
		if err := s.seedProjects(); err != nil {
			t.Fatalf("run %d projects: %v", run, err)
		}
		if err := s.seedSales(); err != nil {
			t.Fatalf("run %d sales: %v", run, err)
		}
		if run == 2 && len(s.tally.created) != 0 {
			t.Fatalf("second run created %v, want nothing", s.tally.created)
		}
	}

	for _, expected := range []map[string]int{expectedProjectCalls(), expectedSalesCalls()} {
		for table, want := range expected {
			if got := store.calls[table]; got != want {
				t.Errorf("%s calls = %d, want %d", table, got, want)
			}
		}
	}
}

func TestDemoContractsCoverEveryStatusWithOneOverdueInstallment(t *testing.T) {
	statuses := map[string]bool{}
	overdue := 0
	today := 0
	for _, contract := range demoContracts() {
		statuses[contract.status] = true
		for _, installment := range contract.installments {
			if !installment.paid && installment.due < today {
				overdue++
			}
		}
	}
	for _, status := range []string{"draft", contractActive, contractPaid, contractCancelled} {
		if !statuses[status] {
			t.Errorf("no demo contract with status %s", status)
		}
	}
	if overdue != 1 {
		t.Errorf("overdue installments = %d, want 1", overdue)
	}
}

func TestDemoCodesAreAllPrefixed(t *testing.T) {
	codes := []string{}
	for _, seed := range demoProjects() {
		codes = append(codes, seed.code)
	}
	for _, seed := range demoUnits() {
		codes = append(codes, seed.code)
	}
	for _, seed := range demoContracts() {
		codes = append(codes, seed.number)
	}
	for _, seed := range demoMaterials() {
		codes = append(codes, seed.code)
	}
	for _, seed := range demoEmployees() {
		codes = append(codes, seed.identity)
	}
	for _, code := range codes {
		if !strings.HasPrefix(code, demoPrefix) {
			t.Errorf("code %s is not prefixed with %s", code, demoPrefix)
		}
	}
}

func TestResetDeletesChildrenBeforeParents(t *testing.T) {
	order := []string{}
	for _, step := range demoPurges(time.Date(2026, time.October, 3, 0, 0, 0, 0, time.UTC)) {
		order = append(order, step.table)
	}
	before := [][2]string{
		{"comparison", "saved_location"},
		{"receivable", "invoice"},
		{"payable", "purchase_order"},
		{"goods_receipt", "stock_movement"},
		{"stock_movement", "material"},
		{"purchase_order", "purchase_request"},
		{"purchase_order", "vendor"},
		{"payable", "vendor"},
		{"construction_contract", "vendor"},
		{"contract", "unit"},
		{"contract", "customer"},
		{"lead", "customer"},
		{"unit", "project"},
		{"purchase_request", "project"},
	}
	for _, pair := range before {
		if slices.Index(order, pair[0]) > slices.Index(order, pair[1]) {
			t.Errorf("%s must be purged before %s", pair[0], pair[1])
		}
	}
	if order[len(order)-1] != "project" {
		t.Errorf("project must be purged last, got %s", order[len(order)-1])
	}
}

func TestResetCoversEveryCashRowWithoutProject(t *testing.T) {
	s := newTestSeeder(newFakeStore())
	purgeable := companyCashNotes(s.today.AddDate(0, 2, 0))
	for _, seed := range s.cashSchedule() {
		if seed.project == "" && !slices.Contains(purgeable, seed.note) {
			t.Errorf("cash note %q has no project and would survive --reset-demo", seed.note)
		}
	}
}

func TestGormLookupFindsRowByNaturalKey(t *testing.T) {
	id := uuid.New()
	db, recorder := databasetest.Open(t, databasetest.Result{Columns: []string{"id"}, Rows: [][]driver.Value{{id.String()}}})

	found, ok, err := gormLookup{db: db}.find(context.Background(), "project", map[string]any{"code": projectGriyaAsri})
	if err != nil || !ok || found != id {
		t.Fatalf("find = %v, %v, %v; want %v", found, ok, err, id)
	}
	queries := recorder.Queries()
	if len(queries) != 1 || !strings.Contains(queries[0].SQL, `FROM "project" WHERE "project"."code" = $1`) {
		t.Fatalf("unexpected query %+v", queries)
	}
}

func TestMonthDayNeverPointsToTheFuture(t *testing.T) {
	s := newTestSeeder(newFakeStore())
	if got := s.monthDay(0, 25); !got.Equal(s.today) {
		t.Errorf("monthDay(0, 25) = %s, want today %s", got, s.today)
	}
	if got := s.monthDay(2, 5); got != time.Date(2026, time.August, 5, 0, 0, 0, 0, time.UTC) {
		t.Errorf("monthDay(2, 5) = %s", got)
	}
	if label := s.monthLabel(4); label != "Juni 2026" {
		t.Errorf("monthLabel(4) = %s", label)
	}
}

func TestDemoSitePhotosAreEmbeddedAndLabelled(t *testing.T) {
	covered := map[string]bool{}
	for _, seed := range demoSitePhotos() {
		content, err := demoPhotoFiles.ReadFile("photos/" + seed.file)
		if err != nil || len(content) == 0 {
			t.Fatalf("photo %s is not embedded: %v", seed.file, err)
		}
		if len(content) > project.SitePhotoMaxBytes {
			t.Errorf("photo %s is %d bytes, over the upload limit", seed.file, len(content))
		}
		if seed.taken >= 0 {
			t.Errorf("photo %s must be dated in the past", seed.file)
		}
		covered[seed.project] = true
	}
	for _, seed := range demoProjects() {
		if !covered[seed.code] {
			t.Errorf("project %s has no demo site photo", seed.code)
		}
	}
	if !strings.HasPrefix(demoPhotoNote, "Ilustrasi") {
		t.Errorf("demo photos must be labelled as illustrations")
	}
}
