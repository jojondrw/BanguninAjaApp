package auth

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name         string    `gorm:"type:varchar(120);not null"`
	Email        string    `gorm:"type:varchar(160);not null;uniqueIndex:idx_users_email"`
	PasswordHash string    `gorm:"type:varchar(120);not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u *User) BeforeCreate(*gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_refresh_tokens_user_id"`
	TokenHash string    `gorm:"type:varchar(64);not null;uniqueIndex:idx_refresh_tokens_hash"`
	ExpiresAt time.Time `gorm:"not null;index:idx_refresh_tokens_expires_at"`
	RevokedAt *time.Time
	CreatedAt time.Time
	User      User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (t *RefreshToken) BeforeCreate(*gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (t RefreshToken) IsUsable(now time.Time) bool {
	return t.RevokedAt == nil && now.Before(t.ExpiresAt)
}

func Entities() []any {
	return []any{&User{}, &RefreshToken{}}
}
