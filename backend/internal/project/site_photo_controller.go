package project

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httprequest"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httpresponse"
)

const (
	sitePhotoField      = "file"
	sitePhotoFormSlack  = 64 << 10
	sitePhotoDateLayout = "2006-01-02"
)

var errSitePhotoDate = apperror.BadRequest("site_photo_date", "Tanggal foto harus berformat YYYY-MM-DD")

func (c *Controller) ListSitePhotos(ctx *gin.Context) {
	projectID, err := httprequest.PathID(ctx, projectParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	photos, err := c.service.ListSitePhotos(ctx.Request.Context(), projectID)
	httpresponse.Respond(ctx, photos, err)
}

func (c *Controller) CreateSitePhoto(ctx *gin.Context) {
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
	upload, err := readSitePhotoUpload(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	photo, err := c.service.CreateSitePhoto(ctx.Request.Context(), projectID, actorID, upload)
	httpresponse.RespondCreated(ctx, photo, err)
}

func (c *Controller) SitePhotoFile(ctx *gin.Context) {
	projectID, photoID, err := nestedIDs(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	file, err := c.service.SitePhotoFile(ctx.Request.Context(), projectID, photoID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.Header("Cache-Control", "private, max-age=86400, immutable")
	ctx.Header("X-Content-Type-Options", "nosniff")
	ctx.Data(http.StatusOK, file.ContentType, file.Content)
}

func (c *Controller) DeleteSitePhoto(ctx *gin.Context) {
	projectID, photoID, err := nestedIDs(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeleteSitePhoto(ctx.Request.Context(), projectID, photoID))
}

func readSitePhotoUpload(ctx *gin.Context) (SitePhotoUpload, error) {
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, SitePhotoMaxBytes+sitePhotoFormSlack)
	header, err := ctx.FormFile(sitePhotoField)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return SitePhotoUpload{}, errSitePhotoTooLarge
		}
		return SitePhotoUpload{}, errSitePhotoEmpty
	}
	if header.Size > SitePhotoMaxBytes {
		return SitePhotoUpload{}, errSitePhotoTooLarge
	}
	file, err := header.Open()
	if err != nil {
		return SitePhotoUpload{}, apperror.Internal(err)
	}
	defer file.Close()
	content, err := io.ReadAll(io.LimitReader(file, SitePhotoMaxBytes+1))
	if err != nil {
		return SitePhotoUpload{}, apperror.Internal(err)
	}

	upload := SitePhotoUpload{Caption: ctx.PostForm("caption"), Content: content}
	if raw := ctx.PostForm("takenOn"); raw != "" {
		takenOn, err := time.Parse(sitePhotoDateLayout, raw)
		if err != nil {
			return SitePhotoUpload{}, errSitePhotoDate
		}
		upload.TakenOn = &takenOn
	}
	return upload, nil
}
