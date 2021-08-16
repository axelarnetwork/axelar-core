package types

import (
	"crypto/sha256"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewLinkedAddress creates a new address to make a deposit which can be transferred to another blockchain
func NewLinkedAddress(chain, symbol, recipientAddr string) sdk.AccAddress {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s_%s_%s", chain, symbol, recipientAddr)))
	return hash[:20]
}

// GetEscrowAddress creates an address for an ibc denomination
func GetEscrowAddress(denom string) sdk.AccAddress {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s", denom)))
	return hash[:20]
}
