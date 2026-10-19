package billing

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/finance"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/ethereum"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/middleware"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
)

type Module struct {
	controller *Controller
	onchain    *OnchainController
	gateway    *GatewayController
	tokens     *token.Manager
}

func NewModule(db *gorm.DB, tokens *token.Manager, onchain config.Onchain, gateway config.Gateway) *Module {
	repository := NewRepository(db, NewFinanceCashBook)
	return &Module{
		controller: NewController(NewService(repository)),
		onchain:    NewOnchainController(NewOnchainService(repository, ethereum.NewClient(onchain.RPCURL), onchain)),
		gateway:    NewGatewayController(NewGatewayService(repository, NewPaymentGateway(gateway), gateway)),
		tokens:     tokens,
	}
}

type financeCashBook struct {
	ledger finance.BillingCashLedger
}

func NewFinanceCashBook(db *gorm.DB) CashBook {
	return financeCashBook{ledger: finance.NewBillingCashLedger(finance.NewRepository(db))}
}

func (b financeCashBook) Record(ctx context.Context, entry CashEntry) error {
	return b.ledger.Record(ctx, finance.BillingCash{
		PaymentID: entry.PaymentID,
		Incoming:  entry.Incoming,
		ProjectID: entry.ProjectID,
		Date:      entry.Date,
		Amount:    entry.Amount,
		Note:      entry.Note,
	})
}

func (m *Module) RegisterRoutes(router gin.IRouter) {
	routes := router.Group("/billing", middleware.Authentication(m.tokens))

	invoices := routes.Group("/invoices")
	invoices.GET("", m.controller.ListInvoices)
	invoices.POST("", m.controller.CreateInvoice)
	invoices.GET("/:id", m.controller.GetInvoice)
	invoices.PUT("/:id", m.controller.UpdateInvoice)
	invoices.DELETE("/:id", m.controller.DeleteInvoice)
	invoices.POST("/:id/payments", m.controller.PayInvoice)
	invoices.GET("/:id/payments", m.controller.InvoicePayments)
	invoices.POST("/:id/onchain-requests", m.onchain.CreateInvoiceRequest)
	invoices.GET("/:id/onchain-requests", m.onchain.InvoiceRequests)
	invoices.POST("/:id/gateway-requests", m.gateway.CreateInvoiceRequest)
	invoices.GET("/:id/gateway-requests", m.gateway.InvoiceRequests)

	receivables := routes.Group("/receivables")
	receivables.GET("", m.controller.ListReceivables)
	receivables.POST("", m.controller.CreateReceivable)
	receivables.GET("/:id", m.controller.GetReceivable)
	receivables.PUT("/:id", m.controller.UpdateReceivable)
	receivables.DELETE("/:id", m.controller.DeleteReceivable)
	receivables.POST("/:id/payments", m.controller.PayReceivable)
	receivables.GET("/:id/payments", m.controller.ReceivablePayments)
	receivables.POST("/:id/onchain-requests", m.onchain.CreateReceivableRequest)
	receivables.GET("/:id/onchain-requests", m.onchain.ReceivableRequests)
	receivables.POST("/:id/gateway-requests", m.gateway.CreateReceivableRequest)
	receivables.GET("/:id/gateway-requests", m.gateway.ReceivableRequests)

	payables := routes.Group("/payables")
	payables.GET("", m.controller.ListPayables)
	payables.POST("", m.controller.CreatePayable)
	payables.GET("/:id", m.controller.GetPayable)
	payables.PUT("/:id", m.controller.UpdatePayable)
	payables.DELETE("/:id", m.controller.DeletePayable)
	payables.POST("/:id/payments", m.controller.PayPayable)
	payables.GET("/:id/payments", m.controller.PayablePayments)

	routes.GET("/reconciliation", m.controller.ListUnreconciledPairs)
	routes.GET("/summary", m.controller.Summarize)

	onchain := routes.Group("/onchain")
	onchain.GET("/config", m.onchain.Config)
	onchain.POST("/requests/:id/cancel", m.onchain.CancelRequest)

	gateway := routes.Group("/gateway")
	gateway.GET("/config", m.gateway.Config)
	gateway.POST("/requests/:id/cancel", m.gateway.CancelRequest)

	pay := router.Group("/pay")
	pay.GET("/:token", m.onchain.PublicPayment)
	pay.POST("/:token/confirm", m.onchain.ConfirmPayment)
	pay.GET("/qris/:token", m.gateway.PublicPayment)
	pay.POST("/qris/:token/refresh", m.gateway.RefreshPayment)
	pay.POST("/qris/:token/simulate", m.gateway.SimulatePayment)
	pay.POST("/xendit/webhook", m.gateway.Webhook)
}
