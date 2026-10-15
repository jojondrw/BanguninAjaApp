package reporting

import (
	"mime"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httprequest"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httpresponse"
)

const reportParam = "id"

type Controller struct {
	service Service
}

func NewController(service Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) ListReports(ctx *gin.Context) {
	var query ReportQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListReports(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) GetReport(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, reportParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	report, err := c.service.GetReport(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, report, err)
}

func (c *Controller) CreateReport(ctx *gin.Context) {
	userID, err := httprequest.UserID(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request ReportRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	report, err := c.service.CreateReport(ctx.Request.Context(), userID, request)
	httpresponse.RespondCreated(ctx, report, err)
}

func (c *Controller) DeleteReport(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, reportParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeleteReport(ctx.Request.Context(), id))
}

func (c *Controller) PreviewReport(ctx *gin.Context) {
	var query PreviewQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	preview, err := c.service.PreviewReport(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, preview, err)
}

func (c *Controller) DownloadReport(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, reportParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	file, err := c.service.DownloadReport(ctx.Request.Context(), id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": file.Name}))
	ctx.Header("Access-Control-Expose-Headers", "Content-Disposition")
	ctx.Header("Cache-Control", "no-store")
	ctx.Data(http.StatusOK, file.ContentType, file.Content)
}
