package main

import (
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/inventory"
)

const (
	warehouseCentral  = "DEMO-GD-PST"
	warehouseGriya    = "DEMO-GD-GAB"
	warehouseHospital = "DEMO-GD-RSIA"

	materialCement     = "DEMO-SMN-50"
	materialRebar10    = "DEMO-BSI-D10"
	materialRebar13    = "DEMO-BSI-D13"
	materialReadymix   = "DEMO-RMX-K300"
	materialLightBrick = "DEMO-BTR-10"
	materialSand       = "DEMO-PSR-BTN"
	materialTile       = "DEMO-KRM-4040"
	materialPaint      = "DEMO-CAT-INT"
	materialPipe       = "DEMO-PVC-3"
	materialCable      = "DEMO-KBL-NYM"
	materialTruss      = "DEMO-BJR-C75"
)

type materialSeed struct {
	code     string
	name     string
	category string
	unit     string
	minimum  float64
	price    int64
}

type warehouseSeed struct {
	code    string
	name    string
	project string
}

type movementSeed struct {
	reference string
	day       int
	kind      string
	material  string
	quantity  float64
	source    string
	target    string
}

func demoMaterials() []materialSeed {
	return append(coreMaterials(), portfolioMaterials()...)
}

func coreMaterials() []materialSeed {
	return []materialSeed{
		{materialCement, "Semen Portland 50 kg", "Material struktur", "sak", 400, 72_000},
		{materialRebar10, "Besi beton ulir D10", "Material struktur", "batang", 300, 98_000},
		{materialRebar13, "Besi beton ulir D13", "Material struktur", "batang", 200, 165_000},
		{materialReadymix, "Beton readymix K-300", "Material struktur", "m3", 0, 1_150_000},
		{materialLightBrick, "Bata ringan 10 cm", "Material dinding", "m3", 40, 650_000},
		{materialSand, "Pasir beton", "Material struktur", "m3", 30, 285_000},
		{materialTile, "Keramik lantai 40×40", "Finishing", "m2", 300, 68_000},
		{materialPaint, "Cat tembok interior", "Finishing", "liter", 200, 38_000},
		{materialPipe, "Pipa PVC 3 inci", "Plumbing", "batang", 60, 115_000},
		{materialCable, "Kabel NYM 2×1,5 mm", "Elektrikal", "m", 500, 9_500},
		{materialTruss, "Baja ringan C75", "Rangka atap", "batang", 150, 92_000},
	}
}

func demoWarehouses() []warehouseSeed {
	return append(coreWarehouses(), portfolioWarehouses()...)
}

func coreWarehouses() []warehouseSeed {
	return []warehouseSeed{
		{warehouseCentral, "Gudang Pusat Cikarang", ""},
		{warehouseGriya, "Gudang Proyek Griya Asri", projectGriyaAsri},
		{warehouseHospital, "Gudang Proyek RSIA Cirebon", projectHospital},
	}
}

func demoMovements() []movementSeed {
	return append(coreMovements(), portfolioMovements()...)
}

func coreMovements() []movementSeed {
	opening := []movementSeed{
		{"DEMO-MUT-001", -300, "in", materialCement, 600, "", warehouseCentral},
		{"DEMO-MUT-002", -300, "in", materialRebar10, 500, "", warehouseCentral},
		{"DEMO-MUT-003", -300, "in", materialRebar13, 300, "", warehouseCentral},
		{"DEMO-MUT-004", -300, "in", materialLightBrick, 60, "", warehouseCentral},
		{"DEMO-MUT-005", -300, "in", materialSand, 50, "", warehouseCentral},
		{"DEMO-MUT-006", -300, "in", materialTile, 120, "", warehouseCentral},
		{"DEMO-MUT-007", -300, "in", materialPaint, 60, "", warehouseCentral},
		{"DEMO-MUT-008", -300, "in", materialPipe, 80, "", warehouseCentral},
		{"DEMO-MUT-009", -300, "in", materialCable, 400, "", warehouseCentral},
		{"DEMO-MUT-010", -300, "in", materialTruss, 200, "", warehouseCentral},
	}
	return append(opening,
		movementSeed{"DEMO-MUT-011", -120, "transfer", materialCement, 250, warehouseCentral, warehouseGriya},
		movementSeed{"DEMO-MUT-012", -95, "transfer", materialRebar13, 150, warehouseCentral, warehouseHospital},
		movementSeed{"DEMO-MUT-013", -30, "out", materialCement, 180, warehouseGriya, ""},
		movementSeed{"DEMO-MUT-014", -21, "out", materialRebar13, 60, warehouseHospital, ""},
		movementSeed{"DEMO-MUT-015", -7, "adjustment", materialSand, 2, warehouseCentral, ""},
		movementSeed{"DEMO-MUT-016", -4, "transfer", materialCable, 300, warehouseCentral, warehouseHospital},
	)
}

func (s *seeder) seedInventory() error {
	for _, seed := range demoMaterials() {
		if err := s.seedMaterial(seed); err != nil {
			return err
		}
	}
	for _, seed := range demoWarehouses() {
		if err := s.seedWarehouse(seed); err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) seedMaterial(seed materialSeed) error {
	unitID, err := s.reference("unit_of_measure", seed.unit)
	if err != nil {
		return err
	}
	id, _, err := s.ensure("material", map[string]any{"code": seed.code}, func() (uuid.UUID, error) {
		response, err := s.services.inventory.CreateMaterial(s.ctx, inventory.MaterialRequest{
			Code: seed.code, Name: seed.name, Category: seed.category, UnitOfMeasureID: unitID,
			MinimumStock: seed.minimum, LastPrice: seed.price,
		})
		return response.ID, err
	})
	if err != nil {
		return err
	}
	s.remember("material", seed.code, id)
	s.remember("material_unit", seed.code, unitID)
	return nil
}

func (s *seeder) seedWarehouse(seed warehouseSeed) error {
	projectID, err := s.knownPointer("project", seed.project)
	if err != nil {
		return err
	}
	id, _, err := s.ensure("warehouse", map[string]any{"code": seed.code}, func() (uuid.UUID, error) {
		response, err := s.services.inventory.CreateWarehouse(s.ctx, inventory.WarehouseRequest{
			Code: seed.code, Name: seed.name, ProjectID: projectID,
		})
		return response.ID, err
	})
	if err != nil {
		return err
	}
	s.remember("warehouse", seed.code, id)
	return nil
}

func (s *seeder) seedStockMovements() error {
	for _, seed := range demoMovements() {
		if err := s.seedMovement(seed); err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) seedMovement(seed movementSeed) error {
	materialID, err := s.known("material", seed.material)
	if err != nil {
		return err
	}
	source, err := s.knownPointer("warehouse", seed.source)
	if err != nil {
		return err
	}
	target, err := s.knownPointer("warehouse", seed.target)
	if err != nil {
		return err
	}
	_, _, err = s.ensure("stock_movement", map[string]any{"reference": seed.reference}, func() (uuid.UUID, error) {
		response, err := s.services.inventory.RecordStockMovement(s.ctx, inventory.StockMovementRequest{
			Date: s.day(seed.day), Type: seed.kind, MaterialID: materialID, Quantity: seed.quantity,
			SourceWarehouseID: source, TargetWarehouseID: target, Reference: seed.reference,
		})
		return response.ID, err
	})
	return err
}
