package keeper

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewAddressValidator returns the callback for validating cosmos SDK addresses
func NewAddressValidator(k Keeper) nexus.AddressValidator {
	return func(ctx sdk.Context, address nexus.CrossChainAddress) error {
		var addrPrefix string
		var ok bool
		if address.Chain == exported.Axelarnet {
			addrPrefix = sdk.GetConfig().GetBech32AccountAddrPrefix()
			ok = true
		} else {
			var chain types.CosmosChain
			chain, ok = k.GetCosmosChainByName(ctx, address.Chain.Name)
			addrPrefix = chain.AddrPrefix
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
