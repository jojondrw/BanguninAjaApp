package scoring

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

func (c *Controller) ListDimensions(ctx *gin.Context) {
	dimensions, err := c.service.ListDimensions(ctx.Request.Context())
	httpresponse.Respond(ctx, dimensions, err)
}

func (c *Controller) GetDimension(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	dimension, err := c.service.GetDimension(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, dimension, err)
}

func (c *Controller) CreateDimension(ctx *gin.Context) {
	var request DimensionRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	dimension, err := c.service.CreateDimension(ctx.Request.Context(), request)
	httpresponse.RespondCreated(ctx, dimension, err)
}

func (c *Controller) UpdateDimension(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request DimensionRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	dimension, err := c.service.UpdateDimension(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, dimension, err)
}

func (c *Controller) DeleteDimension(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeleteDimension(ctx.Request.Context(), id))
}

func (c *Controller) ListBuildingProfiles(ctx *gin.Context) {
	profiles, err := c.service.ListBuildingProfiles(ctx.Request.Context())
	httpresponse.Respond(ctx, profiles, err)
}

func (c *Controller) GetBuildingProfile(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	profile, err := c.service.GetBuildingProfile(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, profile, err)
}

func (c *Controller) CreateBuildingProfile(ctx *gin.Context) {
	var request BuildingProfileRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	profile, err := c.service.CreateBuildingProfile(ctx.Request.Context(), request)
	httpresponse.RespondCreated(ctx, profile, err)
}

func (c *Controller) UpdateBuildingProfile(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request BuildingProfileRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	profile, err := c.service.UpdateBuildingProfile(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, profile, err)
}

func (c *Controller) DeleteBuildingProfile(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeleteBuildingProfile(ctx.Request.Context(), id))
}

func (c *Controller) ListWeights(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	weights, err := c.service.ListWeights(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, weights, err)
}

func (c *Controller) ReplaceWeights(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request WeightsRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	weights, err := c.service.ReplaceWeights(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, weights, err)
}
