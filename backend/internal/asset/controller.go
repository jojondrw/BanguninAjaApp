package asset

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

func (c *Controller) ListAssets(ctx *gin.Context) {
	var query AssetQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListAssets(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) GetAsset(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	asset, err := c.service.GetAsset(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, asset, err)
}

func (c *Controller) CreateAsset(ctx *gin.Context) {
	var request AssetRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	asset, err := c.service.CreateAsset(ctx.Request.Context(), request)
	httpresponse.RespondCreated(ctx, asset, err)
}

func (c *Controller) UpdateAsset(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request AssetRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	asset, err := c.service.UpdateAsset(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, asset, err)
}

func (c *Controller) DeleteAsset(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeleteAsset(ctx.Request.Context(), id))
}

func (c *Controller) ListEquipment(ctx *gin.Context) {
	var query EquipmentQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListEquipment(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) GetEquipment(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	equipment, err := c.service.GetEquipment(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, equipment, err)
}

func (c *Controller) CreateEquipment(ctx *gin.Context) {
	var request EquipmentRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	equipment, err := c.service.CreateEquipment(ctx.Request.Context(), request)
	httpresponse.RespondCreated(ctx, equipment, err)
}

func (c *Controller) UpdateEquipment(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request EquipmentRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	equipment, err := c.service.UpdateEquipment(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, equipment, err)
}

func (c *Controller) DeleteEquipment(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeleteEquipment(ctx.Request.Context(), id))
}
