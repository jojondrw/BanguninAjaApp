package sdm

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

type Karyawan struct {
	entity.Base
	Nik           string     `gorm:"type:varchar(20);not null;uniqueIndex:uq_karyawan_nik"`
	Nama          string     `gorm:"type:varchar(160);not null"`
	Jabatan       string     `gorm:"type:varchar(80);not null"`
	ProyekID      *uuid.UUID `gorm:"type:uuid;index:idx_karyawan_proyek"`
	UserID        *uuid.UUID `gorm:"type:uuid;index:idx_karyawan_user"`
	StatusKerja   string     `gorm:"type:varchar(20);not null;index:idx_karyawan_status"`
	TanggalMasuk  time.Time  `gorm:"type:date;not null"`
	TanggalKeluar *time.Time `gorm:"type:date"`
	GajiPokok     int64      `gorm:"not null;default:0"`
}

func (Karyawan) TableName() string {
	return "karyawan"
}

type Absensi struct {
	entity.Base
	KaryawanID uuid.UUID  `gorm:"type:uuid;not null;index:idx_absensi_karyawan"`
	Tanggal    time.Time  `gorm:"type:date;not null;index:idx_absensi_tanggal"`
	JamMasuk   *time.Time `gorm:"type:time"`
	JamPulang  *time.Time `gorm:"type:time"`
	Status     string     `gorm:"type:varchar(20);not null;index:idx_absensi_status"`
}

func (Absensi) TableName() string {
	return "absensi"
}

type Gaji struct {
	entity.Base
	KaryawanID uuid.UUID `gorm:"type:uuid;not null;index:idx_gaji_karyawan"`
	Periode    string    `gorm:"type:varchar(7);not null;index:idx_gaji_periode"`
	Pokok      int64     `gorm:"not null;default:0"`
	Tunjangan  int64     `gorm:"not null;default:0"`
	Potongan   int64     `gorm:"not null;default:0"`
	Diterima   int64     `gorm:"not null;default:0"`
	Dibayar    bool      `gorm:"not null;default:false"`
}

func (Gaji) TableName() string {
	return "gaji"
}

func Entities() []any {
	return []any{&Karyawan{}, &Absensi{}, &Gaji{}}
}

func Indexes() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_karyawan_nama_trgm ON karyawan USING gin (nama gin_trgm_ops)`,
		`CREATE INDEX IF NOT EXISTS idx_absensi_karyawan_tanggal ON absensi (karyawan_id, tanggal DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_karyawan_aktif ON karyawan (proyek_id) WHERE tanggal_keluar IS NULL`,
	}
}

func Constraints() []string {
	return []string{
		database.ForeignKey("karyawan", "proyek_id", "proyek", database.DeleteSetNull),
		database.ForeignKey("karyawan", "user_id", "users", database.DeleteSetNull),
		database.ForeignKey("absensi", "karyawan_id", "karyawan", database.DeleteCascade),
		database.ForeignKey("gaji", "karyawan_id", "karyawan", database.DeleteCascade),
		database.Check("karyawan", "status_kerja", "status_kerja IN ('tetap','kontrak','harian','subkontraktor')"),
		database.Check("karyawan", "masa_kerja", "tanggal_keluar IS NULL OR tanggal_keluar >= tanggal_masuk"),
		database.Check("karyawan", "gaji_pokok", "gaji_pokok >= 0"),
		database.Check("absensi", "status", "status IN ('hadir','izin','sakit','alpa','libur')"),
		database.Check("absensi", "jam", "jam_pulang IS NULL OR jam_masuk IS NULL OR jam_pulang >= jam_masuk"),
		database.Check("gaji", "periode", "periode ~ '^[0-9]{4}-[0-9]{2}$'"),
		database.Check("gaji", "nilai", "pokok >= 0 AND tunjangan >= 0 AND potongan >= 0 AND diterima >= 0"),
		database.Unique("absensi", "karyawan_tanggal", "karyawan_id, tanggal"),
		database.Unique("gaji", "karyawan_periode", "karyawan_id, periode"),
	}
}
