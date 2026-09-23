package location

import (
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

func (c *Controller) ListSavedLocations(ctx *gin.Context) {
	userID, err := httprequest.UserID(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var query SavedLocationQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListSavedLocations(ctx.Request.Context(), userID, query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) GetSavedLocation(ctx *gin.Context) {
	userID, id, err := ownerAndID(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	location, err := c.service.GetSavedLocation(ctx.Request.Context(), userID, id)
	httpresponse.Respond(ctx, location, err)
}

func (c *Controller) CreateSavedLocation(ctx *gin.Context) {
	userID, err := httprequest.UserID(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request SavedLocationRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	location, err := c.service.CreateSavedLocation(ctx.Request.Context(), userID, request)
	httpresponse.RespondCreated(ctx, location, err)
}

func (c *Controller) UpdateSavedLocation(ctx *gin.Context) {
	userID, id, err := ownerAndID(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request SavedLocationRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	location, err := c.service.UpdateSavedLocation(ctx.Request.Context(), userID, id, request)
	httpresponse.Respond(ctx, location, err)
}

func (c *Controller) DeleteSavedLocation(ctx *gin.Context) {
	userID, id, err := ownerAndID(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeleteSavedLocation(ctx.Request.Context(), userID, id))
}

func (c *Controller) ListComparisons(ctx *gin.Context) {
	userID, err := httprequest.UserID(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var query pagination.Query
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListComparisons(ctx.Request.Context(), userID, query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) GetComparison(ctx *gin.Context) {
	userID, id, err := ownerAndID(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	comparison, err := c.service.GetComparison(ctx.Request.Context(), userID, id)
	httpresponse.Respond(ctx, comparison, err)
}

func (c *Controller) CreateComparison(ctx *gin.Context) {
	userID, err := httprequest.UserID(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request ComparisonRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	comparison, err := c.service.CreateComparison(ctx.Request.Context(), userID, request)
	httpresponse.RespondCreated(ctx, comparison, err)
}

func (c *Controller) UpdateComparison(ctx *gin.Context) {
	userID, id, err := ownerAndID(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request ComparisonRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	comparison, err := c.service.UpdateComparison(ctx.Request.Context(), userID, id, request)
	httpresponse.Respond(ctx, comparison, err)
}

func (c *Controller) DeleteComparison(ctx *gin.Context) {
	userID, id, err := ownerAndID(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeleteComparison(ctx.Request.Context(), userID, id))
}

func ownerAndID(ctx *gin.Context) (uuid.UUID, uuid.UUID, error) {
	userID, err := httprequest.UserID(ctx)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return userID, id, nil
}
