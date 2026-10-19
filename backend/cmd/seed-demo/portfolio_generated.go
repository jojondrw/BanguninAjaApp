package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/sales"
)

type kindTemplate struct {
	kind     string
	phases   []string
	items    []string
	unitType string
	unitArea float64
	profile  string
	photo    string
	sale     bool
	works    []string
	crew     int
}

var kindTemplates = map[string]kindTemplate{
	"perumahan": {"Perumahan", []string{"Pematangan lahan", "Jalan & drainase", "Pembangunan unit", "Fasilitas umum"},
		[]string{"Cut and fill lahan", "Jalan paving & saluran", "Pembangunan rumah", "Fasum & taman"}, "Rumah tipe 45/90", 90, "housing", "gab_2.jpg", true,
		[]string{"Pemasangan dinding bata ringan", "Pengecoran sloof dan kolom praktis", "Pemasangan rangka atap baja ringan", "Pemasangan paving jalan lingkungan"}, 36},
	"apartemen": {"Apartemen", []string{"Pondasi & basement", "Struktur tower", "MEP & fasad", "Interior & serah terima"},
		[]string{"Tiang pancang & basement", "Beton struktur tower", "Fasad kaca & MEP", "Interior unit"}, "Unit 2 kamar 54 m²", 54, "housing", "rsia_2.jpg", true,
		[]string{"Pengecoran pelat lantai tower", "Pembesian kolom dan dinding geser", "Pemasangan curtain wall", "Instalasi pipa air bersih dan sprinkler"}, 110},
	"ruko": {"Ruko", []string{"Pondasi", "Struktur 3 lantai", "Fasad & finishing", "Parkir & serah terima"},
		[]string{"Pondasi tiang strauss", "Struktur beton ruko", "Fasad & finishing", "Area parkir"}, "Ruko 3 lantai 5×15", 225, "mall", "gab_1.jpg", true,
		[]string{"Pengecoran pelat lantai 2", "Pemasangan fasad ACP", "Plesteran dinding lantai 3", "Pemasangan rolling door"}, 30},
	"rumah sakit": {"Rumah Sakit", []string{"Pondasi bore pile", "Struktur gedung", "MEP & gas medis", "Arsitektur & alat medis"},
		[]string{"Pondasi bore pile", "Beton struktur", "Instalasi gas medis & listrik", "Interior ruang rawat"}, "", 0, "hospital", "rsia_1.jpg", false,
		[]string{"Pengecoran kolom lantai atas", "Instalasi pipa gas medis", "Pemasangan tray kabel", "Pemasangan plafon ruang rawat"}, 72},
	"sekolah": {"Sekolah", []string{"Pondasi", "Struktur", "Arsitektur & MEP", "Lansekap"},
		[]string{"Pondasi footplat", "Beton struktur", "Arsitektur ruang kelas", "Lapangan & taman"}, "", 0, "housing", "rsia_1.jpg", false,
		[]string{"Pengecoran balok lantai 2", "Pemasangan dinding ruang kelas", "Pemasangan kusen aluminium", "Instalasi titik lampu"}, 40},
	"gudang": {"Gudang", []string{"Perbaikan tanah", "Struktur baja", "Atap & lantai kerja", "Utilitas"},
		[]string{"Perbaikan tanah & pondasi", "Struktur baja WF", "Atap & floor hardener", "Dock leveler & hydrant"}, "", 0, "mall", "rsm_2.jpg", false,
		[]string{"Ereksi kolom baja", "Pemasangan atap spandek", "Pengecoran lantai kerja", "Instalasi hydrant"}, 44},
	"hotel": {"Hotel", []string{"Pondasi & basement", "Struktur", "MEP & interior kamar", "Uji fungsi"},
		[]string{"Pondasi bore pile", "Beton struktur hotel", "Interior kamar & MEP", "Commissioning"}, "", 0, "entertainment", "cld_1.jpg", false,
		[]string{"Pengecoran pelat lantai kamar", "Pemasangan ducting AC", "Pemasangan interior kamar contoh", "Pengecatan fasad"}, 66},
	"kantor": {"Kantor", []string{"Persiapan & dewatering", "Pondasi & basement", "Struktur", "Fasad & fit-out"},
		[]string{"Dewatering & sheet pile", "Pondasi bore pile", "Beton struktur", "Fasad & fit-out"}, "", 0, "mall", "gab_1.jpg", false,
		[]string{"Pengeboran bore pile", "Galian basement", "Pengecoran core wall", "Pemasangan fasad kaca"}, 58},
	"klinik": {"Klinik", []string{"Pondasi", "Struktur 2 lantai", "MEP & interior"},
		[]string{"Pondasi", "Struktur beton", "MEP & interior klinik"}, "", 0, "hospital", "rsia_2.jpg", false,
		[]string{"Pengecoran pelat lantai 2", "Pemasangan dinding partisi", "Instalasi listrik"}, 22},
	"campuran": {"Kawasan Campuran", []string{"Pematangan & infrastruktur", "Struktur podium", "Struktur tower", "Fasad & fit-out"},
		[]string{"Infrastruktur kawasan", "Struktur podium", "Struktur tower", "Fasad & fit-out"}, "Unit hunian 45 m²", 45, "mall", "rsm_1.jpg", true,
		[]string{"Pengecoran podium zona B", "Pembesian tower hunian", "Pekerjaan jalan kawasan", "Pemasangan fasad podium"}, 140},
}

