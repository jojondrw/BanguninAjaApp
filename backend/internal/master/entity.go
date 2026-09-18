package master

import (
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

type Wilayah struct {
	entity.Base
	Kode    string     `gorm:"type:varchar(20);not null;uniqueIndex:uq_wilayah_kode"`
	Nama    string     `gorm:"type:varchar(120);not null;index:idx_wilayah_nama"`
	Jenis   string     `gorm:"type:varchar(20);not null;index:idx_wilayah_jenis"`
	IndukID *uuid.UUID `gorm:"type:uuid;index:idx_wilayah_induk"`
}

func (Wilayah) TableName() string {
	return "wilayah"
}

type Satuan struct {
	entity.Base
	Kode string `gorm:"type:varchar(20);not null;uniqueIndex:uq_satuan_kode"`
	Nama string `gorm:"type:varchar(60);not null"`
}

func (Satuan) TableName() string {
	return "satuan"
}

type Akun struct {
	entity.Base
	Kode    string     `gorm:"type:varchar(20);not null;uniqueIndex:uq_akun_kode"`
	Nama    string     `gorm:"type:varchar(120);not null;index:idx_akun_nama"`
	Jenis   string     `gorm:"type:varchar(20);not null;index:idx_akun_jenis"`
	IndukID *uuid.UUID `gorm:"type:uuid;index:idx_akun_induk"`
}

func (Akun) TableName() string {
	return "akun"
}

func Entities() []any {
	return []any{&Wilayah{}, &Satuan{}, &Akun{}}
}

func Indexes() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_wilayah_nama_trgm ON wilayah USING gin (nama gin_trgm_ops)`,
		`CREATE INDEX IF NOT EXISTS idx_akun_nama_trgm ON akun USING gin (nama gin_trgm_ops)`,
	}
}

func Constraints() []string {
	return []string{
		database.ForeignKey("wilayah", "induk_id", "wilayah", database.DeleteRestrict),
		database.ForeignKey("akun", "induk_id", "akun", database.DeleteRestrict),
		database.Check("wilayah", "jenis", "jenis IN ('provinsi','kota','kabupaten','kecamatan')"),
		database.Check("akun", "jenis", "jenis IN ('aset','kewajiban','modal','pendapatan','beban')"),
	}
}
