package main

import (
	"fmt"
	"log/slog"
	"os"

	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

type seedStep struct {
	name string
	run  func(*gorm.DB) error
}

func main() {
	if err := run(); err != nil {
		slog.Error("seed failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("seed finished")
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

	return db.Transaction(seedAll)
}

func seedAll(tx *gorm.DB) error {
	steps := []seedStep{
		{"regions", seedRegions},
		{"units of measure", seedUnits},
		{"accounts", seedAccounts},
		{"scoring", seedScoring},
	}

	for _, step := range steps {
		if err := step.run(tx); err != nil {
			return fmt.Errorf("seed %s: %w", step.name, err)
		}
		slog.Info("seeded", slog.String("step", step.name))
	}

	return nil
}
