package keeper

import (
	"errors"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/axelar-core/x/reward/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Migrate2To3 returns the handler that performs in-place store migrations
func Migrate2To3(k Keeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		return migrateKeyPrefixes(k, ctx)
	}
}

func migrateKeyPrefixes(k Keeper, ctx sdk.Context) error {
	store := k.getStore(ctx)

	if err := migrateKeys(store, key.FromStr(poolNamePrefixOld), func(pool *types.Pool) { k.setPool(ctx, *pool) }); err != nil {
		return err
	}

	iter:= store.IteratorNew(key.FromStr(pendingRefundPrefixOld))
	if iter.Valid(){
		return errors.New("there should not be any refundable messages in the store")
	}

	return nil
}

func migrateKeys[T utils.ValidatedProtoMarshaler](store utils.KVStore, originalKey key.Key, migrateValue func(T)) error {
	// migrate in batches so memory pressure doesn't become too large
	keysToDelete := make([][]byte, 0, 1000)
	for {
		iter := store.IteratorNew(originalKey)

		if !iter.Valid() {
			return iter.Close()
		}

		var value T
		for ; iter.Valid() && len(keysToDelete) < 1000; iter.Next() {
			iter.UnmarshalValue(value)
			migrateValue(value)

			keysToDelete = append(keysToDelete, iter.Key())
		}

		if err := iter.Close(); err != nil {
			return err
		}

		for _, poolKey := range keysToDelete {
			store.DeleteRaw(poolKey)
		}

		keysToDelete = keysToDelete[:0]
	}
}
