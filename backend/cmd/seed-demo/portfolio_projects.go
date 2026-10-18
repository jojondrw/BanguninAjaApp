package main

const (
	projectSkyGarden  = "DEMO-SKY"
	projectTembalang  = "DEMO-TBL"
	projectBukitHijau = "DEMO-BHB"
	projectMall       = "DEMO-MBT"
	projectSchool     = "DEMO-SBP"
	projectWarehouse  = "DEMO-GLM"
	projectHotel      = "DEMO-HPC"
	projectOffice     = "DEMO-KBK"
	projectBuahBatu   = "DEMO-RBB"
)

func portfolioProjects() []projectSeed {
	return []projectSeed{
		{
			code: projectSkyGarden, name: "Apartemen Sky Garden Surabaya", region: "3578", kind: "Apartemen",
			status: "ongoing", startMonth: -14, endMonth: 14, contractValue: 128_000_000_000,
			phases: []phaseSeed{
				{"Pondasi tiang pancang & basement", -14, -8, 100},
				{"Struktur tower lantai 1–12", -9, -1, 90},
				{"Struktur tower lantai 13–24", -2, 6, 20},
				{"MEP & fasad kaca", 1, 11, 0},
				{"Interior unit & serah terima", 8, 14, 0},
			},
			items: []budgetItemSeed{
				{"1.1", "Tiang pancang spun pile 50 cm", 4_200, "m", 950_000, 0},
				{"2.1", "Beton struktur K-400 lantai 1–12", 9_800, "m3", 1_450_000, 1},
				{"2.2", "Beton struktur K-400 lantai 13–24", 8_700, "m3", 1_450_000, 2},
				{"3.1", "Fasad curtain wall kaca", 9_600, "m2", 2_350_000, 3},
				{"3.2", "Lift penumpang 6 unit", 6, "unit", 1_850_000_000, 3},
				{"4.1", "Interior unit apartemen", 480, "unit", 38_000_000, 4},
			},
			permits: []permitSeed{
				{"Persetujuan Lingkungan (AMDAL)", "AMDAL/DLH-SBY/031/2025", -450, 5, "issued"},
				{"Persetujuan Bangunan Gedung (PBG)", "PBG-357801-0412-2025", -430, 5, "issued"},
				{"Sertifikat Laik Fungsi (SLF) Tower A", "SLF-357801-PENGAJUAN-022", -5, 0, "submitted"},
			},
		},
		{
			code: projectTembalang, name: "Townhouse Tembalang Semarang", region: "3374", kind: "Perumahan",
			status: "ongoing", startMonth: -7, endMonth: 9, contractValue: 34_500_000_000,
			phases: []phaseSeed{
				{"Pematangan lahan & talud", -7, -5, 100},
				{"Infrastruktur jalan & drainase", -5, -1, 80},
				{"Pembangunan 24 unit townhouse", -3, 6, 35},
				{"Clubhouse & taman", 5, 9, 0},
			},
			items: []budgetItemSeed{
				{"1.1", "Cut and fill & talud batu kali", 8_400, "m3", 210_000, 0},
				{"2.1", "Jalan beton K-250", 3_200, "m2", 320_000, 1},
				{"3.1", "Townhouse 2 lantai tipe 70/90", 24, "unit", 780_000_000, 2},
				{"4.1", "Clubhouse & kolam renang", 1, "ls", 2_150_000_000, 3},
			},
			permits: []permitSeed{
				{"PKKPR", "PKKPR-3374-0219-2025", -230, 3, "issued"},
				{"Persetujuan Bangunan Gedung (PBG)", "PBG-337401-0307-2025", -200, 5, "issued"},
			},
		},
		{
			code: projectBukitHijau, name: "Perumahan Bukit Hijau Bogor", region: "3271", kind: "Perumahan",
			status: "completed", startMonth: -32, endMonth: -6, contractValue: 41_200_000_000,
			phases: []phaseSeed{
				{"Pematangan lahan", -32, -28, 100},
				{"Pembangunan 30 unit", -28, -10, 100},
				{"Serah terima & fasilitas umum", -10, -6, 100},
			},
			items: []budgetItemSeed{
				{"1.1", "Pematangan lahan", 15_000, "m2", 60_000, 0},
				{"2.1", "Rumah tipe 54/120", 30, "unit", 520_000_000, 1},
				{"3.1", "Fasilitas umum & masjid", 1, "ls", 1_450_000_000, 2},
			},
			permits: []permitSeed{
				{"Persetujuan Bangunan Gedung (PBG)", "PBG-327101-0098-2024", -960, 5, "issued"},
				{"Sertifikat Laik Fungsi (SLF)", "SLF-327101-0041-2026", -190, 5, "issued"},
			},
		},
		{
			code: projectMall, name: "Mal Bandung Timur", region: "3273", kind: "Mall",
			status: "planning", startMonth: 4, endMonth: 30, contractValue: 215_000_000_000,
			phases: []phaseSeed{
				{"Studi kelayakan & desain", -5, 2, 45},
				{"Perizinan & AMDAL", -3, 4, 20},
				{"Struktur & basement", 4, 18, 0},
				{"Fit-out tenant", 16, 30, 0},
			},
			items: []budgetItemSeed{
				{"1.1", "Desain arsitektur & struktur", 1, "ls", 6_800_000_000, 0},
				{"2.1", "Struktur beton mal 5 lantai", 42_000, "m3", 1_420_000, 2},
				{"3.1", "Fit-out area umum", 38_000, "m2", 2_600_000, 3},
			},
			permits: []permitSeed{
				{"PKKPR", "PKKPR-3273-PENGAJUAN-058", -50, 0, "submitted"},
				{"Persetujuan Lingkungan (AMDAL)", "AMDAL/DLH-BDG/PENGAJUAN-012", -20, 0, "submitted"},
			},
		},
		{
			code: projectSchool, name: "Sekolah Bina Prestasi Depok", region: "3276", kind: "Sekolah",
			status: "ongoing", startMonth: -5, endMonth: 7, contractValue: 14_800_000_000,
			phases: []phaseSeed{
				{"Pondasi & sloof", -5, -3, 100},
				{"Struktur 3 lantai", -3, 1, 70},
				{"Arsitektur & MEP", 0, 5, 10},
				{"Lapangan & lansekap", 4, 7, 0},
			},
			items: []budgetItemSeed{
				{"1.1", "Pondasi footplat", 1, "ls", 1_250_000_000, 0},
				{"2.1", "Beton struktur K-300", 1_650, "m3", 1_250_000, 1},
				{"3.1", "Pekerjaan arsitektur ruang kelas", 4_800, "m2", 650_000, 2},
				{"4.1", "Lapangan olahraga & lansekap", 1, "ls", 980_000_000, 3},
			},
			permits: []permitSeed{
				{"Persetujuan Bangunan Gedung (PBG)", "PBG-327601-0311-2026", -170, 5, "issued"},
			},
		},
		{
			code: projectWarehouse, name: "Gudang Logistik Makassar", region: "7371", kind: "Gudang",
			status: "ongoing", startMonth: -9, endMonth: 2, contractValue: 26_300_000_000,
			phases: []phaseSeed{
				{"Perbaikan tanah & pondasi", -9, -6, 100},
				{"Struktur baja gudang", -6, -2, 100},
				{"Atap, dinding & lantai kerja", -3, 1, 75},
				{"Dock leveler & utilitas", 0, 2, 15},
			},
			items: []budgetItemSeed{
				{"1.1", "Perbaikan tanah dengan PVD", 22_000, "m2", 145_000, 0},
				{"2.1", "Struktur baja WF", 780, "ton", 24_500_000, 1},
				{"3.1", "Lantai beton floor hardener", 18_000, "m2", 420_000, 2},
				{"4.1", "Dock leveler hidrolik", 12, "unit", 185_000_000, 3},
			},
			permits: []permitSeed{
				{"Persetujuan Bangunan Gedung (PBG)", "PBG-737101-0144-2025", -290, 5, "issued"},
			},
		},
		{
			code: projectHotel, name: "Hotel Pantai Cirebon", region: "3274", kind: "Hotel",
			status: "completed", startMonth: -30, endMonth: -2, contractValue: 58_600_000_000,
			phases: []phaseSeed{
				{"Pondasi & basement", -30, -24, 100},
				{"Struktur 8 lantai", -24, -14, 100},
				{"MEP & interior kamar", -14, -4, 100},
				{"Uji fungsi & serah terima", -4, -2, 100},
			},
			items: []budgetItemSeed{
				{"1.1", "Pondasi bore pile", 1_800, "m", 1_400_000, 0},
				{"2.1", "Beton struktur hotel", 9_200, "m3", 1_380_000, 1},
				{"3.1", "Interior 120 kamar", 120, "unit", 95_000_000, 2},
				{"4.1", "Uji fungsi & commissioning", 1, "ls", 650_000_000, 3},
			},
			permits: []permitSeed{
				{"Persetujuan Bangunan Gedung (PBG)", "PBG-327401-0019-2024", -900, 5, "issued"},
				{"Sertifikat Laik Fungsi (SLF)", "SLF-327401-0077-2026", -50, 5, "issued"},
			},
		},
		{
			code: projectOffice, name: "Kantor Bekasi Square", region: "3275", kind: "Kantor",
			status: "ongoing", startMonth: -3, endMonth: 15, contractValue: 72_000_000_000,
			phases: []phaseSeed{
				{"Pekerjaan persiapan & dewatering", -3, -1, 100},
				{"Pondasi & basement 2 lapis", -1, 4, 30},
				{"Struktur 15 lantai", 3, 11, 0},
				{"Fasad, MEP & fit-out", 9, 15, 0},
			},
			items: []budgetItemSeed{
				{"1.1", "Dewatering & sheet pile", 1, "ls", 2_400_000_000, 0},
				{"2.1", "Pondasi bore pile 80 cm", 3_600, "m", 1_950_000, 1},
				{"3.1", "Beton struktur K-400", 21_000, "m3", 1_450_000, 2},
				{"4.1", "Fasad ACP & kaca", 14_500, "m2", 1_950_000, 3},
			},
			permits: []permitSeed{
				{"Persetujuan Lingkungan (AMDAL)", "AMDAL/DLH-BKS/008/2026", -150, 5, "issued"},
				{"Persetujuan Bangunan Gedung (PBG)", "PBG-327501-0063-2026", -120, 5, "issued"},
			},
		},
		{
			code: projectBuahBatu, name: "Ruko Buah Batu Bandung", region: "3273", kind: "Ruko",
			status: "ongoing", startMonth: -6, endMonth: 6, contractValue: 19_400_000_000,
			phases: []phaseSeed{
				{"Pondasi & sloof", -6, -4, 100},
				{"Struktur 3 lantai", -4, 0, 85},
				{"Fasad & finishing", -1, 4, 25},
				{"Parkir & serah terima", 3, 6, 0},
			},
			items: []budgetItemSeed{
				{"1.1", "Pondasi tiang strauss", 1, "ls", 850_000_000, 0},
				{"2.1", "Struktur beton ruko", 1_320, "m3", 1_300_000, 1},
				{"3.1", "Fasad & finishing ruko", 3_600, "m2", 950_000, 2},
				{"4.1", "Area parkir paving", 1_800, "m2", 190_000, 3},
			},
			permits: []permitSeed{
				{"Persetujuan Bangunan Gedung (PBG)", "PBG-327301-0402-2026", -200, 5, "issued"},
			},
		},
	}
}

