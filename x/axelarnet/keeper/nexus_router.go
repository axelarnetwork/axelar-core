package keeper

import (
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewNexusHandler returns the handler for validating cosmos SDK addresses
func NewNexusHandler() nexus.Handler {
	return func(ctx sdk.Context, address nexus.CrossChainAddress) error {
		if _, err := sdk.AccAddressFromBech32(address.Address); err != nil {

		}

		return nil
	}
}