var saleShare = map[string]int64{"perumahan": 85, "ruko": 85, "apartemen": 60, "campuran": 25}

type generatedSeed struct {
	suffix   string
	name     string
	region   string
	city     string
	kind     string
	status   string
	start    int
	duration int
	value    int64
	pace     float64
	lat      float64
	lon      float64
	units    int
	sold     int
	late     bool
	permit   string
}

func generatedSeeds() []generatedSeed {
	return []generatedSeed{
		{"JKS", "Apartemen Kemang Signature Jakarta Selatan", "3171", "Jakarta Selatan", "apartemen", "ongoing", -16, 34, 312_000_000_000, 0.95, -6.2615, 106.8137, 60, 41, false, "issued"},
		{"JKT", "Perumahan Cakung Asri Jakarta Timur", "3172", "Jakarta Timur", "perumahan", "ongoing", -10, 22, 58_000_000_000, 0.7, -6.1838, 106.9431, 40, 18, false, "issued"},
		{"JKB", "Ruko Puri Kembangan Jakarta Barat", "3174", "Jakarta Barat", "ruko", "completed", -26, 20, 44_500_000_000, 1, -6.1898, 106.7362, 16, 16, false, "issued"},
		{"TGR", "Gudang Distribusi Batuceper Tangerang", "3671", "Tangerang", "gudang", "ongoing", -8, 13, 37_800_000_000, 1.15, -6.1636, 106.6596, 0, 0, true, "issued"},
		{"BSD", "Kantor Teknologi Serpong Tangerang Selatan", "3674", "Tangerang Selatan", "kantor", "ongoing", -12, 26, 146_000_000_000, 0.8, -6.3017, 106.6526, 0, 0, false, "issued"},
		{"BDK", "Perumahan Soreang Indah Kabupaten Bandung", "3204", "Kabupaten Bandung", "perumahan", "ongoing", -9, 18, 27_600_000_000, 1.1, -7.0256, 107.5262, 36, 9, false, "issued"},
		{"CMH", "Klinik Pratama Cimahi Sehat", "3277", "Cimahi", "klinik", "ongoing", -4, 9, 6_400_000_000, 1, -6.8722, 107.5425, 0, 0, false, "issued"},
		{"BDA", "Rumah Sakit Umum Antapani Bandung", "3273", "Bandung", "rumah sakit", "ongoing", -11, 24, 96_000_000_000, 0.85, -6.9147, 107.6618, 0, 0, true, "issued"},
		{"BDS", "Sekolah Dasar Islam Terpadu Sukajadi Bandung", "3273", "Bandung", "sekolah", "completed", -20, 14, 9_300_000_000, 1, -6.8868, 107.5953, 0, 0, false, "issued"},
		{"CRK", "Perumahan Kalijaga Permai Cirebon", "3274", "Cirebon", "perumahan", "ongoing", -13, 20, 23_400_000_000, 0.9, -6.7363, 108.5418, 32, 32, false, "issued"},
		{"CRG", "Gudang Pendingin Pelabuhan Cirebon", "3274", "Cirebon", "gudang", "on_hold", -9, 12, 18_900_000_000, 0.5, -6.7096, 108.5702, 0, 0, false, "expired"},
		{"DPS", "Apartemen Margonda Studentpark Depok", "3276", "Depok", "apartemen", "ongoing", -7, 26, 118_000_000_000, 1.05, -6.3729, 106.8340, 80, 6, false, "issued"},
		{"DPK", "Klinik Utama Sawangan Depok", "3276", "Depok", "klinik", "ongoing", -3, 8, 4_850_000_000, 0.9, -6.4017, 106.7638, 0, 0, false, "issued"},
		{"MKP", "Ruko Pettarani Business Center Makassar", "7371", "Makassar", "ruko", "ongoing", -8, 15, 31_200_000_000, 0.95, -5.1477, 119.4327, 20, 12, false, "issued"},
		{"MKH", "Hotel Losari Bay Makassar", "7371", "Makassar", "hotel", "ongoing", -15, 28, 134_000_000_000, 0.75, -5.1430, 119.4065, 0, 0, true, "issued"},
		{"SBY", "Perumahan Citraland Barat Surabaya", "3578", "Surabaya", "perumahan", "completed", -30, 24, 61_000_000_000, 1, -7.2884, 112.6378, 45, 43, false, "issued"},
		{"MLG", "Kampus Politeknik Swasta Malang", "3573", "Malang", "sekolah", "ongoing", -6, 16, 42_000_000_000, 1, -7.9539, 112.6143, 0, 0, false, "issued"},
		{"SLO", "Hotel Heritage Laweyan Surakarta", "3372", "Surakarta", "hotel", "planning", 2, 22, 72_000_000_000, 1, -7.5700, 110.7932, 0, 0, false, "submitted"},
		{"YOG", "Apartemen Kaliurang Residence Yogyakarta", "3471", "Yogyakarta", "apartemen", "cancelled", -2, 30, 88_000_000_000, 0.2, -7.7559, 110.3934, 50, 0, false, "rejected"},
		{"DPR", "Villa Kawasan Sanur Denpasar", "5171", "Denpasar", "perumahan", "ongoing", -5, 14, 47_500_000_000, 1.2, -8.6883, 115.2583, 18, 11, false, "issued"},
		{"MDN", "Kawasan Campuran Medan Sunggal", "1275", "Medan", "campuran", "ongoing", -18, 42, 640_000_000_000, 0.8, -3.5797, 98.6309, 120, 57, false, "issued"},
		{"PLB", "Rumah Sakit Ibu dan Anak Palembang", "1671", "Palembang", "rumah sakit", "ongoing", -6, 18, 54_000_000_000, 1, -2.9761, 104.7754, 0, 0, false, "issued"},
		{"BPN", "Gudang Logistik Tambang Balikpapan", "6471", "Balikpapan", "gudang", "completed", -14, 11, 29_000_000_000, 1, -1.2420, 116.8890, 0, 0, true, "issued"},
		{"PDG", "Ruko Khatib Sulaiman Padang", "1371", "Padang", "ruko", "planning", 3, 12, 15_600_000_000, 1, -0.9178, 100.3615, 10, 0, false, "submitted"},
		{"PKU", "Perumahan Panam Lestari Pekanbaru", "1471", "Pekanbaru", "perumahan", "ongoing", -14, 20, 32_800_000_000, 0.6, 0.4650, 101.3734, 40, 22, false, "issued"},
		{"BJM", "Kantor Pemerintah Daerah Banjarmasin", "6371", "Banjarmasin", "kantor", "on_hold", -10, 18, 41_000_000_000, 0.45, -3.3194, 114.5908, 0, 0, true, "issued"},
		{"MND", "Hotel Boulevard Manado", "7171", "Manado", "hotel", "planning", 1, 24, 66_000_000_000, 1, 1.4931, 124.8413, 0, 0, false, "submitted"},
		{"BTM", "Superblok Batam Center: Apartemen, Mal, Hotel, dan Perkantoran Terpadu Tahap Pertama", "2171", "Batam", "campuran", "ongoing", -20, 48, 1_240_000_000_000, 0.9, 1.1301, 104.0529, 150, 88, false, "issued"},
		{"RNV", "Renovasi Kantor Pemasaran Pusat", "3273", "Bandung", "kantor", "completed", -5, 3, 850_000_000, 1, -6.9175, 107.6191, 0, 0, false, "issued"},
		{"MKS", "Perumahan Antang Griya Makassar", "7371", "Makassar", "perumahan", "ongoing", -6, 16, 21_800_000_000, 1, -5.1736, 119.4893, 30, 0, false, "issued"},
	}
}

