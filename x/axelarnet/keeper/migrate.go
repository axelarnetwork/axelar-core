package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
)

// GetMigrationHandler returns the handler that performs in-place store migrations from v0.22 to v0.23. The
// migration includes
// - set existing IBC transfers status to completed
// - remove transfer under failedTransfer prefix
// - remove nonce
func GetMigrationHandler(k Keeper) func(_ sdk.Context) error {
	return func(ctx sdk.Context) error {
		migrationRouteIBCTransferQueue(ctx, k)
		setIBCTransfersCompleted(ctx, k)
		removeFailedTransfers(ctx, k)
		removeNonce(ctx, k)

		return nil
	}
}

func setIBCTransfersCompleted(ctx sdk.Context, k Keeper) {
	for _, t := range k.getIBCTransfers(ctx) {
		t.Status = types.TransferCompleted
		k.setTransfer(ctx, t)
	}
}

func removeFailedTransfers(ctx sdk.Context, k Keeper) {
	iter := k.getStore(ctx).IteratorNew(failedTransferPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		k.getStore(ctx).DeleteRaw(iter.Key())
	}
}

func removeNonce(ctx sdk.Context, k Keeper) {
	k.getStore(ctx).DeleteRaw(nonceKey.Bytes())
}

func migrationRouteIBCTransferQueue(ctx sdk.Context, k Keeper) {
	oldQueue := GetOldIBCTransferQueue(ctx, k)
	for !oldQueue.IsEmpty() {
		var t types.IBCTransfer
		oldQueue.Dequeue(&t)
		// enqueue should overwrite with the transfer status pending
		k.GetIBCTransferQueue(ctx).Enqueue(getTransferKey(t.ID), &t)
	}
}

// GetOldIBCTransferQueue returns the queue of IBC transfers
func GetOldIBCTransferQueue(ctx sdk.Context, keeper Keeper) utils.KVQueue {
	return utils.NewGeneralKVQueue(
		"ibc_transfer_queue",
		keeper.getStore(ctx),
		keeper.Logger(ctx),
		func(value codec.ProtoMarshaler) utils.Key {
			transfer := value.(*types.IBCTransfer)
			return utils.KeyFromBz(transfer.ID.Bytes())
		},
	)
}
