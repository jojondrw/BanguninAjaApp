package project

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
	projects := router.Group("/projects", middleware.Authentication(m.tokens))
	projects.GET("", m.controller.ListProjects)
	projects.POST("", m.controller.CreateProject)
	projects.GET("/:id", m.controller.GetProject)
	projects.PUT("/:id", m.controller.UpdateProject)
	projects.PATCH("/:id/status", m.controller.UpdateProjectStatus)
	projects.DELETE("/:id", m.controller.DeleteProject)

	projects.GET("/:id/phases", m.controller.ListPhases)
	projects.POST("/:id/phases", m.controller.CreatePhase)
	projects.PUT("/:id/phases/:childId", m.controller.UpdatePhase)
	projects.DELETE("/:id/phases/:childId", m.controller.DeletePhase)

	projects.GET("/:id/budget-items", m.controller.ListBudgetItems)
	projects.POST("/:id/budget-items", m.controller.CreateBudgetItem)
	projects.PUT("/:id/budget-items/:childId", m.controller.UpdateBudgetItem)
	projects.DELETE("/:id/budget-items/:childId", m.controller.DeleteBudgetItem)

	projects.GET("/:id/permits", m.controller.ListPermits)
	projects.POST("/:id/permits", m.controller.CreatePermit)
	projects.PUT("/:id/permits/:childId", m.controller.UpdatePermit)
	projects.DELETE("/:id/permits/:childId", m.controller.DeletePermit)
}