func portfolioBudgets() []budgetSeed {
	return []budgetSeed{
		{projectSkyGarden, -1, 38_000_000_000, "Anggaran pondasi & struktur bawah"},
		{projectSkyGarden, 0, 52_000_000_000, "Anggaran struktur tower & fasad"},
		{projectTembalang, 0, 24_500_000_000, "Anggaran infrastruktur & unit townhouse"},
		{projectBukitHijau, -1, 9_800_000_000, "Anggaran penyelesaian unit & fasum"},
		{projectMall, 0, 8_500_000_000, "Anggaran desain & perizinan"},
		{projectMall, 1, 64_000_000_000, "Anggaran struktur & basement"},
		{projectSchool, 0, 11_600_000_000, "Anggaran konstruksi gedung sekolah"},
		{projectWarehouse, 0, 21_400_000_000, "Anggaran struktur baja & lantai kerja"},
		{projectHotel, -1, 18_200_000_000, "Anggaran MEP & interior kamar"},
		{projectOffice, 0, 14_000_000_000, "Anggaran persiapan & pondasi"},
		{projectOffice, 1, 41_000_000_000, "Anggaran struktur 15 lantai"},
		{projectBuahBatu, 0, 15_800_000_000, "Anggaran konstruksi ruko"},
	}
}

