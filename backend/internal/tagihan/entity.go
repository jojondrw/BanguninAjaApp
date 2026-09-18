package tagihan

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

type Tagihan struct {
	entity.Base
	Nomor         string     `gorm:"type:varchar(40);not null;uniqueIndex:uq_tagihan_nomor"`
	Keterangan    string     `gorm:"type:varchar(200)"`
	JenisPihak    string     `gorm:"type:varchar(20);not null"`
	PihakID       uuid.UUID  `gorm:"type:uuid;not null;index:idx_tagihan_pihak"`
	ProyekID      *uuid.UUID `gorm:"type:uuid;index:idx_tagihan_proyek"`
	JatuhTempo    time.Time  `gorm:"type:date;not null;index:idx_tagihan_jatuh_tempo"`
	Jumlah        int64      `gorm:"not null;default:0"`
	JumlahDibayar int64      `gorm:"not null;default:0"`
	Status        string     `gorm:"type:varchar(20);not null;index:idx_tagihan_status"`
}

func (Tagihan) TableName() string {
	return "tagihan"
}

type Piutang struct {
	entity.Base
	PembeliID     uuid.UUID `gorm:"type:uuid;not null;index:idx_piutang_pembeli"`
	Referensi     string    `gorm:"type:varchar(60);not null"`
	JatuhTempo    time.Time `gorm:"type:date;not null;index:idx_piutang_jatuh_tempo"`
	Jumlah        int64     `gorm:"not null;default:0"`
	JumlahDibayar int64     `gorm:"not null;default:0"`
	Status        string    `gorm:"type:varchar(20);not null;index:idx_piutang_status"`
}

func (Piutang) TableName() string {
	return "piutang"
}

type Hutang struct {
	entity.Base
	VendorID      uuid.UUID `gorm:"type:uuid;not null;index:idx_hutang_vendor"`
	Referensi     string    `gorm:"type:varchar(60);not null"`
	JatuhTempo    time.Time `gorm:"type:date;not null;index:idx_hutang_jatuh_tempo"`
	Jumlah        int64     `gorm:"not null;default:0"`
	JumlahDibayar int64     `gorm:"not null;default:0"`
	Status        string    `gorm:"type:varchar(20);not null;index:idx_hutang_status"`
}

func (Hutang) TableName() string {
	return "hutang"
}

func Entities() []any {
	return []any{&Tagihan{}, &Piutang{}, &Hutang{}}
}

func Indexes() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_tagihan_belum_lunas ON tagihan (jatuh_tempo) WHERE status <> 'lunas'`,
		`CREATE INDEX IF NOT EXISTS idx_piutang_belum_lunas ON piutang (jatuh_tempo) WHERE status <> 'lunas'`,
		`CREATE INDEX IF NOT EXISTS idx_hutang_belum_lunas ON hutang (jatuh_tempo) WHERE status <> 'lunas'`,
	}
}

func Constraints() []string {
	statuses := "status IN ('belum_jatuh_tempo','jatuh_tempo','lunas','menunggak')"

	return []string{
		database.ForeignKey("tagihan", "proyek_id", "proyek", database.DeleteSetNull),
		database.ForeignKey("piutang", "pembeli_id", "pembeli", database.DeleteRestrict),
		database.ForeignKey("hutang", "vendor_id", "vendor", database.DeleteRestrict),
		database.Check("tagihan", "jenis_pihak", "jenis_pihak IN ('pembeli','vendor')"),
		database.Check("tagihan", "status", statuses),
		database.Check("piutang", "status", statuses),
		database.Check("hutang", "status", statuses),
		database.Check("tagihan", "jumlah", "jumlah >= 0 AND jumlah_dibayar >= 0 AND jumlah_dibayar <= jumlah"),
		database.Check("piutang", "jumlah", "jumlah >= 0 AND jumlah_dibayar >= 0 AND jumlah_dibayar <= jumlah"),
		database.Check("hutang", "jumlah", "jumlah >= 0 AND jumlah_dibayar >= 0 AND jumlah_dibayar <= jumlah"),
	}
}
