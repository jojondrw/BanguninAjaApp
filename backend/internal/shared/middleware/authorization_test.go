package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/access"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
)

type fakeSubjects struct {
	subjects map[uuid.UUID]access.Subject
	err      error
	lookups  int
}

func (f *fakeSubjects) Subject(_ context.Context, userID uuid.UUID) (access.Subject, error) {
	f.lookups++
	if f.err != nil {
		return access.Subject{}, f.err
	}
	subject, ok := f.subjects[userID]
	if !ok {
		return access.Subject{}, access.ErrUnknownSubject
	}
	return subject, nil
}

type authorizationHarness struct {
	router   *gin.Engine
	tokens   *token.Manager
	subjects *fakeSubjects
}

var guardedRoutes = []string{
	"/projects", "/finance", "/billing", "/sales", "/procurement", "/inventory", "/assets", "/equipment",
	"/hr", "/locations", "/site", "/scoring", "/master", "/reports", "/regulation", "/news", "/users", "/brand-new",
}

func newAuthorizationHarness(t *testing.T) authorizationHarness {
	t.Helper()
	gin.SetMode(gin.TestMode)

	tokens := token.NewManager(config.Token{AccessSecret: "authorization-test", AccessTTL: time.Minute, Issuer: "authorization-test"})
	subjects := &fakeSubjects{subjects: map[uuid.UUID]access.Subject{}}

	router := gin.New()
	router.Use(ErrorHandler())
	api := router.Group("/api", Authorization(tokens, subjects))
	reply := func(c *gin.Context) {
		role, _ := AuthenticatedRole(c)
		c.JSON(http.StatusOK, gin.H{"role": role})
	}
	api.POST("/auth/login", reply)
	api.GET("/auth/me", Authentication(tokens), reply)
	for _, prefix := range guardedRoutes {
		group := api.Group(prefix, Authentication(tokens))
		group.GET("/:id", reply)
		group.POST("/:id", reply)
		group.PUT("/:id", reply)
		group.PATCH("/:id", reply)
		group.DELETE("/:id", reply)
	}

	return authorizationHarness{router: router, tokens: tokens, subjects: subjects}
}

func (h authorizationHarness) userWith(t *testing.T, role string, active bool) string {
	t.Helper()
	userID := uuid.New()
	h.subjects.subjects[userID] = access.Subject{Role: role, Active: active}
	issued, err := h.tokens.IssueAccess(userID)
	if err != nil {
		t.Fatalf("issue access token: %v", err)
	}
	return issued.Value
}

func (h authorizationHarness) call(method, path, accessToken string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, nil)
	if accessToken != "" {
		request.Header.Set("Authorization", "Bearer "+accessToken)
	}
	response := httptest.NewRecorder()
	h.router.ServeHTTP(response, request)
	return response
}

func errorCode(t *testing.T, response *httptest.ResponseRecorder) (string, string) {
	t.Helper()
	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode %s: %v", response.Body.String(), err)
	}
	return body.Error.Code, body.Error.Message
}

func TestAuthorizationFollowsTheMatrixForEveryRoleMethodAndPrefix(t *testing.T) {
	harness := newAuthorizationHarness(t)
	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}

	for _, role := range access.RoleCodes() {
		accessToken := harness.userWith(t, role, true)
		for _, prefix := range guardedRoutes {
			module, _ := access.ModuleForPath("/api" + prefix + "/x")
			for _, method := range methods {
				response := harness.call(method, "/api"+prefix+"/x", accessToken)
				allowed := access.Allows(role, module, access.LevelForMethod(method))
				if allowed && response.Code != http.StatusOK {
					t.Errorf("%s %s %s: got %d, want 200", role, method, prefix, response.Code)
				}
				if !allowed && response.Code != http.StatusForbidden {
					t.Errorf("%s %s %s: got %d, want 403", role, method, prefix, response.Code)
				}
			}
		}
	}
}

func TestAuthorizationSpotChecksTheRequestedRoles(t *testing.T) {
	harness := newAuthorizationHarness(t)
	cases := []struct {
		role   string
		method string
		path   string
		status int
	}{
		{"keuangan", http.MethodPost, "/api/finance/x", http.StatusOK},
		{"keuangan", http.MethodGet, "/api/sales/x", http.StatusOK},
		{"keuangan", http.MethodPut, "/api/sales/x", http.StatusForbidden},
		{"keuangan", http.MethodGet, "/api/hr/x", http.StatusForbidden},
		{"penjualan", http.MethodGet, "/api/billing/x", http.StatusOK},
		{"penjualan", http.MethodPost, "/api/billing/x", http.StatusForbidden},
		{"pengadaan", http.MethodDelete, "/api/equipment/x", http.StatusOK},
		{"manajer_proyek", http.MethodPost, "/api/site/x", http.StatusOK},
		{"manajer_proyek", http.MethodPatch, "/api/procurement/x", http.StatusForbidden},
		{"sdm", http.MethodPatch, "/api/hr/x", http.StatusOK},
		{"sdm", http.MethodGet, "/api/finance/x", http.StatusForbidden},
		{"pengamat", http.MethodGet, "/api/hr/x", http.StatusOK},
		{"pengamat", http.MethodDelete, "/api/projects/x", http.StatusForbidden},
		{"pengamat", http.MethodGet, "/api/users/x", http.StatusForbidden},
		{"admin", http.MethodPatch, "/api/users/x", http.StatusOK},
		{"admin", http.MethodGet, "/api/brand-new/x", http.StatusOK},
		{"pengamat", http.MethodGet, "/api/brand-new/x", http.StatusForbidden},
	}
	for _, current := range cases {
		response := harness.call(current.method, current.path, harness.userWith(t, current.role, true))
		if response.Code != current.status {
			t.Errorf("%s %s %s: got %d, want %d", current.role, current.method, current.path, response.Code, current.status)
		}
	}
}

