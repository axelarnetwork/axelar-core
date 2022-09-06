package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
)

// GetMigrationHandler returns the handler that performs in-place store migrations from v0.24 to v0.25
// The migration includes
// - add transfer limit
// - add end blocker limit
func GetMigrationHandler(k Keeper) func(_ sdk.Context) error {
	return func(ctx sdk.Context) error {
		addTransferLimitParam(ctx, k)
		addEndBlockerLimitParam(ctx, k)

		return nil
	}
}

func addTransferLimitParam(ctx sdk.Context, k Keeper) {
	k.params.Set(ctx, types.KeyTransferLimit, types.DefaultParams().TransferLimit)
}

func addEndBlockerLimitParam(ctx sdk.Context, k Keeper) {
	k.params.Set(ctx, types.KeyEndBlockerLimit, types.DefaultParams().EndBlockerLimit)
}
