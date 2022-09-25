package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetMigrationHandler returns the migration handler for the reward module
func GetMigrationHandler(keeper Keeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		return nil
	}
}
