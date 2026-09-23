package finance

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
	routes := router.Group("/finance", middleware.Authentication(m.tokens))

	budgets := routes.Group("/budgets")
	budgets.GET("", m.controller.ListBudgets)
	budgets.POST("", m.controller.CreateBudget)
	budgets.GET("/:id", m.controller.GetBudget)
	budgets.PUT("/:id", m.controller.UpdateBudget)
	budgets.DELETE("/:id", m.controller.DeleteBudget)

	cash := routes.Group("/cash-transactions")
	cash.GET("", m.controller.ListCashTransactions)
	cash.POST("", m.controller.CreateCashTransaction)
	cash.GET("/:id", m.controller.GetCashTransaction)
	cash.PUT("/:id", m.controller.UpdateCashTransaction)
	cash.DELETE("/:id", m.controller.DeleteCashTransaction)

	routes.GET("/cash-flow", m.controller.SummarizeCashFlow)

	journals := routes.Group("/journal-entries")
	journals.GET("", m.controller.ListJournalEntries)
	journals.POST("", m.controller.RecordJournalEntry)
	journals.GET("/:id", m.controller.GetJournalEntry)

	routes.GET("/ledger", m.controller.ReadLedger)
}
