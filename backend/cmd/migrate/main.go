package main

import (
	"log/slog"
	"os"

	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/auth"
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

	return migrate(db)
}

func migrate(db *gorm.DB) error {
	return db.AutoMigrate(registeredEntities()...)
}

func registeredEntities() []any {
	slices := [][]any{
		auth.Entities(),
	}

	entities := make([]any, 0)
	for _, slice := range slices {
		entities = append(entities, slice...)
	}

	return entities
}
