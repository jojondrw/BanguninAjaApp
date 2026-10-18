package main

import (
	"fmt"
	"strings"
)

const (
	vendorPiling   = "DEMO-V09"
	vendorFacade   = "DEMO-V10"
	vendorLift     = "DEMO-V11"
	vendorSteelWF  = "DEMO-V12"
	vendorGeneral  = "DEMO-V13"
	vendorSurabaya = "DEMO-V14"
	vendorMakassar = "DEMO-V15"
	vendorRental   = "DEMO-V16"

	materialGlass     = "DEMO-KCA-8MM"
	materialWF        = "DEMO-WF-300"
	materialHardener  = "DEMO-FHD-25"
	materialACP       = "DEMO-ACP-4MM"
	materialSprinkler = "DEMO-SPR-1"
	materialPaving    = "DEMO-PVG-8"

	warehouseSky    = "DEMO-GD-SKY"
	warehouseSchool = "DEMO-GD-SBP"
	warehouseGLM    = "DEMO-GD-GLM"
	warehouseOffice = "DEMO-GD-KBK"
)

type activeProject struct {
	code     string
	label    string
	start    int
	end      int
	monthly  int64
	salesIn  int64
	overhead int64
}

func portfolioActivity() []activeProject {
	return append([]activeProject{
		{projectSkyGarden, "Sky Garden", -14, 0, 3_100_000_000, 1_450_000_000, 46_000_000},
		{projectTembalang, "Townhouse Tembalang", -7, 0, 1_250_000_000, 640_000_000, 21_000_000},
		{projectBukitHijau, "Bukit Hijau", -18, -6, 980_000_000, 1_150_000_000, 18_000_000},
		{projectSchool, "Sekolah Bina Prestasi", -5, 0, 820_000_000, 0, 0},
		{projectWarehouse, "Gudang Logistik Makassar", -9, 0, 1_700_000_000, 0, 0},
		{projectHotel, "Hotel Pantai Cirebon", -18, -2, 1_650_000_000, 0, 0},
		{projectOffice, "Kantor Bekasi Square", -3, 0, 2_200_000_000, 0, 0},
		{projectBuahBatu, "Ruko Buah Batu", -6, 0, 1_050_000_000, 760_000_000, 24_000_000},
	}, generatedActivity()...)
}

func portfolioOneOffCash() []cashSeed {
	return []cashSeed{
		{30, 4, cashIn, "3100", "", 150_000_000_000, "Setoran modal pemegang saham untuk ekspansi portofolio"},
		{ledgerMonths - 2, 15, cashIn, "2300", projectSkyGarden, 45_000_000_000, "Pencairan kredit konstruksi bank untuk Sky Garden"},
		{ledgerMonths - 9, 11, cashIn, "2300", projectTembalang, 12_000_000_000, "Pencairan kredit konstruksi bank untuk Townhouse Tembalang"},
		{3, 9, cashIn, "2300", projectOffice, 22_000_000_000, "Pencairan kredit konstruksi bank untuk Kantor Bekasi Square"},
		{ledgerMonths - 1, 20, cashOut, "5200", projectMall, 38_500_000_000, "Pembelian lahan Mal Bandung Timur"},
		{4, 6, cashOut, "5300", projectMall, 640_000_000, "Biaya kajian AMDAL dan PKKPR Mal Bandung Timur"},
		{2, 18, cashOut, "5300", projectOffice, 410_000_000, "Biaya retribusi PBG Kantor Bekasi Square"},
		{5, 22, cashOut, "5300", projectSchool, 125_000_000, "Biaya retribusi PBG Sekolah Bina Prestasi"},
		{1, 27, cashIn, "4200", projectBukitHijau, 185_000_000, "Pendapatan sewa kios fasum Bukit Hijau"},
	}
}

