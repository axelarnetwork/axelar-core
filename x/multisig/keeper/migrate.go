package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetMigrationHandler returns the handler that performs in-place store migrations
func GetMigrationHandler() func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		return nil
	}
}
