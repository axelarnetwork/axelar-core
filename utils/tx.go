package utils

import (
	"crypto/sha256"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetTxHash returns the hash of the current transaction, and whether it's a tx, i.e if the current context is a transaction or begin/end blocker
func GetTxHash(ctx sdk.Context) ([]byte, bool) {
	tx := ctx.TxBytes()
	hash := sha256.Sum256(tx)
	return hash[:], len(tx) > 0
}
