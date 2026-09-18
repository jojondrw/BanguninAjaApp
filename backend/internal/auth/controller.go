package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/cookie"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httpresponse"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/middleware"
)

type Controller struct {
	service       Service
	refreshCookie cookie.RefreshWriter
}

func NewController(service Service, refreshCookie cookie.RefreshWriter) *Controller {
	return &Controller{service: service, refreshCookie: refreshCookie}
}

func (c *Controller) Register(ctx *gin.Context) {
	var request RegisterRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		_ = ctx.Error(invalidPayload(err))
		return
	}

	user, err := c.service.Register(ctx.Request.Context(), request)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	httpresponse.Created(ctx, user)
}

func (c *Controller) Login(ctx *gin.Context) {
	var request LoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		_ = ctx.Error(invalidPayload(err))
		return
	}

	session, err := c.service.Login(ctx.Request.Context(), request)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	c.refreshCookie.Set(ctx, session.RefreshToken, session.RefreshExpiresAt)
	httpresponse.OK(ctx, session.Response)
}

func (c *Controller) Refresh(ctx *gin.Context) {
	session, err := c.service.Refresh(ctx.Request.Context(), c.refreshCookie.Read(ctx))
	if err != nil {
		c.refreshCookie.Clear(ctx)
		_ = ctx.Error(err)
		return
	}

	c.refreshCookie.Set(ctx, session.RefreshToken, session.RefreshExpiresAt)
	httpresponse.OK(ctx, session.Response)
}

func (c *Controller) Logout(ctx *gin.Context) {
	if err := c.service.Logout(ctx.Request.Context(), c.refreshCookie.Read(ctx)); err != nil {
		_ = ctx.Error(err)
		return
	}

	c.refreshCookie.Clear(ctx)
	ctx.Status(http.StatusNoContent)
}

func (c *Controller) Profile(ctx *gin.Context) {
	userID, ok := middleware.AuthenticatedUserID(ctx)
	if !ok {
		_ = ctx.Error(apperror.Unauthorized("missing_access_token", "Token akses tidak ditemukan"))
		return
	}

	user, err := c.service.Profile(ctx.Request.Context(), userID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	httpresponse.OK(ctx, user)
}

func invalidPayload(err error) error {
	return apperror.BadRequest("invalid_payload", "Data yang dikirim belum lengkap atau formatnya salah").WithCause(err)
}
