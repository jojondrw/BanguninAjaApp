package middleware

import (
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

func TestAuthorizationLetsThePublicPayLinkThroughWithoutAToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens := token.NewManager(config.Token{AccessSecret: "public-test", AccessTTL: time.Minute, Issuer: "public-test"})
	subjects := &fakeSubjects{subjects: map[uuid.UUID]access.Subject{}}
	router := gin.New()
	router.Use(ErrorHandler())
	api := router.Group("/api", Authorization(tokens, subjects))
	reply := func(c *gin.Context) { c.Status(http.StatusOK) }
	api.GET("/pay/:token", reply)
	api.POST("/pay/:token/confirm", reply)
	api.GET("/billing/onchain/config", Authentication(tokens), reply)

	for _, current := range []struct{ method, path string }{
		{http.MethodGet, "/api/pay/AbCdEfGhIjKlMnOpQrStUvWxYz012345"},
		{http.MethodPost, "/api/pay/AbCdEfGhIjKlMnOpQrStUvWxYz012345/confirm"},
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(current.method, current.path, nil))
		if response.Code != http.StatusOK {
			t.Errorf("%s %s: got %d", current.method, current.path, response.Code)
		}
	}
	if subjects.lookups != 0 {
		t.Fatalf("the public link must not look up a role, got %d lookups", subjects.lookups)
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/billing/onchain/config", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("staff routes must still need a token, got %d", response.Code)
	}
}
