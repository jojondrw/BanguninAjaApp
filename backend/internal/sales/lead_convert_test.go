package sales

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

type leadRepository struct {
	Repository
	lead          Lead
	createErr     error
	savedLead     *Lead
	createdClient *Customer
}

func (f *leadRepository) Transaction(_ context.Context, work func(Repository) error) error {
	return work(f)
}

func (f *leadRepository) LockLead(context.Context, uuid.UUID) (Lead, error) {
	return f.lead, nil
}

func (f *leadRepository) CreateCustomer(_ context.Context, customer *Customer) error {
	if f.createErr != nil {
		return f.createErr
	}
	customer.ID = uuid.New()
	f.createdClient = customer
	return nil
}

func (f *leadRepository) SaveLead(_ context.Context, lead *Lead) error {
	f.savedLead = lead
	return nil
}

func conversionRequest() CustomerRequest {
	return CustomerRequest{Name: "Budi Prospek", Contact: "0812", IdentityNumber: "3273010101010001"}
}

func TestConvertLeadCreatesCustomerAndLinksLead(t *testing.T) {
	repository := &leadRepository{lead: Lead{Name: "Budi Prospek", Stage: "negotiating"}}

	response, err := NewService(repository).ConvertLead(context.Background(), uuid.New(), conversionRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repository.createdClient == nil || repository.savedLead == nil {
		t.Fatal("customer was not created or lead was not saved")
	}
	if repository.savedLead.CustomerID == nil || *repository.savedLead.CustomerID != repository.createdClient.ID {
		t.Fatal("lead is not linked to the new customer")
	}
	if repository.savedLead.Stage != leadWon {
		t.Fatalf("stage = %q, want %q", repository.savedLead.Stage, leadWon)
	}
	if response.Customer.ID != repository.createdClient.ID || response.Lead.CustomerID == nil {
		t.Fatalf("response does not carry the link: %+v", response)
	}
}

func TestConvertLeadRejectsAlreadyConvertedAndCancelled(t *testing.T) {
	linked := uuid.New()
	cases := map[string]Lead{
		"already converted": {Stage: "won", CustomerID: &linked},
		"cancelled":         {Stage: leadCancelled},
	}
	for name, lead := range cases {
		repository := &leadRepository{lead: lead}
		_, err := NewService(repository).ConvertLead(context.Background(), uuid.New(), conversionRequest())
		var appErr *apperror.Error
		if !errors.As(err, &appErr) {
			t.Fatalf("%s: want app error, got %v", name, err)
		}
		if repository.createdClient != nil {
			t.Fatalf("%s: customer must not be created", name)
		}
	}
}

func TestConvertLeadReportsDuplicateIdentityNumber(t *testing.T) {
	repository := &leadRepository{lead: Lead{Stage: "new"}, createErr: database.ErrDuplicate}

	_, err := NewService(repository).ConvertLead(context.Background(), uuid.New(), conversionRequest())
	if !errors.Is(err, errCustomerIdentityUsed) {
		t.Fatalf("want identity-number conflict, got %v", err)
	}
	if repository.savedLead != nil {
		t.Fatal("lead must not change when the customer could not be created")
	}
}
