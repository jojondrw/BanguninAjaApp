package inventory

import (
	"github.com/gin-gonic/gin"

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

func (c *Controller) ListMaterials(ctx *gin.Context) {
	var query MaterialQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListMaterials(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) ListLowStockMaterials(ctx *gin.Context) {
	var query pagination.Query
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListLowStockMaterials(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) GetMaterial(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	material, err := c.service.GetMaterial(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, material, err)
}

func (c *Controller) CreateMaterial(ctx *gin.Context) {
	var request MaterialRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	material, err := c.service.CreateMaterial(ctx.Request.Context(), request)
	httpresponse.RespondCreated(ctx, material, err)
}

func (c *Controller) UpdateMaterial(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request MaterialRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	material, err := c.service.UpdateMaterial(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, material, err)
}

func (c *Controller) DeleteMaterial(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeleteMaterial(ctx.Request.Context(), id))
}

func (c *Controller) ListWarehouses(ctx *gin.Context) {
	var query WarehouseQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListWarehouses(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) GetWarehouse(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	warehouse, err := c.service.GetWarehouse(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, warehouse, err)
}

func (c *Controller) CreateWarehouse(ctx *gin.Context) {
	var request WarehouseRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	warehouse, err := c.service.CreateWarehouse(ctx.Request.Context(), request)
	httpresponse.RespondCreated(ctx, warehouse, err)
}

func (c *Controller) UpdateWarehouse(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request WarehouseRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	warehouse, err := c.service.UpdateWarehouse(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, warehouse, err)
}

func (c *Controller) DeleteWarehouse(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeleteWarehouse(ctx.Request.Context(), id))
}

func (c *Controller) ListStocks(ctx *gin.Context) {
	var query StockQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListStocks(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) ListStockMovements(ctx *gin.Context) {
	var query StockMovementQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListStockMovements(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) GetStockMovement(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	movement, err := c.service.GetStockMovement(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, movement, err)
}

func (c *Controller) RecordStockMovement(ctx *gin.Context) {
	var request StockMovementRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	movement, err := c.service.RecordStockMovement(ctx.Request.Context(), request)
	httpresponse.RespondCreated(ctx, movement, err)
}
