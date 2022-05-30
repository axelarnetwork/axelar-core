package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// GetMigrationHandler returns the handler that performs in-place store migrations from v0.19 to v0.20. The
// migration includes:
// - delete all polls
func GetMigrationHandler(k Keeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		emptyPollQueue(ctx, k)
		deleteAllPolls(ctx, k)

		return nil
	}
}

func emptyPollQueue(ctx sdk.Context, k Keeper) {
	pollQueue := k.GetPollQueue(ctx)

	var pollMeta exported.PollMetadata
	for pollQueue.Dequeue(&pollMeta) {
	}
}

func deleteAllPolls(ctx sdk.Context, k Keeper) {
	var keys [][]byte

	iter := k.getKVStore(ctx).Iterator(pollPrefix)
	for ; iter.Valid(); iter.Next() {
		keys = append(keys, iter.Key())
	}
	utils.CloseLogError(iter, k.Logger(ctx))

	for _, key := range keys {
		k.getKVStore(ctx).DeleteRaw(key)
	}
}
