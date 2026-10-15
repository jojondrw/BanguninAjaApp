package ethereum

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
)

const (
	hexBase        = 16
	decimalBase    = 10
	etherDecimals  = 18
	maxQuantityHex = 64
)

var (
	ErrInvalidQuantity = errors.New("quantity must be 0x followed by hexadecimal digits")

	weiPerEther = new(big.Int).Exp(big.NewInt(decimalBase), big.NewInt(etherDecimals), nil)
)

func ParseQuantity(raw string) (*big.Int, error) {
	body, hasPrefix := strings.CutPrefix(raw, hexPrefix)
	if !hasPrefix || body == "" || len(body) > maxQuantityHex || !isHex(body) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidQuantity, raw)
	}
	value, ok := new(big.Int).SetString(body, hexBase)
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrInvalidQuantity, raw)
	}
	return value, nil
}

func ParseUint64Quantity(raw string) (uint64, error) {
	value, err := ParseQuantity(raw)
	if err != nil {
		return 0, err
	}
	if !value.IsUint64() {
		return 0, fmt.Errorf("%w: %q does not fit 64 bits", ErrInvalidQuantity, raw)
	}
	return value.Uint64(), nil
}

func EncodeQuantity(value uint64) string {
	return hexPrefix + new(big.Int).SetUint64(value).Text(hexBase)
}

func WeiForAmount(amount, amountPerEther int64) *big.Int {
	numerator := new(big.Int).Mul(big.NewInt(amount), weiPerEther)
	divisor := big.NewInt(amountPerEther)
	quotient, remainder := new(big.Int).QuoRem(numerator, divisor, new(big.Int))
	if remainder.Sign() > 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	return quotient
}

func FormatEther(wei *big.Int) string {
	whole, fraction := new(big.Int).QuoRem(wei, weiPerEther, new(big.Int))
	if fraction.Sign() == 0 {
		return whole.String()
	}
	digits := fraction.String()
	padded := strings.Repeat("0", etherDecimals-len(digits)) + digits
	return whole.String() + "." + strings.TrimRight(padded, "0")
}