func portfolioDailyReportPlans() []dailyReportPlan {
	return []dailyReportPlan{
		{projectSkyGarden, 124, "2 tower crane, 2 concrete pump, 4 truk mixer, passenger hoist", []string{
			"Pengecoran pelat lantai 14 zona utara",
			"Pembesian kolom dan dinding geser lantai 15",
			"Pemasangan bracket curtain wall lantai 3 sampai 6",
			"Instalasi pipa sprinkler dan riser air bersih lantai 2 sampai 5",
		}},
		{projectTembalang, 38, "1 excavator mini, 1 molen beton, scaffolding 60 set", []string{
			"Pemasangan dinding bata ringan townhouse nomor 1 sampai 6",
			"Pengecoran pelat lantai 2 townhouse nomor 7 sampai 10",
			"Pengaspalan jalan utama sepanjang 120 m",
			"Pemasangan saluran drainase sisi barat",
		}},
		{projectSchool, 46, "1 concrete pump, 2 truk mixer, scaffolding 80 set", []string{
			"Pengecoran balok dan pelat lantai 3",
			"Pemasangan dinding bata ruang kelas lantai 1",
			"Pemasangan kusen aluminium lantai 1",
			"Instalasi jalur listrik dan titik lampu lantai 1",
		}},
		{projectWarehouse, 52, "1 mobile crane 50 ton, 1 power trowel, 2 truk mixer", []string{
			"Pengecoran lantai kerja floor hardener zona C",
			"Pemasangan atap metal spandek bentang 4 sampai 6",
			"Pemasangan dinding panel bagian timur",
			"Instalasi pompa hydrant dan jalur pipa utama",
		}},
		{projectOffice, 64, "2 bore pile rig, 1 crawler crane, pompa dewatering 6 unit", []string{
			"Pengeboran bore pile as 4 sampai 7",
			"Pemasangan sheet pile sisi selatan",
			"Galian basement lapis 1 zona barat",
			"Monitoring muka air tanah dan penurunan sekitar proyek",
		}},
		{projectBuahBatu, 33, "1 molen beton, scaffolding 50 set, 1 truk material", []string{
			"Pemasangan fasad ACP ruko nomor 1 sampai 4",
			"Plesteran dan acian lantai 3",
			"Pemasangan rolling door lantai dasar",
			"Pengecatan dinding luar ruko nomor 5 sampai 8",
		}},
	}
}

func portfolioSitePhotos() []sitePhotoSeed {
	return []sitePhotoSeed{
		{projectSkyGarden, "rsia_2.jpg", -9, "struktur tower sampai lantai 14"},
		{projectTembalang, "gab_2.jpg", -12, "dinding townhouse deret pertama"},
		{projectBukitHijau, "cld_2.jpg", -200, "perumahan selesai dan dihuni"},
		{projectMall, "rsm_1.jpg", -25, "pengukuran topografi lahan mal"},
		{projectSchool, "rsia_1.jpg", -14, "pengecoran kolom lantai 2 gedung sekolah"},
		{projectWarehouse, "rsm_2.jpg", -60, "uji tanah sebelum perbaikan dengan PVD"},
		{projectHotel, "cld_1.jpg", -70, "pengecatan fasad dan penataan halaman hotel"},
		{projectOffice, "gab_1.jpg", -18, "galian dan pondasi basement"},
		{projectBuahBatu, "gab_2.jpg", -8, "dinding ruko sampai lantai 3"},
	}
}
