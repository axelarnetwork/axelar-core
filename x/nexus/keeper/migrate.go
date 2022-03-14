package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

// GetMigrationHandler returns the handler that performs in-place store migrations from v0.14 to v0.15. The
// migration includes:
// - migrate linked addresses key
func GetMigrationHandler(k Keeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		migrateLinkedAddressesKey(ctx, k)
		removeLatestDepositAddress(ctx, k)
		return nil
	}

}

func migrateLinkedAddressesKey(ctx sdk.Context, k Keeper) {
	var results []types.LinkedAddresses

	iter := k.getStore(ctx).Iterator(linkedAddressesPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var linkedAddresses types.LinkedAddresses
		iter.UnmarshalValue(&linkedAddresses)
		k.getStore(ctx).Delete(iter.GetKey())
		results = append(results, linkedAddresses)
	}

	k.Logger(ctx).Info(fmt.Sprintf("migrating %d linked address key", len(results)))
	for _, r := range results {
		k.setLinkedAddresses(ctx, r)
	}
	k.Logger(ctx).Info("migration finished")
}

func removeLatestDepositAddress(ctx sdk.Context, k Keeper) {
	iter := k.getStore(ctx).Iterator(latestDepositAddressPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	k.Logger(ctx).Info("removing latest deposit address key")
	for ; iter.Valid(); iter.Next() {
		k.getStore(ctx).Delete(iter.GetKey())
	}
}