func generatedCode(seed generatedSeed) string {
	return "DEMO-G" + seed.suffix
}

func generatedProgress(seed generatedSeed, from, to int) int {
	switch seed.status {
	case "completed":
		return 100
	case "planning":
		if from >= 0 {
			return 0
		}
	}
	if from >= 0 {
		return 0
	}
	share := float64(-from) / float64(to-from) * 100 * seed.pace
	return int(math.Min(100, math.Max(0, math.Round(share))))
}

func generatedPhases(seed generatedSeed, template kindTemplate) []phaseSeed {
	count := len(template.phases)
	phases := make([]phaseSeed, count)
	for index, name := range template.phases {
		from := seed.start + seed.duration*index*85/(count*100)
		to := seed.start + seed.duration*(index+1)/count
		if seed.status == "completed" {
			to = min(to, -1)
			from = min(from, to-1)
		}
		phases[index] = phaseSeed{name, from, to, generatedProgress(seed, from, to)}
	}
	return phases
}

func generatedItems(seed generatedSeed, template kindTemplate) []budgetItemSeed {
	weights := []int64{12, 38, 32, 18}
	items := make([]budgetItemSeed, len(template.items))
	for index, description := range template.items {
		share := seed.value * 72 / 100 * weights[index%len(weights)] / 100
		items[index] = budgetItemSeed{fmt.Sprintf("%d.1", index+1), description, 1, "ls", share, index}
	}
	return items
}

