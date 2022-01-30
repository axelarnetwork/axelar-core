package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/legacy"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

// GetMigrationHandler returns the handler that performs in-place store migrations from v0.13 to v0.14. The
// migration includes:
// - migrate assets from CosmosChain struct to ChainState
// - migrate token min amount from TokenMetaData to ChainState
func GetMigrationHandler(k Keeper, a types.AxelarnetKeeper, e types.EVMBaseKeeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		err := migrateChainState(ctx, k)
		if err != nil {
			return err
		}

		err = migrateCosmosChainAssets(ctx, k, a)
		if err != nil {
			return err
		}

		err = migrateEvmAssets(ctx, k, e)
		if err != nil {
			return err
		}

		return nil
	}

}

func migrateChainState(ctx sdk.Context, k Keeper) error {

	for _, chain := range legacy.GetChains(k.getStore(ctx)) {
		oldChainState, ok := legacy.GetChainState(k.getStore(ctx), chain)
		if !ok {
			return fmt.Errorf("failed to get chain state %s", chain.Name)
		}

		newChainState := types.ChainState{
			Chain: exported.Chain{
				Name:                  chain.Name,
				SupportsForeignAssets: chain.SupportsForeignAssets,
				KeyType:               chain.KeyType,
				Module:                chain.Module,
			},
			Maintainers: oldChainState.Maintainers,
			Activated:   oldChainState.Activated,
		}

		deleteChainState(ctx, k, chain.Name)
		k.setChainState(ctx, newChainState)
	}

	return nil
}

func migrateCosmosChainAssets(ctx sdk.Context, k Keeper, a types.AxelarnetKeeper) error {
	// get all cosmos chains
	for _, chainName := range a.GetCosmosChains(ctx) {
		cosmosChain, ok := a.GetCosmosChainByName(ctx, chainName)
		if !ok {
			return fmt.Errorf("failed to find cosmos chain %s", cosmosChain.Name)
		}

		chain, ok := k.GetChain(ctx, cosmosChain.Name)
		if !ok {
			return fmt.Errorf("failed to find chain %s", cosmosChain.Name)
		}

		for _, asset := range cosmosChain.Assets {
			// register asset as native asset in chain state
			err := k.RegisterAsset(ctx, chain, exported.NewAsset(asset.Denom, asset.MinAmount, true))
			if err != nil {
				return err
			}

			// register native assets from cosmos chains to Axelarnet. Axelarnet is a router between EVM <-> Cosmos chains
			if chain.Name != axelarnet.Axelarnet.Name {
				err = k.RegisterAsset(ctx, axelarnet.Axelarnet, exported.NewAsset(asset.Denom, asset.MinAmount, false))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func migrateEvmAssets(ctx sdk.Context, k Keeper, e types.EVMBaseKeeper) error {

	for _, chain := range k.GetChains(ctx) {
		if chain.Module != evmtypes.ModuleName {
			continue
		}

		// get chain keeper
		chainK := e.ForChain(chain.Name)

		for _, token := range chainK.GetTokens(ctx) {
			// register to chain stats
			err := k.RegisterAsset(ctx, chain, exported.NewAsset(token.GetAsset(), token.GetMinAmount(), false))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func deleteChainState(ctx sdk.Context, k Keeper, chain string) {
	k.getStore(ctx).Delete(chainStatePrefix.Append(utils.LowerCaseKey(chain)))
}
