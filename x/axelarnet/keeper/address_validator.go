package keeper

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewAddressValidator returns the callback for validating cosmos SDK addresses
func NewAddressValidator(getter types.CosmosChainGetter) nexus.AddressValidator {
	return func(ctx sdk.Context, address nexus.CrossChainAddress) error {
		chain, ok := getter(ctx, address.Chain.Name)
		if !ok {
			return fmt.Errorf("no known prefix for chain %s", address.Chain.String())
		}

		bz, err := sdk.GetFromBech32(address.Address, chain.AddrPrefix)
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
