package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
)

// Migrate7to8 returns the handler that performs in-place store migrations
func Migrate7to8(k Keeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		params := k.GetParams(ctx)
		params.RouteTimeoutWindow = types.DefaultParams().RouteTimeoutWindow

		k.SetParams(ctx, params)
		return nil
	}
}
