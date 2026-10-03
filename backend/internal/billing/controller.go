package billing

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httprequest"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httpresponse"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

const idParam = "id"

type Controller struct {
	service Service
}

func NewController(service Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) ListInvoices(ctx *gin.Context) {
	var query InvoiceQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListInvoices(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) GetInvoice(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	invoice, err := c.service.GetInvoice(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, invoice, err)
}

func (c *Controller) CreateInvoice(ctx *gin.Context) {
	var request InvoiceRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	invoice, err := c.service.CreateInvoice(ctx.Request.Context(), request)
	httpresponse.RespondCreated(ctx, invoice, err)
}

func (c *Controller) UpdateInvoice(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request InvoiceRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	invoice, err := c.service.UpdateInvoice(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, invoice, err)
}

func (c *Controller) DeleteInvoice(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeleteInvoice(ctx.Request.Context(), id))
}

func (c *Controller) PayInvoice(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request PaymentRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	invoice, err := c.service.PayInvoice(ctx.Request.Context(), id, request, actorOf(ctx))
	httpresponse.Respond(ctx, invoice, err)
}

func (c *Controller) ListReceivables(ctx *gin.Context) {
	var query ReceivableQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListReceivables(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) GetReceivable(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	receivable, err := c.service.GetReceivable(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, receivable, err)
}

func (c *Controller) CreateReceivable(ctx *gin.Context) {
	var request ReceivableRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	receivable, err := c.service.CreateReceivable(ctx.Request.Context(), request)
	httpresponse.RespondCreated(ctx, receivable, err)
}

func (c *Controller) UpdateReceivable(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request ReceivableRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	receivable, err := c.service.UpdateReceivable(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, receivable, err)
}

func (c *Controller) DeleteReceivable(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeleteReceivable(ctx.Request.Context(), id))
}

func (c *Controller) PayReceivable(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request PaymentRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	receivable, err := c.service.PayReceivable(ctx.Request.Context(), id, request, actorOf(ctx))
	httpresponse.Respond(ctx, receivable, err)
}

func (c *Controller) ListPayables(ctx *gin.Context) {
	var query PayableQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListPayables(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) GetPayable(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	payable, err := c.service.GetPayable(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, payable, err)
}

func (c *Controller) CreatePayable(ctx *gin.Context) {
	var request PayableRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	payable, err := c.service.CreatePayable(ctx.Request.Context(), request)
	httpresponse.RespondCreated(ctx, payable, err)
}

func (c *Controller) UpdatePayable(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request PayableRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	payable, err := c.service.UpdatePayable(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, payable, err)
}

func (c *Controller) DeletePayable(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeletePayable(ctx.Request.Context(), id))
}

func (c *Controller) PayPayable(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request PaymentRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	payable, err := c.service.PayPayable(ctx.Request.Context(), id, request, actorOf(ctx))
	httpresponse.Respond(ctx, payable, err)
}

func (c *Controller) InvoicePayments(ctx *gin.Context) {
	c.listPayments(ctx, c.service.InvoicePayments)
}

func (c *Controller) ReceivablePayments(ctx *gin.Context) {
	c.listPayments(ctx, c.service.ReceivablePayments)
}

func (c *Controller) PayablePayments(ctx *gin.Context) {
	c.listPayments(ctx, c.service.PayablePayments)
}

type paymentLister func(ctx context.Context, id uuid.UUID, query PaymentQuery) (pagination.Page[PaymentResponse], error)

func (c *Controller) listPayments(ctx *gin.Context, list paymentLister) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var query PaymentQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := list(ctx.Request.Context(), id, query)
	httpresponse.Respond(ctx, page, err)
}

func actorOf(ctx *gin.Context) *uuid.UUID {
	userID, err := httprequest.UserID(ctx)
	if err != nil {
		return nil
	}
	return &userID
}
