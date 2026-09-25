package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/regulation/importer"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

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
		`C:\Users\ACER NITRO V15\Downloads\BanguninAja\data\raw\rdtr`,
		"directory containing RDTR JSON files",
	)

	file := flag.String(
		"file",
		"",
		"optional RDTR JSON filename; if empty, import all 10 files",
	)

	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := database.Open(cfg)
	if err != nil {
		return err
	}
	defer database.Close(db)

	return importer.Import(context.Background(), db, *dir, *file)
}