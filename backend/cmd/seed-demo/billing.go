package main

import (
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/billing"
)

type paymentSeed struct {
	amount    int64
	day       int
	method    string
	reference string
	note      string
}

type invoiceSeed struct {
	number    string
	note      string
	partyType string
	party     string
	project   string
	due       int
	amount    int64
	payments  []paymentSeed
}

type receivableSeed struct {
	reference string
	customer  string
	project   string
	contract  string
	invoice   string
	due       int
	amount    int64
	payments  []paymentSeed
}

type payableSeed struct {
	reference string
	vendor    string
	project   string
	order     string
	due       int
	amount    int64
	payments  []paymentSeed
}

func demoInvoices() []invoiceSeed {
	return []invoiceSeed{
		{"DEMO-INV-001", "Pekerjaan tambah ruang NICU RSIA Cirebon", "customer", hospitalClient, projectHospital, -25, 2_500_000_000, []paymentSeed{
			{2_500_000_000, -27, "transfer", "TRF-BJB-778120", "Pelunasan pekerjaan tambah ruang NICU"},
		}},
		{"DEMO-INV-002", "Pekerjaan tambah instalasi gas medis RSIA Cirebon", "customer", hospitalClient, projectHospital, 14, 3_000_000_000, []paymentSeed{
			{1_000_000_000, -3, "transfer", "TRF-BJB-781455", "Pembayaran sebagian pekerjaan tambah instalasi gas medis"},
		}},
		{"DEMO-INV-003", "Cicilan ke-5 unit DEMO-GAB-A01", "customer", installmentBuyer, projectGriyaAsri, -12, 48_500_000, nil},
		{"DEMO-INV-004", "Klaim penggantian keramik pecah saat pengiriman", "vendor", vendorFinish, projectGriyaAsri, 21, 3_250_000, nil},
	}
}

func demoReceivables() []receivableSeed {
	return []receivableSeed{
		{"DEMO-PIU-001", hospitalClient, projectHospital, "", "DEMO-INV-002", 14, 3_000_000_000, nil},
		{"DEMO-PIU-002", installmentBuyer, projectGriyaAsri, contractInstallment, "DEMO-INV-003", -12, 48_500_000, nil},
		{"DEMO-PIU-003", mortgageBuyer, projectGriyaAsri, contractMortgage, "", 20, 432_000_000, nil},
		{"DEMO-PIU-004", clusterBuyer, projectCluster, "", "", -40, 18_500_000, []paymentSeed{
			{18_500_000, -41, "tunai", "KW-CLD-0912", "Biaya AJB & BPHTB unit DEMO-CLD-C03"},
		}},
	}
}

func demoPayables() []payableSeed {
	return append(append(corePayables(), portfolioPayables()...), generatedPayables()...)
}

func corePayables() []payableSeed {
	return []payableSeed{
		{"DEMO-HUT-001", vendorBuilding, projectGriyaAsri, orderCement, -180, 58_200_000, []paymentSeed{
			{30_000_000, -190, "transfer", "TRF-BCA-120044", "Pembayaran pertama"},
			{28_200_000, -178, "transfer", "TRF-BCA-120391", "Pelunasan"},
		}},
		{"DEMO-HUT-002", vendorConcrete, projectHospital, orderConcrete, -10, 138_000_000, []paymentSeed{
			{60_000_000, -15, "cek", "CEK-MDR-004471", "Pembayaran sebagian setelah pengecoran tahap 1"},
		}},
		{"DEMO-HUT-003", vendorElectric, projectHospital, orderElectric, 5, 14_100_000, nil},
		{"DEMO-HUT-004", vendorSteel, projectGriyaAsri, "", -10, 45_000_000, nil},
	}
}

