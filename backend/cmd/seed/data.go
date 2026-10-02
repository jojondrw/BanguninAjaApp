package main

import (
	"embed"
	"encoding/csv"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//go:embed data
var seedFiles embed.FS

type codeID struct {
	ID   uuid.UUID
	Code string
}

func readCSV(name string) ([]map[string]string, error) {
	file, err := seedFiles.Open("data/" + name)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", name, err)
	}
	defer func() { _ = file.Close() }()

	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", name, err)
	}

	return toRows(records), nil
}

func readJSON(name string, target any) error {
	content, err := seedFiles.ReadFile("data/" + name)
	if err != nil {
		return fmt.Errorf("open %s: %w", name, err)
	}

	if err := json.Unmarshal(content, target); err != nil {
		return fmt.Errorf("decode %s: %w", name, err)
	}

	return nil
}

func toRows(records [][]string) []map[string]string {
	if len(records) == 0 {
		return nil
	}

	header := records[0]
	rows := make([]map[string]string, 0, len(records)-1)
	for _, record := range records[1:] {
		rows = append(rows, toRow(header, record))
	}

	return rows
}

func toRow(header, record []string) map[string]string {
	row := make(map[string]string, len(header))
	for index, column := range header {
		row[column] = record[index]
	}

	return row
}

func upsertByCode(tx *gorm.DB, record any, columns ...string) error {
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "code"}},
		DoUpdates: clause.AssignmentColumns(append(columns, "updated_at")),
	}).Create(record).Error
}

func loadCodeIDs(tx *gorm.DB, table string) (map[string]uuid.UUID, error) {
	var pairs []codeID
	if err := tx.Table(table).Select("id, code").Scan(&pairs).Error; err != nil {
		return nil, fmt.Errorf("load %s ids: %w", table, err)
	}

	ids := make(map[string]uuid.UUID, len(pairs))
	for _, pair := range pairs {
		ids[pair.Code] = pair.ID
	}

	return ids, nil
}
