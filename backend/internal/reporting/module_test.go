package reporting

import (
	"archive/zip"
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/middleware"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
)

func serveReports(t *testing.T, path string, results ...databasetest.Result) (*httptest.ResponseRecorder, *databasetest.Recorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, recorder := databasetest.Open(t, results...)
	tokens := token.NewManager(config.Token{AccessSecret: "reporting-module-test", AccessTTL: time.Minute, Issuer: "reporting-module-test"})
	access, err := tokens.IssueAccess(uuid.New())
	if err != nil {
		t.Fatalf("issue access token: %v", err)
	}

	router := gin.New()
	router.Use(middleware.ErrorHandler())
	NewModule(db, tokens).RegisterRoutes(router.Group("/api"))

	request := httptest.NewRequest(http.MethodGet, path, nil)
	request.Header.Set("Authorization", "Bearer "+access.Value)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response, recorder
}

func emptyCashFlowResults() []databasetest.Result {
	return []databasetest.Result{
		{Columns: []string{"period", "cash_in", "cash_out"}},
		{Columns: []string{"balance"}, Rows: [][]driver.Value{{int64(0)}}},
		{Columns: []string{"date", "type", "amount", "note", "account_code", "account_name", "project_code", "project_name"}},
	}
}

func TestDownloadStreamsTheFileWithIndonesianFileName(t *testing.T) {
	cases := []struct {
		format      string
		contentType string
		signature   []byte
	}{
		{formatPDF, "application/pdf", []byte("%PDF-")},
		{formatXLSX, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", []byte("PK")},
		{formatCSV, "text/csv; charset=utf-8", []byte("\xEF\xBB\xBF")},
	}
	for _, tc := range cases {
		t.Run(tc.format, func(t *testing.T) {
			reportID := uuid.New()
			results := append([]databasetest.Result{{Columns: reportRowColumns, Rows: [][]driver.Value{reportRowValues(reportID, uuid.New(), tc.format)}}},
				emptyCashFlowResults()...)
			response, recorder := serveReports(t, "/api/reports/"+reportID.String()+"/download", results...)

			if response.Code != http.StatusOK {
				t.Fatalf("got status %d: %s", response.Code, response.Body.String())
			}
			if got := response.Header().Get("Content-Type"); got != tc.contentType {
				t.Fatalf("got content type %q", got)
			}
			want := "attachment; filename=laporan-arus-kas-2026-10." + tc.format
			if got := response.Header().Get("Content-Disposition"); got != want {
				t.Fatalf("got disposition %q, want %q", got, want)
			}
			if !strings.Contains(response.Header().Get("Access-Control-Expose-Headers"), "Content-Disposition") {
				t.Fatalf("the browser must be allowed to read the file name")
			}
			if !bytes.HasPrefix(response.Body.Bytes(), tc.signature) {
				t.Fatalf("body does not start with %q", tc.signature)
			}
			if queries := recorder.Queries(); len(queries) != 4 || !strings.Contains(queries[0].SQL, "report.id = $1") {
				t.Fatalf("got queries %+v", queries)
			}
		})
	}
}

func TestDownloadedXLSXOpensAsZip(t *testing.T) {
	reportID := uuid.New()
	results := append([]databasetest.Result{{Columns: reportRowColumns, Rows: [][]driver.Value{reportRowValues(reportID, uuid.New(), formatXLSX)}}},
		emptyCashFlowResults()...)
	response, _ := serveReports(t, "/api/reports/"+reportID.String()+"/download", results...)

	if _, err := zip.NewReader(bytes.NewReader(response.Body.Bytes()), int64(response.Body.Len())); err != nil {
		t.Fatalf("xlsx is not a zip: %v", err)
	}
}

func TestDownloadOfMissingReportIsNotFound(t *testing.T) {
	response, _ := serveReports(t, "/api/reports/"+uuid.NewString()+"/download", databasetest.Result{Columns: reportRowColumns})

	if response.Code != http.StatusNotFound || !strings.Contains(response.Body.String(), "report_not_found") {
		t.Fatalf("got %d %s", response.Code, response.Body.String())
	}
}

func TestPreviewRouteIsNotTakenForAReportID(t *testing.T) {
	response, _ := serveReports(t, "/api/reports/preview?type=arus_kas&periodStart=2026-10-01&periodEnd=2026-10-31", emptyCashFlowResults()...)

	if response.Code != http.StatusOK {
		t.Fatalf("got status %d: %s", response.Code, response.Body.String())
	}
	var body struct {
		Data PreviewResponse `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Data.Type != typeCashFlow || len(body.Data.Sections) != 2 || body.Data.Sections[0].TotalRows != 2 {
		t.Fatalf("got %+v", body.Data)
	}
}

func TestPreviewRejectsUnknownType(t *testing.T) {
	response, recorder := serveReports(t, "/api/reports/preview?type=neraca")

	if response.Code != http.StatusUnprocessableEntity || !strings.Contains(response.Body.String(), "report_type_invalid") {
		t.Fatalf("got %d %s", response.Code, response.Body.String())
	}
	if len(recorder.Queries()) != 0 {
		t.Fatalf("unknown types must not query the database")
	}
}
