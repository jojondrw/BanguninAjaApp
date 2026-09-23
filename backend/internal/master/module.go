package master

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
	routes := router.Group("/master", middleware.Authentication(m.tokens))

	regions := routes.Group("/regions")
	regions.GET("", m.controller.ListRegions)
	regions.POST("", m.controller.CreateRegion)
	regions.GET("/:id", m.controller.GetRegion)
	regions.PUT("/:id", m.controller.UpdateRegion)
	regions.DELETE("/:id", m.controller.DeleteRegion)

	units := routes.Group("/units-of-measure")
	units.GET("", m.controller.ListUnitsOfMeasure)
	units.POST("", m.controller.CreateUnitOfMeasure)
	units.GET("/:id", m.controller.GetUnitOfMeasure)
	units.PUT("/:id", m.controller.UpdateUnitOfMeasure)
	units.DELETE("/:id", m.controller.DeleteUnitOfMeasure)

	accounts := routes.Group("/accounts")
	accounts.GET("", m.controller.ListAccounts)
	accounts.POST("", m.controller.CreateAccount)
	accounts.GET("/:id", m.controller.GetAccount)
	accounts.PUT("/:id", m.controller.UpdateAccount)
	accounts.DELETE("/:id", m.controller.DeleteAccount)
}
