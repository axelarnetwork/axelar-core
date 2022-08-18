package keeper

import (
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
