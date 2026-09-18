package database

import (
	"fmt"

	"gorm.io/gorm"
)

func ApplyIndexes(db *gorm.DB, statements []string) error {
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return fmt.Errorf("apply index: %w", err)
		}
	}
	return nil
}
