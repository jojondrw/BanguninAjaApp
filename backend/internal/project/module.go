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
	projects.GET("/:id", m.controller.GetProject)
	projects.PUT("/:id", m.controller.UpdateProject)
	projects.PATCH("/:id/status", m.controller.UpdateProjectStatus)
	projects.DELETE("/:id", m.controller.DeleteProject)

	projects.GET("/:id/phases", m.controller.ListPhases)
	projects.POST("/:id/phases", m.controller.CreatePhase)
	projects.PUT("/:id/phases/:childId", m.controller.UpdatePhase)
	projects.DELETE("/:id/phases/:childId", m.controller.DeletePhase)

	projects.GET("/:id/progress-curve", m.controller.ProgressCurve)
	projects.GET("/:id/budget-items", m.controller.ListBudgetItems)
	projects.POST("/:id/budget-items", m.controller.CreateBudgetItem)
	projects.PUT("/:id/budget-items/:childId", m.controller.UpdateBudgetItem)
	projects.DELETE("/:id/budget-items/:childId", m.controller.DeleteBudgetItem)

	projects.GET("/:id/daily-reports", m.controller.ListDailyReports)
	projects.POST("/:id/daily-reports", m.controller.CreateDailyReport)
	projects.PUT("/:id/daily-reports/:childId", m.controller.UpdateDailyReport)
	projects.DELETE("/:id/daily-reports/:childId", m.controller.DeleteDailyReport)

	projects.GET("/:id/payment-schedule", m.controller.PaymentSchedule)
	projects.PUT("/:id/construction-contract", m.controller.SaveContract)
	projects.POST("/:id/payment-terms", m.controller.CreatePaymentTerm)
	projects.DELETE("/:id/payment-terms/:childId", m.controller.DeletePaymentTerm)
	projects.POST("/:id/payment-terms/:childId/claim", m.controller.ClaimPaymentTerm)
	projects.POST("/:id/retention-release", m.controller.ReleaseRetention)

	projects.GET("/:id/permits", m.controller.ListPermits)
	projects.POST("/:id/permits", m.controller.CreatePermit)
	projects.PUT("/:id/permits/:childId", m.controller.UpdatePermit)
	projects.DELETE("/:id/permits/:childId", m.controller.DeletePermit)

	projects.GET("/:id/members", m.controller.ListMembers)
	projects.POST("/:id/members", m.controller.AddMember)
	projects.PATCH("/:id/members/:childId", m.controller.UpdateMember)
	projects.DELETE("/:id/members/:childId", m.controller.RemoveMember)
	projects.GET("/:id/member-candidates", m.controller.ListMemberCandidates)
}
