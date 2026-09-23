package asset

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
	authenticated := middleware.Authentication(m.tokens)

	assets := router.Group("/assets", authenticated)
	assets.GET("", m.controller.ListAssets)
	assets.POST("", m.controller.CreateAsset)
	assets.GET("/:id", m.controller.GetAsset)
	assets.PUT("/:id", m.controller.UpdateAsset)
	assets.DELETE("/:id", m.controller.DeleteAsset)

	equipment := router.Group("/equipment", authenticated)
	equipment.GET("", m.controller.ListEquipment)
	equipment.POST("", m.controller.CreateEquipment)
	equipment.GET("/:id", m.controller.GetEquipment)
	equipment.PUT("/:id", m.controller.UpdateEquipment)
	equipment.DELETE("/:id", m.controller.DeleteEquipment)
}
