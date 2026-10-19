package billing

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httprequest"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httpresponse"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/ratelimit"
)

const (
	callbackTokenHeader = "x-callback-token"

	publicRefreshLimit  = 40
	publicSimulateLimit = 5
	webhookLimit        = 120
	webhookBodyLimit    = 64 << 10
)

type GatewayController struct {
	service   GatewayService
	views     *ratelimit.Window
	refreshes *ratelimit.Window
	simulates *ratelimit.Window
	webhooks  *ratelimit.Window
}

func NewGatewayController(service GatewayService) *GatewayController {
	return &GatewayController{
		service:   service,
		views:     ratelimit.NewWindow(publicViewLimit, publicLimitWindow),
		refreshes: ratelimit.NewWindow(publicRefreshLimit, publicLimitWindow),
		simulates: ratelimit.NewWindow(publicSimulateLimit, publicLimitWindow),
		webhooks:  ratelimit.NewWindow(webhookLimit, publicLimitWindow),
	}
}

func (c *GatewayController) Config(ctx *gin.Context) {
	httpresponse.OK(ctx, c.service.Config())
}

func (c *GatewayController) CreateInvoiceRequest(ctx *gin.Context) {
	c.create(ctx, c.service.CreateInvoiceRequest)
}

func (c *GatewayController) CreateReceivableRequest(ctx *gin.Context) {
	c.create(ctx, c.service.CreateReceivableRequest)
}

func (c *GatewayController) InvoiceRequests(ctx *gin.Context) {
	c.list(ctx, c.service.InvoiceRequests)
}

func (c *GatewayController) ReceivableRequests(ctx *gin.Context) {
	c.list(ctx, c.service.ReceivableRequests)
}

func (c *GatewayController) CancelRequest(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	request, err := c.service.CancelRequest(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, request, err)
}

func (c *GatewayController) PublicPayment(ctx *gin.Context) {
	c.public(ctx, c.views, c.service.PublicPayment)
}

func (c *GatewayController) RefreshPayment(ctx *gin.Context) {
	c.public(ctx, c.refreshes, c.service.RefreshPayment)
}

func (c *GatewayController) SimulatePayment(ctx *gin.Context) {
	c.public(ctx, c.simulates, c.service.SimulatePayment)
}

func (c *GatewayController) Webhook(ctx *gin.Context) {
	if !c.webhooks.Allow(ctx.ClientIP()) {
		_ = ctx.Error(errTooManyAttempts)
		return
	}
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, webhookBodyLimit)
	var body GatewayWebhook
	if err := ctx.ShouldBindJSON(&body); err != nil {
		_ = ctx.Error(apperror.BadRequest("invalid_webhook", "Isi callback tidak bisa dibaca"))
		return
	}
	if err := c.service.HandleWebhook(ctx.Request.Context(), ctx.GetHeader(callbackTokenHeader), body); err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.OK(ctx, gin.H{"received": true})
}

type gatewayPublicAction func(ctx context.Context, token string) (GatewayPublicResponse, error)

func (c *GatewayController) public(ctx *gin.Context, limiter *ratelimit.Window, action gatewayPublicAction) {
	token := ctx.Param(tokenParam)
	if !IsPaymentToken(token) {
		_ = ctx.Error(errGatewayNotFound)
		return
	}
	if !limiter.Allow(token) {
		_ = ctx.Error(errTooManyAttempts)
		return
	}
	payment, err := action(ctx.Request.Context(), token)
	httpresponse.Respond(ctx, payment, err)
}

type gatewayCreator func(ctx context.Context, id uuid.UUID, request GatewayRequestCreate, actor *uuid.UUID) (GatewayRequestResponse, error)

func (c *GatewayController) create(ctx *gin.Context, create gatewayCreator) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request GatewayRequestCreate
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	created, err := create(ctx.Request.Context(), id, request, actorOf(ctx))
	httpresponse.RespondCreated(ctx, created, err)
}

type gatewayLister func(ctx context.Context, id uuid.UUID, query OnchainRequestQuery) (pagination.Page[GatewayRequestResponse], error)

func (c *GatewayController) list(ctx *gin.Context, list gatewayLister) {
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
