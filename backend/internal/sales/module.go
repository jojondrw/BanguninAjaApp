package sales

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
	routes := router.Group("/sales", middleware.Authentication(m.tokens))

	customers := routes.Group("/customers")
	customers.GET("", m.controller.ListCustomers)
	customers.POST("", m.controller.CreateCustomer)
	customers.GET("/:id", m.controller.GetCustomer)
	customers.PUT("/:id", m.controller.UpdateCustomer)
	customers.DELETE("/:id", m.controller.DeleteCustomer)

	units := routes.Group("/units")
	units.GET("", m.controller.ListUnits)
	units.POST("", m.controller.CreateUnit)
	units.GET("/summary", m.controller.SummarizeUnits)
	units.GET("/:id", m.controller.GetUnit)
	units.PUT("/:id", m.controller.UpdateUnit)
	units.DELETE("/:id", m.controller.DeleteUnit)

	leads := routes.Group("/leads")
	leads.GET("", m.controller.ListLeads)
	leads.POST("", m.controller.CreateLead)
	leads.GET("/:id", m.controller.GetLead)
	leads.PUT("/:id", m.controller.UpdateLead)
	leads.DELETE("/:id", m.controller.DeleteLead)

	contracts := routes.Group("/contracts")
	contracts.GET("", m.controller.ListContracts)
	contracts.POST("", m.controller.CreateContract)
	contracts.GET("/:id", m.controller.GetContract)
	contracts.PUT("/:id", m.controller.UpdateContract)
	contracts.PATCH("/:id/status", m.controller.UpdateContractStatus)
	contracts.DELETE("/:id", m.controller.DeleteContract)
	contracts.POST("/:id/installments", m.controller.CreateInstallment)
	contracts.PUT("/:id/installments/:childId", m.controller.UpdateInstallment)
	contracts.DELETE("/:id/installments/:childId", m.controller.DeleteInstallment)
	contracts.POST("/:id/installments/:childId/payment", m.controller.PayInstallment)

	routes.GET("/installments", m.controller.ListInstallments)
}
