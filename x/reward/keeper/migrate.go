package keeper

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/utils"
	multisigTypes "github.com/axelarnetwork/axelar-core/x/multisig/types"
	"github.com/axelarnetwork/axelar-core/x/reward/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func GetMigrationHandler(keeper Keeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		pool, ok := keeper.getPool(ctx, tss.ModuleName)
		if !ok {
			return fmt.Errorf("could not find tss reward pool")
		}
		pool.Name = utils.NormalizeString(multisigTypes.ModuleName)
		keeper.setPool(ctx, pool)
		keeper.deletePool(ctx, tss.ModuleName)
		return nil
	}
}

func (k Keeper) getPool(ctx sdk.Context, name string) (types.Pool, bool) {
	var pool types.Pool
	key := poolNamePrefix.Append(utils.LowerCaseKey(name))
	ok := k.getStore(ctx).Get(key, &pool)
	return pool, ok
}

func (k Keeper) deletePool(ctx sdk.Context, name string) {
	key := poolNamePrefix.Append(utils.LowerCaseKey(name))
	k.getStore(ctx).Delete(key)
}
