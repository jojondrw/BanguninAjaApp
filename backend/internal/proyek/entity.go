package proyek

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

type Proyek struct {
	entity.Base
	Kode          string     `gorm:"type:varchar(20);not null;uniqueIndex:uq_proyek_kode"`
	Nama          string     `gorm:"type:varchar(160);not null"`
	WilayahID     *uuid.UUID `gorm:"type:uuid;index:idx_proyek_wilayah"`
	Jenis         string     `gorm:"type:varchar(40);not null"`
	Status        string     `gorm:"type:varchar(20);not null;index:idx_proyek_status"`
	Progres       int        `gorm:"not null;default:0"`
	TanggalMulai  *time.Time `gorm:"type:date"`
	TargetSelesai *time.Time `gorm:"type:date"`
	NilaiKontrak  int64      `gorm:"not null;default:0"`
}

func (Proyek) TableName() string {
	return "proyek"
}

type TahapProyek struct {
	entity.Base
	ProyekID      uuid.UUID  `gorm:"type:uuid;not null;index:idx_tahap_proyek"`
	Nama          string     `gorm:"type:varchar(160);not null"`
	Urutan        int        `gorm:"not null;default:1"`
	TanggalMulai  *time.Time `gorm:"type:date"`
	TargetSelesai *time.Time `gorm:"type:date"`
	Progres       int        `gorm:"not null;default:0"`
	Status        string     `gorm:"type:varchar(20);not null"`
}

func (TahapProyek) TableName() string {
	return "tahap_proyek"
}

type RabItem struct {
	entity.Base
	ProyekID    uuid.UUID `gorm:"type:uuid;not null;index:idx_rab_item_proyek"`
	Kode        string    `gorm:"type:varchar(20);not null"`
	Uraian      string    `gorm:"type:varchar(200);not null"`
	Volume      float64   `gorm:"type:numeric(14,2);not null;default:0"`
	SatuanID    uuid.UUID `gorm:"type:uuid;not null;index:idx_rab_item_satuan"`
	HargaSatuan int64     `gorm:"not null;default:0"`
	Jumlah      int64     `gorm:"not null;default:0"`
}

func (RabItem) TableName() string {
	return "rab_item"
}

type DokumenIzin struct {
	entity.Base
	ProyekID      uuid.UUID  `gorm:"type:uuid;not null;index:idx_dokumen_izin_proyek"`
	Jenis         string     `gorm:"type:varchar(60);not null"`
	Nomor         string     `gorm:"type:varchar(80);not null"`
	TanggalTerbit *time.Time `gorm:"type:date"`
	BerlakuSampai *time.Time `gorm:"type:date;index:idx_dokumen_izin_berlaku"`
	Status        string     `gorm:"type:varchar(20);not null;index:idx_dokumen_izin_status"`
}

func (DokumenIzin) TableName() string {
	return "dokumen_izin"
}

func Entities() []any {
	return []any{&Proyek{}, &TahapProyek{}, &RabItem{}, &DokumenIzin{}}
}

func Indexes() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_proyek_nama_trgm ON proyek USING gin (nama gin_trgm_ops)`,
		`CREATE INDEX IF NOT EXISTS idx_proyek_kode_pattern ON proyek (kode text_pattern_ops)`,
	}
}

func Constraints() []string {
	return []string{
		database.ForeignKey("proyek", "wilayah_id", "wilayah", database.DeleteRestrict),
		database.ForeignKey("tahap_proyek", "proyek_id", "proyek", database.DeleteCascade),
		database.ForeignKey("rab_item", "proyek_id", "proyek", database.DeleteCascade),
		database.ForeignKey("rab_item", "satuan_id", "satuan", database.DeleteRestrict),
		database.ForeignKey("dokumen_izin", "proyek_id", "proyek", database.DeleteCascade),
		database.Check("proyek", "progres", "progres BETWEEN 0 AND 100"),
		database.Check("proyek", "status", "status IN ('perencanaan','berjalan','tertunda','selesai','batal')"),
		database.Check("proyek", "nilai_kontrak", "nilai_kontrak >= 0"),
		database.Check("proyek", "tanggal", "target_selesai IS NULL OR tanggal_mulai IS NULL OR target_selesai >= tanggal_mulai"),
		database.Check("tahap_proyek", "progres", "progres BETWEEN 0 AND 100"),
		database.Check("rab_item", "volume", "volume >= 0"),
		database.Check("rab_item", "harga_satuan", "harga_satuan >= 0"),
		database.Check("dokumen_izin", "status", "status IN ('diajukan','terbit','kedaluwarsa','ditolak')"),
		database.Unique("rab_item", "proyek_kode", "proyek_id, kode"),
	}
}