func (s *seeder) portfolioMonthlyCash(monthsAgo int) []cashSeed {
	label := s.monthLabel(monthsAgo)
	entries := []cashSeed{}
	for index, project := range portfolioActivity() {
		if -monthsAgo < project.start || -monthsAgo > project.end {
			continue
		}
		wave := int64((monthsAgo+index)%4) * project.monthly / 20
		entries = append(entries, cashSeed{monthsAgo, 6 + index%5, cashOut, "5100", project.code, project.monthly + wave, "Termin kontraktor " + project.label + " — " + label})
		if project.salesIn > 0 {
			entries = append(entries, cashSeed{monthsAgo, 14 + index%7, cashIn, "4100", project.code, project.salesIn + wave/2, "Penerimaan penjualan unit " + project.label + " — " + label})
		}
		if project.overhead > 0 {
			entries = append(entries, cashSeed{monthsAgo, 21, cashOut, "5400", project.code, project.overhead + wave/50, "Biaya pemasaran " + project.label + " — " + label})
		}
	}
	return entries
}

func portfolioVendors() []vendorSeed {
	return []vendorSeed{
		{vendorPiling, "PT Pancang Bumi Perkasa", "Subkontraktor pondasi", "031-5520-718", "09.012.345.6-611.000", 30, "good", true},
		{vendorFacade, "PT Kaca Fasad Indonesia", "Fasad & kaca", "021-4588-210", "10.123.456.7-021.000", 45, "good", true},
		{vendorLift, "PT Elevasi Teknik Mandiri", "Lift & eskalator", "021-6230-917", "11.234.567.8-031.000", 60, "fair", true},
		{vendorSteelWF, "PT Baja Struktur Celebes", "Besi & baja", "0411-872-330", "12.345.678.9-801.000", 30, "good", true},
		{vendorGeneral, "PT Wira Konstruksi Utama", "Kontraktor umum", "021-2988-114", "13.456.789.0-015.000", 30, "good", true},
		{vendorSurabaya, "CV Mitra Material Surabaya", "Material bangunan", "031-7740-552", "14.567.890.1-614.000", 21, "fair", true},
		{vendorMakassar, "UD Sinar Bangunan Makassar", "Material bangunan", "0411-455-091", "15.678.901.2-805.000", 14, "new", true},
		{vendorRental, "PT Sewa Alat Berat Nusantara", "Sewa alat berat", "021-8833-406", "16.789.012.3-407.000", 30, "poor", true},
	}
}

func portfolioMaterials() []materialSeed {
	return []materialSeed{
		{materialGlass, "Kaca tempered 8 mm", "Fasad", "m2", 120, 485_000},
		{materialWF, "Baja WF 300×150", "Material struktur", "batang", 40, 4_850_000},
		{materialHardener, "Floor hardener 25 kg", "Finishing", "sak", 150, 215_000},
		{materialACP, "Aluminium composite panel 4 mm", "Fasad", "lembar", 80, 690_000},
		{materialSprinkler, "Kepala sprinkler 1/2 inci", "Mekanikal", "buah", 300, 62_000},
		{materialPaving, "Paving block 8 cm", "Material jalan", "m2", 400, 98_000},
	}
}

func portfolioWarehouses() []warehouseSeed {
	return []warehouseSeed{
		{warehouseSky, "Gudang Proyek Sky Garden", projectSkyGarden},
		{warehouseSchool, "Gudang Proyek Sekolah Depok", projectSchool},
		{warehouseGLM, "Gudang Proyek Logistik Makassar", projectWarehouse},
		{warehouseOffice, "Gudang Proyek Bekasi Square", projectOffice},
	}
}

