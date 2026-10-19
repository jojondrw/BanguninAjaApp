package main

import (
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/sales"
)

const (
	contractActive    = "active"
	contractPaid      = "paid"
	contractCancelled = "cancelled"

	hospitalClient   = "DEMO-YYS-0001"
	convertedBuyer   = "DEMO-KTP-0009"
	installmentBuyer = "DEMO-KTP-0001"
	mortgageBuyer    = "DEMO-KTP-0002"
	clusterBuyer     = "DEMO-KTP-0007"

	contractInstallment = "DEMO-KTR-GAB-001"
	contractMortgage    = "DEMO-KTR-GAB-002"
	overdueInstallment  = 6
)

type unitSeed struct {
	code     string
	project  string
	unitType string
	area     float64
	price    int64
	status   string
}

type leadSeed struct {
	name        string
	contact     string
	project     string
	source      string
	stage       string
	contactedAt int
	convertTo   *sales.CustomerRequest
}

type installmentSeed struct {
	number int
	due    int
	amount int64
	paid   bool
}

type contractSeed struct {
	number       string
	customer     string
	unit         string
	kind         string
	value        int64
	signed       int
	status       string
	installments []installmentSeed
}

var contractStatusPath = map[string][]string{
	"draft":           nil,
	contractActive:    {contractActive},
	contractPaid:      {contractActive, contractPaid},
	contractCancelled: {contractCancelled},
}

func demoCustomers() []sales.CustomerRequest {
	customers := append(coreCustomers(), portfolioBuyers()...)
	customers = append(customers, portfolioClients()...)
	return append(customers, generatedClients()...)
}

func coreCustomers() []sales.CustomerRequest {
	return []sales.CustomerRequest{
		{Name: "Budi Santoso", Contact: "0812-2045-7781", Email: "budi.santoso@contoh.id", IdentityNumber: installmentBuyer, Address: "Jl. Soekarno-Hatta No. 412, Bandung"},
		{Name: "Siti Rahmawati", Contact: "0813-9420-1156", Email: "siti.rahmawati@contoh.id", IdentityNumber: mortgageBuyer, Address: "Jl. Buah Batu No. 88, Bandung"},
		{Name: "Andi Pratama", Contact: "0821-6610-3349", Email: "andi.pratama@contoh.id", IdentityNumber: "DEMO-KTP-0003", Address: "Jl. Cibaduyut Raya No. 21, Bandung"},
		{Name: "Dewi Lestari", Contact: "0857-2231-9087", Email: "dewi.lestari@contoh.id", IdentityNumber: "DEMO-KTP-0004", Address: "Jl. Setiabudi No. 155, Bandung"},
		{Name: "Hendra Gunawan", Contact: "0811-8890-2210", Email: "hendra.gunawan@contoh.id", IdentityNumber: "DEMO-KTP-0005", Address: "Jl. Margonda Raya No. 301, Depok"},
		{Name: "Maya Sari", Contact: "0838-1172-4455", Email: "maya.sari@contoh.id", IdentityNumber: "DEMO-KTP-0006", Address: "Jl. Juanda No. 17, Depok"},
		{Name: "Rudi Hartono", Contact: "0815-7703-6612", Email: "rudi.hartono@contoh.id", IdentityNumber: clusterBuyer, Address: "Jl. Raya Sawangan No. 9, Depok"},
		{Name: "Agus Setiawan", Contact: "0812-3301-4478", Email: "agus.setiawan@contoh.id", IdentityNumber: "DEMO-KTP-0010", Address: "Jl. Kiaracondong No. 140, Bandung"},
		{Name: "Ratna Dewi", Contact: "0813-5520-9914", Email: "ratna.dewi@contoh.id", IdentityNumber: "DEMO-KTP-0011", Address: "Jl. Pasirkoja No. 62, Bandung"},
		{Name: "Bayu Saputra", Contact: "0821-7745-1023", Email: "bayu.saputra@contoh.id", IdentityNumber: "DEMO-KTP-0012", Address: "Jl. Cijerah No. 19, Bandung"},
		{Name: "Indah Permatasari", Contact: "0857-3390-6671", Email: "indah.permatasari@contoh.id", IdentityNumber: "DEMO-KTP-0013", Address: "Jl. Moh. Toha No. 205, Bandung"},
		{Name: "Dimas Aditya", Contact: "0812-9087-2245", Email: "dimas.aditya@contoh.id", IdentityNumber: "DEMO-KTP-0014", Address: "Jl. Terusan Jakarta No. 77, Bandung"},
		{Name: "Lina Susanti", Contact: "0838-2214-5590", Email: "lina.susanti@contoh.id", IdentityNumber: "DEMO-KTP-0015", Address: "Jl. Kopo No. 318, Bandung"},
		{Name: "Eko Wibowo", Contact: "0815-6630-1187", Email: "eko.wibowo@contoh.id", IdentityNumber: "DEMO-KTP-0016", Address: "Jl. Gatot Subroto No. 51, Bandung"},
		{Name: "Putri Ayu Lestari", Contact: "0813-4471-8820", Email: "putri.ayu@contoh.id", IdentityNumber: "DEMO-KTP-0017", Address: "Jl. Arcamanik No. 12, Bandung"},
		{Name: "Fikri Ramadhan", Contact: "0822-1198-3346", Email: "fikri.ramadhan@contoh.id", IdentityNumber: "DEMO-KTP-0018", Address: "Jl. Cibiru No. 90, Bandung"},
		{Name: "Sri Wahyuni", Contact: "0811-2290-7734", Email: "sri.wahyuni@contoh.id", IdentityNumber: "DEMO-KTP-0019", Address: "Jl. Nusantara No. 44, Depok"},
		{Name: "Yayasan Kesehatan Permata Bunda Cirebon", Contact: "0231-488-120", Email: "pengadaan@permatabunda.contoh.id", IdentityNumber: hospitalClient, Address: "Jl. Tuparev No. 45, Kota Cirebon"},
	}
}

