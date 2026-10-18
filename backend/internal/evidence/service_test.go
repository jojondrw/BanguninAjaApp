package evidence

import (
	"context"
	"encoding/hex"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/ethereum"
)

type memoryRepository struct {
	records   []Record
	batches   map[uuid.UUID]*Batch
	approvers map[uuid.UUID]Approver
	claims    map[uuid.UUID]*ClaimLink
}

func newMemory() *memoryRepository {
	return &memoryRepository{batches: map[uuid.UUID]*Batch{}, approvers: map[uuid.UUID]Approver{}, claims: map[uuid.UUID]*ClaimLink{}}
}

func (m *memoryRepository) Transaction(_ context.Context, work func(Repository) error) error {
	return work(m)
}

func (m *memoryRepository) RecordExists(_ context.Context, kind string, subjectID uuid.UUID, digest string) (bool, error) {
	for _, record := range m.records {
		if record.Kind == kind && record.SubjectID == subjectID && record.Digest == digest {
			return true, nil
		}
	}
	return false, nil
}

func (m *memoryRepository) CreateRecord(_ context.Context, record *Record) error {
	record.ID = uuid.New()
	m.records = append(m.records, *record)
	return nil
}

func (m *memoryRepository) RecordsFor(_ context.Context, projectID *uuid.UUID) ([]Record, error) {
	var found []Record
	for _, record := range m.records {
		if projectID == nil || (record.ProjectID != nil && *record.ProjectID == *projectID) {
			found = append(found, record)
		}
	}
	return found, nil
}

func (m *memoryRepository) UnbatchedRecords(context.Context) ([]Record, error) {
	var found []Record
	for _, record := range m.records {
		if record.BatchID == nil {
			found = append(found, record)
		}
	}
	return found, nil
}

func (m *memoryRepository) PendingBatch(context.Context) (Batch, bool, error) {
	for _, batch := range m.batches {
		if batch.TxHash == nil {
			return *batch, true, nil
		}
	}
	return Batch{}, false, nil
}

func (m *memoryRepository) ReleaseBatch(_ context.Context, batchID uuid.UUID) error {
	for index := range m.records {
		if m.records[index].BatchID != nil && *m.records[index].BatchID == batchID {
			m.records[index].BatchID = nil
		}
	}
	delete(m.batches, batchID)
	return nil
}

func (m *memoryRepository) CreateBatch(_ context.Context, batch *Batch, records []Record) error {
	batch.ID = uuid.New()
	m.batches[batch.ID] = batch
	for _, record := range records {
		for index := range m.records {
			if m.records[index].ID == record.ID {
				id := batch.ID
				m.records[index].BatchID = &id
			}
		}
	}
	return nil
}

func (m *memoryRepository) FindBatch(_ context.Context, id uuid.UUID) (Batch, error) {
	if batch, ok := m.batches[id]; ok {
		return *batch, nil
	}
	return Batch{}, database.ErrNotFound
}

func (m *memoryRepository) SaveBatch(_ context.Context, batch *Batch) error {
	copied := *batch
	m.batches[batch.ID] = &copied
	return nil
}

func (m *memoryRepository) AnchoredBatches(_ context.Context, ids []uuid.UUID) (map[uuid.UUID]Batch, error) {
	found := map[uuid.UUID]Batch{}
	for _, id := range ids {
		if batch, ok := m.batches[id]; ok && batch.TxHash != nil {
			found[id] = *batch
		}
	}
	return found, nil
}

func (m *memoryRepository) ListBatches(context.Context) ([]Batch, error) { return nil, nil }

func (m *memoryRepository) FindApprover(_ context.Context, projectID uuid.UUID) (Approver, bool, error) {
	approver, ok := m.approvers[projectID]
	return approver, ok, nil
}

func (m *memoryRepository) SaveApprover(_ context.Context, approver *Approver) error {
	m.approvers[approver.ProjectID] = *approver
	return nil
}

func (m *memoryRepository) FindClaimByTerm(_ context.Context, termID uuid.UUID) (ClaimLink, bool, error) {
	for _, claim := range m.claims {
		if claim.TermID == termID {
			return *claim, true, nil
		}
	}
	return ClaimLink{}, false, nil
}

func (m *memoryRepository) FindClaimByToken(_ context.Context, token string) (ClaimLink, error) {
	for _, claim := range m.claims {
		if claim.Token == token {
			return *claim, nil
		}
	}
	return ClaimLink{}, database.ErrNotFound
}

