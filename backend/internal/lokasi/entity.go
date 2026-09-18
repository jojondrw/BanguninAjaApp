package lokasi

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/geo"
)

type Incaran struct {
	entity.Base
	UserID           uuid.UUID  `gorm:"type:uuid;not null;index:idx_incaran_user"`
	Nama             string     `gorm:"type:varchar(160);not null"`
	WilayahID        *uuid.UUID `gorm:"type:uuid;index:idx_incaran_wilayah"`
	ProfilBangunanID *uuid.UUID `gorm:"type:uuid;index:idx_incaran_profil"`
	Titik            geo.Point  `gorm:"not null"`
	LuasM2           float64    `gorm:"type:numeric(14,2);not null;default:0"`
	HargaTanahPerM2  int64      `gorm:"not null;default:0"`
	Skor             int        `gorm:"not null;default:0;index:idx_incaran_skor"`
	IndeksBanjir     float64    `gorm:"type:numeric(4,3);not null;default:0"`
	IndeksGempa      float64    `gorm:"type:numeric(4,3);not null;default:0"`
	Catatan          string     `gorm:"type:text"`
	DisimpanPada     time.Time  `gorm:"not null;default:now()"`
}

func (Incaran) TableName() string {
	return "incaran"
}

type SkorDimensi struct {
	entity.Base
	IncaranID uuid.UUID `gorm:"type:uuid;not null;index:idx_skor_dimensi_incaran"`
	DimensiID uuid.UUID `gorm:"type:uuid;not null;index:idx_skor_dimensi_dimensi"`
	Nilai     int       `gorm:"not null;default:0"`
}

func (SkorDimensi) TableName() string {
	return "skor_dimensi"
}

type Perbandingan struct {
	entity.Base
	UserID uuid.UUID `gorm:"type:uuid;not null;index:idx_perbandingan_user"`
	Nama   string    `gorm:"type:varchar(160);not null"`
}

func (Perbandingan) TableName() string {
	return "perbandingan"
}

type PerbandinganItem struct {
	entity.Base
	PerbandinganID uuid.UUID `gorm:"type:uuid;not null;index:idx_perbandingan_item_perbandingan"`
	IncaranID      uuid.UUID `gorm:"type:uuid;not null;index:idx_perbandingan_item_incaran"`
	Urutan         int       `gorm:"not null;default:1"`
}

func (PerbandinganItem) TableName() string {
	return "perbandingan_item"
}

func Entities() []any {
	return []any{&Incaran{}, &SkorDimensi{}, &Perbandingan{}, &PerbandinganItem{}}
}

func Indexes() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_incaran_titik ON incaran USING gist (titik)`,
		`CREATE INDEX IF NOT EXISTS idx_incaran_nama_trgm ON incaran USING gin (nama gin_trgm_ops)`,
		`CREATE INDEX IF NOT EXISTS idx_incaran_user_skor ON incaran (user_id, skor DESC)`,
	}
}

func Constraints() []string {
	return []string{
		database.ForeignKey("incaran", "user_id", "users", database.DeleteCascade),
		database.ForeignKey("incaran", "wilayah_id", "wilayah", database.DeleteRestrict),
		database.ForeignKey("incaran", "profil_bangunan_id", "profil_bangunan", database.DeleteSetNull),
		database.ForeignKey("skor_dimensi", "incaran_id", "incaran", database.DeleteCascade),
		database.ForeignKey("skor_dimensi", "dimensi_id", "dimensi", database.DeleteRestrict),
		database.ForeignKey("perbandingan", "user_id", "users", database.DeleteCascade),
		database.ForeignKey("perbandingan_item", "perbandingan_id", "perbandingan", database.DeleteCascade),
		database.ForeignKey("perbandingan_item", "incaran_id", "incaran", database.DeleteCascade),
		database.Check("incaran", "skor", "skor BETWEEN 0 AND 100"),
		database.Check("incaran", "indeks", "indeks_banjir BETWEEN 0 AND 1 AND indeks_gempa BETWEEN 0 AND 1"),
		database.Check("incaran", "luas", "luas_m2 >= 0"),
		database.Check("incaran", "harga", "harga_tanah_per_m2 >= 0"),
		database.Check("skor_dimensi", "nilai", "nilai BETWEEN 0 AND 100"),
		database.Unique("skor_dimensi", "incaran_dimensi", "incaran_id, dimensi_id"),
		database.Unique("perbandingan_item", "perbandingan_incaran", "perbandingan_id, incaran_id"),
	}
}
