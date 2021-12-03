package keeper

import (
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewNexusHandler returns the handler for validating bitcoin addresses
func NewNexusHandler(k Keeper) nexus.Handler {
	return func(ctx sdk.Context, address nexus.CrossChainAddress) error {
		if _, err := btcutil.DecodeAddress(address.Address, k.GetNetwork(ctx).Params()); err != nil {
			return err
		}

		return nil
	}
}
