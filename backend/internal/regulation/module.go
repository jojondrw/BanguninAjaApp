package regulation

import (
	"github.com/gin-gonic/gin"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/middleware"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
	"gorm.io/gorm"
)

type Module struct {
	controller *Controller
	tokens     *token.Manager
}

func NewModule(db *gorm.DB, tokens *token.Manager) *Module {
	repo := NewRepository(db)
	service := NewService(repo)

	return &Module{
		controller: NewController(service),
		tokens:     tokens,
	}
}

func (m *Module) RegisterRoutes(router gin.IRouter) {
	routes := router.Group("/regulation", middleware.Authentication(m.tokens))
	routes.POST("/lookup", m.controller.Lookup)
}
