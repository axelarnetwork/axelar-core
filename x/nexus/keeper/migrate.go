package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetMigrationHandler returns the handler that performs in-place store migrations from v0.24 to v0.25. The
// migration includes:
//   - migrate maintainer states
func GetMigrationHandler(k Keeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		migrateMaintainerStates(ctx, k)

		return nil
	}
}

func migrateMaintainerStates(ctx sdk.Context, k Keeper) {
	for _, chainState := range k.getChainStates(ctx) {
		for _, maintainerState := range chainState.MaintainerStates {
			maintainerState.Chain = chainState.ChainName()
			k.setChainMaintainerState(ctx, &maintainerState)
		}

		chainState.MaintainerStates = nil
		k.setChainState(ctx, chainState)
	}
}
