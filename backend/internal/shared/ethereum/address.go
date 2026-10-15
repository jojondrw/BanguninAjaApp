package ethereum

import (
	"encoding/hex"
	"errors"
	"strings"

	"golang.org/x/crypto/sha3"
)

const (
	hexPrefix          = "0x"
	addressHexLength   = 40
	hashHexLength      = 64
	checksumUpperFloor = '8'
)

var (
	ErrInvalidAddress  = errors.New("address must be 0x followed by 40 hexadecimal characters")
	ErrAddressChecksum = errors.New("address checksum does not match EIP-55, copy it again from the wallet")
	ErrZeroAddress     = errors.New("the zero address cannot receive payments")
)

func ChecksumAddress(raw string) (string, error) {
	body, hasPrefix := strings.CutPrefix(strings.TrimSpace(raw), hexPrefix)
	if !hasPrefix || len(body) != addressHexLength || !isHex(body) {
		return "", ErrInvalidAddress
	}
	lower := strings.ToLower(body)
	if strings.Trim(lower, "0") == "" {
		return "", ErrZeroAddress
	}
	checksummed := checksumBody(lower)
	if isMixedCase(body) && checksummed != body {
		return "", ErrAddressChecksum
	}
	return hexPrefix + checksummed, nil
}

func SameAddress(left, right string) bool {
	return left != "" && strings.EqualFold(left, right)
}

func IsTransactionHash(value string) bool {
	body, hasPrefix := strings.CutPrefix(value, hexPrefix)
	return hasPrefix && len(body) == hashHexLength && isHex(body)
}

func checksumBody(lower string) string {
	hasher := sha3.NewLegacyKeccak256()
	_, _ = hasher.Write([]byte(lower))
	digest := hex.EncodeToString(hasher.Sum(nil))

	var checksummed strings.Builder
	for index, char := range lower {
		if char >= 'a' && digest[index] >= checksumUpperFloor {
			checksummed.WriteRune(char - 'a' + 'A')
			continue
		}
		checksummed.WriteRune(char)
	}
	return checksummed.String()
}

func isHex(value string) bool {
	if value == "" {
		return false
	}
	for _, char := range value {
		if !strings.ContainsRune("0123456789abcdefABCDEF", char) {
			return false
		}
	}
	return true
}

func isMixedCase(value string) bool {
	return strings.ToLower(value) != value && strings.ToUpper(value) != value
}
