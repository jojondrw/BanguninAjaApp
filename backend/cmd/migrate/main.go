package main

import (
	"log/slog"
	"os"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/aset"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/auth"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/keuangan"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/laporan"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/lokasi"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/master"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/pengadaan"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/penilaian"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/penjualan"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/persediaan"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/proyek"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/sdm"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/tagihan"
)

type slice struct {
	nama        string
	entities    []any
	indexes     []string
	constraints []string
}

func registeredSlices() []slice {
	return []slice{
		{"master", master.Entities(), master.Indexes(), master.Constraints()},
		{"auth", auth.Entities(), auth.Indexes(), auth.Constraints()},
		{"proyek", proyek.Entities(), proyek.Indexes(), proyek.Constraints()},
		{"persediaan", persediaan.Entities(), persediaan.Indexes(), persediaan.Constraints()},
		{"pengadaan", pengadaan.Entities(), pengadaan.Indexes(), pengadaan.Constraints()},
		{"aset", aset.Entities(), aset.Indexes(), aset.Constraints()},
		{"penjualan", penjualan.Entities(), penjualan.Indexes(), penjualan.Constraints()},
		{"keuangan", keuangan.Entities(), keuangan.Indexes(), keuangan.Constraints()},
		{"tagihan", tagihan.Entities(), tagihan.Indexes(), tagihan.Constraints()},
		{"sdm", sdm.Entities(), sdm.Indexes(), sdm.Constraints()},
		{"penilaian", penilaian.Entities(), penilaian.Indexes(), penilaian.Constraints()},
		{"lokasi", lokasi.Entities(), lokasi.Indexes(), lokasi.Constraints()},
		{"laporan", laporan.Entities(), laporan.Indexes(), laporan.Constraints()},
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
			slog.String("slice", current.nama),
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
