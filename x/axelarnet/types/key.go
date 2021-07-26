package types

import (
	"crypto/sha256"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of the module
	ModuleName = "axelarnet"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey to be used for routing msgs
	RouterKey = ModuleName

	QuerierRoute = ModuleName

	RestRoute = ModuleName

	Axelar = "axelar"
)

// NewLinkedAddress create a new address to make a deposit which can be transferred to another blockchain
func NewLinkedAddress(chain, symbol, recipientAddr string) sdk.AccAddress {
	preImage := []byte(chain)
	preImage = append(preImage, symbol...)
	preImage = append(preImage, recipientAddr...)
	hash := sha256.Sum256(preImage)
	return hash[:20]
}