func portfolioRequests() []requestSeed {
	return []requestSeed{
		{"DEMO-PR-101", projectSkyGarden, -95, "completed", "Kaca fasad tower lantai 3 sampai 8", []lineSeed{{materialGlass, 1_200, 0}}},
		{"DEMO-PR-102", projectSkyGarden, -40, "approved", "Sprinkler lantai 1 sampai 6", []lineSeed{{materialSprinkler, 900, 0}, {materialPipe, 120, 0}}},
		{"DEMO-PR-103", projectSkyGarden, -6, "submitted", "Besi tambahan kolom lantai 16", []lineSeed{{materialRebar13, 450, 0}}},
		{"DEMO-PR-104", projectWarehouse, -160, "completed", "Baja WF rangka gudang", []lineSeed{{materialWF, 220, 0}}},
		{"DEMO-PR-105", projectWarehouse, -35, "approved", "Floor hardener zona B dan C", []lineSeed{{materialHardener, 600, 0}}},
		{"DEMO-PR-106", projectSchool, -50, "completed", "Semen dan besi struktur lantai 2", []lineSeed{{materialCement, 650, 0}, {materialRebar10, 420, 0}}},
		{"DEMO-PR-107", projectSchool, -4, "draft", "Keramik ruang kelas lantai 1", []lineSeed{{materialTile, 900, 0}}},
		{"DEMO-PR-108", projectOffice, -28, "approved", "Besi bore pile basement", []lineSeed{{materialRebar13, 800, 0}}},
		{"DEMO-PR-109", projectBuahBatu, -22, "completed", "ACP fasad ruko", []lineSeed{{materialACP, 260, 0}}},
		{"DEMO-PR-110", projectTembalang, -18, "approved", "Paving jalan lingkungan", []lineSeed{{materialPaving, 1_400, 0}}},
		{"DEMO-PR-111", projectTembalang, -9, "rejected", "Ditolak: harga keramik di atas anggaran", []lineSeed{{materialTile, 1_100, 0}}},
	}
}

func portfolioOrders() []orderSeed {
	return []orderSeed{
		{"DEMO-PO-101", vendorFacade, projectSkyGarden, "DEMO-PR-101", -90, 45, true, []lineSeed{{materialGlass, 1_200, 470_000}}},
		{"DEMO-PO-102", vendorSurabaya, projectSkyGarden, "DEMO-PR-102", -36, 21, true, []lineSeed{{materialSprinkler, 900, 60_000}, {materialPipe, 120, 112_000}}},
		{"DEMO-PO-103", vendorSteelWF, projectWarehouse, "DEMO-PR-104", -155, 30, true, []lineSeed{{materialWF, 220, 4_800_000}}},
		{"DEMO-PO-104", vendorMakassar, projectWarehouse, "DEMO-PR-105", -30, 14, true, []lineSeed{{materialHardener, 600, 212_000}}},
		{"DEMO-PO-105", vendorBuilding, projectSchool, "DEMO-PR-106", -46, 30, true, []lineSeed{{materialCement, 650, 73_000}, {materialRebar10, 420, 99_000}}},
		{"DEMO-PO-106", vendorSteel, projectOffice, "DEMO-PR-108", -24, 45, true, []lineSeed{{materialRebar13, 800, 168_000}}},
		{"DEMO-PO-107", vendorFacade, projectBuahBatu, "DEMO-PR-109", -19, 45, true, []lineSeed{{materialACP, 260, 685_000}}},
		{"DEMO-PO-108", vendorBuilding, projectTembalang, "DEMO-PR-110", -14, 30, false, []lineSeed{{materialPaving, 1_400, 96_000}}},
	}
}

func portfolioReceipts() []receiptSeed {
	return []receiptSeed{
		{"DEMO-GR-101", "DEMO-PO-101", warehouseSky, -70, "good", "Kaca fasad batch 1", []receiptLineSeed{{materialGlass, 700, 0}}},
		{"DEMO-GR-102", "DEMO-PO-101", warehouseSky, -48, "partially_damaged", "12 lembar kaca retak saat bongkar muat", []receiptLineSeed{{materialGlass, 488, 12}}},
		{"DEMO-GR-103", "DEMO-PO-102", warehouseSky, -20, "good", "Sprinkler dan pipa lengkap", []receiptLineSeed{{materialSprinkler, 900, 0}, {materialPipe, 120, 0}}},
		{"DEMO-GR-104", "DEMO-PO-103", warehouseGLM, -140, "good", "Baja WF lengkap dengan sertifikat pabrik", []receiptLineSeed{{materialWF, 220, 0}}},
		{"DEMO-GR-105", "DEMO-PO-104", warehouseGLM, -22, "good", "Floor hardener batch 1", []receiptLineSeed{{materialHardener, 400, 0}}},
		{"DEMO-GR-106", "DEMO-PO-105", warehouseSchool, -40, "good", "Semen dan besi diterima", []receiptLineSeed{{materialCement, 650, 0}, {materialRebar10, 420, 0}}},
		{"DEMO-GR-107", "DEMO-PO-106", warehouseOffice, -15, "good", "Besi D13 batch 1", []receiptLineSeed{{materialRebar13, 500, 0}}},
		{"DEMO-GR-108", "DEMO-PO-107", warehouseCentral, -10, "good", "ACP diterima di gudang pusat", []receiptLineSeed{{materialACP, 260, 0}}},
	}
}

