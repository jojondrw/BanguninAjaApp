package database

import (
	"fmt"

	"gorm.io/gorm"
)

var requiredExtensions = []string{"postgis", "pg_trgm", "btree_gist"}

func EnsureExtensions(db *gorm.DB) error {
	for _, name := range requiredExtensions {
		statement := fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS %q", name)
		if err := db.Exec(statement).Error; err != nil {
			return fmt.Errorf("create extension %s: %w", name, err)
		}
	}
	return nil
}

func InstalledExtensions(db *gorm.DB) ([]string, error) {
	var names []string
	err := db.Raw("SELECT extname FROM pg_extension ORDER BY extname").Scan(&names).Error
	if err != nil {
		return nil, fmt.Errorf("read installed extensions: %w", err)
	}
	return names, nil
}
