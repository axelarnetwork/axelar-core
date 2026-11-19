package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// ValidateAddress validates the given cross chain address
func (k Keeper) ValidateAddress(ctx sdk.Context, address exported.CrossChainAddress) error {
	validate, err := k.getAddressValidator(address.Chain.Module)
	if err != nil {
		return err
	}

	if err := validate(ctx, address); err != nil {
		return err
	}

	return nil
}
