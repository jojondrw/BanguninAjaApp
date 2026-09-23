package sales

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httprequest"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httpresponse"
)

const (
	idParam    = "id"
	childParam = "childId"
)

type Controller struct {
	service Service
}

func NewController(service Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) ListCustomers(ctx *gin.Context) {
	var query CustomerQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListCustomers(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) GetCustomer(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	customer, err := c.service.GetCustomer(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, customer, err)
}

func (c *Controller) CreateCustomer(ctx *gin.Context) {
	var request CustomerRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	customer, err := c.service.CreateCustomer(ctx.Request.Context(), request)
	httpresponse.RespondCreated(ctx, customer, err)
}

func (c *Controller) UpdateCustomer(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request CustomerRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	customer, err := c.service.UpdateCustomer(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, customer, err)
}

func (c *Controller) DeleteCustomer(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeleteCustomer(ctx.Request.Context(), id))
}

func (c *Controller) ListUnits(ctx *gin.Context) {
	var query UnitQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListUnits(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) SummarizeUnits(ctx *gin.Context) {
	var query UnitSummaryQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	summary, err := c.service.SummarizeUnits(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, summary, err)
}

func (c *Controller) GetUnit(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	unit, err := c.service.GetUnit(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, unit, err)
}

func (c *Controller) CreateUnit(ctx *gin.Context) {
	var request UnitRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	unit, err := c.service.CreateUnit(ctx.Request.Context(), request)
	httpresponse.RespondCreated(ctx, unit, err)
}

func (c *Controller) UpdateUnit(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request UnitRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	unit, err := c.service.UpdateUnit(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, unit, err)
}

func (c *Controller) DeleteUnit(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeleteUnit(ctx.Request.Context(), id))
}

func (c *Controller) ListLeads(ctx *gin.Context) {
	var query LeadQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListLeads(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) GetLead(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	lead, err := c.service.GetLead(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, lead, err)
}

func (c *Controller) CreateLead(ctx *gin.Context) {
	var request LeadRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	lead, err := c.service.CreateLead(ctx.Request.Context(), request)
	httpresponse.RespondCreated(ctx, lead, err)
}

func (c *Controller) UpdateLead(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request LeadRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	lead, err := c.service.UpdateLead(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, lead, err)
}

func (c *Controller) DeleteLead(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeleteLead(ctx.Request.Context(), id))
}

func (c *Controller) ListContracts(ctx *gin.Context) {
	var query ContractQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListContracts(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) GetContract(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	contract, err := c.service.GetContract(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, contract, err)
}

func (c *Controller) CreateContract(ctx *gin.Context) {
	var request ContractRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	contract, err := c.service.CreateContract(ctx.Request.Context(), request)
	httpresponse.RespondCreated(ctx, contract, err)
}

func (c *Controller) UpdateContract(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request ContractRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	contract, err := c.service.UpdateContract(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, contract, err)
}

func (c *Controller) UpdateContractStatus(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request ContractStatusRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	contract, err := c.service.UpdateContractStatus(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, contract, err)
}

func (c *Controller) DeleteContract(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeleteContract(ctx.Request.Context(), id))
}

func (c *Controller) ListInstallments(ctx *gin.Context) {
	var query InstallmentQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListInstallments(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) CreateInstallment(ctx *gin.Context) {
	contractID, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request InstallmentRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	installment, err := c.service.CreateInstallment(ctx.Request.Context(), contractID, request)
	httpresponse.RespondCreated(ctx, installment, err)
}

func (c *Controller) UpdateInstallment(ctx *gin.Context) {
	contractID, installmentID, err := nestedIDs(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request InstallmentRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	installment, err := c.service.UpdateInstallment(ctx.Request.Context(), contractID, installmentID, request)
	httpresponse.Respond(ctx, installment, err)
}

func (c *Controller) DeleteInstallment(ctx *gin.Context) {
	contractID, installmentID, err := nestedIDs(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeleteInstallment(ctx.Request.Context(), contractID, installmentID))
}

func (c *Controller) PayInstallment(ctx *gin.Context) {
	contractID, installmentID, err := nestedIDs(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request InstallmentPaymentRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	installment, err := c.service.PayInstallment(ctx.Request.Context(), contractID, installmentID, request)
	httpresponse.Respond(ctx, installment, err)
}

func nestedIDs(ctx *gin.Context) (uuid.UUID, uuid.UUID, error) {
	parentID, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	childID, err := httprequest.PathID(ctx, childParam)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return parentID, childID, nil
}
