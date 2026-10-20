package evidence

import (
	"github.com/gin-gonic/gin"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httprequest"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httpresponse"
)

const (
	projectParam = "id"
	childParam   = "childId"
	tokenParam   = "token"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) ProjectEvidence(ctx *gin.Context) {
	projectID, err := httprequest.PathID(ctx, projectParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	evidence, err := c.service.ProjectEvidence(ctx.Request.Context(), projectID)
	httpresponse.Respond(ctx, evidence, err)
}

func (c *Controller) SetApprover(ctx *gin.Context) {
	projectID, err := httprequest.PathID(ctx, projectParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request ApproverRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	address, err := c.service.SetApprover(ctx.Request.Context(), projectID, request.Address)
	httpresponse.Respond(ctx, gin.H{"address": address}, err)
}

func (c *Controller) CreateClaimLink(ctx *gin.Context) {
	projectID, err := httprequest.PathID(ctx, projectParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	termID, err := httprequest.PathID(ctx, childParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	claim, err := c.service.CreateClaimLink(ctx.Request.Context(), projectID, termID)
	httpresponse.RespondCreated(ctx, claim, err)
}

func (c *Controller) Config(ctx *gin.Context) {
	httpresponse.OK(ctx, c.service.Config())
}

func (c *Controller) AnchorWithServer(ctx *gin.Context) {
	actorID, err := httprequest.UserID(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	sent, err := c.service.AnchorWithServer(ctx.Request.Context(), &actorID)
	httpresponse.RespondCreated(ctx, sent, err)
}

func (c *Controller) ListBatches(ctx *gin.Context) {
	batches, err := c.service.ListBatches(ctx.Request.Context())
	httpresponse.Respond(ctx, batches, err)
}

func (c *Controller) CreateBatch(ctx *gin.Context) {
	actorID, err := httprequest.UserID(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	batch, err := c.service.CreateBatch(ctx.Request.Context(), actorID)
	httpresponse.RespondCreated(ctx, batch, err)
}

func (c *Controller) Anchor(ctx *gin.Context) {
	batchID, err := httprequest.PathID(ctx, projectParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request AnchorRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	result, err := c.service.Anchor(ctx.Request.Context(), batchID, request.TxHash)
	httpresponse.Respond(ctx, result, err)
}

func (c *Controller) PublicClaim(ctx *gin.Context) {
	claim, err := c.service.PublicClaim(ctx.Request.Context(), ctx.Param(tokenParam))
	httpresponse.Respond(ctx, claim, err)
}

func (c *Controller) Approve(ctx *gin.Context) {
	var request ApproveRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	status, err := c.service.Approve(ctx.Request.Context(), ctx.Param(tokenParam), request.Signature)
	httpresponse.Respond(ctx, status, err)
}
