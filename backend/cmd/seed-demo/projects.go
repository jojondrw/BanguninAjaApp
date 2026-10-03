package main

import (
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/project"
)

const (
	projectGriyaAsri = "DEMO-GAB"
	projectRuko      = "DEMO-RSM"
	projectHospital  = "DEMO-RSIA"
	projectCluster   = "DEMO-CLD"
)

type phaseSeed struct {
	name       string
	startMonth int
	endMonth   int
	progress   int
}

type budgetItemSeed struct {
	code        string
	description string
	volume      float64
	unit        string
	unitPrice   int64
}

type permitSeed struct {
	kind       string
	number     string
	issuedDay  int
	validYears int
	status     string
}

type projectSeed struct {
	code          string
	name          string
	region        string
	kind          string
	status        string
	startMonth    int
	endMonth      int
	contractValue int64
	phases        []phaseSeed
	items         []budgetItemSeed
	permits       []permitSeed
}

var projectStatusPath = map[string][]string{
	"planning":  nil,
	"ongoing":   {"ongoing"},
	"completed": {"ongoing", "completed"},
}

func demoProjects() []projectSeed {
	return []projectSeed{
		{
			code: projectGriyaAsri, name: "Perumahan Griya Asri Bandung", region: "3273", kind: "Perumahan",
			status: "ongoing", startMonth: -11, endMonth: 9, contractValue: 48_500_000_000,
			phases: []phaseSeed{
				{"Pematangan lahan & cut and fill", -11, -8, 100},
				{"Jalan lingkungan & drainase", -9, -3, 85},
				{"Pembangunan unit tahap 1 (Blok A)", -7, 1, 70},
				{"Pembangunan unit tahap 2 (Blok B)", -2, 6, 25},
				{"Fasilitas umum & taman", 4, 9, 0},
			},
			items: []budgetItemSeed{
				{"1.1", "Pembersihan & pematangan lahan", 42_000, "m2", 45_000},
				{"2.1", "Pekerjaan jalan paving block", 6_800, "m2", 185_000},
				{"2.2", "Saluran drainase beton U-ditch", 1_450, "m", 650_000},
				{"3.1", "Pembangunan rumah tipe 36/72", 48, "unit", 245_000_000},
				{"3.2", "Pembangunan rumah tipe 45/90", 36, "unit", 315_000_000},
				{"4.1", "Instalasi listrik & air bersih kawasan", 1, "ls", 1_350_000_000},
			},
			permits: []permitSeed{
				{"PKKPR", "PKKPR-3273-0412-2025", -340, 3, "issued"},
				{"Persetujuan Bangunan Gedung (PBG)", "PBG-327301-0188-2025", -300, 5, "issued"},
				{"Persetujuan Lingkungan (UKL-UPL)", "UKL-UPL/DLH-BDG/077/2025", -320, 5, "issued"},
				{"Sertifikat Laik Fungsi (SLF) Blok A", "SLF-327301-PENGAJUAN-014", -15, 0, "submitted"},
			},
		},
		{
			code: projectRuko, name: "Ruko Sentra Makassar", region: "7371", kind: "Ruko",
			status: "planning", startMonth: 2, endMonth: 16, contractValue: 22_750_000_000,
			phases: []phaseSeed{
				{"Studi kelayakan & desain", -4, 1, 60},
				{"Pengurusan izin", -2, 2, 30},
				{"Konstruksi struktur", 2, 10, 0},
				{"Finishing & serah terima", 10, 16, 0},
			},
			items: []budgetItemSeed{
				{"1.1", "Pekerjaan persiapan & pagar proyek", 1, "ls", 285_000_000},
				{"2.1", "Struktur beton ruko 3 lantai", 1_850, "m3", 4_250_000},
				{"3.1", "Fasad & pekerjaan arsitektur", 2_400, "m2", 1_150_000},
			},
			permits: []permitSeed{
				{"PKKPR", "PKKPR-7371-PENGAJUAN-031", -40, 0, "submitted"},
				{"Persetujuan Bangunan Gedung (PBG)", "PBG-737101-PENGAJUAN-009", -10, 0, "submitted"},
			},
		},
		{
			code: projectHospital, name: "Rumah Sakit Ibu & Anak Cirebon", region: "3274", kind: "Rumah Sakit",
			status: "ongoing", startMonth: -9, endMonth: 8, contractValue: 67_000_000_000,
			phases: []phaseSeed{
				{"Pekerjaan persiapan", -9, -8, 100},
				{"Pondasi bore pile & basement", -8, -5, 100},
				{"Struktur lantai 1–4", -5, 1, 55},
				{"MEP & instalasi gas medis", -1, 5, 10},
				{"Arsitektur & interior", 3, 8, 0},
			},
			items: []budgetItemSeed{
				{"1.1", "Pondasi bore pile diameter 60 cm", 2_160, "m", 1_450_000},
				{"2.1", "Beton struktur K-350", 4_800, "m3", 1_380_000},
				{"2.2", "Pembesian struktur", 620, "ton", 16_500_000},
				{"3.1", "Instalasi gas medis", 1, "ls", 3_850_000_000},
				{"3.2", "Instalasi listrik & genset", 1, "ls", 5_200_000_000},
			},
			permits: []permitSeed{
				{"Persetujuan Bangunan Gedung (PBG)", "PBG-327401-0051-2025", -280, 5, "issued"},
				{"Persetujuan Lingkungan (AMDAL)", "AMDAL/DLH-CRB/019/2025", -300, 5, "issued"},
				{"Rekomendasi Dinas Kesehatan", "DINKES-CRB/REK/442/2026", -60, 2, "issued"},
			},
		},
		{
			code: projectCluster, name: "Cluster Lama Depok", region: "3276", kind: "Perumahan",
			status: "completed", startMonth: -26, endMonth: -3, contractValue: 18_900_000_000,
			phases: []phaseSeed{
				{"Pematangan lahan", -26, -22, 100},
				{"Pembangunan 24 unit", -22, -6, 100},
				{"Serah terima & perbaikan", -6, -3, 100},
			},
			items: []budgetItemSeed{
				{"1.1", "Pematangan lahan", 9_600, "m2", 55_000},
				{"2.1", "Pembangunan rumah tipe 45/84", 24, "unit", 335_000_000},
			},
			permits: []permitSeed{
				{"Persetujuan Bangunan Gedung (PBG)", "PBG-327601-0233-2024", -760, 5, "issued"},
				{"Sertifikat Laik Fungsi (SLF)", "SLF-327601-0098-2026", -100, 5, "issued"},
			},
		},
	}
}