func demoLeads() []leadSeed {
	return append(append(coreLeads(), portfolioLeads()...), generatedLeads()...)
}

func coreLeads() []leadSeed {
	return []leadSeed{
		{"Rina Marlina", "0812-7788-1209", projectGriyaAsri, "Pameran properti", "negotiating", -3, &sales.CustomerRequest{
			Name: "Rina Marlina", Contact: "0812-7788-1209", Email: "rina.marlina@contoh.id",
			IdentityNumber: convertedBuyer, Address: "Jl. Antapani No. 33, Bandung",
		}},
		{"Fajar Nugroho", "0813-4410-2298", projectGriyaAsri, "Instagram", "new", -1, nil},
		{"Lukman Hakim", "0822-3091-7745", projectGriyaAsri, "Referensi pembeli", "interested", -6, nil},
		{"Nia Kurniawati", "0856-2219-8830", projectGriyaAsri, "Iklan Facebook", "cancelled", -40, nil},
		{"Wulan Sari", "0811-4120-6673", projectRuko, "Situs web", "interested", -4, nil},
		{"Ahmad Fauzi", "0812-4290-1180", projectRuko, "Kunjungan kantor pemasaran", "negotiating", -2, nil},
		{"Teguh Prasetyo", "0853-9921-0457", projectRuko, "Brosur", "new", -9, nil},
		{"Yoga Pratama", "0812-5530-7781", projectGriyaAsri, "Pameran properti", "interested", -5, nil},
		{"Citra Anggraini", "0857-6612-0943", projectGriyaAsri, "Instagram", "new", -2, nil},
		{"Hadi Susanto", "0813-2201-5567", projectGriyaAsri, "Referensi pembeli", "negotiating", -7, nil},
		{"Melati Puspita", "0821-4478-3302", projectCluster, "Situs web", "interested", -12, nil},
		{"Rizky Maulana", "0838-9012-6654", projectRuko, "Iklan Facebook", "new", -1, nil},
	}
}

