package main

import (
	"log/slog"
	"os"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/migration"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/clock"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

func main() {
	if err := run(); err != nil {
		slog.Error("migration failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("migration finished")
}

func run() error {
	if _, err := clock.UseAppZone(); err != nil {
		return err
	}
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

	return migration.Apply(db)
}
