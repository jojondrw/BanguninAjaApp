package penjualan

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

type Pembeli struct {
	entity.Base
	Nama   string `gorm:"type:varchar(160);not null"`
	Kontak string `gorm:"type:varchar(60)"`
	Email  string `gorm:"type:varchar(160)"`
	Nik    string `gorm:"type:varchar(20);uniqueIndex:uq_pembeli_nik"`
	Alamat string `gorm:"type:text"`
}

func (Pembeli) TableName() string {
	return "pembeli"
}

type Unit struct {
	entity.Base
	Kode     string    `gorm:"type:varchar(20);not null;uniqueIndex:uq_unit_kode"`
	ProyekID uuid.UUID `gorm:"type:uuid;not null;index:idx_unit_proyek"`
	Tipe     string    `gorm:"type:varchar(60);not null"`
	LuasM2   float64   `gorm:"type:numeric(10,2);not null;default:0"`
	Harga    int64     `gorm:"not null;default:0"`
	Status   string    `gorm:"type:varchar(20);not null;index:idx_unit_status"`
}

func (Unit) TableName() string {
	return "unit"
}

type Prospek struct {
	entity.Base
	Nama              string     `gorm:"type:varchar(160);not null"`
	Kontak            string     `gorm:"type:varchar(60)"`
	ProyekID          *uuid.UUID `gorm:"type:uuid;index:idx_prospek_proyek"`
	Sumber            string     `gorm:"type:varchar(60)"`
	Tahap             string     `gorm:"type:varchar(20);not null;index:idx_prospek_tahap"`
	TerakhirDihubungi *time.Time `gorm:"type:date"`
}

func (Prospek) TableName() string {
	return "prospek"
}

type Kontrak struct {
	entity.Base
	Nomor     string    `gorm:"type:varchar(40);not null;uniqueIndex:uq_kontrak_nomor"`
	PembeliID uuid.UUID `gorm:"type:uuid;not null;index:idx_kontrak_pembeli"`
	UnitID    uuid.UUID `gorm:"type:uuid;not null;index:idx_kontrak_unit"`
	Jenis     string    `gorm:"type:varchar(20);not null"`
	Nilai     int64     `gorm:"not null;default:0"`
	Tanggal   time.Time `gorm:"type:date;not null"`
	Status    string    `gorm:"type:varchar(20);not null;index:idx_kontrak_status"`
}

func (Kontrak) TableName() string {
	return "kontrak"
}

type Cicilan struct {
	entity.Base
	KontrakID    uuid.UUID  `gorm:"type:uuid;not null;index:idx_cicilan_kontrak"`
	AngsuranKe   int        `gorm:"not null"`
	JatuhTempo   time.Time  `gorm:"type:date;not null;index:idx_cicilan_jatuh_tempo"`
	Jumlah       int64      `gorm:"not null;default:0"`
	TanggalBayar *time.Time `gorm:"type:date"`
	Status       string     `gorm:"type:varchar(20);not null;index:idx_cicilan_status"`
}

func (Cicilan) TableName() string {
	return "cicilan"
}

func Entities() []any {
	return []any{&Pembeli{}, &Unit{}, &Prospek{}, &Kontrak{}, &Cicilan{}}
}

func Indexes() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_pembeli_nama_trgm ON pembeli USING gin (nama gin_trgm_ops)`,
		`CREATE INDEX IF NOT EXISTS idx_prospek_nama_trgm ON prospek USING gin (nama gin_trgm_ops)`,
		`CREATE INDEX IF NOT EXISTS idx_unit_proyek_status ON unit (proyek_id, status)`,
		`CREATE INDEX IF NOT EXISTS idx_cicilan_belum_lunas ON cicilan (jatuh_tempo) WHERE status <> 'lunas'`,
	}
}

func Constraints() []string {
	return []string{
		database.ForeignKey("unit", "proyek_id", "proyek", database.DeleteRestrict),
		database.ForeignKey("prospek", "proyek_id", "proyek", database.DeleteSetNull),
		database.ForeignKey("kontrak", "pembeli_id", "pembeli", database.DeleteRestrict),
		database.ForeignKey("kontrak", "unit_id", "unit", database.DeleteRestrict),
		database.ForeignKey("cicilan", "kontrak_id", "kontrak", database.DeleteCascade),
		database.Check("unit", "status", "status IN ('tersedia','dipesan','terjual','ditahan')"),
		database.Check("unit", "harga", "harga >= 0"),
		database.Check("unit", "luas", "luas_m2 > 0"),
		database.Check("prospek", "tahap", "tahap IN ('baru','tertarik','negosiasi','deal','batal')"),
		database.Check("kontrak", "jenis", "jenis IN ('tunai','kpr','bertahap')"),
		database.Check("kontrak", "status", "status IN ('draft','aktif','lunas','batal')"),
		database.Check("kontrak", "nilai", "nilai >= 0"),
		database.Check("cicilan", "angsuran_ke", "angsuran_ke > 0"),
		database.Check("cicilan", "jumlah", "jumlah >= 0"),
		database.Check("cicilan", "status", "status IN ('belum_jatuh_tempo','jatuh_tempo','lunas','menunggak')"),
		database.Unique("kontrak", "unit_aktif", "unit_id, nomor"),
		database.Unique("cicilan", "kontrak_angsuran", "kontrak_id, angsuran_ke"),
	}
}
