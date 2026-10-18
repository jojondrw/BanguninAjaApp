package wageproof

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/evidence"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/hr"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/project"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/middleware"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
)

type Module struct {
	controller *Controller
	projects   *project.Controller
	tokens     *token.Manager
}

func NewModule(db *gorm.DB, tokens *token.Manager, anchors *evidence.Service) *Module {
	source := wageSource{wages: hr.NewWageFeed(db), places: project.NewEvidenceFeed(db)}
	return &Module{
		controller: NewController(NewService(NewRepository(db), source, evidenceAnchors{service: anchors})),
		projects:   project.NewController(project.NewService(project.NewRepository(db, nil))),
		tokens:     tokens,
	}
}

type wageSource struct {
	wages  hr.WageFeed
	places project.EvidenceFeed
}

func (s wageSource) PaidWages(ctx context.Context, projectID uuid.UUID, period string) ([]WageRowInput, error) {
	rows, err := s.wages.PaidWages(ctx, projectID, period)
	if err != nil {
		return nil, err
	}
	inputs := make([]WageRowInput, len(rows))
	for index, row := range rows {
		inputs[index] = WageRowInput{EmployeeID: row.EmployeeID, Name: row.Name, Gross: row.Gross, Days: row.Days}
	}
	return inputs, nil
}

func (s wageSource) Place(ctx context.Context, projectID uuid.UUID) (Place, error) {
	place, err := s.places.Place(ctx, projectID)
	if err != nil {
		return Place{}, err
	}
	return Place{Code: place.Code, Name: place.Name, RegionCode: place.RegionCode, RegionName: place.RegionName}, nil
}

type evidenceAnchors struct {
	service *evidence.Service
}

func (a evidenceAnchors) Status(ctx context.Context, projectID uuid.UUID, kind string, subjectID uuid.UUID) (AnchorStatus, error) {
	if a.service == nil {
		return AnchorStatus{Status: evidence.StatusPending}, nil
	}
	item, err := a.service.SubjectStatus(ctx, projectID, kind, subjectID)
	if err != nil {
		return AnchorStatus{}, err
	}
	return AnchorStatus{Status: item.Status, AnchoredAt: item.AnchoredAt, TxHash: item.TxHash, TxURL: item.TxURL}, nil
}

type rootFacts struct {
	feed RootFeed
}

func NewEvidenceSource(db *gorm.DB) evidence.FactSource {
	return rootFacts{feed: NewRootFeed(NewRepository(db))}
}

func (r rootFacts) Facts(ctx context.Context, projectID *uuid.UUID) ([]evidence.Fact, error) {
	roots, err := r.feed.Roots(ctx, projectID)
	if err != nil {
		return nil, err
	}
	facts := make([]evidence.Fact, len(roots))
	for index, root := range roots {
		owner := root.ProjectID
		facts[index] = evidence.Fact{Kind: KindWageRoot, SubjectID: root.ProofID, ProjectID: &owner, Date: root.Date, Label: root.Label, Fields: root.Fields}
	}
	return facts, nil
}

func (m *Module) RegisterRoutes(router gin.IRouter) {
	routes := router.Group("/evidence/projects/:id/wage-proofs", middleware.Authentication(m.tokens), m.projects.RequireProjectAccess)
	routes.GET("", m.controller.List)
	routes.POST("", m.controller.Create)

	public := router.Group("/verify")
	public.GET("/wages/:token", m.controller.PublicProof)
	public.GET("/receipts/:token", m.controller.PublicReceipt)
}
