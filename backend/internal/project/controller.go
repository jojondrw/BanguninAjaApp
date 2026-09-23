package project

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httprequest"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httpresponse"
)

const (
	projectParam = "id"
	childParam   = "childId"
)

type Controller struct {
	service Service
}

func NewController(service Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) ListProjects(ctx *gin.Context) {
	var query ProjectQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListProjects(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) GetProject(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, projectParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	project, err := c.service.GetProject(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, project, err)
}

func (c *Controller) CreateProject(ctx *gin.Context) {
	var request ProjectRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	project, err := c.service.CreateProject(ctx.Request.Context(), request)
	httpresponse.RespondCreated(ctx, project, err)
}

func (c *Controller) UpdateProject(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, projectParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request ProjectRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	project, err := c.service.UpdateProject(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, project, err)
}

func (c *Controller) UpdateProjectStatus(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, projectParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request ProjectStatusRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	project, err := c.service.UpdateProjectStatus(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, project, err)
}

func (c *Controller) DeleteProject(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, projectParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeleteProject(ctx.Request.Context(), id))
}

func (c *Controller) ListPhases(ctx *gin.Context) {
	projectID, err := httprequest.PathID(ctx, projectParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	phases, err := c.service.ListPhases(ctx.Request.Context(), projectID)
	httpresponse.Respond(ctx, phases, err)
}

func (c *Controller) CreatePhase(ctx *gin.Context) {
	projectID, err := httprequest.PathID(ctx, projectParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request PhaseRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	phase, err := c.service.CreatePhase(ctx.Request.Context(), projectID, request)
	httpresponse.RespondCreated(ctx, phase, err)
}

func (c *Controller) UpdatePhase(ctx *gin.Context) {
	projectID, phaseID, err := nestedIDs(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request PhaseRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	phase, err := c.service.UpdatePhase(ctx.Request.Context(), projectID, phaseID, request)
	httpresponse.Respond(ctx, phase, err)
}

func (c *Controller) DeletePhase(ctx *gin.Context) {
	projectID, phaseID, err := nestedIDs(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeletePhase(ctx.Request.Context(), projectID, phaseID))
}

func (c *Controller) ListBudgetItems(ctx *gin.Context) {
	projectID, err := httprequest.PathID(ctx, projectParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var query BudgetItemQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListBudgetItems(ctx.Request.Context(), projectID, query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) CreateBudgetItem(ctx *gin.Context) {
	projectID, err := httprequest.PathID(ctx, projectParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request BudgetItemRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	item, err := c.service.CreateBudgetItem(ctx.Request.Context(), projectID, request)
	httpresponse.RespondCreated(ctx, item, err)
}

func (c *Controller) UpdateBudgetItem(ctx *gin.Context) {
	projectID, itemID, err := nestedIDs(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request BudgetItemRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	item, err := c.service.UpdateBudgetItem(ctx.Request.Context(), projectID, itemID, request)
	httpresponse.Respond(ctx, item, err)
}

func (c *Controller) DeleteBudgetItem(ctx *gin.Context) {
	projectID, itemID, err := nestedIDs(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeleteBudgetItem(ctx.Request.Context(), projectID, itemID))
}

func (c *Controller) ListPermits(ctx *gin.Context) {
	projectID, err := httprequest.PathID(ctx, projectParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	permits, err := c.service.ListPermits(ctx.Request.Context(), projectID)
	httpresponse.Respond(ctx, permits, err)
}

func (c *Controller) CreatePermit(ctx *gin.Context) {
	projectID, err := httprequest.PathID(ctx, projectParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request PermitRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	permit, err := c.service.CreatePermit(ctx.Request.Context(), projectID, request)
	httpresponse.RespondCreated(ctx, permit, err)
}

func (c *Controller) UpdatePermit(ctx *gin.Context) {
	projectID, permitID, err := nestedIDs(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request PermitRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	permit, err := c.service.UpdatePermit(ctx.Request.Context(), projectID, permitID, request)
	httpresponse.Respond(ctx, permit, err)
}

func (c *Controller) DeletePermit(ctx *gin.Context) {
	projectID, permitID, err := nestedIDs(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeletePermit(ctx.Request.Context(), projectID, permitID))
}

func nestedIDs(ctx *gin.Context) (uuid.UUID, uuid.UUID, error) {
	projectID, err := httprequest.PathID(ctx, projectParam)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	childID, err := httprequest.PathID(ctx, childParam)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return projectID, childID, nil
}
