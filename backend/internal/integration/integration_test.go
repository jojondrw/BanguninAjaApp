//go:build integration

package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/auth"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/master"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/migration"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/project"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/regulation"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/geo"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
)

func newUnit(t *testing.T) uuid.UUID {
	t.Helper()
	unit := master.UnitOfMeasure{Code: uniqueCode("U"), Name: "Satuan uji"}
	if err := db.Create(&unit).Error; err != nil {
		t.Fatalf("create unit: %v", err)
	}
	return unit.ID
}

func newUser(t *testing.T) uuid.UUID {
	t.Helper()
	id := uuid.New()
	err := db.Exec(
		"INSERT INTO users (id, name, email, password_hash, active) VALUES (?, ?, ?, ?, true)",
		id, "Pengawas uji", uniqueCode("pengawas")+"@integrasi.test", "bukan-hash",
	).Error
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return id
}

func appCode(err error) string {
	var appError *apperror.Error
	if errors.As(err, &appError) {
		return appError.Code
	}
	return ""
}

func TestMigrationsCanRunAgainOnAnUpToDateDatabase(t *testing.T) {
	if err := migration.Apply(db); err != nil {
		t.Fatalf("second migration run: %v", err)
	}

	for _, name := range []string{"fk_budget_item_phase_id", "uq_daily_report_project_day", "chk_daily_report_weather", "uq_project_progress_project_day"} {
		if count(t, "SELECT COUNT(*) FROM pg_constraint WHERE conname = ?", name) != 1 {
			t.Errorf("constraint %s is missing after migration", name)
		}
	}
}

func TestWeightedProgressIsStoredAndDeletingAPhaseUnlinksItsBudgetItems(t *testing.T) {
	ctx := context.Background()
	projects := project.NewService(project.NewRepository(db))
	created := newProject(t, projects)
	unit := newUnit(t)

	preparation, err := projects.CreatePhase(ctx, created.ID, project.PhaseRequest{
		Name: "Persiapan", SortOrder: 1, StartDate: date(2026, 1, 1), TargetEndDate: date(2026, 1, 31), Progress: 100,
	})
	if err != nil {
		t.Fatalf("create phase: %v", err)
	}
	structure, err := projects.CreatePhase(ctx, created.ID, project.PhaseRequest{
		Name: "Struktur", SortOrder: 2, StartDate: date(2026, 2, 1), TargetEndDate: date(2026, 12, 31), Progress: 0,
	})
	if err != nil {
		t.Fatalf("create phase: %v", err)
	}
	for _, item := range []struct {
		phase uuid.UUID
		code  string
		price int64
	}{{preparation.ID, "1.1", 1_000_000}, {structure.ID, "2.1", 9_000_000}} {
		phaseID := item.phase
		if _, err := projects.CreateBudgetItem(ctx, created.ID, project.BudgetItemRequest{
			PhaseID: &phaseID, Code: item.code, Description: "Item " + item.code, Volume: 10, UnitOfMeasureID: unit, UnitPrice: item.price,
		}); err != nil {
			t.Fatalf("create budget item %s: %v", item.code, err)
		}
	}

	stored, err := projects.GetProject(ctx, created.ID)
	if err != nil {
		t.Fatalf("get project: %v", err)
	}
	if stored.Progress != 10 {
		t.Fatalf("progress = %d, want 10 (preparation is 10%% of the RAB)", stored.Progress)
	}
	if count(t, "SELECT COUNT(*) FROM project_progress WHERE project_id = ?", created.ID) != 1 {
		t.Fatal("today's progress was not recorded exactly once")
	}

	if err := projects.DeletePhase(ctx, created.ID, structure.ID); err != nil {
		t.Fatalf("delete phase: %v", err)
	}
	if count(t, "SELECT COUNT(*) FROM budget_item WHERE project_id = ? AND phase_id IS NULL", created.ID) != 1 {
		t.Fatal("deleting a phase must keep its budget items and clear their phase")
	}
	after, _ := projects.GetProject(ctx, created.ID)
	if after.Progress != 100 {
		t.Fatalf("progress after deleting the open phase = %d, want 100", after.Progress)
	}
}

func TestDailyReportForADayThatAlreadyHasOneIsRejectedByTheDatabase(t *testing.T) {
	ctx := context.Background()
	projects := project.NewService(project.NewRepository(db))
	created := newProject(t, projects)
	workers := 25
	day := time.Now().AddDate(0, 0, -1)
	request := project.DailyReportRequest{
		ReportDate: &day, WeatherMorning: "sunny", WeatherAfternoon: "rain", Workers: &workers, WorkDone: "Galian pondasi",
	}

	author := newUser(t)
	if _, err := projects.CreateDailyReport(ctx, created.ID, author, request); err != nil {
		t.Fatalf("first report: %v", err)
	}
	_, err := projects.CreateDailyReport(ctx, created.ID, author, request)

	if appCode(err) != "daily_report_date_used" {
		t.Fatalf("second report for the same day: %v, want daily_report_date_used", err)
	}
	if count(t, "SELECT COUNT(*) FROM daily_report WHERE project_id = ?", created.ID) != 1 {
		t.Fatal("the duplicate report was stored")
	}
}

