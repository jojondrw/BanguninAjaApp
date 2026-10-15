package main

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type lookup interface {
	find(ctx context.Context, table string, keys map[string]any) (uuid.UUID, bool, error)
	first(ctx context.Context, table, order string) (uuid.UUID, bool, error)
}

type gormLookup struct {
	db *gorm.DB
}

func (l gormLookup) find(ctx context.Context, table string, keys map[string]any) (uuid.UUID, bool, error) {
	var ids []uuid.UUID
	err := l.db.WithContext(ctx).Table(table).Where(keys).Limit(1).Pluck("id", &ids).Error
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("find %s: %w", table, err)
	}
	if len(ids) == 0 {
		return uuid.Nil, false, nil
	}
	return ids[0], true, nil
}

func (l gormLookup) first(ctx context.Context, table, order string) (uuid.UUID, bool, error) {
	var ids []uuid.UUID
	err := l.db.WithContext(ctx).Table(table).Order(order).Limit(1).Pluck("id", &ids).Error
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("first %s: %w", table, err)
	}
	if len(ids) == 0 {
		return uuid.Nil, false, nil
	}
	return ids[0], true, nil
}

type tally struct {
	created map[string]int
	skipped map[string]int
}

func newTally() *tally {
	return &tally{created: map[string]int{}, skipped: map[string]int{}}
}

func (t *tally) add(table string) {
	t.created[table]++
}

func (t *tally) skip(table string) {
	t.skipped[table]++
}

func (t *tally) report() {
	tables := make([]string, 0, len(t.created)+len(t.skipped))
	seen := map[string]bool{}
	for _, counts := range []map[string]int{t.created, t.skipped} {
		for table := range counts {
			if !seen[table] {
				seen[table] = true
				tables = append(tables, table)
			}
		}
	}
	sort.Strings(tables)
	for _, table := range tables {
		slog.Info("demo data", slog.String("table", table), slog.Int("created", t.created[table]), slog.Int("skipped", t.skipped[table]))
	}
}
