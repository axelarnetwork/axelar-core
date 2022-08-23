package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/utils/slices"
)

const uaxlAsset = "uaxl"

// GetMigrationHandler returns the handler that performs in-place store migrations from v0.14 to v0.15. The
// migration includes:
// - deregister uaxl asset for all EVM chains
// - remove uaxl fee info
// - add module parameters
// - migrate .Maintainers in chain stats into .MaintainerStates
func GetMigrationHandler(k Keeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		deregisterUaxlAsset(ctx, k)
		removeUaxlEVMFeeInfo(ctx, k)
		addModuleParams(ctx, k)

		if err := migrateChainMaintainers(ctx, k); err != nil {
			return err
		}

		return nil
	}
}

func deregisterUaxlAsset(ctx sdk.Context, k Keeper) {
	for _, chain := range k.GetChains(ctx) {
		if chain.Module != evmTypes.ModuleName {
			continue
		}

		chainState, ok := k.getChainState(ctx, chain)
		if !ok {
			continue
		}

		chainState.Assets = slices.Filter(chainState.Assets, func(a exported.Asset) bool {
			return a.Denom != uaxlAsset
		})

		k.SetChainState(ctx, &chainState)
	}
}

func removeUaxlEVMFeeInfo(ctx sdk.Context, k Keeper) {
	for _, feeInfo := range k.getFeeInfos(ctx) {
		if feeInfo.Asset != uaxlAsset {
			continue
		}

		chain, ok := k.GetChain(ctx, feeInfo.Chain)
		if !ok || chain.Module != evmTypes.ModuleName {
			continue
		}

		k.getStore(ctx).Delete(assetFeePrefix.Append(utils.LowerCaseKey(chain.Name.String())).Append(utils.KeyFromStr(uaxlAsset)))
	}
}

func addModuleParams(ctx sdk.Context, k Keeper) {
	defaultParams := types.DefaultParams()
	k.params.Set(ctx, types.KeyChainMaintainerMissingVoteThreshold, defaultParams.ChainMaintainerMissingVoteThreshold)
	k.params.Set(ctx, types.KeyChainMaintainerIncorrectVoteThreshold, defaultParams.ChainMaintainerIncorrectVoteThreshold)
	k.params.Set(ctx, types.KeyChainMaintainerCheckWindow, defaultParams.ChainMaintainerCheckWindow)
}

func migrateChainMaintainers(ctx sdk.Context, k Keeper) error {
	for _, chainState := range k.getChainStates(ctx) {
		for _, maintainer := range chainState.Maintainers {
			err := k.AddChainMaintainer(ctx, chainState.Chain, maintainer)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
