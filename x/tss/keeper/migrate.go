package keeper

import (
	"fmt"

	store "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

// GetMigrationHandler returns a migration handler that deletes all keys in the TSS module store.
func GetMigrationHandler(storeKey store.StoreKey) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		store := ctx.KVStore(storeKey)

		iter := store.Iterator(nil, nil)
		defer iter.Close()

		var keysToDelete [][]byte
		for ; iter.Valid(); iter.Next() {
			keysToDelete = append(keysToDelete, iter.Key())
		}

		for _, key := range keysToDelete {
			store.Delete(key)
		}

		ctx.Logger().Info(fmt.Sprintf("deleted %d keys from %s module store", len(keysToDelete), types.ModuleName))

		return nil
	}
}
