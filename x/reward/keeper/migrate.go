package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/axelarnetwork/axelar-core/utils"
	multisigTypes "github.com/axelarnetwork/axelar-core/x/multisig/types"
	"github.com/axelarnetwork/axelar-core/x/reward/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
)

// GetMigrationHandler returns the migration handler for the reward module
// - migrating params key TssRelativeInflationRate to KeyMgmtRelativeInflationRate
// - migrate tss rewards pool to multisig
func GetMigrationHandler(keeper Keeper, paramStoreKey sdk.StoreKey, paramTStoreKey sdk.StoreKey) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		migrateParamKey(ctx, keeper, paramStoreKey, paramTStoreKey)
		return migrateTssPool(ctx, keeper)
	}
}

func migrateParamKey(ctx sdk.Context, keeper Keeper, paramStoreKey sdk.StoreKey, paramTStoreKey sdk.StoreKey) {
	oldSubspace := paramtypes.NewSubspace(keeper.cdc, types.ModuleCdc.LegacyAmino, paramStoreKey, paramTStoreKey, keeper.paramSpace.Name()).
		WithKeyTable(paramtypes.NewKeyTable().RegisterParamSet(&oldParams{}))
	var params oldParams
	oldSubspace.GetParamSet(ctx, &params)
	store := prefix.NewStore(ctx.KVStore(paramStoreKey), append([]byte(oldSubspace.Name()), '/'))
	store.Delete(keyTssRelativeInflationRate)

	keeper.SetParams(ctx, types.Params(params))
}

func migrateTssPool(ctx sdk.Context, keeper Keeper) error {
	pool, ok := keeper.getPool(ctx, tss.ModuleName)
	if !ok {
		return fmt.Errorf("could not find tss reward pool")
	}
	pool.Name = utils.NormalizeString(multisigTypes.ModuleName)
	keeper.setPool(ctx, pool)
	keeper.deletePool(ctx, tss.ModuleName)
	return nil
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

var keyTssRelativeInflationRate = []byte("TssRelativeInflationRate")

type oldParams types.Params

func (m *oldParams) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(types.KeyExternalChainVotingInflationRate, &m.ExternalChainVotingInflationRate, validateExternalChainVotingInflationRate),
		paramtypes.NewParamSetPair(keyTssRelativeInflationRate, &m.KeyMgmtRelativeInflationRate, validateTSSRelativeInflationRate),
	}
}

func validateExternalChainVotingInflationRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("external chain voting inflation rate cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("external chain voting inflation rate too large: %s", v)
	}

	return nil
}

func validateTSSRelativeInflationRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("tss inflation rate cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("tess inflation rate too large: %s", v)
	}

	return nil
}
