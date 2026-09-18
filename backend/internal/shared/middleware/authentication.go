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
		value := bearerToken(c.GetHeader("Authorization"))
		if value == "" {
			abortUnauthorized(c, "missing_access_token", "Token akses tidak ditemukan")
			return
		}

		userID, err := manager.ParseAccess(value)
		if err != nil {
			abortUnauthorized(c, "invalid_access_token", "Token akses tidak valid atau sudah kedaluwarsa")
			return
		}

		c.Set(authenticatedUserKey, userID)
		c.Next()
	}
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
