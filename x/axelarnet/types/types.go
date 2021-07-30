package types

import (
	"crypto/sha256"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewLinkedAddress creates a new address to make a deposit which can be transferred to another blockchain
func NewLinkedAddress(chain, symbol, recipientAddr string) sdk.AccAddress {
	preImage := []byte(chain)
	preImage = append(preImage, 0)
	preImage = append(preImage, symbol...)
	preImage = append(preImage, 1)
	preImage = append(preImage, recipientAddr...)
	hash := sha256.Sum256(preImage)
	return hash[:20]
}
