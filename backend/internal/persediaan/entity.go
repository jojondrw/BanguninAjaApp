package persediaan

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

type Material struct {
	entity.Base
	Kode          string    `gorm:"type:varchar(20);not null;uniqueIndex:uq_material_kode"`
	Nama          string    `gorm:"type:varchar(160);not null"`
	Kategori      string    `gorm:"type:varchar(60);index:idx_material_kategori"`
	SatuanID      uuid.UUID `gorm:"type:uuid;not null;index:idx_material_satuan"`
	StokMinimum   float64   `gorm:"type:numeric(14,2);not null;default:0"`
	HargaTerakhir int64     `gorm:"not null;default:0"`
}

func (Material) TableName() string {
	return "material"
}

type Gudang struct {
	entity.Base
	Kode     string     `gorm:"type:varchar(20);not null;uniqueIndex:uq_gudang_kode"`
	Nama     string     `gorm:"type:varchar(120);not null"`
	ProyekID *uuid.UUID `gorm:"type:uuid;index:idx_gudang_proyek"`
}

func (Gudang) TableName() string {
	return "gudang"
}

type Stok struct {
	entity.Base
	MaterialID uuid.UUID `gorm:"type:uuid;not null;index:idx_stok_material"`
	GudangID   uuid.UUID `gorm:"type:uuid;not null;index:idx_stok_gudang"`
	Jumlah     float64   `gorm:"type:numeric(14,2);not null;default:0"`
}

func (Stok) TableName() string {
	return "stok"
}

type MutasiStok struct {
	entity.Base
	Tanggal        time.Time  `gorm:"type:date;not null;index:idx_mutasi_stok_tanggal"`
	Jenis          string     `gorm:"type:varchar(20);not null;index:idx_mutasi_stok_jenis"`
	MaterialID     uuid.UUID  `gorm:"type:uuid;not null;index:idx_mutasi_stok_material"`
	Jumlah         float64    `gorm:"type:numeric(14,2);not null"`
	GudangAsalID   *uuid.UUID `gorm:"type:uuid;index:idx_mutasi_stok_gudang_asal"`
	GudangTujuanID *uuid.UUID `gorm:"type:uuid;index:idx_mutasi_stok_gudang_tujuan"`
	Referensi      string     `gorm:"type:varchar(60)"`
}

func (MutasiStok) TableName() string {
	return "mutasi_stok"
}

func Entities() []any {
	return []any{&Material{}, &Gudang{}, &Stok{}, &MutasiStok{}}
}

func Indexes() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_material_nama_trgm ON material USING gin (nama gin_trgm_ops)`,
		`CREATE INDEX IF NOT EXISTS idx_mutasi_stok_material_tanggal ON mutasi_stok (material_id, tanggal DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_stok_menipis ON stok (material_id) WHERE jumlah <= 0`,
	}
}

func Constraints() []string {
	return []string{
		database.ForeignKey("material", "satuan_id", "satuan", database.DeleteRestrict),
		database.ForeignKey("gudang", "proyek_id", "proyek", database.DeleteSetNull),
		database.ForeignKey("stok", "material_id", "material", database.DeleteCascade),
		database.ForeignKey("stok", "gudang_id", "gudang", database.DeleteCascade),
		database.ForeignKey("mutasi_stok", "material_id", "material", database.DeleteRestrict),
		database.ForeignKey("mutasi_stok", "gudang_asal_id", "gudang", database.DeleteRestrict),
		database.ForeignKey("mutasi_stok", "gudang_tujuan_id", "gudang", database.DeleteRestrict),
		database.Check("material", "stok_minimum", "stok_minimum >= 0"),
		database.Check("stok", "jumlah", "jumlah >= 0"),
		database.Check("mutasi_stok", "jumlah", "jumlah > 0"),
		database.Check("mutasi_stok", "jenis", "jenis IN ('masuk','keluar','transfer','penyesuaian')"),
		database.Check("mutasi_stok", "arah", "gudang_asal_id IS NOT NULL OR gudang_tujuan_id IS NOT NULL"),
		database.Unique("stok", "material_gudang", "material_id, gudang_id"),
	}
}
