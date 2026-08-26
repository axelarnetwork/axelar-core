package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/utils/funcs"
)

// Migrate8to9 returns the handler that performs in-place store migrations
func Migrate8to9(k Keeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		shrinkMaintainerStateBitmaps(ctx, k)
		rebuildProcessingMessageIndex(ctx, k)
		return nil
	}
}

// rebuildProcessingMessageIndex clears the old processing index and re-indexes the
// currently-processing messages under the new FIFO order index. It walks the (small)
// processing index, not the all-time message store; the pre-migration value is the
// message id. All old entries are deleted so no pre-migration format lingers under the
// shared prefix; the backlog has no arrival-order data, so it is re-sequenced in id
// order this once.
func rebuildProcessingMessageIndex(ctx sdk.Context, k Keeper) {
	kvStore := k.getStore(ctx)

	var msgs []exported.GeneralMessage
	var oldKeys [][]byte
	iter := kvStore.IteratorNew(processingMessagePrefix)
	for ; iter.Valid(); iter.Next() {
		oldKeys = append(oldKeys, append([]byte(nil), iter.Key()...))
		if m, ok := k.GetMessage(ctx, string(iter.Value())); ok && m.Is(exported.Processing) {
			msgs = append(msgs, m)
		}
	}
	utils.CloseLogError(iter, k.Logger(ctx))

	for _, oldKey := range oldKeys {
		kvStore.DeleteRaw(oldKey)
	}

	for _, m := range msgs {
		funcs.MustNoErr(k.setProcessingMessageID(ctx, m))
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
