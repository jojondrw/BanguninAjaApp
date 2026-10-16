//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/migration"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/project"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/clock"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

const testDatabaseSuffix = "_test"

var db *gorm.DB

func TestMain(m *testing.M) {
	code, err := run(m)
	if err != nil {
		fmt.Fprintln(os.Stderr, "integration setup:", err)
		os.Exit(1)
	}
	os.Exit(code)
}

func run(m *testing.M) (int, error) {
	if _, err := clock.UseAppZone(); err != nil {
		return 0, err
	}
	cfg, err := config.Load()
	if err != nil {
		return 0, err
	}
	if !strings.HasSuffix(cfg.Database.Name, testDatabaseSuffix) {
		return 0, fmt.Errorf("refusing to run against %q: the database name must end with %s", cfg.Database.Name, testDatabaseSuffix)
	}

	db, err = database.Open(cfg)
	if err != nil {
		return 0, err
	}
	defer func() { _ = database.Close(db) }()

	if err := migration.Apply(db); err != nil {
		return 0, fmt.Errorf("migrate: %w", err)
	}
	return m.Run(), nil
}

func uniqueCode(prefix string) string {
	return prefix + "-" + strings.ToUpper(uuid.NewString()[:8])
}

func date(year int, month time.Month, day int) *time.Time {
	value := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return &value
}

func newProject(t *testing.T, projects project.Service) project.ProjectResponse {
	t.Helper()
	created, err := projects.CreateProject(context.Background(), project.ProjectRequest{
		Code:          uniqueCode("IT"),
		Name:          "Proyek uji integrasi",
		Type:          "Perumahan",
		StartDate:     date(2026, 1, 1),
		TargetEndDate: date(2026, 12, 31),
		ContractValue: 10_000_000_000,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	return created
}

func count(t *testing.T, query string, args ...any) int64 {
	t.Helper()
	var total int64
	if err := db.Raw(query, args...).Scan(&total).Error; err != nil {
		t.Fatalf("count %q: %v", query, err)
	}
	return total
}
