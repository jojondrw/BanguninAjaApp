package auth

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/access"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/cookie"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/middleware"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
)

const subjectCacheTTL = 30 * time.Second

type Module struct {
	controller *Controller
	tokens     *token.Manager
	subjects   *access.Directory
}

func NewModule(db *gorm.DB, tokens *token.Manager, refreshCookie cookie.RefreshWriter) *Module {
	repository := NewRepository(db)
	subjects := access.NewDirectory(repository.FindSubject, subjectCacheTTL)
	service := NewService(repository, tokens, subjects)

	return &Module{
		controller: NewController(service, refreshCookie),
		tokens:     tokens,
		subjects:   subjects,
	}
}

func (m *Module) Subjects() *access.Directory {
	return m.subjects
}

func (m *Module) RegisterRoutes(router gin.IRouter) {
	authenticated := middleware.Authentication(m.tokens)

	routes := router.Group("/auth")
	routes.POST("/register", m.controller.Register)
	routes.POST("/login", m.controller.Login)
	routes.POST("/refresh", m.controller.Refresh)
	routes.POST("/logout", m.controller.Logout)
	routes.GET("/me", authenticated, m.controller.Profile)
	routes.PATCH("/me", authenticated, m.controller.UpdateProfile)
	routes.PATCH("/password", authenticated, m.controller.ChangePassword)

	users := router.Group("/users", authenticated)
	users.GET("", m.controller.ListUsers)
	users.GET("/audit", m.controller.ListUserAudit)
	users.PATCH("/:id/role", m.controller.ChangeUserRole)
	users.PATCH("/:id/active", m.controller.ChangeUserActive)
}
