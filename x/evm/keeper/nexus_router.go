package keeper

import (
	"fmt"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
)

// NewNexusHandler returns the handler for validating bitcoin addresses
func NewNexusHandler() nexus.Handler {
	return &handler{}
}

type handler struct{}

func (h *handler) Validate(ctx sdk.Context, address nexus.CrossChainAddress) error {
	if !common.IsHexAddress(address.Address) {
		return fmt.Errorf("not an hex address")
	}

	return nil
}

func (h *handler) With(properties ...nexus.HandlerProperty) nexus.Handler {
	var newHandler nexus.Handler = h
	for _, property := range properties {
		newHandler = property(newHandler)
	}
	return newHandler
}
