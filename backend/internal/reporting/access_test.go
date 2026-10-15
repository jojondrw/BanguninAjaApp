package reporting

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/access"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
)

func reportRowOfType(id uuid.UUID, kind string) []driver.Value {
	values := reportRowValues(id, uuid.New(), formatPDF)
	values[2] = kind
	return values
}

func recordingService(t *testing.T, results ...databasetest.Result) (Service, *databasetest.Recorder) {
	t.Helper()
	db, recorder := databasetest.Open(t, results...)
	return &service{repository: NewRepository(db), now: func() time.Time { return fixedNow }}, recorder
}

func TestEveryReportKindReadsItsDataModule(t *testing.T) {
	want := map[string]access.Module{
		typeCashFlow: access.ModuleFinance,
		typeBudget:   access.ModuleFinance,
		typeProjects: access.ModuleProjects,
		typeAging:    access.ModuleBilling,
		typeStock:    access.ModuleInventory,
		typePayroll:  access.ModuleHR,
		typeSales:    access.ModuleSales,
	}
	if len(reportKinds) != len(want) {
		t.Fatalf("catalog has %d kinds, mapping covers %d", len(reportKinds), len(want))
	}
	for key, kind := range reportKinds {
		if kind.module != want[key] {
			t.Errorf("%s: got module %q, want %q", key, kind.module, want[key])
		}
	}
	if moduleOfType("payroll") != access.ModuleHR || moduleOfType("neraca") != access.ModuleUnmapped {
		t.Fatal("aliases must map like their canonical kind and unknown kinds to no module")
	}
}

func TestReadableKindsFollowTheMatrix(t *testing.T) {
	cases := map[string][]string{
		access.RoleAdmin:          {typeBudget, typeCashFlow, typePayroll, typeSales, typeAging, typeProjects, typeStock},
		access.RoleObserver:       {typeBudget, typeCashFlow, typePayroll, typeSales, typeAging, typeProjects, typeStock},
		access.RoleSales:          {typeSales, typeAging, typeProjects},
		access.RoleHR:             {typePayroll, typeProjects},
		access.RoleProcurement:    {typeProjects, typeStock},
		access.RoleFinance:        {typeBudget, typeCashFlow, typeSales, typeAging, typeProjects},
		access.RoleProjectManager: {typeBudget, typeCashFlow, typeProjects, typeStock},
		"":                        {},
	}
	for role, want := range cases {
		if got := readableKinds(role); !reflect.DeepEqual(got, want) {
			t.Errorf("%q: got %v, want %v", role, got, want)
		}
	}
}

func TestPreviewDeniesKindsOutsideTheRoleBeforeReadingData(t *testing.T) {
	cases := []struct {
		role    string
		kind    string
		message string
	}{
		{access.RoleSales, typePayroll, "Peran Anda tidak punya akses melihat modul SDM"},
		{access.RoleSales, typeCashFlow, "Peran Anda tidak punya akses melihat modul Keuangan"},
		{access.RoleSales, "payroll", "Peran Anda tidak punya akses melihat modul SDM"},
		{access.RoleHR, typeAging, "Peran Anda tidak punya akses melihat modul Tagihan"},
		{access.RoleProcurement, typeSales, "Peran Anda tidak punya akses melihat modul Penjualan"},
		{access.RoleProjectManager, typePayroll, "Peran Anda tidak punya akses melihat modul SDM"},
		{"", typeProjects, "Peran Anda tidak punya akses melihat modul Proyek"},
	}
	for _, tc := range cases {
		reports, recorder := recordingService(t)
		_, err := reports.PreviewReport(context.Background(), tc.role, PreviewQuery{Type: tc.kind})
		assertForbidden(t, err, tc.message)
		if len(recorder.Queries()) != 0 {
			t.Errorf("%s %s: denied preview must not query, got %d queries", tc.role, tc.kind, len(recorder.Queries()))
		}
	}
}

func TestPreviewStillRejectsUnknownKindsAsInvalid(t *testing.T) {
	reports, _ := recordingService(t)

	_, err := reports.PreviewReport(context.Background(), access.RoleSales, PreviewQuery{Type: "neraca"})

	if !errors.Is(err, errReportTypeInvalid) {
		t.Fatalf("got %v", err)
	}
}

func TestCreateDeniesKindsOutsideTheRole(t *testing.T) {
	repository := &fakeRepository{}

	_, err := NewService(repository).CreateReport(context.Background(), uuid.New(), access.RoleProjectManager,
		ReportRequest{Name: "UJI gaji", Type: typePayroll, Format: formatPDF})

	assertForbidden(t, err, "Peran Anda tidak punya akses melihat modul SDM")
	if repository.created != nil {
		t.Fatal("a denied report must not be recorded")
	}
}

