package procurement

import (
	"github.com/gin-gonic/gin"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httprequest"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httpresponse"
)

const idParam = "id"

type Controller struct {
	service Service
}

func NewController(service Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) ListVendors(ctx *gin.Context) {
	var query VendorQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListVendors(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) GetVendor(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	vendor, err := c.service.GetVendor(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, vendor, err)
}

func (c *Controller) CreateVendor(ctx *gin.Context) {
	var request VendorRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	vendor, err := c.service.CreateVendor(ctx.Request.Context(), request)
	httpresponse.RespondCreated(ctx, vendor, err)
}

func (c *Controller) UpdateVendor(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request VendorRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	vendor, err := c.service.UpdateVendor(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, vendor, err)
}

func (c *Controller) DeleteVendor(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeleteVendor(ctx.Request.Context(), id))
}

func (c *Controller) ListPurchaseRequests(ctx *gin.Context) {
	var query PurchaseRequestQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListPurchaseRequests(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) GetPurchaseRequest(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	document, err := c.service.GetPurchaseRequest(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, document, err)
}

func (c *Controller) CreatePurchaseRequest(ctx *gin.Context) {
	requesterID, err := httprequest.UserID(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request PurchaseRequestRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	document, err := c.service.CreatePurchaseRequest(ctx.Request.Context(), requesterID, request)
	httpresponse.RespondCreated(ctx, document, err)
}

func (c *Controller) UpdatePurchaseRequest(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request PurchaseRequestRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	document, err := c.service.UpdatePurchaseRequest(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, document, err)
}

func (c *Controller) UpdatePurchaseRequestStatus(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request PurchaseRequestStatusRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	document, err := c.service.UpdatePurchaseRequestStatus(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, document, err)
}

func (c *Controller) DeletePurchaseRequest(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeletePurchaseRequest(ctx.Request.Context(), id))
}

func (c *Controller) ListPurchaseOrders(ctx *gin.Context) {
	var query PurchaseOrderQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListPurchaseOrders(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) GetPurchaseOrder(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	order, err := c.service.GetPurchaseOrder(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, order, err)
}

func (c *Controller) CreatePurchaseOrder(ctx *gin.Context) {
	var request PurchaseOrderRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	order, err := c.service.CreatePurchaseOrder(ctx.Request.Context(), request)
	httpresponse.RespondCreated(ctx, order, err)
}

func (c *Controller) UpdatePurchaseOrder(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request PurchaseOrderRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	order, err := c.service.UpdatePurchaseOrder(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, order, err)
}

func (c *Controller) UpdatePurchaseOrderStatus(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request PurchaseOrderStatusRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	order, err := c.service.UpdatePurchaseOrderStatus(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, order, err)
}

func (c *Controller) DeletePurchaseOrder(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeletePurchaseOrder(ctx.Request.Context(), id))
}

func (c *Controller) ListGoodsReceipts(ctx *gin.Context) {
	var query GoodsReceiptQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListGoodsReceipts(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) GetGoodsReceipt(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	receipt, err := c.service.GetGoodsReceipt(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, receipt, err)
}

func (c *Controller) RecordGoodsReceipt(ctx *gin.Context) {
	var request GoodsReceiptRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	receipt, err := c.service.RecordGoodsReceipt(ctx.Request.Context(), request)
	httpresponse.RespondCreated(ctx, receipt, err)
}