func demoUnits() []unitSeed {
	units := []unitSeed{
		{"DEMO-GAB-A01", projectGriyaAsri, "Tipe 36/72", 72, 485_000_000, "available"},
		{"DEMO-GAB-A02", projectGriyaAsri, "Tipe 45/90", 90, 540_000_000, "available"},
		{"DEMO-GAB-A03", projectGriyaAsri, "Tipe 36/72", 72, 485_000_000, "available"},
		{"DEMO-GAB-A04", projectGriyaAsri, "Tipe 36/72", 72, 485_000_000, "available"},
		{"DEMO-GAB-A05", projectGriyaAsri, "Tipe 45/90", 90, 545_000_000, "available"},
		{"DEMO-GAB-A06", projectGriyaAsri, "Tipe 45/90", 90, 545_000_000, "on_hold"},
		{"DEMO-GAB-B01", projectGriyaAsri, "Tipe 54/108", 108, 610_000_000, "available"},
		{"DEMO-GAB-B02", projectGriyaAsri, "Tipe 54/108", 108, 610_000_000, "available"},
		{"DEMO-GAB-B03", projectGriyaAsri, "Tipe 54/108", 108, 625_000_000, "available"},
		{"DEMO-GAB-B04", projectGriyaAsri, "Tipe 45/90", 90, 550_000_000, "available"},
		{"DEMO-GAB-C01", projectGriyaAsri, "Tipe 36/72", 72, 485_000_000, "available"},
		{"DEMO-GAB-C02", projectGriyaAsri, "Tipe 45/90", 90, 545_000_000, "available"},
		{"DEMO-GAB-C03", projectGriyaAsri, "Tipe 36/72", 72, 485_000_000, "available"},
		{"DEMO-GAB-C04", projectGriyaAsri, "Tipe 45/90", 90, 540_000_000, "available"},
		{"DEMO-GAB-C05", projectGriyaAsri, "Tipe 54/108", 108, 610_000_000, "available"},
		{"DEMO-GAB-C06", projectGriyaAsri, "Tipe 36/72", 72, 485_000_000, "available"},
		{"DEMO-GAB-C07", projectGriyaAsri, "Tipe 36/72", 72, 485_000_000, "available"},
		{"DEMO-GAB-C08", projectGriyaAsri, "Tipe 45/90", 90, 550_000_000, "available"},
		{"DEMO-GAB-C09", projectGriyaAsri, "Tipe 36/72", 72, 485_000_000, "available"},
		{"DEMO-GAB-C10", projectGriyaAsri, "Tipe 45/90", 90, 545_000_000, "available"},
		{"DEMO-GAB-C11", projectGriyaAsri, "Tipe 54/108", 108, 610_000_000, "available"},
		{"DEMO-GAB-C12", projectGriyaAsri, "Tipe 36/72", 72, 485_000_000, "available"},
		{"DEMO-CLD-C01", projectCluster, "Tipe 45/84", 84, 455_000_000, "available"},
		{"DEMO-CLD-C02", projectCluster, "Tipe 45/84", 84, 455_000_000, "available"},
		{"DEMO-CLD-C03", projectCluster, "Tipe 45/84", 84, 465_000_000, "available"},
		{"DEMO-CLD-C04", projectCluster, "Tipe 45/84", 84, 465_000_000, "available"},
		{"DEMO-CLD-C05", projectCluster, "Tipe 45/84", 84, 470_000_000, "on_hold"},
	}
	for index, code := range []string{"DEMO-RSM-R01", "DEMO-RSM-R02", "DEMO-RSM-R03", "DEMO-RSM-R04", "DEMO-RSM-R05", "DEMO-RSM-R06"} {
		status := "available"
		if index >= 4 {
			status = "on_hold"
		}
		units = append(units, unitSeed{code, projectRuko, "Ruko 3 lantai 4×15", 180, 1_850_000_000 + int64(index)*45_000_000, status})
	}
	return append(units, portfolioUnits()...)
}

func monthlyInstallments(firstDue int, start int, paidBefore int) []installmentSeed {
	installments := []installmentSeed{{1, firstDue, 97_000_000, true}}
	for number := 2; number <= 9; number++ {
		due := start + (number-2)*30
		installments = append(installments, installmentSeed{number, due, 48_500_000, due < paidBefore})
	}
	return installments
}

func demoContracts() []contractSeed {
	return append(coreContracts(), portfolioContracts()...)
}

