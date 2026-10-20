package ethereum

import (
	"encoding/hex"
	"errors"
	"math/big"
	"strings"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

const (
	privateKeyLength   = 32
	dynamicFeeTxType   = 0x02
	compactRecoveryMin = 27
)

var ErrInvalidPrivateKey = errors.New("private key must be 32 bytes of hex, optionally prefixed with 0x")

type Signer struct {
	key     *secp256k1.PrivateKey
	address string
}

func NewSigner(rawKey string) (*Signer, error) {
	raw, err := hex.DecodeString(strings.TrimPrefix(strings.TrimSpace(rawKey), hexPrefix))
	if err != nil || len(raw) != privateKeyLength {
		return nil, ErrInvalidPrivateKey
	}
	key := secp256k1.PrivKeyFromBytes(raw)
	if key.Key.IsZero() {
		return nil, ErrInvalidPrivateKey
	}
	uncompressed := key.PubKey().SerializeUncompressed()
	address := hex.EncodeToString(Keccak256(uncompressed[1:])[12:])
	return &Signer{key: key, address: hexPrefix + checksumBody(address)}, nil
}

func (s *Signer) Address() string {
	return s.address
}

type DynamicFeeTx struct {
	ChainID              *big.Int
	Nonce                uint64
	MaxPriorityFeePerGas *big.Int
	MaxFeePerGas         *big.Int
	Gas                  uint64
	To                   string
	Value                *big.Int
	Data                 []byte
}

type SignedTx struct {
	Raw  string
	Hash string
}

func (tx DynamicFeeTx) fields() (rlpList, error) {
	to, err := hex.DecodeString(strings.TrimPrefix(strings.ToLower(tx.To), hexPrefix))
	if err != nil || len(to) != addressHexLength/2 {
		return nil, ErrInvalidAddress
	}
	return rlpList{
		rlpUint(tx.ChainID),
		rlpUint64(tx.Nonce),
		rlpUint(tx.MaxPriorityFeePerGas),
		rlpUint(tx.MaxFeePerGas),
		rlpUint64(tx.Gas),
		rlpBytes(to),
		rlpUint(tx.Value),
		rlpBytes(tx.Data),
		rlpList{},
	}, nil
}

func (tx DynamicFeeTx) SigningHash() ([]byte, error) {
	fields, err := tx.fields()
	if err != nil {
		return nil, err
	}
	return Keccak256([]byte{dynamicFeeTxType}, fields.encodeRLP()), nil
}

func (s *Signer) Sign(tx DynamicFeeTx) (SignedTx, error) {
	fields, err := tx.fields()
	if err != nil {
		return SignedTx{}, err
	}
	digest := Keccak256([]byte{dynamicFeeTxType}, fields.encodeRLP())
	compact := ecdsa.SignCompact(s.key, digest, false)
	recovery := uint64(compact[0] - compactRecoveryMin)
	r := new(big.Int).SetBytes(compact[1:33])
	sValue := new(big.Int).SetBytes(compact[33:65])
	signed := append(fields, rlpUint64(recovery), rlpUint(r), rlpUint(sValue))
	encoded := append([]byte{dynamicFeeTxType}, signed.encodeRLP()...)
	return SignedTx{
		Raw:  hexPrefix + hex.EncodeToString(encoded),
		Hash: hexPrefix + hex.EncodeToString(Keccak256(encoded)),
	}, nil
}
