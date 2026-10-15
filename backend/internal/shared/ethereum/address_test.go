package ethereum

import (
	"errors"
	"strings"
	"testing"
)

var eip55Vectors = []string{
	"0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed",
	"0xfB6916095ca1df60bB79Ce92cE3Ea74c37c5d359",
	"0xdbF03B407c01E7cD3CBea99509d93f8DDDC8C6FB",
	"0xD1220A0cf47c7B9Be7A2E6BA89F429762e7b9aDb",
}

func TestChecksumAddressMatchesTheEip55Vectors(t *testing.T) {
	for _, vector := range eip55Vectors {
		got, err := ChecksumAddress(strings.ToLower(vector))
		if err != nil || got != vector {
			t.Errorf("lowercase %s: got %q %v", vector, got, err)
		}
		got, err = ChecksumAddress(vector)
		if err != nil || got != vector {
			t.Errorf("checksummed %s: got %q %v", vector, got, err)
		}
	}
}

func TestChecksumAddressAcceptsAllUppercaseBody(t *testing.T) {
	got, err := ChecksumAddress("0x" + strings.ToUpper(eip55Vectors[0][2:]))
	if err != nil || got != eip55Vectors[0] {
		t.Fatalf("got %q %v", got, err)
	}
}

func TestChecksumAddressRejectsBrokenChecksum(t *testing.T) {
	broken := "0x5AAeb6053F3E94C9b9A09f33669435E7Ef1BeAed"
	if _, err := ChecksumAddress(broken); !errors.Is(err, ErrAddressChecksum) {
		t.Fatalf("got %v", err)
	}
}

func TestChecksumAddressRejectsMalformedInput(t *testing.T) {
	cases := []string{
		"",
		"5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed",
		"0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAe",
		"0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAedd",
		"0xZaAeb6053F3E94C9b9A09f33669435E7Ef1BeAed",
	}
	for _, input := range cases {
		if _, err := ChecksumAddress(input); !errors.Is(err, ErrInvalidAddress) {
			t.Errorf("%q: got %v", input, err)
		}
	}
}

func TestChecksumAddressRejectsTheZeroAddress(t *testing.T) {
	if _, err := ChecksumAddress("0x" + strings.Repeat("0", 40)); !errors.Is(err, ErrZeroAddress) {
		t.Fatalf("got %v", err)
	}
}

func TestSameAddressIgnoresCase(t *testing.T) {
	if !SameAddress(eip55Vectors[0], strings.ToLower(eip55Vectors[0])) {
		t.Fatal("addresses differing only in case are the same account")
	}
	if SameAddress("", "") || SameAddress(eip55Vectors[0], eip55Vectors[1]) {
		t.Fatal("empty or different addresses are not the same account")
	}
}

func TestIsTransactionHash(t *testing.T) {
	valid := "0xc9be6b0adf4a05bb83cd9768fabf03e46e489ee6193c7b841668a9236ca62b0f"
	if !IsTransactionHash(valid) {
		t.Fatal("a real Sepolia hash must be accepted")
	}
	for _, invalid := range []string{"", valid[2:], valid[:65], valid + "0", "0x" + strings.Repeat("g", 64)} {
		if IsTransactionHash(invalid) {
			t.Errorf("%q must be rejected", invalid)
		}
	}
}
