package procurement

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/inventory"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/middleware"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
)

type Module struct {
	controller *Controller
	tokens     *token.Manager
}

func NewModule(db *gorm.DB, tokens *token.Manager) *Module {
	return &Module{
		controller: NewController(NewService(NewRepository(db, newInventoryLedger))),
		tokens:     tokens,
	}
}

type inventoryLedger struct {
	ledger inventory.StockLedger
}

func newInventoryLedger(db *gorm.DB) StockLedger {
	return inventoryLedger{ledger: inventory.NewStockLedger(inventory.NewRepository(db))}
}

func (l inventoryLedger) RecordIncoming(ctx context.Context, stock IncomingStock) (uuid.UUID, error) {
	return l.ledger.RecordIncoming(ctx, inventory.IncomingStock{
		Date:        stock.Date,
		MaterialID:  stock.MaterialID,
		WarehouseID: stock.WarehouseID,
		Quantity:    stock.Quantity,
		Reference:   stock.Reference,
	})
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
