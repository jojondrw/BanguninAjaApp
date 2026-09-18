package pengadaan

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

type Vendor struct {
	entity.Base
	Kode       string `gorm:"type:varchar(20);not null;uniqueIndex:uq_vendor_kode"`
	Nama       string `gorm:"type:varchar(160);not null"`
	Kategori   string `gorm:"type:varchar(60);not null;index:idx_vendor_kategori"`
	Kontak     string `gorm:"type:varchar(60)"`
	Npwp       string `gorm:"type:varchar(25)"`
	TerminHari int    `gorm:"not null;default:0"`
	Penilaian  string `gorm:"type:varchar(20);not null;default:'baru'"`
	Aktif      bool   `gorm:"not null;default:true"`
}

func (Vendor) TableName() string {
	return "vendor"
}

type Permintaan struct {
	entity.Base
	Nomor     string    `gorm:"type:varchar(40);not null;uniqueIndex:uq_permintaan_nomor"`
	ProyekID  uuid.UUID `gorm:"type:uuid;not null;index:idx_permintaan_proyek"`
	PemohonID uuid.UUID `gorm:"type:uuid;not null;index:idx_permintaan_pemohon"`
	Tanggal   time.Time `gorm:"type:date;not null;index:idx_permintaan_tanggal"`
	Status    string    `gorm:"type:varchar(20);not null;index:idx_permintaan_status"`
	Catatan   string    `gorm:"type:text"`
}

func (Permintaan) TableName() string {
	return "permintaan"
}

type PermintaanItem struct {
	entity.Base
	PermintaanID uuid.UUID `gorm:"type:uuid;not null;index:idx_permintaan_item_permintaan"`
	MaterialID   uuid.UUID `gorm:"type:uuid;not null;index:idx_permintaan_item_material"`
	Jumlah       float64   `gorm:"type:numeric(14,2);not null;default:0"`
	SatuanID     uuid.UUID `gorm:"type:uuid;not null"`
}

func (PermintaanItem) TableName() string {
	return "permintaan_item"
}

type Pesanan struct {
	entity.Base
	Nomor        string     `gorm:"type:varchar(40);not null;uniqueIndex:uq_pesanan_nomor"`
	VendorID     uuid.UUID  `gorm:"type:uuid;not null;index:idx_pesanan_vendor"`
	ProyekID     uuid.UUID  `gorm:"type:uuid;not null;index:idx_pesanan_proyek"`
	PermintaanID *uuid.UUID `gorm:"type:uuid;index:idx_pesanan_permintaan"`
	Tanggal      time.Time  `gorm:"type:date;not null"`
	JatuhTempo   *time.Time `gorm:"type:date"`
	Nilai        int64      `gorm:"not null;default:0"`
	Status       string     `gorm:"type:varchar(20);not null"`
}

func (Pesanan) TableName() string {
	return "pesanan"
}

type PesananItem struct {
	entity.Base
	PesananID   uuid.UUID `gorm:"type:uuid;not null;index:idx_pesanan_item_pesanan"`
	MaterialID  uuid.UUID `gorm:"type:uuid;not null;index:idx_pesanan_item_material"`
	Jumlah      float64   `gorm:"type:numeric(14,2);not null;default:0"`
	SatuanID    uuid.UUID `gorm:"type:uuid;not null"`
	HargaSatuan int64     `gorm:"not null;default:0"`
}

func (PesananItem) TableName() string {
	return "pesanan_item"
}

type Penerimaan struct {
	entity.Base
	Nomor     string    `gorm:"type:varchar(40);not null;uniqueIndex:uq_penerimaan_nomor"`
	PesananID uuid.UUID `gorm:"type:uuid;not null;index:idx_penerimaan_pesanan"`
	GudangID  uuid.UUID `gorm:"type:uuid;not null;index:idx_penerimaan_gudang"`
	Tanggal   time.Time `gorm:"type:date;not null;index:idx_penerimaan_tanggal"`
	Kondisi   string    `gorm:"type:varchar(20);not null"`
	Catatan   string    `gorm:"type:text"`
}

