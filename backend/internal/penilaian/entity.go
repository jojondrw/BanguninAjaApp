package penilaian

import (
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

type Dimensi struct {
	entity.Base
	Kode      string `gorm:"type:varchar(30);not null;uniqueIndex:uq_dimensi_kode"`
	Nama      string `gorm:"type:varchar(120);not null"`
	Deskripsi string `gorm:"type:text"`
	Urutan    int    `gorm:"not null;default:1"`
}

func (Dimensi) TableName() string {
	return "dimensi"
}

type ProfilBangunan struct {
	entity.Base
	Kode string `gorm:"type:varchar(30);not null;uniqueIndex:uq_profil_bangunan_kode"`
	Nama string `gorm:"type:varchar(120);not null"`
}

func (ProfilBangunan) TableName() string {
	return "profil_bangunan"
}

type Bobot struct {
	entity.Base
	ProfilBangunanID uuid.UUID `gorm:"type:uuid;not null;index:idx_bobot_profil"`
	DimensiID        uuid.UUID `gorm:"type:uuid;not null;index:idx_bobot_dimensi"`
	Persen           int       `gorm:"not null;default:0"`
}

func (Bobot) TableName() string {
	return "bobot"
}

func Entities() []any {
	return []any{&Dimensi{}, &ProfilBangunan{}, &Bobot{}}
}

func Indexes() []string {
	return []string{}
}

func Constraints() []string {
	return []string{
		database.ForeignKey("bobot", "profil_bangunan_id", "profil_bangunan", database.DeleteCascade),
		database.ForeignKey("bobot", "dimensi_id", "dimensi", database.DeleteCascade),
		database.Check("bobot", "persen", "persen BETWEEN 0 AND 100"),
		database.Unique("bobot", "profil_dimensi", "profil_bangunan_id, dimensi_id"),
	}
}