func (s *seeder) seedBilling() error {
	steps := []func() error{s.seedInvoices, s.seedReceivables, s.seedPayables, s.payLinkedInvoices}
	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) seedInvoices() error {
	for _, seed := range demoInvoices() {
		if err := s.seedInvoice(seed); err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) seedInvoice(seed invoiceSeed) error {
	partyID, err := s.known(seed.partyType, seed.party)
	if err != nil {
		return err
	}
	projectID, err := s.knownPointer("project", seed.project)
	if err != nil {
		return err
	}
	id, created, err := s.ensure("invoice", map[string]any{"number": seed.number}, func() (uuid.UUID, error) {
		response, err := s.services.billing.CreateInvoice(s.ctx, billing.InvoiceRequest{
			Number: seed.number, Note: seed.note, PartyType: seed.partyType, PartyID: partyID,
			ProjectID: projectID, DueDate: s.day(seed.due), Amount: seed.amount,
		})
		return response.ID, err
	})
	if err != nil {
		return err
	}
	s.remember("invoice", seed.number, id)
	if created {
		s.pendingInvoices = append(s.pendingInvoices, seed)
	}
	return nil
}

func (s *seeder) payLinkedInvoices() error {
	for _, seed := range s.pendingInvoices {
		id, err := s.known("invoice", seed.number)
		if err != nil {
			return err
		}
		pay := func(request billing.PaymentRequest) error {
			_, err := s.services.billing.PayInvoice(s.ctx, id, request, &s.owner)
			return err
		}
		if err := s.recordPayments(seed.payments, pay); err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) seedReceivables() error {
	for _, seed := range demoReceivables() {
		if err := s.seedReceivable(seed); err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) seedReceivable(seed receivableSeed) error {
	request, err := s.receivableRequest(seed)
	if err != nil {
		return err
	}
	id, created, err := s.ensure("receivable", map[string]any{"reference": seed.reference}, func() (uuid.UUID, error) {
		response, err := s.services.billing.CreateReceivable(s.ctx, request)
		return response.ID, err
	})
	if err != nil || !created {
		return err
	}
	return s.recordPayments(seed.payments, func(payment billing.PaymentRequest) error {
		_, err := s.services.billing.PayReceivable(s.ctx, id, payment, &s.owner)
		return err
	})
}

func (s *seeder) receivableRequest(seed receivableSeed) (billing.ReceivableRequest, error) {
	customerID, err := s.known("customer", seed.customer)
	if err != nil {
		return billing.ReceivableRequest{}, err
	}
	projectID, err := s.knownPointer("project", seed.project)
	if err != nil {
		return billing.ReceivableRequest{}, err
	}
	contractID, err := s.knownPointer("contract", seed.contract)
	if err != nil {
		return billing.ReceivableRequest{}, err
	}
	invoiceID, err := s.knownPointer("invoice", seed.invoice)
	if err != nil {
		return billing.ReceivableRequest{}, err
	}
	return billing.ReceivableRequest{
		CustomerID: customerID, ProjectID: projectID, ContractID: contractID, InvoiceID: invoiceID,
		Reference: seed.reference, DueDate: s.day(seed.due), Amount: seed.amount,
	}, nil
}

func (s *seeder) seedPayables() error {
	for _, seed := range demoPayables() {
		if err := s.seedPayable(seed); err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) seedPayable(seed payableSeed) error {
	vendorID, err := s.known("vendor", seed.vendor)
	if err != nil {
		return err
	}
	projectID, err := s.knownPointer("project", seed.project)
	if err != nil {
		return err
	}
	orderID, err := s.knownPointer("purchase_order", seed.order)
	if err != nil {
		return err
	}
	id, created, err := s.ensure("payable", map[string]any{"reference": seed.reference}, func() (uuid.UUID, error) {
		response, err := s.services.billing.CreatePayable(s.ctx, billing.PayableRequest{
			VendorID: vendorID, ProjectID: projectID, PurchaseOrderID: orderID,
			Reference: seed.reference, DueDate: s.day(seed.due), Amount: seed.amount,
		})
		return response.ID, err
	})
	if err != nil || !created {
		return err
	}
	return s.recordPayments(seed.payments, func(payment billing.PaymentRequest) error {
		_, err := s.services.billing.PayPayable(s.ctx, id, payment, &s.owner)
		return err
	})
}

func (s *seeder) recordPayments(payments []paymentSeed, pay func(billing.PaymentRequest) error) error {
	for _, payment := range payments {
		paidAt := pastOrToday(s.day(payment.day), s.today)
		err := pay(billing.PaymentRequest{
			Amount: payment.amount, PaidAt: &paidAt, Method: payment.method,
			Reference: payment.reference, Note: payment.note,
		})
		if err != nil {
			return err
		}
		s.tally.add("billing_payment")
	}
	return nil
}