func generatedPermits(seed generatedSeed) []permitSeed {
	number := fmt.Sprintf("PBG-%s01-G%s-2025", seed.region, seed.suffix)
	switch seed.permit {
	case "submitted":
		return []permitSeed{{"Persetujuan Bangunan Gedung (PBG)", strings.Replace(number, "2025", "PENGAJUAN", 1), -25, 0, "submitted"}}
	case "rejected":
		return []permitSeed{{"PKKPR", fmt.Sprintf("PKKPR-%s-G%s-DITOLAK", seed.region, seed.suffix), -40, 0, "rejected"}}
	case "expired":
		return []permitSeed{
			{"Persetujuan Bangunan Gedung (PBG)", number, -30 * (-seed.start + 2), 1, "expired"},
			{"Persetujuan Lingkungan (UKL-UPL)", fmt.Sprintf("UKL-UPL/G%s/2025", seed.suffix), -30 * (-seed.start + 3), 5, "issued"},
		}
	}
	issued := min(-30*(-seed.start+2), -20)
	return []permitSeed{
		{"PKKPR", fmt.Sprintf("PKKPR-%s-G%s-2025", seed.region, seed.suffix), issued - 30, 3, "issued"},
		{"Persetujuan Bangunan Gedung (PBG)", number, issued, 5, "issued"},
	}
}

func generatedProjects() []projectSeed {
	projects := []projectSeed{}
	for _, seed := range generatedSeeds() {
		template := kindTemplates[seed.kind]
		projects = append(projects, projectSeed{
			code: generatedCode(seed), name: seed.name, region: seed.region, kind: template.kind, status: seed.status,
			startMonth: seed.start, endMonth: seed.start + seed.duration, contractValue: seed.value,
			phases: generatedPhases(seed, template), items: generatedItems(seed, template), permits: generatedPermits(seed),
		})
	}
	return projects
}

