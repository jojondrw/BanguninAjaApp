package billing

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httprequest"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httpresponse"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

const (
	tokenParam = "token"

	publicViewLimit    = 60
	publicConfirmLimit = 20
	publicLimitWindow  = time.Minute
)

type OnchainController struct {
	service  OnchainService
	views    *windowLimiter
	confirms *windowLimiter
}

func NewOnchainController(service OnchainService) *OnchainController {
	return &OnchainController{
		service:  service,
		views:    newWindowLimiter(publicViewLimit, publicLimitWindow),
		confirms: newWindowLimiter(publicConfirmLimit, publicLimitWindow),
	}
}

func (c *OnchainController) Config(ctx *gin.Context) {
	httpresponse.OK(ctx, c.service.Config())
}

func (c *OnchainController) CreateInvoiceRequest(ctx *gin.Context) {
	c.create(ctx, c.service.CreateInvoiceRequest)
}

func (c *OnchainController) CreateReceivableRequest(ctx *gin.Context) {
	c.create(ctx, c.service.CreateReceivableRequest)
}

func (c *OnchainController) InvoiceRequests(ctx *gin.Context) {
	c.list(ctx, c.service.InvoiceRequests)
}

func (c *OnchainController) ReceivableRequests(ctx *gin.Context) {
	c.list(ctx, c.service.ReceivableRequests)
}

func (c *OnchainController) CancelRequest(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	request, err := c.service.CancelRequest(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, request, err)
}

func (c *OnchainController) PublicPayment(ctx *gin.Context) {
	token, ok := c.publicToken(ctx, c.views)
	if !ok {
		return
	}
	payment, err := c.service.PublicPayment(ctx.Request.Context(), token)
	httpresponse.Respond(ctx, payment, err)
}

func (c *OnchainController) ConfirmPayment(ctx *gin.Context) {
	token, ok := c.publicToken(ctx, c.confirms)
	if !ok {
		return
	}
	var request OnchainConfirmRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	outcome, err := c.service.ConfirmPayment(ctx.Request.Context(), token, request)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	if outcome.Verification == verificationWaiting {
		ctx.JSON(http.StatusAccepted, httpresponse.Envelope{Data: outcome})
		return
	}
	httpresponse.OK(ctx, outcome)
}

func (c *OnchainController) publicToken(ctx *gin.Context, limiter *windowLimiter) (string, bool) {
	token := ctx.Param(tokenParam)
	if !IsPaymentToken(token) {
		_ = ctx.Error(errOnchainNotFound)
		return "", false
	}
	if !limiter.Allow(token) {
		_ = ctx.Error(errTooManyAttempts)
		return "", false
	}
	return token, true
}

type onchainCreator func(ctx context.Context, id uuid.UUID, request OnchainRequestCreate, actor *uuid.UUID) (OnchainRequestResponse, error)

func (c *OnchainController) create(ctx *gin.Context, create onchainCreator) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request OnchainRequestCreate
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	created, err := create(ctx.Request.Context(), id, request, actorOf(ctx))
	httpresponse.RespondCreated(ctx, created, err)
}

type onchainLister func(ctx context.Context, id uuid.UUID, query OnchainRequestQuery) (pagination.Page[OnchainRequestResponse], error)

func (c *OnchainController) list(ctx *gin.Context, list onchainLister) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var query OnchainRequestQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := list(ctx.Request.Context(), id, query)
	httpresponse.Respond(ctx, page, err)
}
