package master

import (
	"github.com/gin-gonic/gin"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httprequest"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httpresponse"
)

type Controller struct {
	service Service
}

func NewController(service Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) ListRegions(ctx *gin.Context) {
	var query RegionQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListRegions(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) GetRegion(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, "id")
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	region, err := c.service.GetRegion(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, region, err)
}

func (c *Controller) CreateRegion(ctx *gin.Context) {
	var request RegionRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	region, err := c.service.CreateRegion(ctx.Request.Context(), request)
	httpresponse.RespondCreated(ctx, region, err)
}

func (c *Controller) UpdateRegion(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, "id")
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request RegionRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	region, err := c.service.UpdateRegion(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, region, err)
}

func (c *Controller) DeleteRegion(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, "id")
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeleteRegion(ctx.Request.Context(), id))
}

func (c *Controller) ListUnitsOfMeasure(ctx *gin.Context) {
	var query UnitOfMeasureQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListUnitsOfMeasure(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) GetUnitOfMeasure(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, "id")
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	unit, err := c.service.GetUnitOfMeasure(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, unit, err)
}

func (c *Controller) CreateUnitOfMeasure(ctx *gin.Context) {
	var request UnitOfMeasureRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	unit, err := c.service.CreateUnitOfMeasure(ctx.Request.Context(), request)
	httpresponse.RespondCreated(ctx, unit, err)
}

func (c *Controller) UpdateUnitOfMeasure(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, "id")
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request UnitOfMeasureRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	unit, err := c.service.UpdateUnitOfMeasure(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, unit, err)
}

func (c *Controller) DeleteUnitOfMeasure(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, "id")
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeleteUnitOfMeasure(ctx.Request.Context(), id))
}

func (c *Controller) ListAccounts(ctx *gin.Context) {
	var query AccountQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListAccounts(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) GetAccount(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, "id")
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	account, err := c.service.GetAccount(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, account, err)
}

func (c *Controller) CreateAccount(ctx *gin.Context) {
	var request AccountRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	account, err := c.service.CreateAccount(ctx.Request.Context(), request)
	httpresponse.RespondCreated(ctx, account, err)
}

func (c *Controller) UpdateAccount(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, "id")
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request AccountRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	account, err := c.service.UpdateAccount(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, account, err)
}

func (c *Controller) DeleteAccount(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, "id")
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeleteAccount(ctx.Request.Context(), id))
}
