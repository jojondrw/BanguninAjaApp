package hr

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
	routes := router.Group("/hr", middleware.Authentication(m.tokens))

	employees := routes.Group("/employees")
	employees.GET("", m.controller.ListEmployees)
	employees.POST("", m.controller.CreateEmployee)
	employees.GET("/summary", m.controller.SummarizeEmployees)
	employees.GET("/:id", m.controller.GetEmployee)
	employees.PUT("/:id", m.controller.UpdateEmployee)
	employees.DELETE("/:id", m.controller.DeleteEmployee)

	attendances := routes.Group("/attendances")
	attendances.GET("", m.controller.ListAttendances)
	attendances.POST("", m.controller.CreateAttendance)
	attendances.GET("/:id", m.controller.GetAttendance)
	attendances.PUT("/:id", m.controller.UpdateAttendance)
	attendances.DELETE("/:id", m.controller.DeleteAttendance)

	payrolls := routes.Group("/payrolls")
	payrolls.GET("", m.controller.ListPayrolls)
	payrolls.POST("", m.controller.CreatePayroll)
	payrolls.GET("/:id", m.controller.GetPayroll)
	payrolls.PUT("/:id", m.controller.UpdatePayroll)
	payrolls.PATCH("/:id/pay", m.controller.PayPayroll)
	payrolls.DELETE("/:id", m.controller.DeletePayroll)
}
