package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

// Migrate8to9 returns the handler that performs in-place store migrations
func Migrate8to9(k Keeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		shrinkMaintainerStateBitmaps(ctx, k)
		return nil
	}
}

func shrinkMaintainerStateBitmaps(ctx sdk.Context, k Keeper) {
	maxSize := types.MaxBitmapSize()

	for _, chain := range k.GetChains(ctx) {
		for _, ms := range k.getChainMaintainerStates(ctx, chain.Name) {
			ms.MissingVotes.TrueCountCache.SetMaxSize(maxSize)
			ms.IncorrectVotes.TrueCountCache.SetMaxSize(maxSize)
			k.setChainMaintainerState(ctx, &ms)
		}
	}
}
