package ethereum

import (
	"encoding/hex"
	"errors"
	"math/big"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

func TestKeccakMatchesKnownVector(t *testing.T) {
	got := hex.EncodeToString(Keccak256([]byte("")))
	if got != "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470" {
		t.Fatalf("keccak of empty input = %s", got)
	}
}

func TestRecoverAddressReturnsTheSigner(t *testing.T) {
	key, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	uncompressed := key.PubKey().SerializeUncompressed()
	want, _ := ChecksumAddress("0x" + hex.EncodeToString(Keccak256(uncompressed[1:])[12:]))
	digest := Keccak256([]byte("termin 2 DEMO-RSIA"))

	compact := ecdsa.SignCompact(key, digest, false)
	ethereumStyle := append(append([]byte{}, compact[1:]...), compact[0])

	got, err := RecoverAddress(digest, "0x"+hex.EncodeToString(ethereumStyle))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("recovered %s, want %s", got, want)
	}

	other := Keccak256([]byte("termin 3 DEMO-RSIA"))
	if impostor, _ := RecoverAddress(other, "0x"+hex.EncodeToString(ethereumStyle)); impostor == want {
		t.Fatal("a signature must not verify for different data")
	}
}

func TestRecoverAddressRejectsMalformedSignatures(t *testing.T) {
	for _, signature := range []string{"", "0x1234", "0x" + hex.EncodeToString(make([]byte, 65))} {
		if _, err := RecoverAddress(make([]byte, 32), signature); !errors.Is(err, ErrInvalidSignature) {
			t.Errorf("signature %q: got %v, want ErrInvalidSignature", signature, err)
		}
	}
}

func TestWordAndAddressWordArePadded(t *testing.T) {
	if got := hex.EncodeToString(Word(big.NewInt(255))); got != "00000000000000000000000000000000000000000000000000000000000000ff" {
		t.Fatalf("word = %s", got)
	}
	word, err := AddressWord("0x05788638f4e23932884765f666D85654547AecB4")
	if err != nil || hex.EncodeToString(word) != "00000000000000000000000005788638f4e23932884765f666d85654547aecb4" {
		t.Fatalf("address word = %x, err %v", word, err)
	}
}
