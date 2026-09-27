package site

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/middleware"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
)

type Module struct {
	controller *Controller
	tokens     *token.Manager
}

func NewModule(db *gorm.DB, tokens *token.Manager, scoreConfig config.Score) *Module {
	scores := NewScoreClient(scoreConfig.BaseURL, scoreConfig.Timeout)
	service := NewService(NewRepository(db), scores)

	return &Module{
		controller: NewController(service),
		tokens:     tokens,
	}
}

func (m *Module) RegisterRoutes(router gin.IRouter) {
	routes := router.Group("/site", middleware.Authentication(m.tokens))
	routes.POST("/evaluate", m.controller.Evaluate)
}
