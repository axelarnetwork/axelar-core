package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetMigrationHandler returns a migration handler
func GetMigrationHandler(k Keeper) func(sdk.Context) error {
	return func(ctx sdk.Context) error {
		return nil
	}
}
