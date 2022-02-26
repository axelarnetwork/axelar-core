package keeper

import (
	"crypto/sha256"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/reward/exported"
	"github.com/axelarnetwork/axelar-core/x/reward/types"
)

var (
	poolNamePrefix      = utils.KeyFromStr("pool")
	pendingRefundPrefix = utils.KeyFromStr("refund")
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
	key := poolNamePrefix.Append(utils.LowerCaseKey(name))
	ok := k.getStore(ctx).Get(key, &pool)
	if !ok {
		return newPool(ctx, k, k.banker, k.distributor, k.staker, types.NewPool(name))
	}

	return newPool(ctx, k, k.banker, k.distributor, k.staker, pool)
}

func (k Keeper) getPools(ctx sdk.Context) []types.Pool {
	var pools []types.Pool

	store := k.getStore(ctx)
	iter := store.Iterator(poolNamePrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var pool types.Pool
		iter.UnmarshalValue(&pool)

		pools = append(pools, pool)
	}

	return pools
}

func (k Keeper) setPool(ctx sdk.Context, pool types.Pool) {
	key := poolNamePrefix.Append(utils.LowerCaseKey(pool.Name))
	k.getStore(ctx).Set(key, &pool)
}

func (k Keeper) getStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}

// SetPendingRefund saves pending refundable message
func (k Keeper) SetPendingRefund(ctx sdk.Context, req types.RefundMsgRequest, refund types.Refund) error {
	hash := sha256.Sum256(k.cdc.MustMarshalLengthPrefixed(&req))
	k.getStore(ctx).Set(pendingRefundPrefix.Append(utils.KeyFromBz(hash[:])), &refund)
	return nil
}

// GetPendingRefund retrieves a pending refundable message
func (k Keeper) GetPendingRefund(ctx sdk.Context, req types.RefundMsgRequest) (types.Refund, bool) {
	var refund types.Refund
	hash := sha256.Sum256(k.cdc.MustMarshalLengthPrefixed(&req))
	ok := k.getStore(ctx).Get(pendingRefundPrefix.Append(utils.KeyFromBz(hash[:])), &refund)

	return refund, ok
}

// DeletePendingRefund retrieves a pending refundable message
func (k Keeper) DeletePendingRefund(ctx sdk.Context, req types.RefundMsgRequest) {
	hash := sha256.Sum256(k.cdc.MustMarshalLengthPrefixed(&req))
	k.getStore(ctx).Delete(pendingRefundPrefix.Append(utils.KeyFromBz(hash[:])))
}