func (Penerimaan) TableName() string {
	return "penerimaan"
}

type PenerimaanItem struct {
	entity.Base
	PenerimaanID   uuid.UUID `gorm:"type:uuid;not null;index:idx_penerimaan_item_penerimaan"`
	PesananItemID  uuid.UUID `gorm:"type:uuid;not null;index:idx_penerimaan_item_pesanan_item"`
	JumlahDiterima float64   `gorm:"type:numeric(14,2);not null;default:0"`
	JumlahDitolak  float64   `gorm:"type:numeric(14,2);not null;default:0"`
}

func (PenerimaanItem) TableName() string {
	return "penerimaan_item"
}

func Entities() []any {
	return []any{
		&Vendor{}, &Permintaan{}, &PermintaanItem{},
		&Pesanan{}, &PesananItem{}, &Penerimaan{}, &PenerimaanItem{},
	}
}

func Indexes() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_vendor_nama_trgm ON vendor USING gin (nama gin_trgm_ops)`,
		`CREATE INDEX IF NOT EXISTS idx_pesanan_status_tanggal ON pesanan (status, tanggal DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_permintaan_status_tanggal ON permintaan (status, tanggal DESC)`,
	}
}

func Constraints() []string {
	return []string{
		database.ForeignKey("permintaan", "proyek_id", "proyek", database.DeleteRestrict),
		database.ForeignKey("permintaan", "pemohon_id", "users", database.DeleteRestrict),
		database.ForeignKey("permintaan_item", "permintaan_id", "permintaan", database.DeleteCascade),
		database.ForeignKey("permintaan_item", "material_id", "material", database.DeleteRestrict),
		database.ForeignKey("permintaan_item", "satuan_id", "satuan", database.DeleteRestrict),
		database.ForeignKey("pesanan", "vendor_id", "vendor", database.DeleteRestrict),
		database.ForeignKey("pesanan", "proyek_id", "proyek", database.DeleteRestrict),
		database.ForeignKey("pesanan", "permintaan_id", "permintaan", database.DeleteSetNull),
		database.ForeignKey("pesanan_item", "pesanan_id", "pesanan", database.DeleteCascade),
		database.ForeignKey("pesanan_item", "material_id", "material", database.DeleteRestrict),
		database.ForeignKey("pesanan_item", "satuan_id", "satuan", database.DeleteRestrict),
		database.ForeignKey("penerimaan", "pesanan_id", "pesanan", database.DeleteRestrict),
		database.ForeignKey("penerimaan", "gudang_id", "gudang", database.DeleteRestrict),
		database.ForeignKey("penerimaan_item", "penerimaan_id", "penerimaan", database.DeleteCascade),
		database.ForeignKey("penerimaan_item", "pesanan_item_id", "pesanan_item", database.DeleteRestrict),
		database.Check("vendor", "termin_hari", "termin_hari >= 0"),
		database.Check("vendor", "penilaian", "penilaian IN ('baru','baik','cukup','buruk')"),
		database.Check("permintaan", "status", "status IN ('draft','diajukan','disetujui','ditolak','selesai')"),
		database.Check("pesanan", "status", "status IN ('draft','dikirim','diterima_sebagian','selesai','batal')"),
		database.Check("pesanan", "nilai", "nilai >= 0"),
		database.Check("penerimaan", "kondisi", "kondisi IN ('baik','sebagian_rusak','rusak')"),
		database.Check("permintaan_item", "jumlah", "jumlah > 0"),
		database.Check("pesanan_item", "jumlah", "jumlah > 0"),
		database.Check("penerimaan_item", "jumlah", "jumlah_diterima >= 0 AND jumlah_ditolak >= 0"),
		database.Unique("permintaan_item", "permintaan_material", "permintaan_id, material_id"),
		database.Unique("pesanan_item", "pesanan_material", "pesanan_id, material_id"),
	}
}
