package finance

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httprequest"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httpresponse"
)

const (
	budgetTotalsColumns = "COALESCE(SUM(budget.value), 0) AS budget, COALESCE(SUM(spent.amount), 0) AS realized"
	budgetSpentJoin     = "LEFT JOIN LATERAL (SELECT SUM(cash_transaction.amount) AS amount FROM cash_transaction " +
		"WHERE cash_transaction.project_id = budget.project_id AND cash_transaction.type = ? " +
		"AND cash_transaction.date >= make_date(budget.year::int, 1, 1) AND cash_transaction.date < make_date((budget.year + 1)::int, 1, 1)) AS spent ON TRUE"
)

type BudgetSummaryQuery struct {
	ProjectID *uuid.UUID `form:"projectId,parser=encoding.TextUnmarshaler"`
	Year      int        `form:"year" binding:"omitempty,min=2000,max=2100"`
}

type BudgetSummaryResponse struct {
	Budget   int64 `json:"budget"`
	Realized int64 `json:"realized"`
}

type BudgetTotals struct {
	Budget   int64
	Realized int64
}

type BudgetSummaryRepository interface {
	SummarizeBudgets(ctx context.Context, filter BudgetFilter) (BudgetTotals, error)
}

type BudgetSummaryService interface {
	SummarizeBudgets(ctx context.Context, query BudgetSummaryQuery) (BudgetSummaryResponse, error)
}

type budgetSummaryRepository struct {
	db *gorm.DB
}

func NewBudgetSummaryRepository(db *gorm.DB) BudgetSummaryRepository {
	return budgetSummaryRepository{db: db}
}

func (r budgetSummaryRepository) SummarizeBudgets(ctx context.Context, filter BudgetFilter) (BudgetTotals, error) {
	var totals BudgetTotals
	err := r.db.WithContext(ctx).
		Model(&Budget{}).
		Select(budgetTotalsColumns).
		Joins(budgetSpentJoin, cashOut).
		Scopes(filter.apply).
		Scan(&totals).Error
	return totals, database.Translate(err)
}

type budgetSummaryService struct {
	repository BudgetSummaryRepository
}

func NewBudgetSummaryService(repository BudgetSummaryRepository) BudgetSummaryService {
	return budgetSummaryService{repository: repository}
}

func (s budgetSummaryService) SummarizeBudgets(ctx context.Context, query BudgetSummaryQuery) (BudgetSummaryResponse, error) {
	totals, err := s.repository.SummarizeBudgets(ctx, BudgetFilter{ProjectID: query.ProjectID, Year: query.Year})
	if err != nil {
		return BudgetSummaryResponse{}, apperror.Internal(err)
	}
	return BudgetSummaryResponse(totals), nil
}

type BudgetSummaryController struct {
	service BudgetSummaryService
}

func NewBudgetSummaryController(service BudgetSummaryService) *BudgetSummaryController {
	return &BudgetSummaryController{service: service}
}

func newBudgetSummaryController(db *gorm.DB) *BudgetSummaryController {
	return NewBudgetSummaryController(NewBudgetSummaryService(NewBudgetSummaryRepository(db)))
}

func (c *BudgetSummaryController) SummarizeBudgets(ctx *gin.Context) {
	var query BudgetSummaryQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	summary, err := c.service.SummarizeBudgets(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, summary, err)
}
