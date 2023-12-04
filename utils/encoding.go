package utils

import (
	"encoding/hex"
	"strings"
)

// Encode a bytearray as hex string with 0x prefix.
func HexEncode(input []byte) string {
	return "0x" + hex.EncodeToString(input)
}

// Decode a hex string. Hex string can be optionally prefixed with 0x.
func HexDecode(input string) ([]byte, error) {
	return hex.DecodeString(strings.TrimPrefix(input, "0x"))
}
