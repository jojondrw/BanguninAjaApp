package ethereum

import (
	"errors"
	"math/big"
	"testing"
)

func TestParseQuantityReadsHexNumbers(t *testing.T) {
	value, err := ParseQuantity("0x38d7ea4c68000")
	if err != nil || value.String() != "1000000000000000" {
		t.Fatalf("got %v %v", value, err)
	}
	chain, err := ParseUint64Quantity("0xaa36a7")
	if err != nil || chain != 11155111 {
		t.Fatalf("got %d %v", chain, err)
	}
}

func TestParseQuantityRejectsMalformedNumbers(t *testing.T) {
	for _, raw := range []string{"", "0x", "aa36a7", "0xzz", "12"} {
		if _, err := ParseQuantity(raw); !errors.Is(err, ErrInvalidQuantity) {
			t.Errorf("%q: got %v", raw, err)
		}
	}
	if _, err := ParseUint64Quantity("0x1ffffffffffffffff"); !errors.Is(err, ErrInvalidQuantity) {
		t.Fatalf("a value above 64 bits must be rejected, got %v", err)
	}
}

func TestEncodeQuantityWritesMinimalHex(t *testing.T) {
	if got := EncodeQuantity(11921168); got != "0xb5e710" {
		t.Fatalf("got %s", got)
	}
	if got := EncodeQuantity(0); got != "0x0" {
		t.Fatalf("got %s", got)
	}
}

func TestWeiForAmountUsesTheSimulationRate(t *testing.T) {
	cases := []struct {
		amount, perEther int64
		want             string
	}{
		{50_000_000, 1_000_000_000_000, "50000000000000"},
		{1, 1_000_000_000_000, "1000000"},
		{1_000_000_000_000, 1_000_000_000_000, "1000000000000000000"},
		{1, 3, "333333333333333334"},
		{9_000_000_000_000_000, 1, "9000000000000000000000000000000000"},
	}
	for _, current := range cases {
		if got := WeiForAmount(current.amount, current.perEther); got.String() != current.want {
			t.Errorf("%d at %d per ether: got %s, want %s", current.amount, current.perEther, got, current.want)
		}
	}
}

func TestFormatEtherTrimsTrailingZeros(t *testing.T) {
	cases := map[string]string{
		"50000000000000":      "0.00005",
		"1000000000000000000": "1",
		"1500000000000000000": "1.5",
		"1":                   "0.000000000000000001",
		"0":                   "0",
	}
	for wei, want := range cases {
		value, _ := new(big.Int).SetString(wei, 10)
		if got := FormatEther(value); got != want {
			t.Errorf("%s wei: got %s, want %s", wei, got, want)
		}
	}
}
