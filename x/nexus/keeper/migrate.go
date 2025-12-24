package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
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
		linkDeleted := deleteLinkDepositState(ctx, k)
		ctx.Logger().Info(fmt.Sprintf("deleted %d deprecated link-deposit keys from nexus store", linkDeleted))

		rateLimitDeleted := deleteRateLimitState(ctx, k)
		ctx.Logger().Info(fmt.Sprintf("deleted %d deprecated rate-limit keys from nexus store", rateLimitDeleted))

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

// deleteLinkDepositState removes the linked addresses and latest deposit address
// entries that were created by the link-deposit protocol. This data is no longer
// used since the Link command has been removed.
func deleteLinkDepositState(ctx sdk.Context, k Keeper) int {
	store := k.getStore(ctx)
	totalDeleted := 0

	// Delete all entries from linkedAddressesPrefix and latestDepositAddressPrefix
	for _, prefix := range []utils.Key{linkedAddressesPrefix, latestDepositAddressPrefix} {
		iter := store.Iterator(prefix)
		defer utils.CloseLogError(iter, k.Logger(ctx))

		var keysToDelete []utils.Key
		for ; iter.Valid(); iter.Next() {
			keysToDelete = append(keysToDelete, utils.KeyFromBz(iter.Key()))
		}

		for _, k := range keysToDelete {
			store.Delete(k)
		}

		totalDeleted += len(keysToDelete)
	}

	return totalDeleted
}

// deleteRateLimitState removes the rate limit and transfer epoch state.
// Rate limiting has been removed from the protocol.
func deleteRateLimitState(ctx sdk.Context, k Keeper) int {
	store := k.getStore(ctx)
	totalDeleted := 0

	// Delete all entries from rateLimitPrefix and transferEpochPrefix
	for _, prefix := range []key.Key{rateLimitPrefix, transferEpochPrefix} {
		iter := store.IteratorNew(prefix)
		defer utils.CloseLogError(iter, k.Logger(ctx))

		var keysToDelete []key.Key
		for ; iter.Valid(); iter.Next() {
			keysToDelete = append(keysToDelete, key.FromBz(iter.Key()))
		}

		for _, k := range keysToDelete {
			store.DeleteNew(k)
		}

		totalDeleted += len(keysToDelete)
	}

	return totalDeleted
}

func addModuleParamGateway(ctx sdk.Context, k Keeper) {
	k.params.Set(ctx, types.KeyGateway, types.DefaultParams().Gateway)
}

func addModuleParamEndBlockerLimit(ctx sdk.Context, k Keeper) {
	k.params.Set(ctx, types.KeyEndBlockerLimit, types.DefaultParams().EndBlockerLimit)
}
