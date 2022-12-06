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

	if err := utils.MigrateKeys(store, key.FromStr(poolNamePrefixOld),&types.Pool{}, func(pool *types.Pool) error { k.setPool(ctx, *pool); return nil }); err != nil {
		return err
	}

	iter := store.IteratorNew(key.FromStr(pendingRefundPrefixOld))
	if iter.Valid() {
		return errors.New("there should be no refundable messages in the store")
	}

	return nil
}
