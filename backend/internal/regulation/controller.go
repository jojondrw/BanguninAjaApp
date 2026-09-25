package regulation

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

func (c *Controller) Lookup(ctx *gin.Context) {
	var request LookupRequest

	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}

	response, err := c.service.GetByPoint(
		ctx.Request.Context(),
		request.Latitude,
		request.Longitude,
	)

	httpresponse.Respond(ctx, response, err)
}
