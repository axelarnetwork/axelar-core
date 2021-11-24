package utils

import (
	"crypto/sha256"
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
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
