package ethereum

import "math/big"

const (
	rlpShortString = 0x80
	rlpLongString  = 0xb7
	rlpShortList   = 0xc0
	rlpLongList    = 0xf7
	rlpShortLimit  = 55
)

type rlpItem interface {
	encodeRLP() []byte
}

type rlpBytes []byte

type rlpList []rlpItem

func rlpUint(value *big.Int) rlpBytes {
	if value == nil || value.Sign() == 0 {
		return rlpBytes{}
	}
	return rlpBytes(value.Bytes())
}

func rlpUint64(value uint64) rlpBytes {
	return rlpUint(new(big.Int).SetUint64(value))
}

func (b rlpBytes) encodeRLP() []byte {
	if len(b) == 1 && b[0] < rlpShortString {
		return []byte{b[0]}
	}
	return append(rlpHeader(rlpShortString, rlpLongString, len(b)), b...)
}

func (l rlpList) encodeRLP() []byte {
	var payload []byte
	for _, item := range l {
		payload = append(payload, item.encodeRLP()...)
	}
	return append(rlpHeader(rlpShortList, rlpLongList, len(payload)), payload...)
}

func rlpHeader(short, long byte, length int) []byte {
	if length <= rlpShortLimit {
		return []byte{short + byte(length)}
	}
	size := new(big.Int).SetInt64(int64(length)).Bytes()
	return append([]byte{long + byte(len(size))}, size...)
}
