package main

import (
	"log/slog"
	"os"

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
	"github.com/jojondrw/BanguninAjaApp/backend/internal/reporting"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/sales"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/scoring"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
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
		{"reporting", reporting.Entities(), reporting.Indexes(), reporting.Constraints()},
	}
}

func main() {
	if err := run(); err != nil {
		slog.Error("migration failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("migration finished")
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := database.Open(cfg)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := database.Close(db); closeErr != nil {
			slog.Error("close database failed", slog.String("error", closeErr.Error()))
		}
	}()

	if err := database.EnsureExtensions(db); err != nil {
		return err
	}

	installed, err := database.InstalledExtensions(db)
	if err != nil {
		return err
	}
	slog.Info("extensions ready", slog.Any("installed", installed))

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

	return nil
}
