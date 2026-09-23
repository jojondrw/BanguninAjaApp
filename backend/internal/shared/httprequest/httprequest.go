package httprequest

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/middleware"
)

var (
	errInvalidPayload = apperror.BadRequest("invalid_payload", "Data yang dikirim belum lengkap atau formatnya salah")
	errInvalidQuery   = apperror.BadRequest("invalid_query", "Parameter pencarian tidak valid")
	errInvalidID      = apperror.BadRequest("invalid_id", "Format id tidak valid")
	errMissingUser    = apperror.Unauthorized("missing_access_token", "Token akses tidak ditemukan")
)

func BindJSON(c *gin.Context, target any) error {
	if err := c.ShouldBindJSON(target); err != nil {
		return errInvalidPayload.WithCause(err)
	}
	return nil
}

func BindQuery(c *gin.Context, target any) error {
	if err := c.ShouldBindQuery(target); err != nil {
		return errInvalidQuery.WithCause(err)
	}
	return nil
}

func PathID(c *gin.Context, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		return uuid.Nil, errInvalidID.WithCause(err)
	}
	return id, nil
}

func UserID(c *gin.Context) (uuid.UUID, error) {
	userID, ok := middleware.AuthenticatedUserID(c)
	if !ok {
		return uuid.Nil, errMissingUser
	}
	return userID, nil
}
