package utils

import (
	"crypto/sha256"
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
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

// GetNonce deterministically calculates a nonce using the context's header hash and gas meter
func GetNonce(ctx sdk.Context) Nonce {
	bz := make([]byte, 16)
	binary.LittleEndian.PutUint64(bz, uint64(ctx.BlockGasMeter().GasConsumed()))
	bz = append(bz, ctx.HeaderHash()...)
	return sha256.Sum256(bz)
}