func generatedSalePlans() []salePlan {
	plans := []salePlan{}
	for _, seed := range generatedSeeds() {
		template := kindTemplates[seed.kind]
		if !template.sale || seed.units == 0 {
			continue
		}
		code := generatedCode(seed)
		price := seed.value * saleShare[seed.kind] / 100 / int64(seed.units)
		price = price / 5_000_000 * 5_000_000
		status := "available"
		if seed.status == "planning" || seed.status == "cancelled" {
			status = "on_hold"
		}
		units := numberedUnits(code, "G"+seed.suffix+"-U", template.unitType, seed.units, template.unitArea, price, price/40, status)
		first := max(seed.start*30, -720)
		plans = append(plans, salePlan{code, seed.city, "Jl. Raya " + seed.city, units, seed.sold, first, -3})
	}
	return plans
}

func generatedClients() []sales.CustomerRequest {
	clients := []sales.CustomerRequest{}
	for _, seed := range generatedSeeds() {
		if kindTemplates[seed.kind].sale {
			continue
		}
		clients = append(clients, sales.CustomerRequest{
			Name:           "PT Pemilik " + strings.SplitN(seed.name, " ", 2)[1],
			Contact:        personPhone(len(seed.name) * 13),
			Email:          "proyek.g" + strings.ToLower(seed.suffix) + "@pemberikerja.contoh.id",
			IdentityNumber: "DEMO-PT-G" + seed.suffix,
			Address:        "Jl. Utama No. 1, " + seed.city,
		})
	}
	return clients
}

func running(seed generatedSeed) bool {
	return seed.status == "ongoing" || seed.status == "on_hold" || seed.status == "completed"
}

func generatedConstructionContracts() []constructionSeed {
	vendors := []string{vendorGeneral, vendorHousing, vendorStructure, vendorSteelWF, vendorPiling}
	contracts := []constructionSeed{}
	for index, seed := range generatedSeeds() {
		if !running(seed) {
			continue
		}
		terms := []termSeed{{0, 10, true}, {25, 20, true}, {50, 20, true}, {75, 25, true}, {100, 25, true}}
		var changes []changeSeed
		if index%4 == 1 {
			changes = []changeSeed{{"Perubahan spesifikasi material sesuai permintaan pemilik " + seed.suffix, seed.value / 60, 20, []string{"approved", "rejected", ""}[index%3], -40, -30}}
		}
		contracts = append(contracts, constructionSeed{generatedCode(seed), vendors[index%len(vendors)], "DEMO-SPK-G" + seed.suffix, seed.value * 65 / 100, 5, terms, changes})
	}
	return contracts
}

func generatedOwnerContracts() []ownerSeed {
	contracts := []ownerSeed{}
	for _, seed := range generatedSeeds() {
		if kindTemplates[seed.kind].sale || !running(seed) {
			continue
		}
		terms := []termSeed{{0, 10, true}, {25, 20, true}, {50, 20, true}, {75, 25, true}, {100, 25, true}}
		contracts = append(contracts, ownerSeed{generatedCode(seed), "DEMO-PT-G" + seed.suffix, "DEMO-KTR-G" + seed.suffix, seed.value, 5, 30, terms, seed.late})
	}
	return contracts
}

func generatedActivity() []activeProject {
	activity := []activeProject{}
	for _, seed := range generatedSeeds() {
		if !running(seed) {
			continue
		}
		end := min(seed.start+seed.duration, 0)
		if seed.status == "on_hold" {
			end = -2
		}
		monthly := seed.value * 55 / 100 / int64(seed.duration)
		var salesIn int64
		if plan := kindTemplates[seed.kind]; plan.sale && seed.units > 0 {
			salesIn = seed.value * 70 / 100 / int64(seed.duration) * int64(seed.sold) / int64(seed.units)
		}
		activity = append(activity, activeProject{generatedCode(seed), seed.name, max(seed.start, -ledgerMonths+1), end, monthly, salesIn, monthly / 90})
	}
	return activity
}

func generatedEmployees() []employeeSeed {
	employees := []employeeSeed{}
	crew := 0
	for _, seed := range generatedSeeds() {
		if seed.status != "ongoing" {
			continue
		}
		code := generatedCode(seed)
		joined := max(seed.start, -18)
		team := []roleSeed{siteTeam[0], siteTeam[1], siteTeam[4], siteCrew[0], siteCrew[1], siteCrew[4]}
		for _, role := range team {
			name := workerNames[crew%len(workerNames)]
			if role.employment != employmentDaily {
				name = personName(400 + crew)
			}
			employees = append(employees, employeeSeed{
				fmt.Sprintf("DEMO-NIK-%03d", 301+crew), name, role.position, code, role.employment, joined, role.salary, role.allowance,
			})
			crew++
		}
	}
	return employees
}

