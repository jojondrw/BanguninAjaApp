package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/reporting"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/access"
)

type reportSeed struct {
	name      string
	kind      string
	format    string
	project   string
	monthsAgo int
	periodic  bool
	created   int
}

type reportWriter interface {
	backdateReport(ctx context.Context, id uuid.UUID, createdOn time.Time) error
}

type gormReportWriter struct {
	db *gorm.DB
}

func (w gormReportWriter) backdateReport(ctx context.Context, id uuid.UUID, createdOn time.Time) error {
	err := w.db.WithContext(ctx).Model(&reporting.Report{}).Where("id = ?", id).
		Updates(map[string]any{"created_at": createdOn, "updated_at": createdOn}).Error
	if err != nil {
		return fmt.Errorf("backdate report: %w", err)
	}
	return nil
}

func reportSeedsOf(today time.Time) []reportSeed {
	seeds := []reportSeed{}
	formats := []string{"pdf", "xlsx"}
	for monthsAgo := 5; monthsAgo >= 0; monthsAgo-- {
		label := monthLabelOf(today, monthsAgo)
		seeds = append(seeds, reportSeed{"Arus Kas " + label, "cash_flow", formats[monthsAgo%2], "", monthsAgo, true, -(monthsAgo*30 - 2)})
	}
	for monthsAgo := 3; monthsAgo >= 1; monthsAgo-- {
		label := monthLabelOf(today, monthsAgo)
		seeds = append(seeds,
			reportSeed{"Penggajian " + label, "payroll", "xlsx", "", monthsAgo, true, -(monthsAgo*30 - 4)},
			reportSeed{"Umur Piutang dan Utang " + label, "aging", []string{"pdf", "csv"}[monthsAgo%2], "", monthsAgo, true, -(monthsAgo*30 - 3)},
		)
	}
	return append(seeds,
		reportSeed{"Anggaran dan Realisasi Sky Garden", "budget", "xlsx", projectSkyGarden, 0, false, -6},
		reportSeed{"Anggaran dan Realisasi Griya Asri", "budget", "pdf", projectGriyaAsri, 0, false, -18},
		reportSeed{"Anggaran dan Realisasi Semua Proyek", "budget", "xlsx", "", 0, false, -40},
		reportSeed{"Portofolio Proyek Kuartal Ini", "project", "pdf", "", 0, false, -9},
		reportSeed{"Laporan Proyek RSIA Cirebon", "project", "pdf", projectHospital, 0, false, -21},
		reportSeed{"Stok Material Semua Gudang", "stock", "xlsx", "", 0, false, -1},
		reportSeed{"Stok Material Gudang Pusat", "stock", "csv", "", 0, false, -33},
		reportSeed{"Penjualan Apartemen Sky Garden", "sales", "pdf", projectSkyGarden, 0, false, -5},
		reportSeed{"Penjualan Perumahan Griya Asri", "sales", "xlsx", projectGriyaAsri, 0, false, -26},
		reportSeed{"Penjualan Semua Proyek", "sales", "pdf", "", 0, false, -60},
	)
}

func demoReportNames(today time.Time) []string {
	names := []string{}
	for offset := 0; offset <= resetLookbackMonths; offset++ {
		for _, seed := range reportSeedsOf(monthStart(today, offset)) {
			names = append(names, seed.name)
		}
	}
	return names
}

func (s *seeder) seedReports() error {
	service := reporting.NewService(reporting.NewRepository(s.reportDB))
	for _, seed := range reportSeedsOf(s.today) {
		request := reporting.ReportRequest{Name: seed.name, Type: seed.kind, Format: seed.format}
		if seed.project != "" {
			projectID, err := s.known("project", seed.project)
			if err != nil {
				return err
			}
			request.ProjectID = &projectID
		}
		if seed.periodic {
			start := s.monthDay(seed.monthsAgo, 1)
			end := pastOrToday(monthStart(s.today, seed.monthsAgo-1).AddDate(0, 0, -1), s.today)
			request.PeriodStart = &start
			request.PeriodEnd = &end
		}
		id, created, err := s.ensure("report", map[string]any{"name": seed.name, "created_by": s.owner}, func() (uuid.UUID, error) {
			response, err := service.CreateReport(s.ctx, s.owner, access.RoleAdmin, request)
			return response.ID, err
		})
		if err != nil {
			return err
		}
		if created {
			if err := s.reports.backdateReport(s.ctx, id, pastOrToday(s.day(seed.created), s.today)); err != nil {
				return err
			}
		}
	}
	return nil
}
