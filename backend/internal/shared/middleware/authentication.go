package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
)

const (
	authenticatedUserKey = "authenticated_user_id"
	bearerPrefix         = "Bearer "
)

func Authentication(manager *token.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := authenticate(c, manager); ok {
			c.Next()
		}
	}
}

func authenticate(c *gin.Context, manager *token.Manager) (uuid.UUID, bool) {
	if userID, ok := AuthenticatedUserID(c); ok {
		return userID, true
	}

	value := bearerToken(c.GetHeader("Authorization"))
	if value == "" {
		abortUnauthorized(c, "missing_access_token", "Token akses tidak ditemukan")
		return uuid.Nil, false
	}

	userID, err := manager.ParseAccess(value)
	if err != nil {
		abortUnauthorized(c, "invalid_access_token", "Token akses tidak valid atau sudah kedaluwarsa")
		return uuid.Nil, false
	}

	c.Set(authenticatedUserKey, userID)
	return userID, true
}

func AuthenticatedUserID(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get(authenticatedUserKey)
	if !exists {
		return uuid.Nil, false
	}

	userID, ok := value.(uuid.UUID)
	return userID, ok
}

func bearerToken(header string) string {
	if !strings.HasPrefix(header, bearerPrefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, bearerPrefix))
}

func abortUnauthorized(c *gin.Context, code, message string) {
	_ = c.Error(apperror.Unauthorized(code, message))
	c.Abort()
}
