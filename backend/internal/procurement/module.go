package procurement

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
	routes := router.Group("/procurement", middleware.Authentication(m.tokens))

	vendors := routes.Group("/vendors")
	vendors.GET("", m.controller.ListVendors)
	vendors.POST("", m.controller.CreateVendor)
	vendors.GET("/:id", m.controller.GetVendor)
	vendors.PUT("/:id", m.controller.UpdateVendor)
	vendors.DELETE("/:id", m.controller.DeleteVendor)

	requests := routes.Group("/purchase-requests")
	requests.GET("", m.controller.ListPurchaseRequests)
	requests.POST("", m.controller.CreatePurchaseRequest)
	requests.GET("/:id", m.controller.GetPurchaseRequest)
	requests.PUT("/:id", m.controller.UpdatePurchaseRequest)
	requests.PATCH("/:id/status", m.controller.UpdatePurchaseRequestStatus)
	requests.DELETE("/:id", m.controller.DeletePurchaseRequest)

	orders := routes.Group("/purchase-orders")
	orders.GET("", m.controller.ListPurchaseOrders)
	orders.POST("", m.controller.CreatePurchaseOrder)
	orders.GET("/:id", m.controller.GetPurchaseOrder)
	orders.PUT("/:id", m.controller.UpdatePurchaseOrder)
	orders.PATCH("/:id/status", m.controller.UpdatePurchaseOrderStatus)
	orders.DELETE("/:id", m.controller.DeletePurchaseOrder)

	receipts := routes.Group("/goods-receipts")
	receipts.GET("", m.controller.ListGoodsReceipts)
	receipts.POST("", m.controller.RecordGoodsReceipt)
	receipts.GET("/:id", m.controller.GetGoodsReceipt)
}