func portfolioMovements() []movementSeed {
	return []movementSeed{
		{"DEMO-MUT-101", -60, "out", materialGlass, 650, warehouseSky, ""},
		{"DEMO-MUT-102", -12, "out", materialSprinkler, 420, warehouseSky, ""},
		{"DEMO-MUT-103", -120, "out", materialWF, 210, warehouseGLM, ""},
		{"DEMO-MUT-104", -14, "out", materialHardener, 310, warehouseGLM, ""},
		{"DEMO-MUT-105", -30, "out", materialCement, 520, warehouseSchool, ""},
		{"DEMO-MUT-106", -8, "out", materialACP, 180, warehouseCentral, ""},
		{"DEMO-MUT-107", -5, "out", materialRebar13, 260, warehouseOffice, ""},
		{"DEMO-MUT-108", -3, "adjustment", materialGlass, 3, warehouseSky, ""},
	}
}

func portfolioPayables() []payableSeed {
	return []payableSeed{
		{"DEMO-HUT-101", vendorFacade, projectSkyGarden, "DEMO-PO-101", -45, 564_000_000, []paymentSeed{
			{300_000_000, -60, "transfer", "TRF-MDR-550127", "Pembayaran batch 1 kaca fasad"},
			{264_000_000, -44, "transfer", "TRF-MDR-551890", "Pelunasan kaca fasad"},
		}},
		{"DEMO-HUT-102", vendorSurabaya, projectSkyGarden, "DEMO-PO-102", -15, 67_440_000, nil},
		{"DEMO-HUT-103", vendorSteelWF, projectWarehouse, "DEMO-PO-103", -125, 1_056_000_000, []paymentSeed{
			{1_056_000_000, -126, "transfer", "TRF-BNI-310442", "Pelunasan baja WF"},
		}},
		{"DEMO-HUT-104", vendorMakassar, projectWarehouse, "DEMO-PO-104", -16, 127_200_000, []paymentSeed{
			{60_000_000, -18, "cek", "GIRO-BRI-77120", "Pembayaran sebagian floor hardener"},
		}},
		{"DEMO-HUT-105", vendorBuilding, projectSchool, "DEMO-PO-105", 4, 89_030_000, nil},
		{"DEMO-HUT-106", vendorSteel, projectOffice, "DEMO-PO-106", 21, 134_400_000, nil},
		{"DEMO-HUT-107", vendorFacade, projectBuahBatu, "DEMO-PO-107", 26, 178_100_000, nil},
		{"DEMO-HUT-108", vendorRental, projectOffice, "", -6, 96_000_000, nil},
	}
}

