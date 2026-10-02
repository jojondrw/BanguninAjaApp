package main

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/master"
)

type hierarchyRecord func(row map[string]string, parentID *uuid.UUID) any

func seedRegions(tx *gorm.DB) error {
	rows, err := readCSV("regions.csv")
	if err != nil {
		return err
	}

	return upsertHierarchy(tx, "region", rows, func(row map[string]string, parentID *uuid.UUID) any {
		return &master.Region{Code: row["code"], Name: row["name"], Type: row["type"], ParentID: parentID}
	})
}

func seedAccounts(tx *gorm.DB) error {
	rows, err := readCSV("accounts.csv")
	if err != nil {
		return err
	}

	return upsertHierarchy(tx, "account", rows, func(row map[string]string, parentID *uuid.UUID) any {
		return &master.Account{Code: row["code"], Name: row["name"], Type: row["type"], ParentID: parentID}
	})
}

func seedUnits(tx *gorm.DB) error {
	rows, err := readCSV("units.csv")
	if err != nil {
		return err
	}

	for _, row := range rows {
		unit := &master.UnitOfMeasure{Code: row["code"], Name: row["name"]}
		if err := upsertByCode(tx, unit, "name"); err != nil {
			return fmt.Errorf("unit %s: %w", row["code"], err)
		}
	}

	return nil
}

func upsertHierarchy(tx *gorm.DB, table string, rows []map[string]string, build hierarchyRecord) error {
	ids := map[string]uuid.UUID{}
	pending := rows

	for len(pending) > 0 {
		ready, waiting := splitByResolvedParent(pending, ids)
		if len(ready) == 0 {
			return fmt.Errorf("%s: %d rows reference unknown parent codes", table, len(waiting))
		}

		if err := upsertLevel(tx, ready, ids, build); err != nil {
			return fmt.Errorf("%s: %w", table, err)
		}

		loaded, err := loadCodeIDs(tx, table)
		if err != nil {
			return err
		}

		ids = loaded
		pending = waiting
	}

	return nil
}

func splitByResolvedParent(rows []map[string]string, ids map[string]uuid.UUID) (ready, waiting []map[string]string) {
	for _, row := range rows {
		if parentResolved(row["parent_code"], ids) {
			ready = append(ready, row)
			continue
		}
		waiting = append(waiting, row)
	}

	return ready, waiting
}

func parentResolved(parentCode string, ids map[string]uuid.UUID) bool {
	if parentCode == "" {
		return true
	}

	_, found := ids[parentCode]
	return found
}

func upsertLevel(tx *gorm.DB, rows []map[string]string, ids map[string]uuid.UUID, build hierarchyRecord) error {
	for _, row := range rows {
		record := build(row, parentIDOf(row["parent_code"], ids))
		if err := upsertByCode(tx, record, "name", "type", "parent_id"); err != nil {
			return fmt.Errorf("code %s: %w", row["code"], err)
		}
	}

	return nil
}

func parentIDOf(parentCode string, ids map[string]uuid.UUID) *uuid.UUID {
	if parentCode == "" {
		return nil
	}

	id := ids[parentCode]
	return &id
}
