package hr

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/middleware"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
)

func serveHR(t *testing.T, path string, results ...databasetest.Result) (*httptest.ResponseRecorder, *databasetest.Recorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, recorder := databasetest.Open(t, results...)
	tokens := token.NewManager(config.Token{AccessSecret: "hr-module-test", AccessTTL: time.Minute, Issuer: "hr-module-test"})
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

func TestPayrollSummaryRouteIsNotTakenForAPayrollID(t *testing.T) {
	response, recorder := serveHR(t, "/api/hr/payrolls/summary?period=2026-09", payrollTotalsResult)

	if response.Code != http.StatusOK {
		t.Fatalf("got status %d: %s", response.Code, response.Body.String())
	}
	var body struct {
		Data PayrollSummaryResponse `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	want := PayrollSummaryResponse{Count: 3, GrossPay: 21000000, NetPay: 20300000, PaidNetPay: 6850000, UnpaidNetPay: 13450000, UnpaidCount: 2}
	if body.Data != want {
		t.Fatalf("got %+v, want %+v", body.Data, want)
	}
	if queries := recorder.Queries(); len(queries) != 1 || len(queries[0].Args) != 1 || queries[0].Args[0] != "2026-09" {
		t.Fatalf("got queries %+v", queries)
	}
}

func TestPayrollSummaryRejectsMalformedFilters(t *testing.T) {
	for _, path := range []string{
		"/api/hr/payrolls/summary?period=2026-9",
		"/api/hr/payrolls/summary?projectId=bukan-uuid",
	} {
		response, recorder := serveHR(t, path)
		if response.Code != http.StatusBadRequest {
			t.Errorf("%s: got status %d, want 400", path, response.Code)
		}
		if len(recorder.Queries()) != 0 {
			t.Errorf("%s: a rejected filter must not reach the database", path)
		}
	}
}
