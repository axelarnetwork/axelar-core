package types

import (
	"crypto/sha256"
	"fmt"

	"github.com/axelarnetwork/axelar-core/utils"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

// NewLinkedAddress creates a new address to make a deposit which can be transferred to another blockchain
func NewLinkedAddress(ctx sdk.Context, chain, symbol, recipientAddr string) sdk.AccAddress {
	nonce := utils.GetNonce(ctx.HeaderHash(), ctx.BlockGasMeter())
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s_%s_%s_%x", chain, symbol, recipientAddr, nonce)))
	return hash[:address.Len]
}

// GetEscrowAddress creates an address for an ibc denomination
func GetEscrowAddress(denom string) sdk.AccAddress {
	hash := sha256.Sum256([]byte(denom))
	return hash[:address.Len]
}