func (m *memoryRepository) ListClaims(context.Context, uuid.UUID) ([]ClaimLink, error) {
	return nil, nil
}

func (m *memoryRepository) SaveClaim(_ context.Context, claim *ClaimLink) error {
	if claim.ID == uuid.Nil {
		claim.ID = uuid.New()
	}
	copied := *claim
	m.claims[claim.ID] = &copied
	return nil
}

type stubSources struct {
	facts []Fact
	claim Claim
}

func (s *stubSources) ProjectFacts(context.Context, *uuid.UUID) ([]Fact, error) { return s.facts, nil }
func (s *stubSources) PaymentFacts(context.Context, *uuid.UUID) ([]Fact, error) { return nil, nil }
func (s *stubSources) OwnerClaim(context.Context, uuid.UUID, uuid.UUID) (Claim, error) {
	return s.claim, nil
}

type stubChain struct {
	transaction ethereum.Transaction
	seen        bool
	receipt     *ethereum.Receipt
	head        uint64
}

func (c *stubChain) BlockNumber(context.Context) (uint64, error) { return c.head, nil }
func (c *stubChain) TransactionByHash(context.Context, string) (ethereum.Transaction, bool, error) {
	return c.transaction, c.seen, nil
}
func (c *stubChain) TransactionReceipt(context.Context, string) (ethereum.Receipt, bool, error) {
	if c.receipt == nil {
		return ethereum.Receipt{}, false, nil
	}
	return *c.receipt, true, nil
}
func (c *stubChain) BlockTime(context.Context, uint64) (time.Time, error) {
	return time.Date(2026, time.October, 18, 9, 0, 0, 0, time.UTC), nil
}

var testConfig = config.Onchain{ChainID: 11155111, MinConfirmations: 1, ExplorerURL: "https://sepolia.etherscan.io"}

const anchorHash = "0x1111111111111111111111111111111111111111111111111111111111111111"

func reportFact(projectID uuid.UUID, subject uuid.UUID, work string) Fact {
	return Fact{Kind: "daily_report", SubjectID: subject, ProjectID: &projectID, Date: time.Now(), Fields: map[string]any{"workDone": work}}
}

func anchoredService(t *testing.T, sources *stubSources) (*Service, *memoryRepository, *stubChain) {
	t.Helper()
	repository := newMemory()
	chain := &stubChain{}
	service := NewService(repository, sources, chain, testConfig)
	batch, err := service.CreateBatch(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("create batch: %v", err)
	}
	chain.seen = true
	chain.transaction = ethereum.Transaction{Hash: anchorHash, From: "0xabc", ChainID: big.NewInt(11155111), Input: batch.Calldata}
	chain.receipt = &ethereum.Receipt{Status: 1, BlockNumber: 100}
	chain.head = 101
	result, err := service.Anchor(context.Background(), batch.ID, anchorHash)
	if err != nil || result.Status != anchorConfirmed {
		t.Fatalf("anchor: %+v %v", result, err)
	}
	return service, repository, chain
}

func TestEvidenceStatusFollowsTheAnchoredFingerprint(t *testing.T) {
	projectID := uuid.New()
	kept, edited, removed := uuid.New(), uuid.New(), uuid.New()
	sources := &stubSources{facts: []Fact{
		reportFact(projectID, kept, "Pengecoran kolom"),
		reportFact(projectID, edited, "Hujan, pekerjaan berhenti 2 jam"),
		reportFact(projectID, removed, "Bekisting lantai 2"),
	}}
	service, _, _ := anchoredService(t, sources)

	added := uuid.New()
	sources.facts = []Fact{
		reportFact(projectID, kept, "Pengecoran kolom"),
		reportFact(projectID, edited, "Cuaca cerah, pekerjaan lancar"),
		reportFact(projectID, added, "Pemasangan bata"),
	}
	evidence, err := service.ProjectEvidence(context.Background(), projectID)
	if err != nil {
		t.Fatal(err)
	}
	status := map[uuid.UUID]string{}
	for _, item := range evidence.Items {
		status[item.SubjectID] = item.Status
	}
	want := map[uuid.UUID]string{kept: StatusAnchored, edited: StatusChanged, removed: StatusDeleted, added: StatusPending}
	for subject, expected := range want {
		if status[subject] != expected {
			t.Errorf("subject %s: status %q, want %q", subject, status[subject], expected)
		}
	}
	if evidence.Summary != (Summary{Anchored: 1, Pending: 1, Changed: 1, Deleted: 1}) {
		t.Errorf("summary = %+v", evidence.Summary)
	}
}