func portfolioConstructionContracts() []constructionSeed {
	return []constructionSeed{
		{projectSkyGarden, vendorGeneral, "DEMO-SPK-SKY-01", 82_000_000_000, 5, []termSeed{
			{0, 10, true}, {20, 15, true}, {40, 20, false}, {60, 20, false}, {80, 20, false}, {100, 15, false},
		}, []changeSeed{
			{"Penambahan sky lounge di lantai atap tower", 2_850_000_000, 45, "approved", -60, -48},
		}},
		{projectTembalang, vendorHousing, "DEMO-SPK-TBL-01", 21_600_000_000, 5, []termSeed{
			{25, 25, true}, {50, 25, false}, {75, 25, false}, {100, 25, false},
		}, nil},
		{projectBukitHijau, vendorHousing, "DEMO-SPK-BHB-01", 17_400_000_000, 5, []termSeed{
			{30, 30, true}, {60, 30, true}, {100, 40, true},
		}, nil},
		{projectSchool, vendorGeneral, "DEMO-SPK-SBP-01", 10_900_000_000, 5, []termSeed{
			{0, 10, true}, {30, 30, true}, {60, 30, false}, {100, 30, false},
		}, []changeSeed{
			{"Tambahan ruang laboratorium komputer di lantai 3", 480_000_000, 14, "", -5, 0},
		}},
		{projectWarehouse, vendorSteelWF, "DEMO-SPK-GLM-01", 19_200_000_000, 5, []termSeed{
			{0, 10, true}, {30, 30, true}, {60, 30, true}, {100, 30, false},
		}, nil},
		{projectHotel, vendorStructure, "DEMO-SPK-HPC-01", 41_500_000_000, 5, []termSeed{
			{0, 10, true}, {25, 20, true}, {50, 20, true}, {75, 25, true}, {100, 25, true},
		}, []changeSeed{
			{"Upgrade sistem tata udara kamar ke VRF", 1_120_000_000, 21, "approved", -420, -405},
		}},
		{projectOffice, vendorPiling, "DEMO-SPK-KBK-01", 9_400_000_000, 5, []termSeed{
			{0, 10, true}, {50, 40, false}, {100, 50, false},
		}, nil},
		{projectBuahBatu, vendorHousing, "DEMO-SPK-RBB-01", 13_200_000_000, 5, []termSeed{
			{0, 10, true}, {35, 30, true}, {70, 30, false}, {100, 30, false},
		}, nil},
	}
}

func portfolioOwnerContracts() []ownerSeed {
	return []ownerSeed{
		{projectSchool, schoolClient, "DEMO-KTR-SBP-01", 14_800_000_000, 5, 30, []termSeed{
			{0, 10, true}, {25, 20, true}, {50, 20, false}, {75, 25, false}, {100, 25, false},
		}, false},
		{projectWarehouse, warehouseClient, "DEMO-KTR-GLM-01", 26_300_000_000, 5, 45, []termSeed{
			{0, 10, true}, {30, 30, true}, {60, 30, true}, {100, 30, false},
		}, false},
		{projectHotel, hotelClient, "DEMO-KTR-HPC-01", 58_600_000_000, 5, 30, []termSeed{
			{0, 10, true}, {25, 20, true}, {50, 20, true}, {75, 25, true}, {100, 25, true},
		}, false},
		{projectOffice, officeClient, "DEMO-KTR-KBK-01", 72_000_000_000, 5, 30, []termSeed{
			{0, 10, true}, {25, 15, false}, {50, 25, false}, {75, 25, false}, {100, 25, false},
		}, false},
	}
}

func portfolioAssets() []assetSeed {
	return []assetSeed{
		{"DEMO-AST-101", "Tower crane 8 ton", "Alat berat", -16, 3_100_000_000, 120, "in_use"},
		{"DEMO-AST-102", "Concrete pump 42 m", "Alat konstruksi", -20, 2_650_000_000, 96, "in_use"},
		{"DEMO-AST-103", "Bore pile rig", "Alat berat", -8, 4_200_000_000, 120, "in_use"},
		{"DEMO-AST-104", "Mobile crane 50 ton", "Alat berat", -26, 5_600_000_000, 120, "in_use"},
		{"DEMO-AST-105", "Kantor proyek kontainer 6 unit", "Bangunan", -14, 360_000_000, 48, "in_use"},
		{"DEMO-AST-106", "Mobil operasional Innova", "Kendaraan", -22, 410_000_000, 60, "in_use"},
		{"DEMO-AST-107", "Showroom Sky Garden", "Bangunan", -15, 1_250_000_000, 60, "in_use"},
		{"DEMO-AST-108", "Power trowel 2 unit", "Alat konstruksi", -10, 96_000_000, 48, "maintenance"},
	}
}

