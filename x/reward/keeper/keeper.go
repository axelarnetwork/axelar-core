package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/axelar-core/x/reward/exported"
	"github.com/axelarnetwork/axelar-core/x/reward/types"
	"github.com/axelarnetwork/utils/funcs"
)

var (
	poolNamePrefixOld      = "pool"   // Deprecated: migrate to poolNamePrefix
	pendingRefundPrefixOld = "refund" // Deprecated: migrate to pendingRefundPrefix

	poolNamePrefix      = key.RegisterStaticKey(types.ModuleName, 1)
	pendingRefundPrefix = key.RegisterStaticKey(types.ModuleName, 2)
)

var _ types.Rewarder = Keeper{}

// Keeper provides access to all state changes regarding the reward module
type Keeper struct {
	storeKey    sdk.StoreKey
	cdc         codec.BinaryCodec
	paramSpace  paramtypes.Subspace
	banker      types.Banker
	distributor types.Distributor
	staker      types.Staker
}

// NewKeeper returns a new reward keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, paramSpace paramtypes.Subspace, banker types.Banker, distributor types.Distributor, staker types.Staker) Keeper {
	return Keeper{
		cdc:         cdc,
		storeKey:    storeKey,
		paramSpace:  paramSpace.WithKeyTable(types.KeyTable()),
		banker:      banker,
		distributor: distributor,
		staker:      staker,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParams returns the total set of reward parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)

	return params
}

// SetParams sets the total set of reward parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// GetPool returns the reward pool of the given name, or returns an empty reward pool if not found
func (k Keeper) GetPool(ctx sdk.Context, name string) exported.RewardPool {
	var pool types.Pool
	ok := k.getStore(ctx).GetNew(poolNamePrefix.Append(key.FromStrHashed(name)), &pool)
	if !ok {
		return newPool(ctx, k, k.banker, k.distributor, k.staker, types.NewPool(name))
	}

	return newPool(ctx, k, k.banker, k.distributor, k.staker, pool)
}

func (k Keeper) getPools(ctx sdk.Context) []types.Pool {
	var pools []types.Pool

	store := k.getStore(ctx)
	iter := store.IteratorNew(poolNamePrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var pool types.Pool
		iter.UnmarshalValue(&pool)

		pools = append(pools, pool)
	}

	return pools
}

func (k Keeper) setPool(ctx sdk.Context, pool types.Pool) {
	funcs.MustNoErr(k.getStore(ctx).SetNewValidated(poolNamePrefix.Append(key.FromStrHashed(pool.Name)), &pool))
}

func (k Keeper) getStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}

// SetPendingRefund saves pending refundable message
func (k Keeper) SetPendingRefund(ctx sdk.Context, req types.RefundMsgRequest, refund types.Refund) error {
	return k.getStore(ctx).SetNewValidated(pendingRefundPrefix.Append(key.FromBzHashed(k.cdc.MustMarshalLengthPrefixed(&req))), &refund)
}

// GetPendingRefund retrieves a pending refundable message
func (k Keeper) GetPendingRefund(ctx sdk.Context, req types.RefundMsgRequest) (types.Refund, bool) {
	var refund types.Refund
	ok := k.getStore(ctx).GetNew(pendingRefundPrefix.Append(key.FromBzHashed(k.cdc.MustMarshalLengthPrefixed(&req))), &refund)

	return refund, ok
}

// DeletePendingRefund retrieves a pending refundable message
func (k Keeper) DeletePendingRefund(ctx sdk.Context, req types.RefundMsgRequest) {
	k.getStore(ctx).DeleteNew(pendingRefundPrefix.Append(key.FromBzHashed(k.cdc.MustMarshalLengthPrefixed(&req))))
}
