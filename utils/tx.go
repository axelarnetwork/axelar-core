package utils

import (
	"crypto/sha256"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetTxHash returns the hash of the current transaction
func GetTxHash(ctx sdk.Context) []byte {
	tx := ctx.TxBytes()
	hash := sha256.Sum256(tx)
	return hash[:]
}

// GetTxHashAsHex returns the hash of the current transaction as a hex string
func GetTxHashAsHex(ctx sdk.Context) string {
	return HexEncode(GetTxHash(ctx))
}