func TestAnchorChecksTheTransactionData(t *testing.T) {
	projectID := uuid.New()
	repository := newMemory()
	chain := &stubChain{}
	service := NewService(repository, &stubSources{facts: []Fact{reportFact(projectID, uuid.New(), "Galian")}}, chain, testConfig)
	batch, err := service.CreateBatch(context.Background(), uuid.New())
	if err != nil {
		t.Fatal(err)
	}

	result, err := service.Anchor(context.Background(), batch.ID, anchorHash)
	if err != nil || result.Status != anchorWaiting {
		t.Fatalf("unseen transaction: %+v %v, want waiting", result, err)
	}

	chain.seen = true
	chain.transaction = ethereum.Transaction{ChainID: big.NewInt(11155111), Input: "0x424e474e" + "00"}
	if _, err := service.Anchor(context.Background(), batch.ID, anchorHash); !errors.Is(err, errTxWrongData) {
		t.Fatalf("wrong data: got %v", err)
	}

	chain.transaction.Input = batch.Calldata
	chain.receipt = &ethereum.Receipt{Status: 0, BlockNumber: 5}
	if _, err := service.Anchor(context.Background(), batch.ID, anchorHash); !errors.Is(err, errTxReverted) {
		t.Fatalf("reverted: got %v", err)
	}

	chain.receipt = &ethereum.Receipt{Status: 1, BlockNumber: 5}
	chain.head = 5
	result, err = service.Anchor(context.Background(), batch.ID, anchorHash)
	if err != nil || result.Status != anchorConfirmed || result.Batch.TxHash != anchorHash {
		t.Fatalf("confirmed: %+v %v", result, err)
	}
	if _, err := service.CreateBatch(context.Background(), uuid.New()); !errors.Is(err, errNothingToAnchor) {
		t.Fatalf("nothing new: got %v", err)
	}
}

func TestOnlyTheRegisteredOwnerWalletCanApprove(t *testing.T) {
	projectID, termID := uuid.New(), uuid.New()
	sources := &stubSources{
		facts: []Fact{reportFact(projectID, uuid.New(), "Struktur lantai 4")},
		claim: Claim{TermID: termID, ProjectID: projectID, ProjectCode: "DEMO-RSIA", ProjectName: "RSIA Cirebon",
			ContractNumber: "DEMO-KTR-RSIA-01", Label: "Termin 2", Progress: 46, NetAmount: 12_730_000_000,
			ClaimedOn: time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)},
	}
	service, _, _ := anchoredService(t, sources)
	ctx := context.Background()

	owner, _ := secp256k1.GeneratePrivateKey()
	stranger, _ := secp256k1.GeneratePrivateKey()
	address := func(key *secp256k1.PrivateKey) string {
		raw := key.PubKey().SerializeUncompressed()
		return "0x" + hex.EncodeToString(ethereum.Keccak256(raw[1:])[12:])
	}
	if _, err := service.SetApprover(ctx, projectID, address(owner)); err != nil {
		t.Fatal(err)
	}
	link, err := service.CreateClaimLink(ctx, projectID, termID)
	if err != nil {
		t.Fatal(err)
	}
	public, err := service.PublicClaim(ctx, link.Token)
	if err != nil {
		t.Fatal(err)
	}
	sign := func(key *secp256k1.PrivateKey) string {
		compact := ecdsa.SignCompact(key, ApprovalDigest(testConfig.ChainID, public.TypedData.Message), false)
		return "0x" + hex.EncodeToString(append(append([]byte{}, compact[1:]...), compact[0]))
	}

	if _, err := service.Approve(ctx, link.Token, sign(stranger)); !errors.Is(err, errWrongSigner) {
		t.Fatalf("stranger: got %v", err)
	}
	status, err := service.Approve(ctx, link.Token, sign(owner))
	if err != nil || status.ApprovedAt == nil || !status.StillSame {
		t.Fatalf("owner approval: %+v %v", status, err)
	}
	if _, err := service.Approve(ctx, link.Token, sign(owner)); !errors.Is(err, errClaimApproved) {
		t.Fatalf("second approval: got %v", err)
	}
}