func (s *seeder) seedProjects() error {
	for _, seed := range demoProjects() {
		if err := s.seedProject(seed); err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) seedProject(seed projectSeed) error {
	regionID, err := s.reference("region", seed.region)
	if err != nil {
		return err
	}
	id, created, err := s.ensure("project", map[string]any{"code": seed.code}, func() (uuid.UUID, error) {
		response, err := s.services.projects.CreateProject(s.ctx, project.ProjectRequest{
			Code: seed.code, Name: seed.name, RegionID: &regionID, Type: seed.kind,
			StartDate: s.monthOffset(seed.startMonth), TargetEndDate: s.monthOffset(seed.endMonth),
			ContractValue: seed.contractValue,
		})
		return response.ID, err
	})
	if err != nil {
		return err
	}
	s.remember("project", seed.code, id)
	if err := s.seedProjectDetails(id, seed); err != nil {
		return err
	}
	if !created {
		return nil
	}
	return s.advanceProject(id, seed.status)
}

func (s *seeder) seedProjectDetails(id uuid.UUID, seed projectSeed) error {
	for index, phase := range seed.phases {
		if err := s.seedPhase(id, phase, index+1); err != nil {
			return err
		}
	}
	for _, item := range seed.items {
		if err := s.seedBudgetItem(id, item); err != nil {
			return err
		}
	}
	for _, permit := range seed.permits {
		if err := s.seedPermit(id, permit); err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) seedPhase(projectID uuid.UUID, seed phaseSeed, order int) error {
	_, _, err := s.ensure("project_phase", map[string]any{"project_id": projectID, "name": seed.name}, func() (uuid.UUID, error) {
		response, err := s.services.projects.CreatePhase(s.ctx, projectID, project.PhaseRequest{
			Name: seed.name, SortOrder: order, Progress: seed.progress,
			StartDate: s.monthOffset(seed.startMonth), TargetEndDate: s.monthOffset(seed.endMonth),
		})
		return response.ID, err
	})
	return err
}

func (s *seeder) seedBudgetItem(projectID uuid.UUID, seed budgetItemSeed) error {
	unitID, err := s.reference("unit_of_measure", seed.unit)
	if err != nil {
		return err
	}
	_, _, err = s.ensure("budget_item", map[string]any{"project_id": projectID, "code": seed.code}, func() (uuid.UUID, error) {
		response, err := s.services.projects.CreateBudgetItem(s.ctx, projectID, project.BudgetItemRequest{
			Code: seed.code, Description: seed.description, Volume: seed.volume,
			UnitOfMeasureID: unitID, UnitPrice: seed.unitPrice,
		})
		return response.ID, err
	})
	return err
}

func (s *seeder) seedPermit(projectID uuid.UUID, seed permitSeed) error {
	request := project.PermitRequest{Type: seed.kind, Number: seed.number, Status: seed.status}
	if seed.status == "issued" {
		request.IssuedDate = s.dayPointer(seed.issuedDay)
		validUntil := s.day(seed.issuedDay).AddDate(seed.validYears, 0, 0)
		request.ValidUntil = &validUntil
	}
	_, _, err := s.ensure("permit", map[string]any{"project_id": projectID, "number": seed.number}, func() (uuid.UUID, error) {
		response, err := s.services.projects.CreatePermit(s.ctx, projectID, request)
		return response.ID, err
	})
	return err
}

func (s *seeder) advanceProject(id uuid.UUID, status string) error {
	for _, next := range projectStatusPath[status] {
		if _, err := s.services.projects.UpdateProjectStatus(s.ctx, id, project.ProjectStatusRequest{Status: next}); err != nil {
			return err
		}
	}
	return nil
}
