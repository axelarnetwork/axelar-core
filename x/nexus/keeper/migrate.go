package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

// Migrate6to7 returns the handler that performs in-place store migrations
func Migrate6to7(k Keeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		addModuleParamGateway(ctx, k)
		addModuleParamEndBlockerLimit(ctx, k)

		return nil
	}
}

func addModuleParamGateway(ctx sdk.Context, k Keeper) {
	k.params.Set(ctx, types.KeyGateway, types.DefaultParams().Gateway)
}

func addModuleParamEndBlockerLimit(ctx sdk.Context, k Keeper) {
	k.params.Set(ctx, types.KeyEndBlockerLimit, types.DefaultParams().EndBlockerLimit)
}