func coreContracts() []contractSeed {
	installments := monthlyInstallments(-160, -132, -30)
	return []contractSeed{
		{contractInstallment, installmentBuyer, "DEMO-GAB-A01", "installment", 485_000_000, -170, contractActive, installments},
		{contractMortgage, mortgageBuyer, "DEMO-GAB-A02", "mortgage", 540_000_000, -70, contractActive, []installmentSeed{
			{1, -60, 108_000_000, true}, {2, 20, 432_000_000, false},
		}},
		{"DEMO-KTR-GAB-003", "DEMO-KTP-0003", "DEMO-GAB-A03", "cash", 470_000_000, -215, contractPaid, []installmentSeed{
			{1, -210, 10_000_000, true}, {2, -180, 460_000_000, true},
		}},
		{"DEMO-KTR-GAB-004", convertedBuyer, "DEMO-GAB-B01", "mortgage", 610_000_000, -2, "draft", []installmentSeed{
			{1, 10, 122_000_000, false}, {2, 60, 488_000_000, false},
		}},
		{"DEMO-KTR-GAB-005", "DEMO-KTP-0004", "DEMO-GAB-B02", "mortgage", 610_000_000, -75, contractCancelled, nil},
		{"DEMO-KTR-CLD-001", "DEMO-KTP-0005", "DEMO-CLD-C01", "mortgage", 455_000_000, -420, contractPaid, []installmentSeed{
			{1, -410, 91_000_000, true}, {2, -330, 364_000_000, true},
		}},
		{"DEMO-KTR-CLD-002", "DEMO-KTP-0006", "DEMO-CLD-C02", "cash", 450_000_000, -380, contractPaid, []installmentSeed{
			{1, -375, 450_000_000, true},
		}},
		{"DEMO-KTR-CLD-003", clusterBuyer, "DEMO-CLD-C03", "installment", 465_000_000, -330, contractPaid, []installmentSeed{
			{1, -320, 165_000_000, true}, {2, -230, 150_000_000, true}, {3, -140, 150_000_000, true},
		}},
		{"DEMO-KTR-CLD-004", "DEMO-KTP-0019", "DEMO-CLD-C04", "cash", 465_000_000, -200, contractPaid, []installmentSeed{
			{1, -198, 465_000_000, true},
		}},
		{"DEMO-KTR-GAB-006", "DEMO-KTP-0010", "DEMO-GAB-C01", "cash", 485_000_000, -150, contractPaid, []installmentSeed{
			{1, -148, 10_000_000, true}, {2, -118, 475_000_000, true},
		}},
		{"DEMO-KTR-GAB-007", "DEMO-KTP-0011", "DEMO-GAB-C02", "mortgage", 545_000_000, -125, contractPaid, []installmentSeed{
			{1, -120, 109_000_000, true}, {2, -90, 436_000_000, true},
		}},
		{"DEMO-KTR-GAB-008", "DEMO-KTP-0012", "DEMO-GAB-C03", "installment", 485_000_000, -100, contractActive, monthlyInstallments(-95, -65, 0)},
		{"DEMO-KTR-GAB-009", "DEMO-KTP-0013", "DEMO-GAB-C04", "mortgage", 540_000_000, -45, contractActive, []installmentSeed{
			{1, -40, 108_000_000, true}, {2, 30, 432_000_000, false},
		}},
		{"DEMO-KTR-GAB-010", "DEMO-KTP-0014", "DEMO-GAB-C05", "mortgage", 610_000_000, -30, contractActive, []installmentSeed{
			{1, -25, 122_000_000, true}, {2, 45, 488_000_000, false},
		}},
		{"DEMO-KTR-GAB-011", "DEMO-KTP-0015", "DEMO-GAB-C06", "installment", 485_000_000, -20, contractActive, monthlyInstallments(-15, 15, 0)},
		{"DEMO-KTR-GAB-012", "DEMO-KTP-0016", "DEMO-GAB-C07", "cash", 485_000_000, -3, "draft", []installmentSeed{
			{1, 7, 485_000_000, false},
		}},
		{"DEMO-KTR-GAB-013", "DEMO-KTP-0017", "DEMO-GAB-C08", "cash", 550_000_000, -60, contractPaid, []installmentSeed{
			{1, -58, 550_000_000, true},
		}},
		{"DEMO-KTR-GAB-014", "DEMO-KTP-0018", "DEMO-GAB-B03", "mortgage", 625_000_000, -95, contractPaid, []installmentSeed{
			{1, -90, 125_000_000, true}, {2, -50, 500_000_000, true},
		}},
	}
}

func (s *seeder) seedSales() error {
	steps := []func() error{s.seedCustomers, s.seedLeads, s.seedUnits, s.seedContracts}
	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) seedCustomers() error {
	for _, request := range demoCustomers() {
		id, _, err := s.ensure("customer", map[string]any{"identity_number": request.IdentityNumber}, func() (uuid.UUID, error) {
			response, err := s.services.sales.CreateCustomer(s.ctx, request)
			return response.ID, err
		})
		if err != nil {
			return err
		}
		s.remember("customer", request.IdentityNumber, id)
	}
	return nil
}

