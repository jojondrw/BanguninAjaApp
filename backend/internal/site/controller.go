package site

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

// Evaluate handles POST /api/site/evaluate (auth required). See T1 contract §2.
func (c *Controller) Evaluate(ctx *gin.Context) {
	userID, err := httprequest.UserID(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	var request EvaluateRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}

	response, err := c.service.Evaluate(ctx.Request.Context(), userID, request)
	httpresponse.RespondCreated(ctx, response, err)
}
