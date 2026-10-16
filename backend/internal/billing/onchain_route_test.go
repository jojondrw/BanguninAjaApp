package billing

import (
	"context"
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
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/ratelimit"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
)

func servePublicRoute(t *testing.T, router *gin.Engine, method, path, body, accessToken string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		request.Header.Set("Authorization", "Bearer "+accessToken)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func onchainRouter(t *testing.T, results ...databasetest.Result) (*gin.Engine, *databasetest.Recorder, *token.Manager) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, recorder := databasetest.Open(t, results...)
	tokens := token.NewManager(config.Token{AccessSecret: "billing-onchain-test", AccessTTL: time.Minute, Issuer: "billing-onchain-test"})
	router := gin.New()
	router.Use(middleware.ErrorHandler())
	NewModule(db, tokens, config.Onchain{}).RegisterRoutes(router.Group("/api"))
	return router, recorder, tokens
}

func errorCodeOf(t *testing.T, response *httptest.ResponseRecorder) string {
	t.Helper()
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode %s: %v", response.Body.String(), err)
	}
	return body.Error.Code
}

func TestPublicPayRouteNeedsNoLogin(t *testing.T) {
	router, recorder, _ := onchainRouter(t)

	response := servePublicRoute(t, router, http.MethodGet, "/api/pay/"+testToken, "", "")
	if response.Code != http.StatusNotFound || errorCodeOf(t, response) != "onchain_request_not_found" {
		t.Fatalf("got %d %s", response.Code, response.Body.String())
	}
	queries := recorder.Queries()
	if len(queries) != 1 || !strings.Contains(queries[0].SQL, "onchain_payment_request") || queries[0].Args[0] != testToken {
		t.Fatalf("the handler must look the token up, got %+v", queries)
	}
}

func TestStaffOnchainRoutesStillNeedLogin(t *testing.T) {
	router, recorder, tokens := onchainRouter(t)

	for _, path := range []string{"/api/billing/onchain/config", "/api/billing/invoices/" + uuid.NewString() + "/onchain-requests"} {
		response := servePublicRoute(t, router, http.MethodGet, path, "", "")
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("%s: got %d", path, response.Code)
		}
	}
	if len(recorder.Queries()) != 0 {
		t.Fatal("an anonymous call must not reach the database")
	}

	issued, _ := tokens.IssueAccess(uuid.New())
	response := servePublicRoute(t, router, http.MethodGet, "/api/billing/onchain/config", "", issued.Value)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"enabled":false`) {
		t.Fatalf("got %d %s", response.Code, response.Body.String())
	}
}

func TestPublicPayRouteRejectsMalformedInputBeforeTheDatabase(t *testing.T) {
	router, recorder, _ := onchainRouter(t)
	cases := []struct {
		method, path, body string
		status             int
	}{
		{http.MethodGet, "/api/pay/short", "", http.StatusNotFound},
		{http.MethodGet, "/api/pay/" + strings.Repeat("!", 32), "", http.StatusNotFound},
		{http.MethodPost, "/api/pay/" + testToken + "/confirm", `{"txHash":"0x12"}`, http.StatusBadRequest},
		{http.MethodPost, "/api/pay/" + testToken + "/confirm", `{"txHash":"` + testHash[2:] + `00"}`, http.StatusBadRequest},
		{http.MethodPost, "/api/pay/" + testToken + "/confirm", `{"txHash":"0x` + strings.Repeat("g", 64) + `"}`, http.StatusBadRequest},
		{http.MethodPost, "/api/pay/" + testToken + "/confirm", `{}`, http.StatusBadRequest},
	}
	for _, current := range cases {
		response := servePublicRoute(t, router, current.method, current.path, current.body, "")
		if response.Code != current.status {
			t.Errorf("%s %s %s: got %d", current.method, current.path, current.body, response.Code)
		}
	}
	if len(recorder.Queries()) != 0 {
		t.Fatalf("malformed input must not reach the database, got %d queries", len(recorder.Queries()))
	}
}

func TestPublicConfirmIsRateLimitedPerToken(t *testing.T) {
	router, _, _ := onchainRouter(t)
	path := "/api/pay/" + testToken + "/confirm"

	for attempt := 0; attempt < publicConfirmLimit; attempt++ {
		if response := servePublicRoute(t, router, http.MethodPost, path, `{}`, ""); response.Code != http.StatusBadRequest {
			t.Fatalf("attempt %d: got %d", attempt, response.Code)
		}
	}
	response := servePublicRoute(t, router, http.MethodPost, path, `{}`, "")
	if response.Code != http.StatusTooManyRequests || errorCodeOf(t, response) != "too_many_requests" {
		t.Fatalf("got %d %s", response.Code, response.Body.String())
	}
	other := "/api/pay/ZZZZEfGhIjKlMnOpQrStUvWxYz012345/confirm"
	if response := servePublicRoute(t, router, http.MethodPost, other, `{}`, ""); response.Code != http.StatusBadRequest {
		t.Fatalf("another token has its own budget, got %d", response.Code)
	}
}

type waitingService struct {
	OnchainService
	received OnchainConfirmRequest
}

func (s *waitingService) ConfirmPayment(_ context.Context, _ string, request OnchainConfirmRequest) (OnchainConfirmResponse, error) {
	s.received = request
	return OnchainConfirmResponse{Verification: verificationWaiting, Message: waitingForBlock}, nil
}

func TestPublicConfirmAnswersAcceptedWhileWaiting(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &waitingService{}
	controller := NewOnchainController(service)
	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.POST("/api/pay/:token/confirm", controller.ConfirmPayment)

	response := servePublicRoute(t, router, http.MethodPost, "/api/pay/"+testToken+"/confirm", `{"txHash":"`+testHash+`"}`, "")
	if response.Code != http.StatusAccepted || !strings.Contains(response.Body.String(), `"verification":"waiting"`) {
		t.Fatalf("got %d %s", response.Code, response.Body.String())
	}
	if service.received.TxHash != testHash {
		t.Fatalf("got %+v", service.received)
	}
}

func TestWindowLimiterStartsAFreshWindow(t *testing.T) {
	limiter := ratelimit.NewWindow(2, time.Minute)
	current := onchainNow
	limiter.SetClock(func() time.Time { return current })

	if !limiter.Allow("a") || !limiter.Allow("a") || limiter.Allow("a") {
		t.Fatal("the third call inside the window must be refused")
	}
	current = current.Add(time.Minute)
	if !limiter.Allow("a") {
		t.Fatal("a new window must allow calls again")
	}
}
