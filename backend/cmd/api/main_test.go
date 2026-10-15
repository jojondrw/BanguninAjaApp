package main

import (
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/cookie"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
)

func subjectResult(role string, active bool) databasetest.Result {
	return databasetest.Result{Columns: []string{"active", "role_code"}, Rows: [][]driver.Value{{active, role}}}
}

func callRouter(t *testing.T, method, path string, results ...databasetest.Result) (*httptest.ResponseRecorder, *databasetest.Recorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, recorder := databasetest.Open(t, results...)
	tokens := token.NewManager(config.Token{AccessSecret: "router-test", AccessTTL: time.Minute, Issuer: "router-test"})
	router := buildRouter(config.Config{}, db, tokens, cookie.NewRefreshWriter(config.Cookie{}))

	issued, err := tokens.IssueAccess(uuid.New())
	if err != nil {
		t.Fatalf("issue access token: %v", err)
	}
	request := httptest.NewRequest(method, path, strings.NewReader("{}"))
	request.Header.Set("Authorization", "Bearer "+issued.Value)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response, recorder
}

func TestRouterBlocksModulesOutsideTheRoleBeforeTheHandlerRuns(t *testing.T) {
	cases := []struct {
		role   string
		method string
		path   string
	}{
		{"sdm", http.MethodGet, "/api/finance/budgets"},
		{"penjualan", http.MethodPost, "/api/billing/invoices"},
		{"pengamat", http.MethodDelete, "/api/projects/" + uuid.NewString()},
		{"keuangan", http.MethodGet, "/api/users"},
		{"pengadaan", http.MethodPost, "/api/reports"},
	}
	for _, current := range cases {
		response, recorder := callRouter(t, current.method, current.path, subjectResult(current.role, true))
		if response.Code != http.StatusForbidden {
			t.Errorf("%s %s %s: got %d", current.role, current.method, current.path, response.Code)
		}
		if queries := recorder.Queries(); len(queries) != 1 {
			t.Errorf("%s %s: handler must not run, got %d queries", current.role, current.path, len(queries))
		}
	}
}

func TestRouterLetsPermittedRoleReachTheHandler(t *testing.T) {
	response, recorder := callRouter(t, http.MethodGet, "/api/hr/employees", subjectResult("sdm", true))

	if response.Code == http.StatusForbidden || response.Code == http.StatusUnauthorized {
		t.Fatalf("got %d: %s", response.Code, response.Body.String())
	}
	if queries := recorder.Queries(); len(queries) < 2 {
		t.Fatalf("handler did not run, got %d queries", len(queries))
	}
}

func TestRouterKeepsLoginOpenWithoutRoleLookup(t *testing.T) {
	response, recorder := callRouter(t, http.MethodPost, "/api/auth/logout")

	if response.Code != http.StatusNoContent {
		t.Fatalf("got %d", response.Code)
	}
	if queries := recorder.Queries(); len(queries) != 0 {
		t.Fatalf("got %d queries", len(queries))
	}
}
