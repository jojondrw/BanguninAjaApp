package evidence

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/billing"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/project"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/ethereum"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/middleware"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
)

type Module struct {
	service    *Service
	controller *Controller
	projects   *project.Controller
	tokens     *token.Manager
}

type FactSource interface {
	Facts(ctx context.Context, projectID *uuid.UUID) ([]Fact, error)
}

func NewModule(db *gorm.DB, tokens *token.Manager, onchain config.Onchain, extra ...FactSource) *Module {
	sources := feedSources{projects: project.NewEvidenceFeed(db), payments: billing.NewPaymentFeed(db), extra: extra}
	service := NewService(NewRepository(db), sources, ethereum.NewClient(onchain.RPCURL), onchain)
	return &Module{
		service:    service,
		controller: NewController(service),
		projects:   project.NewController(project.NewService(project.NewRepository(db, nil))),
		tokens:     tokens,
	}
}

type feedSources struct {
	projects project.EvidenceFeed
	payments billing.PaymentFeed
	extra    []FactSource
}

func (s feedSources) ProjectFacts(ctx context.Context, projectID *uuid.UUID) ([]Fact, error) {
	found, err := s.projects.Facts(ctx, projectID)
	if err != nil {
		return nil, err
	}
	facts := make([]Fact, len(found))
	for index, fact := range found {
		owner := fact.ProjectID
		facts[index] = Fact{Kind: fact.Kind, SubjectID: fact.SubjectID, ProjectID: &owner, Date: fact.Date, Label: fact.Label, Fields: fact.Fields}
	}
	for _, source := range s.extra {
		more, err := source.Facts(ctx, projectID)
		if err != nil {
			return nil, err
		}
		facts = append(facts, more...)
	}
	return facts, nil
}

func (s feedSources) PaymentFacts(ctx context.Context, projectID *uuid.UUID) ([]Fact, error) {
	found, err := s.payments.Payments(ctx, projectID)
	if err != nil {
		return nil, err
	}
	facts := make([]Fact, len(found))
	for index, payment := range found {
		facts[index] = Fact{
			Kind: KindPayment, SubjectID: payment.ID, ProjectID: payment.ProjectID, Date: payment.PaidAt, Label: payment.Parent,
			Fields: map[string]any{
				"target": payment.Target, "parent": payment.Parent, "amount": payment.Amount,
				"paidAt": payment.PaidAt.Format("2006-01-02"), "method": payment.Method, "reference": payment.Reference,
			},
		}
	}
	return facts, nil
}

func (s feedSources) OwnerClaim(ctx context.Context, projectID, termID uuid.UUID) (Claim, error) {
	claim, err := s.projects.OwnerClaim(ctx, projectID, termID)
	if err != nil {
		return Claim{}, err
	}
	return Claim{
		TermID: claim.TermID, ProjectID: claim.ProjectID, ProjectCode: claim.ProjectCode, ProjectName: claim.ProjectName,
		ContractNumber: claim.ContractNumber, OwnerName: claim.OwnerName, Label: claim.Label, Progress: claim.Progress,
		GrossAmount: claim.GrossAmount, RetentionAmount: claim.RetentionAmount, NetAmount: claim.NetAmount, ClaimedOn: claim.ClaimedOn,
	}, nil
}

func (m *Module) Service() *Service {
	return m.service
}

func (m *Module) RegisterRoutes(router gin.IRouter) {
	routes := router.Group("/evidence", middleware.Authentication(m.tokens))
	routes.GET("/batches", m.controller.ListBatches)
	routes.POST("/batches", m.controller.CreateBatch)
	routes.POST("/batches/:id/anchor", m.controller.Anchor)

	projects := routes.Group("/projects/:id", m.projects.RequireProjectAccess)
	projects.GET("", m.controller.ProjectEvidence)
	projects.PUT("/approver", m.controller.SetApprover)
	projects.POST("/claims/:childId", m.controller.CreateClaimLink)

	public := router.Group("/verify")
	public.GET("/claims/:token", m.controller.PublicClaim)
	public.POST("/claims/:token/approve", m.controller.Approve)
}
