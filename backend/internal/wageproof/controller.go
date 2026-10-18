package wageproof

import (
	"github.com/gin-gonic/gin"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httprequest"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httpresponse"
)

const projectParam = "id"

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) List(ctx *gin.Context) {
	projectID, err := httprequest.PathID(ctx, projectParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	proofs, err := c.service.List(ctx.Request.Context(), projectID)
	httpresponse.Respond(ctx, proofs, err)
}

func (c *Controller) Create(ctx *gin.Context) {
	projectID, err := httprequest.PathID(ctx, projectParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	actorID, err := httprequest.UserID(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request CreateRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	result, err := c.service.Create(ctx.Request.Context(), projectID, actorID, request.Period)
	httpresponse.Respond(ctx, result, err)
}

func (c *Controller) PublicProof(ctx *gin.Context) {
	proof, err := c.service.PublicProof(ctx.Request.Context(), ctx.Param("token"))
	httpresponse.Respond(ctx, proof, err)
}

func (c *Controller) PublicReceipt(ctx *gin.Context) {
	receipt, err := c.service.PublicReceipt(ctx.Request.Context(), ctx.Param("token"))
	httpresponse.Respond(ctx, receipt, err)
}
