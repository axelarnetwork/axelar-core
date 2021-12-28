package utils

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
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

// ValidateString checks if the given string is entirely composed of utf8 runs and if its normalized as NFKC
func ValidateString(str string, canBeEmpty bool) error {
	if !canBeEmpty && len(str) == 0 {
		return fmt.Errorf("string is empty")
	}

	if !utf8.ValidString(str) {
		return fmt.Errorf("not an utf8 string")

	}

	if !norm.NFKC.IsNormalString(str) {
		return fmt.Errorf("wrong normalization")
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
