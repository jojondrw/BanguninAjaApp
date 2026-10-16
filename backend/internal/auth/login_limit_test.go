package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/cookie"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/middleware"
)

type countingLogins struct {
	Service
	calls int
}

func (s *countingLogins) Login(_ context.Context, request LoginRequest) (Session, error) {
	s.calls++
	if request.Password != "kata-sandi-benar" {
		return Session{}, errInvalidCredential
	}
	return Session{RefreshToken: "refresh"}, nil
}

func loginRouter(service Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	controller := NewController(service, cookie.NewRefreshWriter(config.Cookie{Name: "refresh_token", Path: "/api/auth"}))
	router := gin.New()
	router.Use(middleware.SecurityHeaders(), middleware.ErrorHandler())
	router.POST("/api/auth/login", controller.Login)
	return router
}

func login(router *gin.Engine, email, password string) *httptest.ResponseRecorder {
	body := `{"email":"` + email + `","password":"` + password + `"}`
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func TestLoginIsBlockedAfterTenAttemptsForTheSameEmail(t *testing.T) {
	service := &countingLogins{}
	router := loginRouter(service)

	for attempt := 1; attempt <= loginAttemptsPerEmail; attempt++ {
		if response := login(router, "Korban@BanguninAja.id", "tebakan-salah"); response.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d: got %d, want 401", attempt, response.Code)
		}
	}

	blocked := login(router, "KORBAN@banguninaja.id", "kata-sandi-benar")
	if blocked.Code != http.StatusTooManyRequests || !strings.Contains(blocked.Body.String(), "too_many_login_attempts") {
		t.Fatalf("attempt 11: got %d %s, want 429 too_many_login_attempts", blocked.Code, blocked.Body.String())
	}
	if service.calls != loginAttemptsPerEmail {
		t.Fatalf("the service was called %d times, want the blocked attempt to stop before it", service.calls)
	}

	if other := login(router, "lain@banguninaja.id", "kata-sandi-benar"); other.Code != http.StatusOK {
		t.Fatalf("another email got %d, want 200", other.Code)
	}
}

func TestAPIResponsesCarrySecurityHeaders(t *testing.T) {
	response := login(loginRouter(&countingLogins{}), "uji@banguninaja.id", "kata-sandi-benar")

	expected := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}
	for header, want := range expected {
		if got := response.Header().Get(header); got != want {
			t.Errorf("%s = %q, want %q", header, got, want)
		}
	}
}