func generatedDailyReportPlans() []dailyReportPlan {
	plans := []dailyReportPlan{}
	for _, seed := range generatedSeeds() {
		if seed.status != "ongoing" {
			continue
		}
		template := kindTemplates[seed.kind]
		plans = append(plans, dailyReportPlan{generatedCode(seed), template.crew, "Alat sesuai rencana kerja mingguan", template.works})
	}
	return plans
}

var generatedPhotos = map[string]string{
	"JKS": "rsia_2.jpg", "JKT": "gab_1.jpg", "JKB": "ruko_selesai.jpg", "TGR": "gudang_baja.jpg", "BSD": "rs_bertingkat.jpg",
	"BDK": "gab_2.jpg", "CMH": "klinik.jpg", "BDA": "rs_bertingkat.jpg", "BDS": "sekolah.jpg", "CRK": "cld_1.jpg",
	"CRG": "rsm_2.jpg", "DPS": "rsia_1.jpg", "DPK": "gab_1.jpg", "MKP": "ruko_bangun.jpg", "MKH": "hotel_pantai.jpg",
	"SBY": "cld_2.jpg", "MLG": "sekolah.jpg", "SLO": "rsm_1.jpg", "YOG": "rsm_1.jpg", "DPR": "villa_bali.jpg",
	"MDN": "apt_struktur.jpg", "PLB": "rsia_1.jpg", "BPN": "gudang_baja.jpg", "PDG": "rsm_2.jpg", "PKU": "gab_1.jpg",
	"BJM": "kantor_basement.jpg", "MND": "rsm_1.jpg", "BTM": "apt_fasad.jpg", "RNV": "klinik.jpg", "MKS": "gab_1.jpg",
}

func generatedSitePhotos() []sitePhotoSeed {
	photos := []sitePhotoSeed{}
	for index, seed := range generatedSeeds() {
		template := kindTemplates[seed.kind]
		photo := template.photo
		if chosen, found := generatedPhotos[seed.suffix]; found {
			photo = chosen
		}
		photos = append(photos, sitePhotoSeed{generatedCode(seed), photo, -(3 + index%20), "kondisi lapangan " + strings.ToLower(template.kind) + " " + seed.city})
	}
	return photos
}

func generatedSites() []siteSeed {
	sites := []siteSeed{}
	for index, seed := range generatedSeeds() {
		template := kindTemplates[seed.kind]
		base := 60 + index%25
		sites = append(sites, siteSeed{
			"Lahan terpilih " + seed.name, seed.region, template.profile, generatedCode(seed), seed.lat, seed.lon,
			float64(4_000 + (index*3_700)%40_000), int64(2_500_000 + (index*1_350_000)%18_000_000),
			float64(index%9) / 10, float64((index*3)%8) / 10,
			[]int{base, base + 7 - index%5, base + 3, base - 4 + index%9, base - 2},
		})
	}
	return sites
}

func generatedLeads() []leadSeed {
	leads := []leadSeed{}
	n := 600
	for _, plan := range generatedSalePlans() {
		for index := range 4 + len(plan.units)%5 {
			leads = append(leads, leadSeed{
				personName(n), personPhone(n), plan.project, leadSources[n%len(leadSources)],
				leadStages[(n+index)%len(leadStages)], -(1 + (n*3)%60), nil,
			})
			n++
		}
	}
	return leads
}

func generatedLoans() []cashSeed {
	loans := []cashSeed{}
	for index, seed := range generatedSeeds() {
		if !running(seed) {
			continue
		}
		monthsAgo := min(-seed.start, ledgerMonths-1)
		loans = append(loans, cashSeed{monthsAgo, 10 + index%15, cashIn, "2300", generatedCode(seed), seed.value * 40 / 100, "Pencairan kredit konstruksi bank untuk " + seed.name})
	}
	return loans
}
