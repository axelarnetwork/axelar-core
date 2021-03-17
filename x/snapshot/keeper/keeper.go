package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	sdkExported "github.com/cosmos/cosmos-sdk/x/staking/exported"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/snapshot/types"
)

const lastCounterKey = "lastcounter"
const counterPrefix = "counter_"

// Make sure the keeper implements the Snapshotter interface
var _ exported.Snapshotter = Keeper{}

// Keeper represents the snapshot keeper
type Keeper struct {
	storeKey    sdk.StoreKey
	staking     types.StakingKeeper
	slasher     types.Slasher
	broadcaster types.Broadcaster
	cdc         *codec.Codec
	params      subspace.Subspace
}

// NewKeeper creates a new keeper for the staking module
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, paramSpace params.Subspace, broadcaster types.Broadcaster, staking types.StakingKeeper, slasher types.Slasher) Keeper {
	return Keeper{
		storeKey:    key,
		cdc:         cdc,
		staking:     staking,
		params:      paramSpace.WithKeyTable(types.KeyTable()),
		slasher:     slasher,
		broadcaster: broadcaster,
	}
}

// Logger returns the logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetParams sets the module's parameters
func (k Keeper) SetParams(ctx sdk.Context, set types.Params) {
	k.params.SetParamSet(ctx, &set)
}

// GetParams gets the module's parameters
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.params.GetParamSet(ctx, &params)
	return
}

// TakeSnapshot attempts to create a new snapshot
func (k Keeper) TakeSnapshot(ctx sdk.Context) error {
	s, ok := k.GetLatestSnapshot(ctx)

	if !ok {
		k.setLatestCounter(ctx, 0)
		return k.executeSnapshot(ctx, 0)
	}

	lockingPeriod := k.getLockingPeriod(ctx)
	if s.Timestamp.Add(lockingPeriod).After(ctx.BlockTime()) {
		return fmt.Errorf("not enough time has passed since last snapshot, need to wait %s longer",
			s.Timestamp.Add(lockingPeriod).Sub(ctx.BlockTime()).String())
	}

	k.setLatestCounter(ctx, s.Counter+1)
	return k.executeSnapshot(ctx, s.Counter+1)
}

func (k Keeper) getLockingPeriod(ctx sdk.Context) time.Duration {
	var lockingPeriod time.Duration
	k.params.Get(ctx, types.KeyLockingPeriod, &lockingPeriod)
	return lockingPeriod
}

// GetLatestSnapshot retrieves the last created snapshot
func (k Keeper) GetLatestSnapshot(ctx sdk.Context) (exported.Snapshot, bool) {
	r := k.GetLatestCounter(ctx)
	if r == -1 {
		return exported.Snapshot{}, false
	}

	return k.GetSnapshot(ctx, r)
}

// GetSnapshot retrieves a snapshot by counter, if it exists
func (k Keeper) GetSnapshot(ctx sdk.Context, counter int64) (exported.Snapshot, bool) {
	bz := ctx.KVStore(k.storeKey).Get(counterKey(counter))
	if bz == nil {

		return exported.Snapshot{}, false
	}

	var snapshot exported.Snapshot
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &snapshot)

	return snapshot, true
}

// GetLatestCounter returns the latest snapshot counter
func (k Keeper) GetLatestCounter(ctx sdk.Context) int64 {
	bz := ctx.KVStore(k.storeKey).Get([]byte(lastCounterKey))

	if bz == nil {

		return -1
	}

	var i int64
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &i)
	return i
}

func (k Keeper) executeSnapshot(ctx sdk.Context, nextCounter int64) error {
	var validators []exported.Validator
	fnAppend := func(_ int64, v sdkExported.ValidatorI) (stop bool) {
		validators = append(validators, v)
		return false
	}

	k.staking.IterateLastValidators(ctx, fnAppend)

	activeStake := sdk.ZeroInt()
	for _, validator := range validators {
		activeStake = activeStake.AddRaw(validator.GetConsensusPower())
	}

	snapshot := exported.Snapshot{
		Validators: validators,
		Timestamp:  ctx.BlockTime(),
		Height:     ctx.BlockHeight(),
		TotalPower: activeStake,
		Counter:    nextCounter,
	}

	// filters
	filterActive := func(vals []exported.Validator) ([]exported.Validator, error) {
		return utils.FilterActiveValidators(ctx, k.slasher, vals)
	}
	filterProxies := func(vals []exported.Validator) ([]exported.Validator, error) {
		return utils.FilterProxies(ctx, k.broadcaster, vals), nil
	}
	filteredSnapshot, err := snapshot.Filter(filterActive)
	if err != nil {
		return err
	}
	filteredSnapshot, err = filteredSnapshot.Filter(filterProxies)
	if err != nil {
		return err
	}

	ctx.KVStore(k.storeKey).Set(counterKey(nextCounter), k.cdc.MustMarshalBinaryLengthPrefixed(filteredSnapshot))

	return nil
}

func (k Keeper) setLatestCounter(ctx sdk.Context, counter int64) {
	ctx.KVStore(k.storeKey).Set([]byte(lastCounterKey), k.cdc.MustMarshalBinaryLengthPrefixed(counter))
}

func counterKey(counter int64) []byte {
	return []byte(fmt.Sprintf("%s%d", counterPrefix, counter))
}
