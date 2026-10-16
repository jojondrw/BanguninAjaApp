package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/asset"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/billing"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/finance"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/hr"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/inventory"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/location"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/procurement"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/project"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/sales"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/site"
)

const (
	demoPrefix = "DEMO-"
	wibOffset  = 7 * 60 * 60
)

var indonesianMonths = []string{
	"Januari", "Februari", "Maret", "April", "Mei", "Juni",
	"Juli", "Agustus", "September", "Oktober", "November", "Desember",
}

type services struct {
	projects    project.Service
	finance     finance.Service
	sales       sales.Service
	inventory   inventory.Service
	procurement procurement.Service
	billing     billing.Service
	hr          hr.Service
	assets      asset.Service
	locations   location.Service
	sites       site.Service
}

type locationPatcher interface {
	describeLocation(ctx context.Context, id uuid.UUID, patch locationDetails) error
}

type locationDetails struct {
	note     string
	regionID *uuid.UUID
	areaSqm  float64
}

type seeder struct {
	ctx      context.Context
	lookup   lookup
	services services
	patcher  locationPatcher
	history  progressRecorder
	today    time.Time
	owner    uuid.UUID
	ids      map[string]uuid.UUID
	tally    *tally
	scoring  bool

	pendingInvoices []invoiceSeed
}

type seedStep struct {
	name string
	run  func() error
}

func (s *seeder) run() error {
	steps := []seedStep{
		{"projects", s.seedProjects},
		{"progress history", s.seedProgressHistory},
		{"finance", s.seedFinance},
		{"sales", s.seedSales},
		{"inventory", s.seedInventory},
		{"procurement", s.seedProcurement},
		{"stock movements", s.seedStockMovements},
		{"billing", s.seedBilling},
		{"hr", s.seedHR},
		{"assets", s.seedAssets},
		{"locations", s.seedLocations},
	}
	for _, step := range steps {
		if err := step.run(); err != nil {
			return fmt.Errorf("seed demo %s: %w", step.name, err)
		}
		slog.Info("demo seeded", slog.String("step", step.name))
	}
	s.tally.report()
	return nil
}

func (s *seeder) ensure(table string, keys map[string]any, create func() (uuid.UUID, error)) (uuid.UUID, bool, error) {
	id, found, err := s.lookup.find(s.ctx, table, keys)
	if err != nil {
		return uuid.Nil, false, err
	}
	if found {
		s.tally.skip(table)
		return id, false, nil
	}
	id, err = create()
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("create %s %v: %w", table, keys, err)
	}
	s.tally.add(table)
	return id, true, nil
}

func (s *seeder) reference(table, code string) (uuid.UUID, error) {
	cached, known := s.ids[table+":"+code]
	if known {
		return cached, nil
	}
	id, found, err := s.lookup.find(s.ctx, table, map[string]any{"code": code})
	if err != nil {
		return uuid.Nil, err
	}
	if !found {
		return uuid.Nil, fmt.Errorf("%s %s not found, run cmd/seed first", table, code)
	}
	s.remember(table, code, id)
	return id, nil
}

func (s *seeder) remember(table, key string, id uuid.UUID) {
	s.ids[table+":"+key] = id
}

func (s *seeder) known(table, key string) (uuid.UUID, error) {
	id, found := s.ids[table+":"+key]
	if !found {
		return uuid.Nil, fmt.Errorf("demo %s %s has not been seeded", table, key)
	}
	return id, nil
}

func (s *seeder) knownPointer(table, key string) (*uuid.UUID, error) {
	if key == "" {
		return nil, nil
	}
	id, err := s.known(table, key)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (s *seeder) day(offset int) time.Time {
	return s.today.AddDate(0, 0, offset)
}

func (s *seeder) dayPointer(offset int) *time.Time {
	date := s.day(offset)
	return &date
}

func (s *seeder) monthDay(monthsAgo, day int) time.Time {
	return pastOrToday(monthStart(s.today, monthsAgo).AddDate(0, 0, day-1), s.today)
}

func (s *seeder) monthOffset(months int) *time.Time {
	date := monthStart(s.today, -months)
	return &date
}

func (s *seeder) monthLabel(monthsAgo int) string {
	return monthLabelOf(s.today, monthsAgo)
}

func monthStart(today time.Time, monthsAgo int) time.Time {
	return time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -monthsAgo, 0)
}

func monthLabelOf(today time.Time, monthsAgo int) string {
	date := monthStart(today, monthsAgo)
	return fmt.Sprintf("%s %d", indonesianMonths[date.Month()-1], date.Year())
}

func (s *seeder) period(monthsAgo int) string {
	return s.monthDay(monthsAgo, 1).Format("2006-01")
}

func todayInJakarta(now time.Time) time.Time {
	local := now.In(time.FixedZone("WIB", wibOffset))
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
}

func pastOrToday(date, today time.Time) time.Time {
	if date.After(today) {
		return today
	}
	return date
}
