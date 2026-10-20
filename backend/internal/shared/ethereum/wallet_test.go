package ethereum

import (
	"encoding/hex"
	"math/big"
	"strings"
	"testing"
)

func TestRLPMatchesTheSpecificationExamples(t *testing.T) {
	long := "Lorem ipsum dolor sit amet, consectetur adipisicing elit"
	cases := map[string]struct {
		item rlpItem
		want string
	}{
		"dog":          {rlpBytes("dog"), "83646f67"},
		"cat dog list": {rlpList{rlpBytes("cat"), rlpBytes("dog")}, "c88363617483646f67"},
		"empty string": {rlpBytes{}, "80"},
		"empty list":   {rlpList{}, "c0"},
		"zero":         {rlpUint64(0), "80"},
		"fifteen":      {rlpUint64(15), "0f"},
		"1024":         {rlpUint64(1024), "820400"},
		"long string":  {rlpBytes(long), "b838" + hex.EncodeToString([]byte(long))},
		"set theory": {
			rlpList{rlpList{}, rlpList{rlpList{}}, rlpList{rlpList{}, rlpList{rlpList{}}}},
			"c7c0c1c0c3c0c1c0",
		},
	}
	for name, current := range cases {
		if got := hex.EncodeToString(current.item.encodeRLP()); got != current.want {
			t.Errorf("%s: got %s, want %s", name, got, current.want)
		}
	}
}

func TestSignerDerivesTheKnownAddress(t *testing.T) {
	signer, err := NewSigner("0x0000000000000000000000000000000000000000000000000000000000000001")
	if err != nil {
		t.Fatal(err)
	}
	if signer.Address() != "0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf" {
		t.Fatalf("got %s", signer.Address())
	}
	for _, bad := range []string{"", "0x12", "zz" + strings.Repeat("0", 62), "0x" + strings.Repeat("0", 64)} {
		if _, err := NewSigner(bad); err == nil {
			t.Errorf("key %q must be rejected", bad)
		}
	}
}

type decoded struct {
	bytes []byte
	list  []decoded
}

func decodeRLP(t *testing.T, data []byte) (decoded, []byte) {
	t.Helper()
	prefix := data[0]
	switch {
	case prefix < rlpShortString:
		return decoded{bytes: data[:1]}, data[1:]
	case prefix <= rlpLongString:
		size := int(prefix - rlpShortString)
		return decoded{bytes: data[1 : 1+size]}, data[1+size:]
	case prefix < rlpShortList:
		width := int(prefix - rlpLongString)
		size := int(new(big.Int).SetBytes(data[1 : 1+width]).Int64())
		return decoded{bytes: data[1+width : 1+width+size]}, data[1+width+size:]
	default:
		width, size := 0, 0
		if prefix <= rlpLongList {
			size = int(prefix - rlpShortList)
		} else {
			width = int(prefix - rlpLongList)
			size = int(new(big.Int).SetBytes(data[1 : 1+width]).Int64())
		}
		body := data[1+width : 1+width+size]
		var items []decoded
		for len(body) > 0 {
			var item decoded
			item, body = decodeRLP(t, body)
			items = append(items, item)
		}
		return decoded{list: items}, data[1+width+size:]
	}
}

func TestSignedDynamicFeeTxRecoversToTheSigner(t *testing.T) {
	signer, err := NewSigner("0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318")
	if err != nil {
		t.Fatal(err)
	}
	tx := DynamicFeeTx{
		ChainID:              big.NewInt(11155111),
		Nonce:                7,
		MaxPriorityFeePerGas: big.NewInt(1_500_000_000),
		MaxFeePerGas:         big.NewInt(30_000_000_000),
		Gas:                  60_000,
		To:                   "0x05788638f4e23932884765f666D85654547AecB4",
		Value:                big.NewInt(0),
		Data:                 []byte{0xde, 0xad, 0xbe, 0xef},
	}
	if signer.Address() != "0x2c7536E3605D9C16a7a3D7b1898e529396a65c23" {
		t.Fatalf("got %s", signer.Address())
	}
	signed, err := signer.Sign(tx)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := hex.DecodeString(strings.TrimPrefix(signed.Raw, "0x"))
	if raw[0] != dynamicFeeTxType {
		t.Fatalf("typed transaction must start with 0x02, got %x", raw[0])
	}
	if signed.Hash != "0x"+hex.EncodeToString(Keccak256(raw)) {
		t.Fatal("the hash must be keccak256 of the whole typed envelope")
	}
	envelope, rest := decodeRLP(t, raw[1:])
	if len(rest) != 0 || len(envelope.list) != 12 {
		t.Fatalf("expected a 12-field list, got %d fields and %d trailing bytes", len(envelope.list), len(rest))
	}
	fields := envelope.list
	if new(big.Int).SetBytes(fields[0].bytes).Int64() != 11155111 || new(big.Int).SetBytes(fields[1].bytes).Int64() != 7 {
		t.Fatalf("chain id or nonce encoded wrongly: %x %x", fields[0].bytes, fields[1].bytes)
	}
	if hex.EncodeToString(fields[7].bytes) != "deadbeef" || fields[8].list != nil {
		t.Fatalf("data or access list encoded wrongly")
	}

	digest, err := tx.SigningHash()
	if err != nil {
		t.Fatal(err)
	}
	signature := make([]byte, 65)
	new(big.Int).SetBytes(fields[10].bytes).FillBytes(signature[0:32])
	new(big.Int).SetBytes(fields[11].bytes).FillBytes(signature[32:64])
	signature[64] = byte(new(big.Int).SetBytes(fields[9].bytes).Int64())
	recovered, err := RecoverAddress(digest, "0x"+hex.EncodeToString(signature))
	if err != nil {
		t.Fatal(err)
	}
	if recovered != signer.Address() {
		t.Fatalf("signature recovers to %s, want %s", recovered, signer.Address())
	}
}

func TestSigningRejectsABadRecipient(t *testing.T) {
	signer, _ := NewSigner("0x0000000000000000000000000000000000000000000000000000000000000001")
	if _, err := signer.Sign(DynamicFeeTx{ChainID: big.NewInt(1), To: "0x1234"}); err == nil {
		t.Fatal("a short recipient address must be rejected")
	}
}