func TestStoredReportOfUnreadableKindIsDenied(t *testing.T) {
	reportID := uuid.New()
	stored := databasetest.Result{Columns: reportRowColumns, Rows: [][]driver.Value{reportRowOfType(reportID, typePayroll)}}
	calls := map[string]func(Service) error{
		"get": func(reports Service) error {
			_, err := reports.GetReport(context.Background(), access.RoleSales, reportID)
			return err
		},
		"download": func(reports Service) error {
			_, err := reports.DownloadReport(context.Background(), access.RoleSales, reportID)
			return err
		},
		"delete": func(reports Service) error {
			return reports.DeleteReport(context.Background(), access.RoleSales, reportID)
		},
	}
	for name, call := range calls {
		reports, recorder := recordingService(t, stored)
		assertForbidden(t, call(reports), "Peran Anda tidak punya akses melihat modul SDM")
		if queries := recorder.Queries(); len(queries) != 1 {
			t.Errorf("%s: only the report lookup may run, got %d queries", name, len(queries))
		}
	}
}

func TestStoredReportOfUnknownKindIsAdminOnly(t *testing.T) {
	reportID := uuid.New()
	stored := databasetest.Result{Columns: reportRowColumns, Rows: [][]driver.Value{reportRowOfType(reportID, "neraca")}}

	reports, _ := recordingService(t, stored)
	assertForbidden(t, reports.DeleteReport(context.Background(), access.RoleObserver, reportID), "Peran Anda belum diberi akses ke fitur ini")

	reports, recorder := recordingService(t, stored, databasetest.Result{Affected: 1})
	if err := reports.DeleteReport(context.Background(), access.RoleAdmin, reportID); err != nil {
		t.Fatalf("admin delete got %v", err)
	}
	if len(recorder.Queries()) != 2 {
		t.Fatalf("got %d queries", len(recorder.Queries()))
	}
}

func TestDeleteOfReadableKindRemovesTheReport(t *testing.T) {
	reportID := uuid.New()
	reports, recorder := recordingService(t,
		databasetest.Result{Columns: reportRowColumns, Rows: [][]driver.Value{reportRowOfType(reportID, typeSales)}},
		databasetest.Result{Affected: 1},
	)

	if err := reports.DeleteReport(context.Background(), access.RoleSales, reportID); err != nil {
		t.Fatal(err)
	}
	queries := recorder.Queries()
	if len(queries) != 2 {
		t.Fatalf("got %d queries", len(queries))
	}
	assertSQLContains(t, queries[1].SQL, `DELETE FROM "report"`)
}

func TestListOnlyReturnsReadableKinds(t *testing.T) {
	reports, recorder := recordingService(t,
		databasetest.Result{Columns: []string{"count"}, Rows: [][]driver.Value{{int64(0)}}},
		databasetest.Result{Columns: reportRowColumns},
	)

	if _, err := reports.ListReports(context.Background(), access.RoleSales, ReportQuery{}); err != nil {
		t.Fatal(err)
	}
	queries := recorder.Queries()
	if len(queries) != 2 {
		t.Fatalf("got %d queries", len(queries))
	}
	for _, query := range queries {
		assertSQLContains(t, query.SQL, "report.type IN ($1,$2,$3)")
		if !reflect.DeepEqual(query.Args[:3], []any{typeSales, typeAging, typeProjects}) {
			t.Errorf("got args %v", query.Args)
		}
	}
}

func TestListWithoutReadableKindsIsEmptyWithoutQuerying(t *testing.T) {
	reports, recorder := recordingService(t)

	page, err := reports.ListReports(context.Background(), "", ReportQuery{})

	if err != nil || page.TotalItems != 0 || page.Items == nil || len(page.Items) != 0 {
		t.Fatalf("got %+v, %v", page, err)
	}
	if len(recorder.Queries()) != 0 {
		t.Fatalf("got %d queries", len(recorder.Queries()))
	}
}

func TestRoutesUseTheRoleFromAuthorization(t *testing.T) {
	cases := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/reports/preview?type=penggajian"},
		{http.MethodGet, "/api/reports/preview?type=arus_kas"},
	}
	for _, tc := range cases {
		response, recorder := serveReportsAs(t, access.RoleSales, tc.method, tc.path)
		if response.Code != http.StatusForbidden {
			t.Fatalf("%s: got %d %s", tc.path, response.Code, response.Body.String())
		}
		var body struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || body.Error.Code != "forbidden" {
			t.Fatalf("%s: got %s", tc.path, response.Body.String())
		}
		if len(recorder.Queries()) != 0 {
			t.Fatalf("%s: got %d queries", tc.path, len(recorder.Queries()))
		}
	}
}

func TestCreateRouteDeniesUnreadableKind(t *testing.T) {
	response, recorder := serveReportsAs(t, access.RoleFinance, http.MethodPost, "/api/reports")

	if response.Code != http.StatusForbidden || len(recorder.Queries()) != 0 {
		t.Fatalf("got %d %s after %d queries", response.Code, response.Body.String(), len(recorder.Queries()))
	}
}

func assertForbidden(t *testing.T, err error, message string) {
	t.Helper()
	var failure *apperror.Error
	if !errors.As(err, &failure) || failure.Status != http.StatusForbidden || failure.Code != "forbidden" || failure.Message != message {
		t.Errorf("got %v, want 403 forbidden %q", err, message)
	}
}
