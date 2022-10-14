package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewAddressValidator returns the callback for validating cosmos SDK addresses
func NewAddressValidator(keeper types.BaseKeeper, bank types.BankKeeper) nexus.AddressValidator {
	return func(ctx sdk.Context, address nexus.CrossChainAddress) error {
		chain, ok := keeper.GetCosmosChainByName(ctx, address.Chain.Name)
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

		if address.Chain.Name.Equals(exported.Axelarnet.Name) && bank.BlockedAddr(bz) {
			return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", address.Address)
		}

		return nil
	}
}