func TestAuthorizationRespondsForbiddenWithTheMissingAccess(t *testing.T) {
	harness := newAuthorizationHarness(t)

	response := harness.call(http.MethodPost, "/api/finance/x", harness.userWith(t, "pengamat", true))

	code, message := errorCode(t, response)
	if response.Code != http.StatusForbidden || code != "forbidden" {
		t.Fatalf("got %d %s", response.Code, code)
	}
	if message != "Peran Anda tidak punya akses mengubah data modul Keuangan" {
		t.Fatalf("got message %q", message)
	}
}

func TestAuthorizationStoresTheRoleForHandlers(t *testing.T) {
	harness := newAuthorizationHarness(t)

	response := harness.call(http.MethodGet, "/api/hr/x", harness.userWith(t, "sdm", true))

	if response.Code != http.StatusOK || response.Body.String() != `{"role":"sdm"}` {
		t.Fatalf("got %d %s", response.Code, response.Body.String())
	}
}

func TestAuthorizationLeavesAuthRoutesOpen(t *testing.T) {
	harness := newAuthorizationHarness(t)

	if response := harness.call(http.MethodPost, "/api/auth/login", ""); response.Code != http.StatusOK {
		t.Fatalf("login got %d", response.Code)
	}
	if response := harness.call(http.MethodGet, "/api/auth/me", harness.userWith(t, "", true)); response.Code != http.StatusOK {
		t.Fatalf("me got %d", response.Code)
	}
	if harness.subjects.lookups != 0 {
		t.Fatalf("auth routes must not look up roles, got %d lookups", harness.subjects.lookups)
	}
}

func TestAuthorizationRejectsMissingTokenBeforeLookingUpTheRole(t *testing.T) {
	harness := newAuthorizationHarness(t)

	response := harness.call(http.MethodGet, "/api/projects/x", "")

	code, _ := errorCode(t, response)
	if response.Code != http.StatusUnauthorized || code != "missing_access_token" || harness.subjects.lookups != 0 {
		t.Fatalf("got %d %s after %d lookups", response.Code, code, harness.subjects.lookups)
	}
}

func TestAuthorizationLooksUpTheRoleOncePerRequest(t *testing.T) {
	harness := newAuthorizationHarness(t)

	harness.call(http.MethodGet, "/api/projects/x", harness.userWith(t, "admin", true))

	if harness.subjects.lookups != 1 {
		t.Fatalf("got %d lookups", harness.subjects.lookups)
	}
}

func TestAuthorizationRejectsDeactivatedAndUnknownUsers(t *testing.T) {
	harness := newAuthorizationHarness(t)

	inactive := harness.call(http.MethodGet, "/api/projects/x", harness.userWith(t, "admin", false))
	if code, _ := errorCode(t, inactive); inactive.Code != http.StatusUnauthorized || code != "account_inactive" {
		t.Fatalf("inactive got %d %s", inactive.Code, code)
	}

	issued, _ := harness.tokens.IssueAccess(uuid.New())
	unknown := harness.call(http.MethodGet, "/api/projects/x", issued.Value)
	if code, _ := errorCode(t, unknown); unknown.Code != http.StatusUnauthorized || code != "user_not_found" {
		t.Fatalf("unknown got %d %s", unknown.Code, code)
	}
}

func TestAuthorizationTreatsUserWithoutRoleAsNoAccess(t *testing.T) {
	harness := newAuthorizationHarness(t)
	accessToken := harness.userWith(t, "", true)
	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}

	for _, prefix := range guardedRoutes {
		for _, method := range methods {
			response := harness.call(method, "/api"+prefix+"/x", accessToken)
			if code, _ := errorCode(t, response); response.Code != http.StatusForbidden || code != "forbidden" {
				t.Errorf("%s %s: got %d %s", method, prefix, response.Code, code)
			}
		}
	}

	if response := harness.call(http.MethodGet, "/api/auth/me", accessToken); response.Code != http.StatusOK {
		t.Fatalf("a user without role must still read the profile, got %d", response.Code)
	}
}

func TestAuthorizationReportsLookupFailureAsServerError(t *testing.T) {
	harness := newAuthorizationHarness(t)
	accessToken := harness.userWith(t, "admin", true)
	harness.subjects.err = errors.New("database down")

	response := harness.call(http.MethodGet, "/api/projects/x", accessToken)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("got %d", response.Code)
	}
}
