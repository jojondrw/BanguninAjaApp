package project

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/billing"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/middleware"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
)

type Module struct {
	controller *Controller
	tokens     *token.Manager
}

func NewModule(db *gorm.DB, tokens *token.Manager) *Module {
	return &Module{
		controller: NewController(NewService(NewRepository(db, NewBillingPayables))),
		tokens:     tokens,
	}
}

type billingPayables struct {
	ledger billing.PayableLedger
}

func NewBillingPayables(db *gorm.DB) PayableLedger {
	return billingPayables{ledger: billing.NewPayableLedger(billing.NewRepository(db))}
}

func (p billingPayables) Record(ctx context.Context, entry PayableEntry) (uuid.UUID, error) {
	return p.ledger.Record(ctx, billing.PayableEntry{
		VendorID:  entry.VendorID,
		ProjectID: entry.ProjectID,
		Reference: entry.Reference,
		DueDate:   entry.DueDate,
		Amount:    entry.Amount,
	})
}

func (m *Module) RegisterRoutes(router gin.IRouter) {
	projects := router.Group("/projects", middleware.Authentication(m.tokens))
	projects.GET("", m.controller.ListProjects)
	projects.POST("", m.controller.CreateProject)
	projects.GET("/summary", m.controller.SummarizeProjects)

	project := projects.Group("/:id", m.controller.RequireProjectAccess)
	project.GET("", m.controller.GetProject)
	project.PUT("", m.controller.UpdateProject)
	project.PATCH("/status", m.controller.UpdateProjectStatus)
	project.DELETE("", m.controller.DeleteProject)

	project.GET("/phases", m.controller.ListPhases)
	project.POST("/phases", m.controller.CreatePhase)
	project.PUT("/phases/:childId", m.controller.UpdatePhase)
	project.DELETE("/phases/:childId", m.controller.DeletePhase)

	project.GET("/progress-curve", m.controller.ProgressCurve)
	project.GET("/budget-items", m.controller.ListBudgetItems)
	project.POST("/budget-items", m.controller.CreateBudgetItem)
	project.PUT("/budget-items/:childId", m.controller.UpdateBudgetItem)
	project.DELETE("/budget-items/:childId", m.controller.DeleteBudgetItem)

	project.GET("/daily-reports", m.controller.ListDailyReports)
	project.POST("/daily-reports", m.controller.CreateDailyReport)
	project.PUT("/daily-reports/:childId", m.controller.UpdateDailyReport)
	project.DELETE("/daily-reports/:childId", m.controller.DeleteDailyReport)

	project.GET("/payment-schedule", m.controller.PaymentSchedule)
	project.PUT("/construction-contract", m.controller.SaveContract)
	project.POST("/payment-terms", m.controller.CreatePaymentTerm)
	project.DELETE("/payment-terms/:childId", m.controller.DeletePaymentTerm)
	project.POST("/payment-terms/:childId/claim", m.controller.ClaimPaymentTerm)
	project.POST("/retention-release", m.controller.ReleaseRetention)
	project.POST("/contract-changes", m.controller.ProposeContractChange)
	project.PATCH("/contract-changes/:childId", m.controller.DecideContractChange)
	project.DELETE("/contract-changes/:childId", m.controller.DeleteContractChange)

	project.GET("/permits", m.controller.ListPermits)
	project.POST("/permits", m.controller.CreatePermit)
	project.PUT("/permits/:childId", m.controller.UpdatePermit)
	project.DELETE("/permits/:childId", m.controller.DeletePermit)

	project.GET("/members", m.controller.ListMembers)
	project.POST("/members", m.controller.AddMember)
	project.PATCH("/members/:childId", m.controller.UpdateMember)
	project.DELETE("/members/:childId", m.controller.RemoveMember)
	project.GET("/member-candidates", m.controller.ListMemberCandidates)
}
