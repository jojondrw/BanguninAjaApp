package reporting

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
	reports := router.Group("/reports", middleware.Authentication(m.tokens))
	reports.GET("", m.controller.ListReports)
	reports.POST("", m.controller.CreateReport)
	reports.GET("/:id", m.controller.GetReport)
	reports.DELETE("/:id", m.controller.DeleteReport)
}
