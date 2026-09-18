package laporan

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

type Laporan struct {
	entity.Base
	Nama         string     `gorm:"type:varchar(160);not null"`
	Jenis        string     `gorm:"type:varchar(40);not null;index:idx_laporan_jenis"`
	Cakupan      string     `gorm:"type:varchar(120)"`
	ProyekID     *uuid.UUID `gorm:"type:uuid;index:idx_laporan_proyek"`
	PeriodeMulai *time.Time `gorm:"type:date"`
	PeriodeAkhir *time.Time `gorm:"type:date"`
	Format       string     `gorm:"type:varchar(10);not null"`
	DibuatOleh   uuid.UUID  `gorm:"type:uuid;not null;index:idx_laporan_dibuat_oleh"`
}

func (Laporan) TableName() string {
	return "laporan"
}

func Entities() []any {
	return []any{&Laporan{}}
}

func Indexes() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_laporan_dibuat ON laporan (created_at DESC)`,
	}
}

func Constraints() []string {
	return []string{
		database.ForeignKey("laporan", "proyek_id", "proyek", database.DeleteCascade),
		database.ForeignKey("laporan", "dibuat_oleh", "users", database.DeleteRestrict),
		database.Check("laporan", "format", "format IN ('pdf','xlsx','csv')"),
		database.Check("laporan", "periode", "periode_akhir IS NULL OR periode_mulai IS NULL OR periode_akhir >= periode_mulai"),
	}
}
