package scoring

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
	routes := router.Group("/scoring", middleware.Authentication(m.tokens))

	dimensions := routes.Group("/dimensions")
	dimensions.GET("", m.controller.ListDimensions)
	dimensions.POST("", m.controller.CreateDimension)
	dimensions.GET("/:id", m.controller.GetDimension)
	dimensions.PUT("/:id", m.controller.UpdateDimension)
	dimensions.DELETE("/:id", m.controller.DeleteDimension)

	profiles := routes.Group("/building-profiles")
	profiles.GET("", m.controller.ListBuildingProfiles)
	profiles.POST("", m.controller.CreateBuildingProfile)
	profiles.GET("/:id", m.controller.GetBuildingProfile)
	profiles.PUT("/:id", m.controller.UpdateBuildingProfile)
	profiles.DELETE("/:id", m.controller.DeleteBuildingProfile)
	profiles.GET("/:id/weights", m.controller.ListWeights)
	profiles.PUT("/:id/weights", m.controller.ReplaceWeights)
}
