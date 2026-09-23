package billing

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/middleware"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
)

type Module struct {
	controller *Controller
	tokens     *token.Manager
}

func NewModule(db *gorm.DB, tokens *token.Manager) *Module {
	return &Module{
		controller: NewController(NewService(NewRepository(db))),
		tokens:     tokens,
	}
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

	receivables := routes.Group("/receivables")
	receivables.GET("", m.controller.ListReceivables)
	receivables.POST("", m.controller.CreateReceivable)
	receivables.GET("/:id", m.controller.GetReceivable)
	receivables.PUT("/:id", m.controller.UpdateReceivable)
	receivables.DELETE("/:id", m.controller.DeleteReceivable)
	receivables.POST("/:id/payments", m.controller.PayReceivable)

	payables := routes.Group("/payables")
	payables.GET("", m.controller.ListPayables)
	payables.POST("", m.controller.CreatePayable)
	payables.GET("/:id", m.controller.GetPayable)
	payables.PUT("/:id", m.controller.UpdatePayable)
	payables.DELETE("/:id", m.controller.DeletePayable)
	payables.POST("/:id/payments", m.controller.PayPayable)
}
