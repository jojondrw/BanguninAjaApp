package main

import (
	"log/slog"
	"os"

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

	if err := database.EnsureExtensions(db); err != nil {
		return err
	}

	installed, err := database.InstalledExtensions(db)
	if err != nil {
		return err
	}
	slog.Info("extensions ready", slog.Any("installed", installed))

	if err := db.AutoMigrate(registeredEntities()...); err != nil {
		return err
	}

	return database.ApplyIndexes(db, registeredIndexes())
}

func registeredEntities() []any {
	entities := make([]any, 0)
	for _, slice := range [][]any{
		auth.Entities(),
	} {
		entities = append(entities, slice...)
	}

	return entities
}

func registeredIndexes() []string {
	statements := make([]string, 0)
	for _, slice := range [][]string{
		auth.Indexes(),
	} {
		statements = append(statements, slice...)
	}

	return statements
}
