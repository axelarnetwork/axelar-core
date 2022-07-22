package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
)

// GetMigrationHandler returns the handler that performs in-place store migrations from v0.19 to v0.20. The
// migration includes:
// - delete all polls
// - add EndBlockerLimit parameter
func GetMigrationHandler(k Keeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		emptyPollQueue(ctx, k)

		deleteAllWithPrefix(ctx, k, pollPrefix)
		deleteAllWithPrefix(ctx, k, votesPrefix)
		deleteAllWithPrefix(ctx, k, voterPrefix)

		addEndBlockerLimitParam(ctx, k)

		return nil
	}
}

func emptyPollQueue(ctx sdk.Context, k Keeper) {
	pollQueue := k.GetPollQueue(ctx)

	var pollMeta exported.PollMetadata
	for pollQueue.Dequeue(&pollMeta) {
	}
}

func deleteAllWithPrefix(ctx sdk.Context, k Keeper, prefix utils.Key) {
	var keys [][]byte

	iter := k.getKVStore(ctx).Iterator(prefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		keys = append(keys, iter.Key())
	}
	for _, key := range keys {
		k.getKVStore(ctx).DeleteRaw(key)
	}
}

func addEndBlockerLimitParam(ctx sdk.Context, k Keeper) {
	k.paramSpace.Set(ctx, types.KeyEndBlockerLimit, types.DefaultParams().EndBlockerLimit)
}
