package main

import (
	"fmt"
	"strings"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/sales"
)

const (
	schoolClient    = "DEMO-YYS-0002"
	warehouseClient = "DEMO-PT-0001"
	hotelClient     = "DEMO-PT-0002"
	officeClient    = "DEMO-PT-0003"
)

var (
	buyerFirstNames = []string{
		"Arif", "Bella", "Cahya", "Dani", "Erna", "Fauzan", "Gita", "Hana", "Irfan", "Jihan",
		"Kevin", "Laras", "Mario", "Nadia", "Oki", "Prita", "Qori", "Reza", "Salsa", "Taufik",
		"Umi", "Vino", "Wening", "Yusuf", "Zahra", "Adit", "Bunga", "Candra", "Desi", "Elang",
	}
	buyerLastNames = []string{
		"Hidayat", "Kusuma", "Wijaya", "Halim", "Saputri", "Nugraha", "Siregar", "Pohan", "Tanjung", "Lubis",
		"Wibisono", "Prakoso", "Hutapea", "Sihombing", "Marpaung", "Harahap", "Purnomo", "Utami", "Kartika", "Syahputra",
		"Firmansyah", "Rachman", "Daulay", "Simanjuntak", "Wardhana",
	}
	leadSources = []string{"Pameran properti", "Instagram", "Situs web", "Referensi pembeli", "Iklan Facebook", "Kunjungan kantor pemasaran", "Brosur"}
	leadStages  = []string{"new", "interested", "negotiating", "interested", "new", "cancelled", "negotiating"}
)

type salePlan struct {
	project   string
	city      string
	street    string
	units     []unitSeed
	sold      int
	firstSign int
	lastSign  int
}

func personName(n int) string {
	return buyerFirstNames[n%len(buyerFirstNames)] + " " + buyerLastNames[(n*7)%len(buyerLastNames)]
}

func personEmail(name string, n int) string {
	return fmt.Sprintf("%s.%d@contoh.id", strings.ToLower(strings.ReplaceAll(name, " ", ".")), n)
}

func personPhone(n int) string {
	return fmt.Sprintf("08%02d-%04d-%04d", 11+n%28, 1000+(n*373)%9000, 1000+(n*911)%9000)
}

func buyerIdentity(n int) string {
	return fmt.Sprintf("DEMO-KTP-%04d", 1001+n)
}

func skyUnits() []unitSeed {
	kinds := []struct {
		name  string
		area  float64
		price int64
	}{
		{"Studio 24 m²", 24, 685_000_000},
		{"1 kamar 36 m²", 36, 985_000_000},
		{"2 kamar 54 m²", 54, 1_450_000_000},
		{"3 kamar 78 m²", 78, 2_100_000_000},
	}
	units := []unitSeed{}
	for index := range 48 {
		kind := kinds[index%len(kinds)]
		floor := 5 + index/4
		code := fmt.Sprintf("DEMO-SKY-%02d%c", floor, 'A'+index%4)
		units = append(units, unitSeed{code, projectSkyGarden, kind.name, kind.area, kind.price + int64(floor)*4_000_000, "available"})
	}
	return units
}

func numberedUnits(project, prefix, kind string, count int, area float64, price, step int64, status string) []unitSeed {
	units := []unitSeed{}
	for index := range count {
		code := fmt.Sprintf("DEMO-%s%02d", prefix, index+1)
		units = append(units, unitSeed{code, project, kind, area, price + int64(index%4)*step, status})
	}
	return units
}

func salePlans() []salePlan {
	return append([]salePlan{
		{projectSkyGarden, "Surabaya", "Jl. Mayjen Sungkono", skyUnits(), 30, -400, -4},
		{projectTembalang, "Semarang", "Jl. Ngesrep Timur", numberedUnits(projectTembalang, "TBL-T", "Townhouse 70/90", 24, 90, 1_380_000_000, 45_000_000, "available"), 14, -200, -3},
		{projectBukitHijau, "Bogor", "Jl. Pajajaran", numberedUnits(projectBukitHijau, "BHB-H", "Tipe 54/120", 30, 120, 845_000_000, 15_000_000, "available"), 28, -900, -200},
		{projectBuahBatu, "Bandung", "Jl. Buah Batu", numberedUnits(projectBuahBatu, "RBB-R", "Ruko 3 lantai 5×17", 12, 255, 2_650_000_000, 35_000_000, "available"), 7, -160, -12},
		{projectMall, "Bandung", "Jl. Soekarno-Hatta", numberedUnits(projectMall, "MBT-L", "Ruang tenant 120 m²", 10, 120, 3_600_000_000, 150_000_000, "on_hold"), 0, 0, 0},
	}, generatedSalePlans()...)
}

