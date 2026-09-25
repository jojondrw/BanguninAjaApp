package news

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

func (c *Controller) FetchNews(ctx *gin.Context) {
	var query NewsQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}

	news, err := c.service.FetchNews(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, news, err)
}