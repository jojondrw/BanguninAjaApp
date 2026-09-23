package database

import (
	"context"
	"strings"

	"gorm.io/gorm"
)

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

type Listing struct {
	Filter func(*gorm.DB) *gorm.DB
	Order  string
	Offset int
	Limit  int
}

func FindPage[T any](ctx context.Context, db *gorm.DB, listing Listing) ([]T, int64, error) {
	var model T
	var total int64
	if err := db.WithContext(ctx).Model(&model).Scopes(listing.scopes()...).Count(&total).Error; err != nil {
		return nil, 0, Translate(err)
	}

	items := make([]T, 0, listing.Limit)
	err := db.WithContext(ctx).
		Scopes(listing.scopes()...).
		Order(listing.Order).
		Offset(listing.Offset).
		Limit(listing.Limit).
		Find(&items).Error
	return items, total, Translate(err)
}

func ContainsPattern(term string) string {
	return "%" + likeEscaper.Replace(strings.TrimSpace(term)) + "%"
}

func (l Listing) scopes() []func(*gorm.DB) *gorm.DB {
	if l.Filter == nil {
		return nil
	}
	return []func(*gorm.DB) *gorm.DB{l.Filter}
}
