package auth

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

type User struct {
	entity.Base
	Name         string     `gorm:"type:varchar(120);not null"`
	Email        string     `gorm:"type:varchar(160);not null;uniqueIndex:idx_users_email"`
	PasswordHash string     `gorm:"type:varchar(120);not null"`
	RoleID       *uuid.UUID `gorm:"type:uuid;index:idx_users_role"`
	Active       bool       `gorm:"not null;default:true"`
	LastLoginAt  *time.Time
}

type RefreshToken struct {
	entity.Base
	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_refresh_tokens_user_id"`
	TokenHash string    `gorm:"type:varchar(64);not null;uniqueIndex:idx_refresh_tokens_hash"`
	ExpiresAt time.Time `gorm:"not null;index:idx_refresh_tokens_expires_at"`
	RevokedAt *time.Time
}

func (t RefreshToken) IsUsable(now time.Time) bool {
	return t.RevokedAt == nil && now.Before(t.ExpiresAt)
}

type Role struct {
	entity.Base
	Code        string `gorm:"type:varchar(30);not null;uniqueIndex:uq_role_code"`
	Name        string `gorm:"type:varchar(80);not null"`
	Description string `gorm:"type:varchar(200)"`
}

func (Role) TableName() string {
	return "role"
}

type ProjectMember struct {
	entity.Base
	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_project_member_user"`
	ProjectID uuid.UUID `gorm:"type:uuid;not null;index:idx_project_member_project"`
	RoleID    uuid.UUID `gorm:"type:uuid;not null;index:idx_project_member_role"`
}

func (ProjectMember) TableName() string {
	return "project_member"
}

func Entities() []any {
	return []any{&Role{}, &User{}, &RefreshToken{}, &ProjectMember{}}
}

func Indexes() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_active
		   ON refresh_tokens (user_id) WHERE revoked_at IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_users_name_trgm
		   ON users USING gin (name gin_trgm_ops)`,
	}
}

func Constraints() []string {
	return []string{
		database.ForeignKey("users", "role_id", "role", database.DeleteSetNull),
		database.ForeignKey("refresh_tokens", "user_id", "users", database.DeleteCascade),
		database.ForeignKey("project_member", "user_id", "users", database.DeleteCascade),
		database.ForeignKey("project_member", "project_id", "project", database.DeleteCascade),
		database.ForeignKey("project_member", "role_id", "role", database.DeleteRestrict),
		database.Unique("project_member", "user_project", "user_id, project_id"),
	}
}
