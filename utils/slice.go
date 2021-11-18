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

// Nonce calculates a byte slice using the context's tx bytes and gas meter
func Nonce(ctx sdk.Context) [sha256.Size]byte {
	bz := make([]byte, 16)
	binary.LittleEndian.PutUint64(bz, uint64(ctx.BlockGasMeter().GasConsumed()))
	bz = append(bz, ctx.TxBytes()...)
	return sha256.Sum256(bz)
}
