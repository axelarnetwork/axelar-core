package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetMigrationHandler returns the handler that performs in-place store migrations from v0.19 to v0.20. The
// migration includes:
// - delete all polls
// - add EndBlockerLimit parameter
func GetMigrationHandler(k Keeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		return nil
	}
}
