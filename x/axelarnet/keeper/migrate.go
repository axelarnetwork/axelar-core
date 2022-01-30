package keeper

import (
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/legacy"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetMigrationHandler returns the handler that performs in-place store migrations from v0.13 to v0.14. The
// migration includes:
// - remove all path
// - remove all chain <-> asset mapping
func GetMigrationHandler(k Keeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		// delete all path
		deleteByPrefix(ctx, k, legacy.PathPrefix)
		// delete chain by asset
		deleteByPrefix(ctx, k, legacy.ChainByAssetPrefix)
		// delete asset by chain
		deleteByPrefix(ctx, k, legacy.AssetByChainPrefix)

		return nil
	}

}

func deleteByPrefix(ctx sdk.Context, k Keeper, keyPrefix utils.StringKey) {
	iter := k.getStore(ctx).Iterator(keyPrefix)

	for ; iter.Valid(); iter.Next() {
		k.getStore(ctx).Delete(iter.GetKey())
	}
}
