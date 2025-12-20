package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils/key"
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

// Migrate7to8 returns the handler that performs in-place store migrations
func Migrate7to8(k Keeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		deleteLinkDepositDisabledKey(ctx, k)

		return nil
	}
}

// deleteLinkDepositDisabledKey removes the deprecated linkDepositDisabledKey from the store.
// This key was used to toggle the link-deposit protocol on/off, but the feature
// has been permanently disabled and the toggle is no longer needed.
func deleteLinkDepositDisabledKey(ctx sdk.Context, k Keeper) {
	// This was: linkDepositDisabledKey = key.RegisterStaticKey(types.ModuleName, 8)
	deprecatedKey := key.RegisterStaticKey(types.ModuleName, 8)
	k.getStore(ctx).DeleteNew(deprecatedKey)
}

func addModuleParamGateway(ctx sdk.Context, k Keeper) {
	k.params.Set(ctx, types.KeyGateway, types.DefaultParams().Gateway)
}

func addModuleParamEndBlockerLimit(ctx sdk.Context, k Keeper) {
	k.params.Set(ctx, types.KeyEndBlockerLimit, types.DefaultParams().EndBlockerLimit)
}
