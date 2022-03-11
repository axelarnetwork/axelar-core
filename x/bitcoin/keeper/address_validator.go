package keeper

import (
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewAddressValidator returns the callback for validating bitcoin addresses
func NewAddressValidator(k Keeper) nexus.AddressValidator {
	return func(ctx sdk.Context, address nexus.CrossChainAddress) error {
		if _, err := btcutil.DecodeAddress(address.Address, k.GetNetwork(ctx).Params()); err != nil {
			return err
		}

		return nil
	}
}
