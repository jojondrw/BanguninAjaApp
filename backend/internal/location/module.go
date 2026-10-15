package location

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/regulation"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/middleware"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
)

type Module struct {
	controller *Controller
	tokens     *token.Manager
}

func NewModule(db *gorm.DB, tokens *token.Manager) *Module {
	repository := NewRepository(db)
	regulations := regulation.NewService(regulation.NewRepository(db))
	return &Module{
		controller: NewController(NewService(repository), NewReportService(repository, regulations)),
		tokens:     tokens,
	}
}

func (m *Module) RegisterRoutes(router gin.IRouter) {
	routes := router.Group("/locations", middleware.Authentication(m.tokens))

	saved := routes.Group("/saved")
	saved.GET("", m.controller.ListSavedLocations)
	saved.POST("", m.controller.CreateSavedLocation)
	saved.GET("/summary", m.controller.SummarizeSavedLocations)
	saved.GET("/:id", m.controller.GetSavedLocation)
	saved.GET("/:id/report", m.controller.DownloadSiteReport)
	saved.PUT("/:id", m.controller.UpdateSavedLocation)
	saved.DELETE("/:id", m.controller.DeleteSavedLocation)

	comparisons := routes.Group("/comparisons")
	comparisons.GET("", m.controller.ListComparisons)
	comparisons.POST("", m.controller.CreateComparison)
	comparisons.GET("/:id", m.controller.GetComparison)
	comparisons.GET("/:id/report", m.controller.DownloadComparisonReport)
	comparisons.PUT("/:id", m.controller.UpdateComparison)
	comparisons.DELETE("/:id", m.controller.DeleteComparison)
}
