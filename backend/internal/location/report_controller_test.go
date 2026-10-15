package location

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
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

var (
	savedLocationColumns  = []string{"id", "user_id", "name", "score", "land_price_per_sqm", "flood_index", "earthquake_index", "risk_flags", "saved_at"}
	dimensionScoreColumns = []string{"saved_location_id", "dimension_id", "value"}
	dimensionColumns      = []string{"id", "code", "name", "description", "sort_order"}
	savedAtFixture        = time.Date(2026, 10, 5, 4, 27, 59, 0, time.UTC)
)

func serveLocationReport(t *testing.T, userID uuid.UUID, path string, results ...databasetest.Result) (*httptest.ResponseRecorder, *databasetest.Recorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, recorder := databasetest.Open(t, results...)
	tokens := token.NewManager(config.Token{AccessSecret: "location-report-test", AccessTTL: time.Minute, Issuer: "location-report-test"})
	access, err := tokens.IssueAccess(userID)
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

func savedLocationRow(id, userID uuid.UUID, name string, score int64, flags any) []driver.Value {
	return []driver.Value{id.String(), userID.String(), name, score, int64(4500000), 0.12, 0.3, flags, savedAtFixture}
}

func assertErrorCode(t *testing.T, response *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if response.Code != status {
		t.Fatalf("got status %d, want %d: %s", response.Code, status, response.Body.String())
	}
	var envelope struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil || envelope.Error.Code != code {
		t.Fatalf("got body %s, want error code %s", response.Body.String(), code)
	}
}

func assertOwnedBy(t *testing.T, query databasetest.Query, userID uuid.UUID) {
	t.Helper()
	for _, argument := range query.Args {
		if fmt.Sprint(argument) == userID.String() && strings.Contains(query.SQL, "user_id = $") {
			return
		}
	}
	t.Fatalf("query must be limited to the signed-in user:\n%s %v", query.SQL, query.Args)
}

func assertPDFAttachment(t *testing.T, response *httptest.ResponseRecorder, fileName *regexp.Regexp) {
	t.Helper()
	if response.Code != http.StatusOK {
		t.Fatalf("got status %d: %s", response.Code, response.Body.String())
	}
	if got := response.Header().Get("Content-Type"); got != "application/pdf" {
		t.Fatalf("got content type %q", got)
	}
	if got := response.Header().Get("Content-Disposition"); !fileName.MatchString(got) {
		t.Fatalf("got disposition %q, want %s", got, fileName)
	}
	if !strings.Contains(response.Header().Get("Access-Control-Expose-Headers"), "Content-Disposition") {
		t.Fatalf("the file name header must be readable from the browser")
	}
	if !bytes.HasPrefix(response.Body.Bytes(), []byte("%PDF-")) {
		t.Fatalf("body is not a PDF")
	}
}

func TestSiteReportIsNotFoundForUnknownLocation(t *testing.T) {
	userID, locationID := uuid.New(), uuid.New()
	response, recorder := serveLocationReport(t, userID, "/api/locations/saved/"+locationID.String()+"/report",
		databasetest.Result{Columns: savedLocationColumns})

	assertErrorCode(t, response, http.StatusNotFound, "saved_location_not_found")
	queries := recorder.Queries()
	if len(queries) != 1 {
		t.Fatalf("got %d queries, want only the owned lookup", len(queries))
	}
	assertOwnedBy(t, queries[0], userID)
}

func TestSiteReportStreamsAPDFOfTheStoredEvaluation(t *testing.T) {
	userID, locationID, dimensionID := uuid.New(), uuid.New(), uuid.New()
	response, recorder := serveLocationReport(t, userID, "/api/locations/saved/"+locationID.String()+"/report",
		databasetest.Result{Columns: savedLocationColumns, Rows: [][]driver.Value{savedLocationRow(locationID, userID, "UJI Lahan Cibubur", 71, "[]")}},
		databasetest.Result{Columns: dimensionScoreColumns, Rows: [][]driver.Value{{locationID.String(), dimensionID.String(), int64(83)}}},
		databasetest.Result{Columns: dimensionColumns, Rows: [][]driver.Value{{dimensionID.String(), "demografi_sosial", "Demografi & Sosial", "Jumlah penduduk", int64(3)}}},
		databasetest.Result{Columns: []string{"id"}},
	)

	assertPDFAttachment(t, response, regexp.MustCompile(`^attachment; filename=analisis-lokasi-uji-lahan-cibubur-\d{4}-\d{2}-\d{2}\.pdf$`))
	body := response.Body.Bytes()
	for _, fragment := range []string{"(UJI Lahan Cibubur)", "(71 dari 100)", "(Tidak ada penanda risiko.)", "(Rp 4.500.000)", "(83)"} {
		if !bytes.Contains(body, []byte(fragment)) {
			t.Errorf("pdf is missing %s", fragment)
		}
	}
	queries := recorder.Queries()
	if len(queries) != 4 {
		t.Fatalf("got %d queries, want location, scores, dimensions and regulation", len(queries))
	}
	assertOwnedBy(t, queries[0], userID)
	if !strings.Contains(queries[3].SQL, `"regulation"`) {
		t.Fatalf("last query must look up the stored regulation:\n%s", queries[3].SQL)
	}
}

func TestComparisonReportIsNotFoundForUnknownComparison(t *testing.T) {
	userID, comparisonID := uuid.New(), uuid.New()
	response, recorder := serveLocationReport(t, userID, "/api/locations/comparisons/"+comparisonID.String()+"/report",
		databasetest.Result{Columns: []string{"id", "user_id", "name"}})

	assertErrorCode(t, response, http.StatusNotFound, "comparison_not_found")
	assertOwnedBy(t, recorder.Queries()[0], userID)
}

func TestComparisonReportStreamsAPDFWithOneColumnPerLocation(t *testing.T) {
	userID, comparisonID, firstID, secondID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	response, recorder := serveLocationReport(t, userID, "/api/locations/comparisons/"+comparisonID.String()+"/report",
		databasetest.Result{Columns: []string{"id", "user_id", "name"}, Rows: [][]driver.Value{{comparisonID.String(), userID.String(), "UJI Bandingkan Depok"}}},
		databasetest.Result{Columns: []string{"comparison_id", "saved_location_id", "sort_order"}, Rows: [][]driver.Value{
			{comparisonID.String(), secondID.String(), int64(1)},
			{comparisonID.String(), firstID.String(), int64(2)},
		}},
		databasetest.Result{Columns: savedLocationColumns, Rows: [][]driver.Value{
			savedLocationRow(firstID, userID, "UJI Lahan Sawangan", 58, nil),
			savedLocationRow(secondID, userID, "UJI Lahan Cinere", 74, `[{"code":"rawan_banjir","severity":"medium","message":"Sedang"}]`),
		}},
		databasetest.Result{Columns: dimensionScoreColumns},
		databasetest.Result{Columns: dimensionColumns},
	)

	assertPDFAttachment(t, response, regexp.MustCompile(`^attachment; filename=perbandingan-lokasi-uji-bandingkan-depok-\d{4}-\d{2}-\d{2}\.pdf$`))
	body := response.Body.Bytes()
	for _, fragment := range []string{"(UJI Lahan Cinere dengan skor 74 dari 100)", "(UJI Lahan Sawangan)", "(Tidak tercatat)"} {
		if !bytes.Contains(body, []byte(fragment)) {
			t.Errorf("pdf is missing %s", fragment)
		}
	}
	queries := recorder.Queries()
	if len(queries) != 5 {
		t.Fatalf("got %d queries", len(queries))
	}
	assertOwnedBy(t, queries[0], userID)
	assertOwnedBy(t, queries[2], userID)
}
