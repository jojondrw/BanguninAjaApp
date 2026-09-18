package auth

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

type User struct {
	entity.Base
	Name          string     `gorm:"type:varchar(120);not null"`
	Email         string     `gorm:"type:varchar(160);not null;uniqueIndex:idx_users_email"`
	PasswordHash  string     `gorm:"type:varchar(120);not null"`
	PeranID       *uuid.UUID `gorm:"type:uuid;index:idx_users_peran"`
	Aktif         bool       `gorm:"not null;default:true"`
	TerakhirMasuk *time.Time
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

type Peran struct {
	entity.Base
	Kode       string `gorm:"type:varchar(30);not null;uniqueIndex:uq_peran_kode"`
	Nama       string `gorm:"type:varchar(80);not null"`
	Keterangan string `gorm:"type:varchar(200)"`
}

func (Peran) TableName() string {
	return "peran"
}

type PenggunaProyek struct {
	entity.Base
	UserID   uuid.UUID `gorm:"type:uuid;not null;index:idx_pengguna_proyek_user"`
	ProyekID uuid.UUID `gorm:"type:uuid;not null;index:idx_pengguna_proyek_proyek"`
	PeranID  uuid.UUID `gorm:"type:uuid;not null;index:idx_pengguna_proyek_peran"`
}

func (PenggunaProyek) TableName() string {
	return "pengguna_proyek"
}

func Entities() []any {
	return []any{&Peran{}, &User{}, &RefreshToken{}, &PenggunaProyek{}}
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
		database.ForeignKey("users", "peran_id", "peran", database.DeleteSetNull),
		database.ForeignKey("refresh_tokens", "user_id", "users", database.DeleteCascade),
		database.ForeignKey("pengguna_proyek", "user_id", "users", database.DeleteCascade),
		database.ForeignKey("pengguna_proyek", "proyek_id", "proyek", database.DeleteCascade),
		database.ForeignKey("pengguna_proyek", "peran_id", "peran", database.DeleteRestrict),
		database.Unique("pengguna_proyek", "user_proyek", "user_id, proyek_id"),
	}
}
