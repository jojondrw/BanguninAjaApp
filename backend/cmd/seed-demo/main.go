package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/asset"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/auth"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/billing"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/finance"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/hr"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/inventory"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/location"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/procurement"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/project"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/sales"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/access"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/clock"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

const (
	demoAccountEmail = "demo@banguninaja.id"
	demoAccountName  = "Tim Demo BanguninAja"
	demoManagerEmail = "manajer@banguninaja.id"
	demoManagerName  = "Rina Wulandari"
)

var errNoOwner = errors.New("no user to own the demo data: register an account first, pass --owner, or set DEMO_USER_PASSWORD to create " + demoAccountEmail)

type options struct {
	reset      bool
	ownerEmail string
}

func main() {
	if err := run(); err != nil {
		slog.Error("seed demo failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Info("seed demo finished")
}

func parseOptions(args []string) (options, error) {
	flags := flag.NewFlagSet("seed-demo", flag.ContinueOnError)
	reset := flags.Bool("reset-demo", false, "hapus semua data demo berawalan DEMO- lalu berhenti")
	owner := flags.String("owner", os.Getenv("DEMO_OWNER_EMAIL"), "email akun pemilik lokasi tersimpan dan permintaan pembelian demo")
	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
	return options{reset: *reset, ownerEmail: strings.ToLower(strings.TrimSpace(*owner))}, nil
}

func run() error {
	if _, err := clock.UseAppZone(); err != nil {
		return err
	}
	opts, err := parseOptions(os.Args[1:])
	if err != nil {
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

	ctx := context.Background()
	db = db.Session(&gorm.Session{Logger: db.Logger.LogMode(logger.Warn)})
	if opts.reset {
		return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			return resetDemo(ctx, tx, todayInJakarta(time.Now()))
		})
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return seedDemo(ctx, tx, cfg, opts)
	})
}

func seedDemo(ctx context.Context, tx *gorm.DB, cfg config.Config, opts options) error {
	finder := gormLookup{db: tx}
	accounts := auth.NewService(auth.NewRepository(tx), nil, noSubjectCache{})
	owner, err := resolveOwner(ctx, finder, accounts, opts.ownerEmail)
	if err != nil {
		return err
	}
	manager, err := resolveManager(ctx, finder, accounts)
	if err != nil {
		return err
	}
	s := &seeder{
		ctx:      ctx,
		lookup:   finder,
		services: newServices(ctx, tx, cfg),
		patcher:  gormLocationPatcher{db: tx},
		history:  gormProgressRecorder{db: tx},
		terms:    gormTermWriter{db: tx},
		today:    todayInJakarta(time.Now()),
		owner:    owner,
		manager:  manager,
		ids:      map[string]uuid.UUID{},
		tally:    newTally(),
	}
	if s.services.sites == nil {
		slog.Warn("score service unreachable, saved locations get sample values marked as contoh")
	}
	return s.run()
}

func newServices(ctx context.Context, tx *gorm.DB, cfg config.Config) services {
	return services{
		projects:    project.NewService(project.NewRepository(tx, project.NewBillingLedger)),
		finance:     finance.NewService(finance.NewRepository(tx)),
		sales:       sales.NewService(sales.NewRepository(tx)),
		inventory:   inventory.NewService(inventory.NewRepository(tx)),
		procurement: procurement.NewService(procurement.NewRepository(tx, newInventoryLedger)),
		billing:     billing.NewService(billing.NewRepository(tx, billing.NewFinanceCashBook)),
		hr:          hr.NewService(hr.NewRepository(tx)),
		assets:      asset.NewService(asset.NewRepository(tx)),
		locations:   location.NewService(location.NewRepository(tx)),
		sites:       newSiteService(ctx, tx, cfg.Score),
	}
}

func resolveOwner(ctx context.Context, finder lookup, accounts auth.Service, email string) (uuid.UUID, error) {
	if email != "" {
		return requireUser(ctx, finder, email)
	}
	id, found, err := finder.find(ctx, "users", map[string]any{"email": demoAccountEmail})
	if err != nil {
		return uuid.Nil, err
	}
	if found {
		return id, makeAdmin(ctx, accounts, id)
	}
	if password := os.Getenv("DEMO_USER_PASSWORD"); password != "" {
		created, err := accounts.Register(ctx, auth.RegisterRequest{Name: demoAccountName, Email: demoAccountEmail, Password: password})
		if err != nil {
			return uuid.Nil, err
		}
		return created.ID, makeAdmin(ctx, accounts, created.ID)
	}
	id, found, err = finder.first(ctx, "users", "created_at ASC")
	if err != nil {
		return uuid.Nil, err
	}
	if !found {
		return uuid.Nil, errNoOwner
	}
	return id, nil
}

func resolveManager(ctx context.Context, finder lookup, accounts auth.Service) (uuid.UUID, error) {
	id, found, err := finder.find(ctx, "users", map[string]any{"email": demoManagerEmail})
	if err != nil {
		return uuid.Nil, err
	}
	if !found {
		password := os.Getenv("DEMO_USER_PASSWORD")
		if password == "" {
			return uuid.Nil, nil
		}
		created, err := accounts.Register(ctx, auth.RegisterRequest{Name: demoManagerName, Email: demoManagerEmail, Password: password})
		if err != nil {
			return uuid.Nil, err
		}
		id = created.ID
	}
	if _, err := accounts.ChangeRole(ctx, uuid.Nil, id, auth.ChangeRoleRequest{Role: access.RoleProjectManager}); err != nil {
		return uuid.Nil, fmt.Errorf("make %s project manager: %w", demoManagerEmail, err)
	}
	return id, nil
}

type noSubjectCache struct{}

func (noSubjectCache) Forget(uuid.UUID) {}

func makeAdmin(ctx context.Context, accounts auth.Service, userID uuid.UUID) error {
	if _, err := accounts.ChangeRole(ctx, uuid.Nil, userID, auth.ChangeRoleRequest{Role: access.RoleAdmin}); err != nil {
		return fmt.Errorf("make %s admin: %w", demoAccountEmail, err)
	}
	return nil
}

func requireUser(ctx context.Context, finder lookup, email string) (uuid.UUID, error) {
	id, found, err := finder.find(ctx, "users", map[string]any{"email": email})
	if err != nil {
		return uuid.Nil, err
	}
	if !found {
		return uuid.Nil, fmt.Errorf("owner %s is not registered", email)
	}
	return id, nil
}
