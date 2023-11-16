package utils

import (
	"encoding/hex"
	"strings"
)

// Decode a hex string. Hex string can be optionally prefixed with 0x.
func HexDecode(input string) ([]byte, error) {
	if strings.HasPrefix(input, "0x") {
		input = input[2:]
	}

	return hex.DecodeString(input)
}
