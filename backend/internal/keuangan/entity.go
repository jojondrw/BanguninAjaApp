package keuangan

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

type Anggaran struct {
	entity.Base
	ProyekID   uuid.UUID `gorm:"type:uuid;not null;index:idx_anggaran_proyek"`
	Tahun      int       `gorm:"not null;index:idx_anggaran_tahun"`
	Nilai      int64     `gorm:"not null;default:0"`
	Keterangan string    `gorm:"type:varchar(200)"`
}

func (Anggaran) TableName() string {
	return "anggaran"
}

type TransaksiKas struct {
	entity.Base
	Tanggal    time.Time  `gorm:"type:date;not null;index:idx_transaksi_kas_tanggal"`
	Jenis      string     `gorm:"type:varchar(20);not null;index:idx_transaksi_kas_jenis"`
	AkunID     uuid.UUID  `gorm:"type:uuid;not null;index:idx_transaksi_kas_akun"`
	ProyekID   *uuid.UUID `gorm:"type:uuid;index:idx_transaksi_kas_proyek"`
	Jumlah     int64      `gorm:"not null;default:0"`
	Keterangan string     `gorm:"type:varchar(200)"`
}

func (TransaksiKas) TableName() string {
	return "transaksi_kas"
}

type Jurnal struct {
	entity.Base
	Nomor      string    `gorm:"type:varchar(40);not null;uniqueIndex:uq_jurnal_nomor"`
	Tanggal    time.Time `gorm:"type:date;not null;index:idx_jurnal_tanggal"`
	Keterangan string    `gorm:"type:varchar(200)"`
	Sumber     string    `gorm:"type:varchar(40)"`
}

func (Jurnal) TableName() string {
	return "jurnal"
}

type JurnalDetail struct {
	entity.Base
	JurnalID uuid.UUID `gorm:"type:uuid;not null;index:idx_jurnal_detail_jurnal"`
	AkunID   uuid.UUID `gorm:"type:uuid;not null;index:idx_jurnal_detail_akun"`
	Debit    int64     `gorm:"not null;default:0"`
	Kredit   int64     `gorm:"not null;default:0"`
}

func (JurnalDetail) TableName() string {
	return "jurnal_detail"
}

func Entities() []any {
	return []any{&Anggaran{}, &TransaksiKas{}, &Jurnal{}, &JurnalDetail{}}
}

func Indexes() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_jurnal_detail_akun_jurnal ON jurnal_detail (akun_id, jurnal_id)`,
		`CREATE INDEX IF NOT EXISTS idx_transaksi_kas_proyek_tanggal ON transaksi_kas (proyek_id, tanggal DESC)`,
	}
}

func Constraints() []string {
	return []string{
		database.ForeignKey("anggaran", "proyek_id", "proyek", database.DeleteCascade),
		database.ForeignKey("transaksi_kas", "akun_id", "akun", database.DeleteRestrict),
		database.ForeignKey("transaksi_kas", "proyek_id", "proyek", database.DeleteSetNull),
		database.ForeignKey("jurnal_detail", "jurnal_id", "jurnal", database.DeleteCascade),
		database.ForeignKey("jurnal_detail", "akun_id", "akun", database.DeleteRestrict),
		database.Check("anggaran", "nilai", "nilai >= 0"),
		database.Check("anggaran", "tahun", "tahun BETWEEN 2000 AND 2100"),
		database.Check("transaksi_kas", "jenis", "jenis IN ('masuk','keluar')"),
		database.Check("transaksi_kas", "jumlah", "jumlah > 0"),
		database.Check("jurnal_detail", "nilai", "debit >= 0 AND kredit >= 0"),
		database.Check("jurnal_detail", "satu_sisi", "(debit = 0) <> (kredit = 0)"),
		database.Unique("anggaran", "proyek_tahun", "proyek_id, tahun"),
	}
}
