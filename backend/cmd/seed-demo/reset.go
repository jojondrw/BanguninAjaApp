package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"
)

const demoPattern = demoPrefix + "%"

type purge struct {
	table string
	where string
	args  []any
}

func demoPurges(today time.Time) []purge {
	demoProjects := gorm.Expr("SELECT id FROM project WHERE code LIKE ?", demoPattern)
	demoCustomers := gorm.Expr("SELECT id FROM customer WHERE identity_number LIKE ?", demoPattern)
	demoMaterials := gorm.Expr("SELECT id FROM material WHERE code LIKE ?", demoPattern)
	demoWarehouses := gorm.Expr("SELECT id FROM warehouse WHERE code LIKE ?", demoPattern)
	return []purge{
		{"comparison", "name = ?", []any{comparisonName}},
		{"saved_location", "name IN ?", []any{demoSiteNames()}},
		{"receivable", "reference LIKE ?", []any{demoPattern}},
		{"payable", "reference LIKE ?", []any{demoPattern}},
		{"invoice", "number LIKE ?", []any{demoPattern}},
		{"employee", "identity_number LIKE ?", []any{demoPattern}},
		{"equipment", "code LIKE ?", []any{demoPattern}},
		{"asset", "code LIKE ?", []any{demoPattern}},
		{"goods_receipt", "number LIKE ?", []any{demoPattern}},
		{"stock_movement", "reference LIKE ? OR material_id IN (?) OR source_warehouse_id IN (?) OR target_warehouse_id IN (?)", []any{demoPattern, demoMaterials, demoWarehouses, demoWarehouses}},
		{"purchase_order", "number LIKE ?", []any{demoPattern}},
		{"purchase_request", "number LIKE ?", []any{demoPattern}},
		{"vendor", "code LIKE ?", []any{demoPattern}},
		{"material", "code LIKE ?", []any{demoPattern}},
		{"warehouse", "code LIKE ?", []any{demoPattern}},
		{"cash_transaction", "project_id IN (?) OR note IN ?", []any{demoProjects, companyCashNotes(today)}},
		{"contract", "number LIKE ?", []any{demoPattern}},
		{"lead", "project_id IN (?) OR customer_id IN (?)", []any{demoProjects, demoCustomers}},
		{"customer", "identity_number LIKE ?", []any{demoPattern}},
		{"unit", "code LIKE ?", []any{demoPattern}},
		{"project", "code LIKE ?", []any{demoPattern}},
	}
}

func resetDemo(ctx context.Context, tx *gorm.DB, today time.Time) error {
	for _, step := range demoPurges(today) {
		result := tx.WithContext(ctx).Table(step.table).Where(step.where, step.args...).Delete(map[string]any{})
		if result.Error != nil {
			return fmt.Errorf("reset demo %s: %w", step.table, result.Error)
		}
		slog.Info("demo data removed", slog.String("table", step.table), slog.Int64("rows", result.RowsAffected))
	}
	return nil
}
