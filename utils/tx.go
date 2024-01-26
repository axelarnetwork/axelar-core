package utils

import (
	"crypto/sha256"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetTxHash returns the hash of the current transaction. If it's called inside begin/end block, it returns the hash of empty bytes
func GetTxHash(ctx sdk.Context) []byte {
	hash := sha256.Sum256(ctx.TxBytes())
	return hash[:]
}
