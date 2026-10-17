package migration

import (
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/asset"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/auth"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/billing"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/finance"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/hr"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/inventory"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/location"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/master"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/procurement"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/project"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/regulation"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/reporting"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/sales"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/scoring"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/duedate"
)

type slice struct {
	name        string
	entities    []any
	indexes     []string
	constraints []string
}

func registeredSlices() []slice {
	return []slice{
		{"master", master.Entities(), master.Indexes(), master.Constraints()},
		{"auth", auth.Entities(), auth.Indexes(), auth.Constraints()},
		{"project", project.Entities(), project.Indexes(), project.Constraints()},
		{"inventory", inventory.Entities(), inventory.Indexes(), inventory.Constraints()},
		{"procurement", procurement.Entities(), procurement.Indexes(), procurement.Constraints()},
		{"asset", asset.Entities(), asset.Indexes(), asset.Constraints()},
		{"sales", sales.Entities(), sales.Indexes(), sales.Constraints()},
		{"finance", finance.Entities(), finance.Indexes(), finance.Constraints()},
		{"billing", billing.Entities(), billing.Indexes(), billing.Constraints()},
		{"hr", hr.Entities(), hr.Indexes(), hr.Constraints()},
		{"scoring", scoring.Entities(), scoring.Indexes(), scoring.Constraints()},
		{"location", location.Entities(), location.Indexes(), location.Constraints()},
		{"regulation", regulation.Entities(), regulation.Indexes(), regulation.Constraints()},
		{"reporting", reporting.Entities(), reporting.Indexes(), reporting.Constraints()},
	}
}

func Apply(db *gorm.DB) error {
	if err := database.EnsureExtensions(db); err != nil {
		return err
	}

	installed, err := database.InstalledExtensions(db)
	if err != nil {
		return err
	}
	slog.Info("extensions ready", slog.Any("installed", installed))

	if err := database.ApplyIndexes(db, hr.Repairs()); err != nil {
		return err
	}

	slices := registeredSlices()

	for _, current := range slices {
		if err := db.AutoMigrate(current.entities...); err != nil {
			return err
		}
		slog.Info("tables ready",
			slog.String("slice", current.name),
			slog.Int("entities", len(current.entities)))
	}

	for _, current := range slices {
		if err := database.ApplyIndexes(db, current.indexes); err != nil {
			return err
		}
		if err := database.ApplyIndexes(db, current.constraints); err != nil {
			return err
		}
	}

	return backfill(db)
}

func backfill(db *gorm.DB) error {
	if err := prepareRoles(db); err != nil {
		return err
	}

	posted, err := finance.PostMissingCashJournals(context.Background(), finance.NewRepository(db))
	if err != nil {
		return err
	}
	slog.Info("cash journals ready", slog.Int("posted", posted))

	payments, err := billing.BackfillPayments(context.Background(), billing.NewRepository(db, nil))
	if err != nil {
		return err
	}
	slog.Info("billing payment history ready", slog.Int64("backfilled", payments))

	projects, err := project.RefreshAllProgress(context.Background(), project.NewRepository(db, nil), duedate.Today(time.Now()))
	if err != nil {
		return err
	}
	slog.Info("weighted project progress ready", slog.Int("projects", projects))
	return nil
}

func prepareRoles(db *gorm.DB) error {
	repository := auth.NewRepository(db)

	seeded, err := auth.EnsureRoles(context.Background(), repository)
	if err != nil {
		return err
	}
	slog.Info("roles ready", slog.Int("roles", seeded))

	assigned, err := auth.PromoteLegacyUsersToAdmin(context.Background(), repository)
	if err != nil {
		return err
	}
	slog.Info("users from before roles promoted to admin", slog.Int64("users", assigned))
	return nil
}
