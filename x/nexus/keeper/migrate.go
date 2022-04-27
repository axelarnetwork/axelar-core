package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/slices"
)

const uaxlAsset = "uaxl"

// GetMigrationHandler returns the handler that performs in-place store migrations from v0.14 to v0.15. The
// migration includes:
// - deregister uaxl asset for all EVM chains
func GetMigrationHandler(k Keeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		deregisterUaxlAsset(ctx, k)

		return nil
	}
}

func deregisterUaxlAsset(ctx sdk.Context, k Keeper) {
	for _, chain := range k.GetChains(ctx) {
		if chain.Module != evmTypes.ModuleName {
			continue
		}

		chainState, _ := k.getChainState(ctx, chain)
		chainState.Chain = chain
		chainState.Assets = slices.Filter(chainState.Assets, func(a exported.Asset) bool {
			return a.Denom != uaxlAsset
		})

		k.setChainState(ctx, chainState)
	}
}