func portfolioEquipment() []equipmentSeed {
	return []equipmentSeed{
		{"DEMO-EQP-101", "Tower crane TC-SKY-01", "DEMO-AST-101", projectSkyGarden, 2_840, serviceIn(9), "operating"},
		{"DEMO-EQP-102", "Concrete pump CP-02", "DEMO-AST-102", projectSkyGarden, 1_920, serviceIn(16), "operating"},
		{"DEMO-EQP-103", "Bore pile rig BP-01", "DEMO-AST-103", projectOffice, 640, serviceIn(25), "operating"},
		{"DEMO-EQP-104", "Mobile crane MC-50", "DEMO-AST-104", projectWarehouse, 3_380, serviceIn(-2), "maintenance"},
		{"DEMO-EQP-105", "Power trowel PT-01", "DEMO-AST-108", projectWarehouse, 870, serviceIn(4), "maintenance"},
		{"DEMO-EQP-106", "Passenger hoist PH-01", "", projectSkyGarden, 1_150, serviceIn(30), "operating"},
		{"DEMO-EQP-107", "Molen beton 350 L #2", "", projectSchool, 520, serviceIn(45), "operating"},
		{"DEMO-EQP-108", "Pompa dewatering 6 unit", "", projectOffice, 1_760, serviceIn(7), "operating"},
		{"DEMO-EQP-109", "Excavator mini PC-30", "", projectTembalang, 940, serviceIn(18), "operating"},
		{"DEMO-EQP-110", "Scaffolding 300 set", "", projectBuahBatu, 0, nil, "idle"},
	}
}

func portfolioSites() []siteSeed {
	return []siteSeed{
		{"Kandidat apartemen Mayjen Sungkono, Surabaya", "3578", "housing", projectSkyGarden, -7.2905, 112.7176, 6_800, 18_500_000, 0.21, 0.18, []int{74, 88, 83, 79, 71}},
		{"Kandidat townhouse Tembalang, Semarang", "3374", "housing", projectTembalang, -7.0528, 110.4386, 16_000, 3_900_000, 0.16, 0.22, []int{77, 74, 81, 70, 72}},
		{"Kandidat perumahan Bukit Hijau, Bogor", "3271", "housing", projectBukitHijau, -6.5658, 106.8064, 15_000, 4_600_000, 0.12, 0.33, []int{82, 73, 78, 69, 74}},
		{"Kandidat mal Soekarno-Hatta, Bandung Timur", "3273", "mall", projectMall, -6.9408, 107.6720, 32_000, 8_900_000, 0.38, 0.29, []int{66, 84, 82, 76, 61}},
		{"Kandidat sekolah Cimanggis, Depok", "3276", "housing", projectSchool, -6.3836, 106.8829, 7_500, 6_200_000, 0.19, 0.23, []int{78, 80, 85, 66, 63}},
		{"Kandidat gudang Kawasan Industri Makassar", "7371", "mall", projectWarehouse, -5.1102, 119.4876, 24_000, 2_400_000, 0.44, 0.11, []int{61, 79, 64, 81, 77}},
		{"Kandidat hotel pesisir Kejaksan, Cirebon", "3274", "entertainment", projectHotel, -6.7158, 108.5704, 5_400, 6_800_000, 0.52, 0.18, []int{58, 81, 72, 78, 70}},
		{"Kandidat kantor Ahmad Yani, Bekasi", "3275", "mall", projectOffice, -6.2475, 107.0011, 4_800, 21_000_000, 0.33, 0.21, []int{69, 90, 84, 80, 66}},
		{"Kandidat ruko Buah Batu, Bandung", "3273", "mall", projectBuahBatu, -6.9476, 107.6338, 3_200, 9_800_000, 0.27, 0.31, []int{73, 82, 80, 77, 68}},
		{"Kandidat perumahan Pakuwon, Surabaya Timur", "3578", "housing", "", -7.2732, 112.7999, 28_000, 7_200_000, 0.47, 0.17, []int{60, 78, 79, 73, 65}},
		{"Kandidat apartemen Simpang Lima, Semarang", "3374", "housing", "", -6.9903, 110.4229, 3_900, 16_400_000, 0.36, 0.20, []int{65, 89, 83, 74, 59}},
	}
}

func ownerReference(number string, position int) string {
	return fmt.Sprintf("TRF-%s-%d", strings.TrimPrefix(number, "DEMO-KTR-"), position)
}
