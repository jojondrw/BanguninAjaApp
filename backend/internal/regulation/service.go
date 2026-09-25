package regulation

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Service interface {
	GetByPoint(ctx context.Context, latitude, longitude float64) (*Response, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetByPoint(ctx context.Context, latitude, longitude float64) (*Response, error) {
	regulation, err := s.repo.GetByPoint(ctx, latitude, longitude)

	if err == nil {
		return &Response{
			ZoneName:       regulation.ZoneName,
			ZoneCode:       regulation.ZoneCode,
			SubZoneName:    regulation.SubZoneName,
			SubZoneCode:    regulation.SubZoneCode,
			District:       regulation.District,
			Village:        regulation.Village,
			KDB:            regulation.KDB,
			KLB:            regulation.KLB,
			KDH:            regulation.KDH,
			GSB:            regulation.GSB,
			MaxHeight:      regulation.MaxHeight,
			RegulationNote: regulation.RegulationNote,
			IsSimulated:    regulation.IsSimulated,
			Source:         regulation.Source,
		}, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return &Response{
		ZoneName:       "Kawasan Budidaya",
		ZoneCode:       "SIM",
		SubZoneName:    "Permukiman/Komersial",
		KDB:            60,
		KLB:            2.4,
		KDH:            30,
		MaxHeight:      "Simulasi",
		IsSimulated:    true,
		Source:         "Simulated - national land-use context",
		RegulationNote: "Data RDTR belum tersedia untuk lokasi ini; nilai bersifat simulasi dan tidak digunakan dalam scoring.",
	}, nil
}