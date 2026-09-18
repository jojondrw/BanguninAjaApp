package auth

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/cookie"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/middleware"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
)

type Module struct {
	controller *Controller
	tokens     *token.Manager
}

func NewModule(db *gorm.DB, tokens *token.Manager, refreshCookie cookie.RefreshWriter) *Module {
	repository := NewRepository(db)
	service := NewService(repository, tokens)

	return &Module{
		controller: NewController(service, refreshCookie),
		tokens:     tokens,
	}
}

func (m *Module) RegisterRoutes(router gin.IRouter) {
	routes := router.Group("/auth")
	routes.POST("/register", m.controller.Register)
	routes.POST("/login", m.controller.Login)
	routes.POST("/refresh", m.controller.Refresh)
	routes.POST("/logout", m.controller.Logout)
	routes.GET("/me", middleware.Authentication(m.tokens), m.controller.Profile)
}