func portfolioUnits() []unitSeed {
	units := []unitSeed{}
	for _, plan := range salePlans() {
		units = append(units, plan.units...)
	}
	return units
}

func portfolioBuyers() []sales.CustomerRequest {
	buyers := []sales.CustomerRequest{}
	n := 0
	for _, plan := range salePlans() {
		for index := range plan.sold {
			name := personName(n)
			buyers = append(buyers, sales.CustomerRequest{
				Name: name, Contact: personPhone(n), Email: personEmail(name, n), IdentityNumber: buyerIdentity(n),
				Address: fmt.Sprintf("%s No. %d, %s", plan.street, 10+index*7, plan.city),
			})
			n++
		}
	}
	return buyers
}

func portfolioClients() []sales.CustomerRequest {
	return []sales.CustomerRequest{
		{Name: "Yayasan Pendidikan Bina Prestasi", Contact: "021-7720-118", Email: "sarpras@binaprestasi.contoh.id", IdentityNumber: schoolClient, Address: "Jl. Raya Bogor Km 31, Depok"},
		{Name: "PT Lintas Logistik Sulawesi", Contact: "0411-588-204", Email: "proyek@lintaslogistik.contoh.id", IdentityNumber: warehouseClient, Address: "Kawasan Industri Makassar Blok C, Makassar"},
		{Name: "PT Pesona Pantai Hotelindo", Contact: "0231-209-771", Email: "development@pesonapantai.contoh.id", IdentityNumber: hotelClient, Address: "Jl. Yos Sudarso No. 8, Kota Cirebon"},
		{Name: "PT Bekasi Square Properti", Contact: "021-8895-3320", Email: "owner.rep@bekasisquare.contoh.id", IdentityNumber: officeClient, Address: "Jl. Ahmad Yani No. 1, Kota Bekasi"},
	}
}

func shares(value int64, percents []int64) []int64 {
	amounts := make([]int64, len(percents))
	var assigned int64
	for index, percent := range percents {
		amounts[index] = value * percent / 100
		assigned += amounts[index]
	}
	amounts[len(amounts)-1] += value - assigned
	return amounts
}

func saleInstallments(kind string, value int64, signed int, draft bool) []installmentSeed {
	var dues []int
	var percents []int64
	if draft {
		signed = 0
	}
	switch kind {
	case "cash":
		dues, percents = []int{signed + 7, signed + 37}, []int64{10, 90}
	case "mortgage":
		dues, percents = []int{signed + 14, signed + 75}, []int64{20, 80}
	default:
		dues, percents = []int{signed + 7}, []int64{20}
		for month := range 8 {
			dues = append(dues, signed+37+month*30)
			percents = append(percents, 10)
		}
	}
	amounts := shares(value, percents)
	installments := make([]installmentSeed, len(dues))
	for index, due := range dues {
		installments[index] = installmentSeed{index + 1, due, amounts[index], !draft && due < 0}
	}
	return installments
}

func saleStatus(installments []installmentSeed, draft bool) string {
	if draft {
		return "draft"
	}
	for _, installment := range installments {
		if !installment.paid {
			return contractActive
		}
	}
	return contractPaid
}

func portfolioContracts() []contractSeed {
	kinds := []string{"mortgage", "cash", "installment"}
	contracts := []contractSeed{}
	n := 0
	for _, plan := range salePlans() {
		for index := range plan.sold {
			unit := plan.units[index]
			signed := plan.firstSign + (plan.lastSign-plan.firstSign)*index/max(plan.sold-1, 1) - (index*17)%9
			signed = min(signed, plan.lastSign)
			number := fmt.Sprintf("DEMO-KTR-%s-%03d", strings.TrimPrefix(plan.project, demoPrefix), index+1)
			value := unit.price - int64(index%4)*5_000_000
			if index == 3 {
				contracts = append(contracts, contractSeed{number, buyerIdentity(n), unit.code, "mortgage", value, signed, contractCancelled, nil})
				n++
				continue
			}
			kind := kinds[index%len(kinds)]
			draft := signed > -10
			installments := saleInstallments(kind, value, signed, draft)
			contracts = append(contracts, contractSeed{number, buyerIdentity(n), unit.code, kind, value, signed, saleStatus(installments, draft), installments})
			n++
		}
	}
	return contracts
}

func portfolioLeads() []leadSeed {
	projects := []string{projectSkyGarden, projectTembalang, projectBuahBatu, projectMall, projectGriyaAsri, projectSkyGarden}
	leads := []leadSeed{}
	for index := range 48 {
		n := 100 + index
		leads = append(leads, leadSeed{
			personName(n), personPhone(n), projects[index%len(projects)],
			leadSources[index%len(leadSources)], leadStages[index%len(leadStages)], -(1 + (index*5)%45), nil,
		})
	}
	return leads
}
