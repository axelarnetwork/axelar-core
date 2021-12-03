package keeper

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewNexusHandler returns the handler for validating cosmos SDK addresses
func NewNexusHandler(k Keeper) nexus.Handler {
	return func(ctx sdk.Context, address nexus.CrossChainAddress) error {
		var addrPrefix string
		var ok bool
		if address.Chain == exported.Axelarnet {
			addrPrefix = sdk.GetConfig().GetBech32AccountAddrPrefix()
			ok = true
		} else {
			addrPrefix, ok = k.GetCosmosChainAddrPrefix(ctx, address.Chain.Name)
		}

		if !ok {
			return fmt.Errorf("no known prefix for chain %s", address.Chain.String())
		}

		bz, err := sdk.GetFromBech32(address.Address, addrPrefix)
		if err != nil {
			return err
		}

		err = sdk.VerifyAddressFormat(bz)
		if err != nil {
			return err
		}

		return nil
	}
}
