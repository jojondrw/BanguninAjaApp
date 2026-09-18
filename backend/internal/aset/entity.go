package aset

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

type Aset struct {
	entity.Base
	Kode                string     `gorm:"type:varchar(20);not null;uniqueIndex:uq_aset_kode"`
	Nama                string     `gorm:"type:varchar(160);not null"`
	Kategori            string     `gorm:"type:varchar(60);not null;index:idx_aset_kategori"`
	TanggalPerolehan    *time.Time `gorm:"type:date"`
	NilaiPerolehan      int64      `gorm:"not null;default:0"`
	AkumulasiPenyusutan int64      `gorm:"not null;default:0"`
	UmurManfaatBulan    int        `gorm:"not null;default:0"`
	Status              string     `gorm:"type:varchar(20);not null;default:'digunakan'"`
}

func (Aset) TableName() string {
	return "aset"
}

type Alat struct {
	entity.Base
	Kode             string     `gorm:"type:varchar(20);not null;uniqueIndex:uq_alat_kode"`
	Nama             string     `gorm:"type:varchar(160);not null"`
	AsetID           *uuid.UUID `gorm:"type:uuid;index:idx_alat_aset"`
	ProyekID         *uuid.UUID `gorm:"type:uuid;index:idx_alat_proyek"`
	JamOperasi       int        `gorm:"not null;default:0"`
	ServisBerikutnya *time.Time `gorm:"type:date;index:idx_alat_servis"`
	Status           string     `gorm:"type:varchar(20);not null;index:idx_alat_status"`
}

func (Alat) TableName() string {
	return "alat"
}

func Entities() []any {
	return []any{&Aset{}, &Alat{}}
}

func Indexes() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_aset_nama_trgm ON aset USING gin (nama gin_trgm_ops)`,
		`CREATE INDEX IF NOT EXISTS idx_alat_nama_trgm ON alat USING gin (nama gin_trgm_ops)`,
	}
}

func Constraints() []string {
	return []string{
		database.ForeignKey("alat", "aset_id", "aset", database.DeleteSetNull),
		database.ForeignKey("alat", "proyek_id", "proyek", database.DeleteSetNull),
		database.Check("aset", "nilai", "nilai_perolehan >= 0 AND akumulasi_penyusutan >= 0"),
		database.Check("aset", "penyusutan_wajar", "akumulasi_penyusutan <= nilai_perolehan"),
		database.Check("aset", "status", "status IN ('digunakan','perawatan','dijual','dilepas')"),
		database.Check("alat", "jam_operasi", "jam_operasi >= 0"),
		database.Check("alat", "status", "status IN ('beroperasi','idle','perawatan','rusak')"),
	}
}