func TestDeletingAProjectRemovesItsDailyReportsAndProgressHistory(t *testing.T) {
	ctx := context.Background()
	projects := project.NewService(project.NewRepository(db))
	created := newProject(t, projects)
	workers := 10
	day := time.Now().AddDate(0, 0, -2)
	if _, err := projects.CreatePhase(ctx, created.ID, project.PhaseRequest{Name: "Persiapan", SortOrder: 1, Progress: 40}); err != nil {
		t.Fatalf("create phase: %v", err)
	}
	if _, err := projects.CreateDailyReport(ctx, created.ID, newUser(t), project.DailyReportRequest{
		ReportDate: &day, WeatherMorning: "sunny", WeatherAfternoon: "sunny", Workers: &workers, WorkDone: "Pagar proyek",
	}); err != nil {
		t.Fatalf("create report: %v", err)
	}

	if err := projects.DeleteProject(ctx, created.ID); err != nil {
		t.Fatalf("delete project: %v", err)
	}

	if count(t, "SELECT COUNT(*) FROM daily_report WHERE project_id = ?", created.ID) != 0 ||
		count(t, "SELECT COUNT(*) FROM project_progress WHERE project_id = ?", created.ID) != 0 {
		t.Fatal("reports or progress history outlived their project")
	}
}

func TestRegulationLookupUsesTheNearestZoneWithinOneHundredFiftyMetres(t *testing.T) {
	near := regulation.Regulation{
		RTRID: uniqueCode("R"), RegionCode: "3273", ZoneName: "Perumahan Kepadatan Sedang", District: "Coblong",
		Point: geo.Point{Lon: 107.6100, Lat: -6.8900}, KDB: 60, KLB: 1.8, KDH: 20,
	}
	farther := regulation.Regulation{
		RTRID: uniqueCode("R"), RegionCode: "3273", ZoneName: "Perdagangan dan Jasa", District: "Coblong",
		Point: geo.Point{Lon: 107.6109, Lat: -6.8900}, KDB: 80, KLB: 3.2, KDH: 10,
	}
	if err := db.Create(&near).Error; err != nil {
		t.Fatalf("insert regulation: %v", err)
	}
	if err := db.Create(&farther).Error; err != nil {
		t.Fatalf("insert regulation: %v", err)
	}
	lookup := regulation.NewService(regulation.NewRepository(db))

	found, err := lookup.GetByPoint(context.Background(), -6.89001, 107.61001)
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if found.IsSimulated || found.ZoneName != "Perumahan Kepadatan Sedang" {
		t.Fatalf("got %+v, want the zone about 1 m away rather than the one about 100 m away", found)
	}

	outside, err := lookup.GetByPoint(context.Background(), -6.8900, 107.6200)
	if err != nil {
		t.Fatalf("lookup outside: %v", err)
	}
	if !outside.IsSimulated {
		t.Fatalf("a point 1 km from any zone got %+v, want the simulated fallback", outside)
	}
}

type ignoreForget struct{}

func (ignoreForget) Forget(uuid.UUID) {}

func TestReusingARefreshTokenEndsEverySessionOfThatUser(t *testing.T) {
	ctx := context.Background()
	tokens := token.NewManager(config.Token{
		AccessSecret: "rahasia-integrasi-yang-cukup-panjang-untuk-hs256", AccessTTL: 15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour, Issuer: "banguninaja-integrasi",
	})
	accounts := auth.NewService(auth.NewRepository(db), tokens, ignoreForget{})
	email := uniqueCode("uji") + "@integrasi.test"
	if _, err := accounts.Register(ctx, auth.RegisterRequest{Name: "Uji Integrasi", Email: email, Password: "rahasia-123"}); err != nil {
		t.Fatalf("register: %v", err)
	}

	first, err := accounts.Login(ctx, auth.LoginRequest{Email: email, Password: "rahasia-123"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	second, err := accounts.Refresh(ctx, first.RefreshToken)
	if err != nil {
		t.Fatalf("first refresh: %v", err)
	}

	if _, err := accounts.Refresh(ctx, first.RefreshToken); err == nil {
		t.Fatal("an already rotated refresh token was accepted again")
	}
	if _, err := accounts.Refresh(ctx, second.RefreshToken); err == nil {
		t.Fatal("after token reuse, the newest session must also be revoked")
	}
	if count(t, "SELECT COUNT(*) FROM refresh_tokens r JOIN users u ON u.id = r.user_id WHERE u.email = ? AND r.revoked_at IS NULL", email) != 0 {
		t.Fatal("some refresh tokens of the user are still active")
	}
}
