package keeper

import (
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewNexusHandler returns the handler for validating bitcoin addresses
func NewNexusHandler(k Keeper) nexus.Handler {
	return &handler{keeper: k}
}

type handler struct {
	keeper Keeper
}

func (h *handler) Validate(ctx sdk.Context, address nexus.CrossChainAddress) error {
	if _, err := btcutil.DecodeAddress(address.Address, h.keeper.GetNetwork(ctx).Params()); err != nil {
		return err
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
