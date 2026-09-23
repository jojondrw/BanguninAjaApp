package finance

import (
	"github.com/gin-gonic/gin"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httprequest"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httpresponse"
)

const idParam = "id"

type Controller struct {
	service Service
}

func NewController(service Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) ListBudgets(ctx *gin.Context) {
	var query BudgetQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListBudgets(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) GetBudget(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	budget, err := c.service.GetBudget(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, budget, err)
}

func (c *Controller) CreateBudget(ctx *gin.Context) {
	var request BudgetRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	budget, err := c.service.CreateBudget(ctx.Request.Context(), request)
	httpresponse.RespondCreated(ctx, budget, err)
}

func (c *Controller) UpdateBudget(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request BudgetRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	budget, err := c.service.UpdateBudget(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, budget, err)
}

func (c *Controller) DeleteBudget(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeleteBudget(ctx.Request.Context(), id))
}

func (c *Controller) ListCashTransactions(ctx *gin.Context) {
	var query CashTransactionQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListCashTransactions(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) GetCashTransaction(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	transaction, err := c.service.GetCashTransaction(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, transaction, err)
}

func (c *Controller) CreateCashTransaction(ctx *gin.Context) {
	var request CashTransactionRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	transaction, err := c.service.CreateCashTransaction(ctx.Request.Context(), request)
	httpresponse.RespondCreated(ctx, transaction, err)
}

func (c *Controller) UpdateCashTransaction(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request CashTransactionRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	transaction, err := c.service.UpdateCashTransaction(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, transaction, err)
}

func (c *Controller) DeleteCashTransaction(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeleteCashTransaction(ctx.Request.Context(), id))
}

func (c *Controller) SummarizeCashFlow(ctx *gin.Context) {
	var query CashFlowQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	summary, err := c.service.SummarizeCashFlow(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, summary, err)
}

func (c *Controller) ListJournalEntries(ctx *gin.Context) {
	var query JournalEntryQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListJournalEntries(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) GetJournalEntry(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	entry, err := c.service.GetJournalEntry(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, entry, err)
}

func (c *Controller) RecordJournalEntry(ctx *gin.Context) {
	var request JournalEntryRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	entry, err := c.service.RecordJournalEntry(ctx.Request.Context(), request)
	httpresponse.RespondCreated(ctx, entry, err)
}

func (c *Controller) ReadLedger(ctx *gin.Context) {
	var query LedgerQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	ledger, err := c.service.ReadLedger(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, ledger, err)
}
