package utils

import (
	"encoding/hex"
	"strings"
)

// Decode a hex string. Hex string can be optionally prefixed with 0x.
func HexDecode(input string) ([]byte, error) {
	return hex.DecodeString(strings.TrimPrefix(input, "0x"))
}
