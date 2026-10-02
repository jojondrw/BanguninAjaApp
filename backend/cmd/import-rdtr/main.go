package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/regulation/importer"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

const rdtrDirEnv = "RDTR_DIR"

var errMissingDir = errors.New("set -dir or " + rdtrDirEnv + " to the folder holding rdtr_*.json")

func main() {
	if err := run(); err != nil {
		slog.Error("RDTR import failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("RDTR import finished")
}

func run() error {
	dir := flag.String(
		"dir",
		os.Getenv(rdtrDirEnv),
		"directory containing rdtr_*.json files (defaults to $"+rdtrDirEnv+")",
	)

	file := flag.String(
		"file",
		"",
		"optional RDTR JSON filename; if empty, import every rdtr_*.json in the directory",
	)

	flag.Parse()

	if *dir == "" {
		return errMissingDir
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

	return importer.Import(context.Background(), db, *dir, *file)
}
