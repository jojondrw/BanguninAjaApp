package middleware

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/access"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
)

const authenticatedRoleKey = "authenticated_role"

type SubjectSource interface {
	Subject(ctx context.Context, userID uuid.UUID) (access.Subject, error)
}

func Authorization(manager *token.Manager, subjects SubjectSource) gin.HandlerFunc {
	return func(c *gin.Context) {
		module, guarded := access.ModuleForPath(routePath(c))
		if !guarded {
			c.Next()
			return
		}

		userID, ok := authenticate(c, manager)
		if !ok {
			return
		}

		subject, ok := resolveSubject(c, subjects, userID)
		if !ok {
			return
		}

		needed := access.LevelForMethod(c.Request.Method)
		if !access.Allows(subject.Role, module, needed) {
			_ = c.Error(apperror.Forbidden("forbidden", access.DeniedMessage(module, needed)))
			c.Abort()
			return
		}

		c.Set(authenticatedRoleKey, subject.Role)
		c.Next()
	}
}

func AuthenticatedRole(c *gin.Context) (string, bool) {
	value, exists := c.Get(authenticatedRoleKey)
	if !exists {
		return "", false
	}
	role, ok := value.(string)
	return role, ok
}

func routePath(c *gin.Context) string {
	if path := c.FullPath(); path != "" {
		return path
	}
	return c.Request.URL.Path
}

func resolveSubject(c *gin.Context, subjects SubjectSource, userID uuid.UUID) (access.Subject, bool) {
	subject, err := subjects.Subject(c.Request.Context(), userID)
	if errors.Is(err, access.ErrUnknownSubject) {
		abortUnauthorized(c, "user_not_found", "Akun tidak ditemukan")
		return access.Subject{}, false
	}
	if err != nil {
		_ = c.Error(apperror.Internal(err))
		c.Abort()
		return access.Subject{}, false
	}
	if !subject.Active {
		abortUnauthorized(c, "account_inactive", "Akun ini sudah dinonaktifkan, hubungi admin")
		return access.Subject{}, false
	}
	return subject, true
}
