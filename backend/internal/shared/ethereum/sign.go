package ethereum

import (
	"encoding/hex"
	"errors"
	"math/big"
	"strings"

	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"golang.org/x/crypto/sha3"
)

const (
	signatureLength   = 65
	recoveryOffset    = 27
	compactHeaderBase = 27
	wordLength        = 32
)

var ErrInvalidSignature = errors.New("signature must be 0x followed by 65 bytes (r, s, v)")

func Keccak256(parts ...[]byte) []byte {
	hasher := sha3.NewLegacyKeccak256()
	for _, part := range parts {
		_, _ = hasher.Write(part)
	}
	return hasher.Sum(nil)
}

func Word(value *big.Int) []byte {
	word := make([]byte, wordLength)
	return value.FillBytes(word)
}

func AddressWord(address string) ([]byte, error) {
	raw, err := hex.DecodeString(strings.TrimPrefix(strings.ToLower(address), hexPrefix))
	if err != nil || len(raw) != addressHexLength/2 {
		return nil, ErrInvalidAddress
	}
	word := make([]byte, wordLength)
	copy(word[wordLength-len(raw):], raw)
	return word, nil
}

func RecoverAddress(digest []byte, signature string) (string, error) {
	raw, err := hex.DecodeString(strings.TrimPrefix(signature, hexPrefix))
	if err != nil || len(raw) != signatureLength {
		return "", ErrInvalidSignature
	}
	v := raw[signatureLength-1]
	if v >= recoveryOffset {
		v -= recoveryOffset
	}
	if v > 1 {
		return "", ErrInvalidSignature
	}
	compact := make([]byte, signatureLength)
	compact[0] = compactHeaderBase + v
	copy(compact[1:], raw[:signatureLength-1])
	key, _, err := ecdsa.RecoverCompact(compact, digest)
	if err != nil {
		return "", ErrInvalidSignature
	}
	uncompressed := key.SerializeUncompressed()
	address := hex.EncodeToString(Keccak256(uncompressed[1:])[12:])
	return hexPrefix + checksumBody(address), nil
}
