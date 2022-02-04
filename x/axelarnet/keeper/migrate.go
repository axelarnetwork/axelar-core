package keeper

import (
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/legacy"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strings"
)

// GetMigrationHandler returns the handler that performs in-place store migrations from v0.13 to v0.14. The
// migration includes:
// - remove all path
// - remove all chain <-> asset mapping
func GetMigrationHandler(k Keeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		// rebuild cosmos chain
		rebuildCosmosChain(ctx, k)
		// delete all path
		deleteByPrefix(ctx, k, legacy.PathPrefix)
		// delete chain by asset
		deleteByPrefix(ctx, k, legacy.ChainByAssetPrefix)
		// delete asset by chain
		deleteByPrefix(ctx, k, legacy.AssetByChainPrefix)

		return nil
	}

}

func rebuildCosmosChain(ctx sdk.Context, k Keeper) {
	for _, chainName := range k.GetCosmosChains(ctx) {
		ibcPath := getIBCPath(ctx, k, chainName)
		assets := getAssets(ctx, k, chainName)
		chain, _ := k.GetCosmosChainByName(ctx, chainName)
		k.SetCosmosChain(ctx, types.CosmosChain{
			Name: chainName,
			// used in migration
			Assets:     append(chain.Assets, assets...),
			IBCPath:    ibcPath,
			AddrPrefix: chain.AddrPrefix,
		})

	}
}

func deleteByPrefix(ctx sdk.Context, k Keeper, keyPrefix utils.StringKey) {
	iter := k.getStore(ctx).Iterator(keyPrefix)

	for ; iter.Valid(); iter.Next() {
		k.getStore(ctx).Delete(iter.GetKey())
	}
}

func getIBCPath(ctx sdk.Context, k Keeper, chain string) string {
	bz := k.getStore(ctx).GetRaw(legacy.PathPrefix.Append(utils.LowerCaseKey(chain)))
	if bz == nil {
		return ""
	}

	return string(bz)
}

func getAssets(ctx sdk.Context, k Keeper, chain string) []types.Asset {
	iter := k.getStore(ctx).Iterator(legacy.AssetByChainPrefix.Append(utils.LowerCaseKey(chain)))

	var assets []types.Asset
	for ; iter.Valid(); iter.Next() {
		var asset types.Asset
		iter.UnmarshalValue(&asset)
		// uaxl accidentally registered as native asset from terra
		if asset.Denom == exported.NativeAsset && !strings.EqualFold(chain, exported.Axelarnet.Name) {
			continue
		}
		assets = append(assets, asset)
	}

	return assets
}
