package project

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

type dailyReportRepository struct {
	fakeRepository
	created     *DailyReport
	createErr   error
	recentSince time.Time
}

func (f *dailyReportRepository) CreateDailyReport(_ context.Context, report *DailyReport) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created = report
	return nil
}

func (f *dailyReportRepository) ListDailyReports(context.Context, DailyReportFilter) ([]DailyReport, int64, error) {
	return []DailyReport{{WorkDone: "Pengecoran kolom lantai 2"}}, 1, nil
}

func (f *dailyReportRepository) CountRecentDailyReports(_ context.Context, _ uuid.UUID, since time.Time) (int64, int64, error) {
	f.recentSince = since
	return 26, 4, nil
}

func dailyService(repository *dailyReportRepository, today time.Time) *service {
	evaluator := NewService(repository).(*service)
	evaluator.now = func() time.Time { return today }
	return evaluator
}

func dailyRequest(day time.Time) DailyReportRequest {
	workers := 42
	return DailyReportRequest{
		ReportDate:       &day,
		WeatherMorning:   "sunny",
		WeatherAfternoon: "rain",
		Workers:          &workers,
		Equipment:        "  1 excavator, 2 dump truck  ",
		WorkDone:         "  Pengecoran kolom lantai 2  ",
		Issues:           "",
	}
}

func TestDailyReportIsStoredForTheDayWithTrimmedTextAndItsAuthor(t *testing.T) {
	repository := &dailyReportRepository{fakeRepository: fakeRepository{project: Project{Status: statusOngoing}}}
	author := uuid.New()
	reportedAt := time.Date(2026, 10, 16, 15, 30, 0, 0, time.UTC)

	report, err := dailyService(repository, time.Date(2026, 10, 16, 18, 0, 0, 0, time.UTC)).
		CreateDailyReport(context.Background(), uuid.New(), author, dailyRequest(reportedAt))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !report.ReportDate.Equal(date(2026, 10, 16)) {
		t.Fatalf("report date = %s, want the day only", report.ReportDate)
	}
	if report.WorkDone != "Pengecoran kolom lantai 2" || report.Equipment != "1 excavator, 2 dump truck" {
		t.Fatalf("text not trimmed: %q / %q", report.WorkDone, report.Equipment)
	}
	if report.Workers != 42 || report.WeatherAfternoon != "rain" {
		t.Fatalf("got %+v", report)
	}
	if repository.created == nil || repository.created.CreatedBy == nil || *repository.created.CreatedBy != author {
		t.Fatal("the author of the report was not stored")
	}
}

func TestDailyReportCannotBeWrittenForADayThatHasNotHappened(t *testing.T) {
	repository := &dailyReportRepository{fakeRepository: fakeRepository{project: Project{Status: statusOngoing}}}
	today := time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)

	_, err := dailyService(repository, today).
		CreateDailyReport(context.Background(), uuid.New(), uuid.New(), dailyRequest(date(2026, 10, 17)))

	var appError *apperror.Error
	if !errors.As(err, &appError) || appError.Code != "daily_report_in_future" {
		t.Fatalf("err = %v, want daily_report_in_future", err)
	}
	if repository.created != nil {
		t.Fatal("a future report was stored")
	}

	if _, err := dailyService(repository, today).
		CreateDailyReport(context.Background(), uuid.New(), uuid.New(), dailyRequest(date(2026, 10, 16))); err != nil {
		t.Fatalf("a report for today was rejected: %v", err)
	}
}

func TestSecondDailyReportForTheSameDayIsAConflict(t *testing.T) {
	repository := &dailyReportRepository{
		fakeRepository: fakeRepository{project: Project{Status: statusOngoing}},
		createErr:      database.ErrDuplicate,
	}

	_, err := dailyService(repository, time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)).
		CreateDailyReport(context.Background(), uuid.New(), uuid.New(), dailyRequest(date(2026, 10, 15)))

	var appError *apperror.Error
	if !errors.As(err, &appError) || appError.Code != "daily_report_date_used" || appError.Status != 409 {
		t.Fatalf("err = %v, want 409 daily_report_date_used", err)
	}
}

func TestClosedProjectTakesNoMoreDailyReports(t *testing.T) {
	repository := &dailyReportRepository{fakeRepository: fakeRepository{project: Project{Status: statusCompleted}}}

	_, err := dailyService(repository, time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)).
		CreateDailyReport(context.Background(), uuid.New(), uuid.New(), dailyRequest(date(2026, 10, 16)))

	if !errors.Is(err, errProjectClosed) {
		t.Fatalf("err = %v, want errProjectClosed", err)
	}
}

func TestDailyReportListCountsRainDaysInTheLastThirtyDays(t *testing.T) {
	repository := &dailyReportRepository{fakeRepository: fakeRepository{project: Project{Status: statusOngoing}}}

	page, err := dailyService(repository, time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)).
		ListDailyReports(context.Background(), uuid.New(), DailyReportQuery{Query: pagination.Query{Page: 1}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !repository.recentSince.Equal(date(2026, 9, 17)) {
		t.Fatalf("window starts %s, want 2026-09-17 so today is the 30th day", repository.recentSince.Format(time.DateOnly))
	}
	if page.ReportsLast30 != 26 || page.RainDaysLast30 != 4 || page.TotalItems != 1 {
		t.Fatalf("got %+v", page)
	}
}
