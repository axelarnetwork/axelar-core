package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
)

// GetMigrationHandler returns the handler that performs in-place store migrations from v0.17 to v0.18. The
// migration includes:
// - delete all pending polls
// - migrate all completed polls
func GetMigrationHandler(k Keeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		deleteAllPendingPolls(ctx, k)
		migrateAllCompletedPolls(ctx, k)

		return nil
	}
}

func deleteAllPendingPolls(ctx sdk.Context, k Keeper) {
	iter := k.getKVStore(ctx).Iterator(pollPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var pollMetadata exported.PollMetadata
		iter.UnmarshalValue(&pollMetadata)

		if !pollMetadata.Is(exported.Pending) {
			continue
		}

		k.newPollStore(ctx, pollMetadata.Key).DeletePoll()
	}
}

func migrateAllCompletedPolls(ctx sdk.Context, k Keeper) {
	iter := k.getKVStore(ctx).Iterator(pollPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var pollMetadata exported.PollMetadata
		iter.UnmarshalValue(&pollMetadata)

		if !pollMetadata.Is(exported.Completed) {
			continue
		}

		poll := k.newPollStore(ctx, pollMetadata.Key)
		voterIter := k.getKVStore(ctx).Iterator(voterPrefix.AppendStr(poll.key.String()))
		for ; voterIter.Valid(); voterIter.Next() {
			poll.KVStore.Set(voterIter.GetKey(), &types.VoteRecord{IsLate: false})
		}
		utils.CloseLogError(voterIter, k.Logger(ctx))
		// The actual completed at cannot be retrieved anymore, but need to
		// make it valid
		pollMetadata.CompletedAt = 1
		poll.SetMetadata(pollMetadata)
	}
}
