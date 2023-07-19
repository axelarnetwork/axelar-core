package utils

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strings"
	"unicode/utf8"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	"golang.org/x/text/unicode/norm"
)

// IndexOf returns the index of str in the slice; -1 if not found
func IndexOf(strs []string, str string) int {
	for i := range strs {
		if strs[i] == str {
			return i
		}
	}

	return -1
}

// NormalizeString normalizes a string as NFKC
func NormalizeString(str string) string {
	return norm.NFKC.String(str)
}

// ValidateString checks if the given string is:
//
// 1. non-empty
// 2. entirely composed of utf8 runes
// 3. normalized as NFKC
// 4. does not contain any forbidden Unicode code points
func ValidateString(str string, forbidden ...string) error {
	var f string
	if len(forbidden) == 0 {
		f = DefaultDelimiter
	} else {
		f = strings.Join(forbidden, "")
	}

	return validateString(str, false, f)
}

// ValidateStringAllowEmpty checks if the given string is:
//
// 1. entirely composed of utf8 runes
// 2. normalized as NFKC
// 3. does not contain any forbidden Unicode code points
func ValidateStringAllowEmpty(str string, forbidden string) error {
	return validateString(str, true, forbidden)
}

func validateString(str string, canBeEmpty bool, forbidden string) error {
	if !canBeEmpty && len(str) == 0 {
		return fmt.Errorf("string is empty")
	}

	if !utf8.ValidString(str) {
		return fmt.Errorf("not an utf8 string")
	}

	if !norm.NFKC.IsNormalString(str) {
		return fmt.Errorf("wrong normalization")
	}

	if len(forbidden) == 0 {
		return nil
	}

	forbidden = norm.NFKC.String(forbidden)
	if strings.ContainsAny(str, forbidden) {
		return fmt.Errorf("string '%s' must not contain any of '%s'", str, forbidden)
	}

	return nil
}

// Nonce defines a 32 byte array representing a deterministically-generated nonce
type Nonce [sha256.Size]byte

// GetNonce deterministically calculates a nonce using a hash and gas meter
func GetNonce(hash tmbytes.HexBytes, gasMeter sdk.GasMeter) Nonce {
	bz := make([]byte, 16)
	if gasMeter != nil {
		binary.LittleEndian.PutUint64(bz, uint64(gasMeter.GasConsumed()))
		bz = append(bz, hash...)
	}
	return sha256.Sum256(bz)
}
