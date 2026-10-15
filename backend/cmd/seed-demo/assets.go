package main

import (
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/asset"
)

type assetSeed struct {
	code       string
	name       string
	category   string
	acquired   int
	value      int64
	lifeMonths int
	status     string
}

type equipmentSeed struct {
	code        string
	name        string
	asset       string
	project     string
	hours       int
	serviceDays *int
	status      string
}

func serviceIn(days int) *int {
	return &days
}

func demoAssets() []assetSeed {
	return []assetSeed{
		{"DEMO-AST-001", "Excavator 20 ton", "Alat berat", -30, 1_850_000_000, 96, "in_use"},
		{"DEMO-AST-002", "Dump truck 8 m³", "Kendaraan", -24, 620_000_000, 60, "in_use"},
		{"DEMO-AST-003", "Concrete mixer 350 liter", "Alat konstruksi", -18, 48_000_000, 48, "maintenance"},
		{"DEMO-AST-004", "Tower crane 6 ton", "Alat berat", -14, 2_400_000_000, 120, "in_use"},
		{"DEMO-AST-005", "Mobil pikap operasional", "Kendaraan", -40, 235_000_000, 60, "for_sale"},
		{"DEMO-AST-006", "Genset 100 kVA", "Alat konstruksi", -70, 310_000_000, 60, "disposed"},
		{"DEMO-AST-007", "Kantor pemasaran Griya Asri", "Bangunan", -11, 420_000_000, 60, "in_use"},
	}
}

func demoEquipment() []equipmentSeed {
	return []equipmentSeed{
		{"DEMO-EQP-001", "Excavator unit 1", "DEMO-AST-001", projectGriyaAsri, 3_420, serviceIn(12), "operating"},
		{"DEMO-EQP-002", "Dump truck D 9123 KX", "DEMO-AST-002", projectGriyaAsri, 2_180, serviceIn(30), "operating"},
		{"DEMO-EQP-003", "Concrete mixer #1", "DEMO-AST-003", projectHospital, 1_150, serviceIn(3), "maintenance"},
		{"DEMO-EQP-004", "Tower crane TC-01", "DEMO-AST-004", projectHospital, 1_960, serviceIn(20), "operating"},
		{"DEMO-EQP-005", "Vibrator beton", "", projectHospital, 640, serviceIn(-5), "broken"},
		{"DEMO-EQP-006", "Stamper kodok", "", projectCluster, 880, nil, "idle"},
		{"DEMO-EQP-007", "Scaffolding 200 set", "", projectGriyaAsri, 0, nil, "idle"},
	}
}

func (s *seeder) seedAssets() error {
	for _, seed := range demoAssets() {
		if err := s.seedAsset(seed); err != nil {
			return err
		}
	}
	for _, seed := range demoEquipment() {
		if err := s.seedEquipment(seed); err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) seedAsset(seed assetSeed) error {
	id, _, err := s.ensure("asset", map[string]any{"code": seed.code}, func() (uuid.UUID, error) {
		response, err := s.services.assets.CreateAsset(s.ctx, asset.AssetRequest{
			Code: seed.code, Name: seed.name, Category: seed.category,
			AcquisitionDate: s.monthOffset(seed.acquired), AcquisitionValue: seed.value,
			AccumulatedDepreciation: depreciationOf(seed), UsefulLifeMonths: seed.lifeMonths, Status: seed.status,
		})
		return response.ID, err
	})
	if err != nil {
		return err
	}
	s.remember("asset", seed.code, id)
	return nil
}

func depreciationOf(seed assetSeed) int64 {
	accumulated := seed.value * int64(-seed.acquired) / int64(seed.lifeMonths)
	return min(accumulated, seed.value)
}

func (s *seeder) seedEquipment(seed equipmentSeed) error {
	assetID, err := s.knownPointer("asset", seed.asset)
	if err != nil {
		return err
	}
	projectID, err := s.knownPointer("project", seed.project)
	if err != nil {
		return err
	}
	request := asset.EquipmentRequest{
		Code: seed.code, Name: seed.name, AssetID: assetID, ProjectID: projectID,
		OperatingHours: seed.hours, Status: seed.status,
	}
	if seed.serviceDays != nil {
		request.NextServiceDate = s.dayPointer(*seed.serviceDays)
	}
	_, _, err = s.ensure("equipment", map[string]any{"code": seed.code}, func() (uuid.UUID, error) {
		response, err := s.services.assets.CreateEquipment(s.ctx, request)
		return response.ID, err
	})
	return err
}
