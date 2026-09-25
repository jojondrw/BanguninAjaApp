package news

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
		controller: NewController(NewService(NewClient())),
		tokens:     tokens,
	}
}

func (m *Module) RegisterRoutes(router gin.IRouter) {
	news := router.Group("/news", middleware.Authentication(m.tokens))
	news.GET("", m.controller.FetchNews)
}