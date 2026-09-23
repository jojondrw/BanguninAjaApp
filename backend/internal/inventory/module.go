package inventory

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
	routes := router.Group("/inventory", middleware.Authentication(m.tokens))

	materials := routes.Group("/materials")
	materials.GET("", m.controller.ListMaterials)
	materials.POST("", m.controller.CreateMaterial)
	materials.GET("/low-stock", m.controller.ListLowStockMaterials)
	materials.GET("/:id", m.controller.GetMaterial)
	materials.PUT("/:id", m.controller.UpdateMaterial)
	materials.DELETE("/:id", m.controller.DeleteMaterial)

	warehouses := routes.Group("/warehouses")
	warehouses.GET("", m.controller.ListWarehouses)
	warehouses.POST("", m.controller.CreateWarehouse)
	warehouses.GET("/:id", m.controller.GetWarehouse)
	warehouses.PUT("/:id", m.controller.UpdateWarehouse)
	warehouses.DELETE("/:id", m.controller.DeleteWarehouse)

	routes.GET("/stocks", m.controller.ListStocks)

	movements := routes.Group("/stock-movements")
	movements.GET("", m.controller.ListStockMovements)
	movements.POST("", m.controller.RecordStockMovement)
	movements.GET("/:id", m.controller.GetStockMovement)
}
