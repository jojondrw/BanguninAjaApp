package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/cookie"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httprequest"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httpresponse"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/middleware"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/ratelimit"
)

const (
	loginAttemptsPerEmail = 10
	loginAttemptWindow    = 15 * time.Minute
)

var errTooManyLogins = apperror.New(
	http.StatusTooManyRequests,
	"too_many_login_attempts",
	"Terlalu banyak percobaan masuk untuk email ini. Coba lagi dalam 15 menit",
)

type Controller struct {
	service       Service
	refreshCookie cookie.RefreshWriter
	logins        *ratelimit.Window
}

func NewController(service Service, refreshCookie cookie.RefreshWriter) *Controller {
	return &Controller{
		service:       service,
		refreshCookie: refreshCookie,
		logins:        ratelimit.NewWindow(loginAttemptsPerEmail, loginAttemptWindow),
	}
}

func (c *Controller) Register(ctx *gin.Context) {
	var request RegisterRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
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
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	if !c.logins.Allow(normalizeEmail(request.Email)) {
		_ = ctx.Error(errTooManyLogins)
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

func (c *Controller) UpdateProfile(ctx *gin.Context) {
	userID, err := httprequest.UserID(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	var request UpdateProfileRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}

	user, err := c.service.UpdateProfile(ctx.Request.Context(), userID, request)
	httpresponse.Respond(ctx, user, err)
}

func (c *Controller) ChangePassword(ctx *gin.Context) {
	userID, err := httprequest.UserID(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	var request ChangePasswordRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}

	err = c.service.ChangePassword(ctx.Request.Context(), userID, c.refreshCookie.Read(ctx), request)
	httpresponse.RespondNoContent(ctx, err)
}

func (c *Controller) ListUsers(ctx *gin.Context) {
	var query UserQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}

	page, err := c.service.ListUsers(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) ChangeUserRole(ctx *gin.Context) {
	userID, err := httprequest.PathID(ctx, "id")
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	var request ChangeRoleRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}

	user, err := c.service.ChangeRole(ctx.Request.Context(), userID, request)
	httpresponse.Respond(ctx, user, err)
}

func (c *Controller) ChangeUserActive(ctx *gin.Context) {
	userID, err := httprequest.PathID(ctx, "id")
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	var request ChangeActiveRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}

	user, err := c.service.ChangeActive(ctx.Request.Context(), userID, request)
	httpresponse.Respond(ctx, user, err)
}
