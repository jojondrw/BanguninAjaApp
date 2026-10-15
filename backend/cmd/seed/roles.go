package main

import (
	"context"

	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/auth"
)

func seedRoles(tx *gorm.DB) error {
	_, err := auth.EnsureRoles(context.Background(), auth.NewRepository(tx))
	return err
}