func (s *seeder) seedLeads() error {
	for _, seed := range demoLeads() {
		if err := s.seedLead(seed); err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) seedLead(seed leadSeed) error {
	projectID, err := s.knownPointer("project", seed.project)
	if err != nil {
		return err
	}
	id, created, err := s.ensure("lead", map[string]any{"name": seed.name, "project_id": *projectID}, func() (uuid.UUID, error) {
		response, err := s.services.sales.CreateLead(s.ctx, sales.LeadRequest{
			Name: seed.name, Contact: seed.contact, ProjectID: projectID, Source: seed.source,
			Stage: seed.stage, LastContactedAt: s.dayPointer(seed.contactedAt),
		})
		return response.ID, err
	})
	if err != nil || seed.convertTo == nil {
		return err
	}
	return s.convertLead(id, created, *seed.convertTo)
}

func (s *seeder) convertLead(leadID uuid.UUID, created bool, request sales.CustomerRequest) error {
	if created {
		response, err := s.services.sales.ConvertLead(s.ctx, leadID, request)
		if err != nil {
			return err
		}
		s.tally.add("customer")
		s.remember("customer", request.IdentityNumber, response.Customer.ID)
		return nil
	}
	id, found, err := s.lookup.find(s.ctx, "customer", map[string]any{"identity_number": request.IdentityNumber})
	if err != nil {
		return err
	}
	if found {
		s.remember("customer", request.IdentityNumber, id)
	}
	return nil
}

func (s *seeder) seedUnits() error {
	for _, seed := range demoUnits() {
		projectID, err := s.known("project", seed.project)
		if err != nil {
			return err
		}
		id, _, err := s.ensure("unit", map[string]any{"code": seed.code}, func() (uuid.UUID, error) {
			response, err := s.services.sales.CreateUnit(s.ctx, sales.UnitRequest{
				Code: seed.code, ProjectID: projectID, UnitType: seed.unitType,
				AreaSqm: seed.area, Price: seed.price, Status: seed.status,
			})
			return response.ID, err
		})
		if err != nil {
			return err
		}
		s.remember("unit", seed.code, id)
	}
	return nil
}

func (s *seeder) seedContracts() error {
	for _, seed := range demoContracts() {
		if err := s.seedContract(seed); err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) seedContract(seed contractSeed) error {
	customerID, err := s.known("customer", seed.customer)
	if err != nil {
		return err
	}
	unitID, err := s.known("unit", seed.unit)
	if err != nil {
		return err
	}
	id, created, err := s.ensure("contract", map[string]any{"number": seed.number}, func() (uuid.UUID, error) {
		response, err := s.services.sales.CreateContract(s.ctx, sales.ContractRequest{
			Number: seed.number, CustomerID: customerID, UnitID: unitID,
			Type: seed.kind, Value: seed.value, Date: s.day(seed.signed),
		})
		return response.ID, err
	})
	if err != nil {
		return err
	}
	s.remember("contract", seed.number, id)
	if !created {
		return nil
	}
	return s.scheduleContract(id, seed)
}

func (s *seeder) scheduleContract(contractID uuid.UUID, seed contractSeed) error {
	installmentIDs := map[int]uuid.UUID{}
	for _, installment := range seed.installments {
		response, err := s.services.sales.CreateInstallment(s.ctx, contractID, sales.InstallmentRequest{
			InstallmentNumber: installment.number, DueDate: s.day(installment.due), Amount: installment.amount,
		})
		if err != nil {
			return err
		}
		s.tally.add("installment")
		installmentIDs[installment.number] = response.ID
	}
	return s.settleContract(contractID, seed, installmentIDs)
}

func (s *seeder) settleContract(contractID uuid.UUID, seed contractSeed, installmentIDs map[int]uuid.UUID) error {
	for _, next := range contractStatusPath[seed.status] {
		if next == contractPaid {
			if err := s.payInstallments(contractID, seed, installmentIDs); err != nil {
				return err
			}
		}
		if _, err := s.services.sales.UpdateContractStatus(s.ctx, contractID, sales.ContractStatusRequest{Status: next}); err != nil {
			return err
		}
	}
	if seed.status == contractActive {
		return s.payInstallments(contractID, seed, installmentIDs)
	}
	return nil
}

func (s *seeder) payInstallments(contractID uuid.UUID, seed contractSeed, installmentIDs map[int]uuid.UUID) error {
	for _, installment := range seed.installments {
		if !installment.paid {
			continue
		}
		paidDate := pastOrToday(s.day(installment.due-2), s.today)
		_, err := s.services.sales.PayInstallment(s.ctx, contractID, installmentIDs[installment.number], sales.InstallmentPaymentRequest{PaidDate: paidDate})
		if err != nil {
			return err
		}
	}
	return nil
}
