package evidence

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/ethereum"
)

func TestEncodingMatchesTheEIP712SpecificationVector(t *testing.T) {
	text := func(value string) []byte { return ethereum.Keccak256([]byte(value)) }
	address := func(value string) []byte {
		word, err := ethereum.AddressWord(value)
		if err != nil {
			t.Fatal(err)
		}
		return word
	}
	verifying := address("0xCcCCccccCCCCcCCCCCCcCcCccCcCCCcCcccccccC")
	domain := ethereum.Keccak256(
		text("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"),
		text("Ether Mail"), text("1"), ethereum.Word(big.NewInt(1)), verifying,
	)
	personType := text("Person(string name,address wallet)")
	person := func(name, wallet string) []byte {
		return ethereum.Keccak256(personType, text(name), address(wallet))
	}
	mail := ethereum.Keccak256(
		text("Mail(Person from,Person to,string contents)Person(string name,address wallet)"),
		person("Cow", "0xCD2a3d9F938E13CD947Ec05AbC7FE734Df8DD826"),
		person("Bob", "0xbBbBBBBbbBBBbbbBbbBbbbbBBbBbbbbBbBbbBBbB"),
		text("Hello, Bob!"),
	)
	digest := ethereum.Keccak256([]byte{0x19, 0x01}, domain, mail)

	if got := hex.EncodeToString(domain); got != "f2cee375fa42b42143804025fc449deafd50cc031ca257e0b194a650a912090f" {
		t.Errorf("domain separator = %s", got)
	}
	if got := hex.EncodeToString(digest); got != "be609aee343fb3c4b28e1df9e632fca64fcfaede20f02e86244efddf30957bd2" {
		t.Errorf("digest = %s", got)
	}
}

func sampleMessage() ApprovalMessage {
	return ApprovalMessage{
		Project: "DEMO-RSIA Rumah Sakit Ibu & Anak Cirebon", Contract: "DEMO-KTR-RSIA-01", Term: "Termin 2",
		Progress: "46", Amount: "12730000000", ClaimedOn: "2026-10-01",
		EvidenceRoot: "0x" + hex.EncodeToString(ethereum.Keccak256([]byte("bukti"))),
	}
}

func TestApprovalSignatureRecoversTheOwnerAndBindsEveryField(t *testing.T) {
	key, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	uncompressed := key.PubKey().SerializeUncompressed()
	owner, _ := ethereum.ChecksumAddress("0x" + hex.EncodeToString(ethereum.Keccak256(uncompressed[1:])[12:]))

	message := sampleMessage()
	compact := ecdsa.SignCompact(key, ApprovalDigest(11155111, message), false)
	signature := "0x" + hex.EncodeToString(append(append([]byte{}, compact[1:]...), compact[0]))

	signer, err := ethereum.RecoverAddress(ApprovalDigest(11155111, message), signature)
	if err != nil || signer != owner {
		t.Fatalf("recovered %s (err %v), want %s", signer, err, owner)
	}

	tampered := []func(*ApprovalMessage){
		func(m *ApprovalMessage) { m.Amount = "13730000000" },
		func(m *ApprovalMessage) { m.Progress = "50" },
		func(m *ApprovalMessage) {
			m.EvidenceRoot = "0x" + hex.EncodeToString(ethereum.Keccak256([]byte("lain")))
		},
		func(m *ApprovalMessage) { m.Term = "Termin 3" },
	}
	for index, change := range tampered {
		changed := sampleMessage()
		change(&changed)
		if got, _ := ethereum.RecoverAddress(ApprovalDigest(11155111, changed), signature); got == owner {
			t.Errorf("change %d still verifies as the owner", index)
		}
	}
	if got, _ := ethereum.RecoverAddress(ApprovalDigest(1, message), signature); got == owner {
		t.Error("a signature for Sepolia must not verify on another chain")
	}
}

func TestTypedDataDescribesTheSameStructAsTheDigest(t *testing.T) {
	data := NewTypedData(11155111, sampleMessage())
	fields := data.Types[primaryType]
	signature := primaryType + "("
	for index, field := range fields {
		if index > 0 {
			signature += ","
		}
		signature += field.Type + " " + field.Name
	}
	if signature+")" != approvalType {
		t.Fatalf("typed data fields %q do not match the hashed type %q", signature+")", approvalType)
	}
}
