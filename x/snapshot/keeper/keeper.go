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
	tss         types.Tss
	cdc         *codec.Codec
	params      subspace.Subspace
}

// NewKeeper creates a new keeper for the staking module
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, paramSpace params.Subspace, broadcaster types.Broadcaster, staking types.StakingKeeper, slasher types.Slasher, tss types.Tss) Keeper {
	return Keeper{
		storeKey:    key,
		cdc:         cdc,
		staking:     staking,
		params:      paramSpace.WithKeyTable(types.KeyTable()),
		slasher:     slasher,
		broadcaster: broadcaster,
		tss:         tss,
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
func (k Keeper) TakeSnapshot(ctx sdk.Context, validatorCount int64) error {
	s, ok := k.GetLatestSnapshot(ctx)

	if !ok {
		k.setLatestCounter(ctx, 0)
		return k.executeSnapshot(ctx, 0, validatorCount)
	}

	lockingPeriod := k.getLockingPeriod(ctx)
	if s.Timestamp.Add(lockingPeriod).After(ctx.BlockTime()) {
		return fmt.Errorf("not enough time has passed since last snapshot, need to wait %s longer",
			s.Timestamp.Add(lockingPeriod).Sub(ctx.BlockTime()).String())
	}

	k.setLatestCounter(ctx, s.Counter+1)
	return k.executeSnapshot(ctx, s.Counter+1, validatorCount)
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

func isActive(ctx sdk.Context, slasher types.Slasher, validator exported.Validator) bool {
	signingInfo, found := slasher.GetValidatorSigningInfo(ctx, validator.GetConsAddr())

	return found && !signingInfo.Tombstoned && signingInfo.MissedBlocksCounter <= 0 && !validator.IsJailed()
}

func hasProxyRegistered(ctx sdk.Context, broadcaster types.Broadcaster, validator exported.Validator) bool {
	return broadcaster.GetProxy(ctx, validator.GetOperator()) != nil
}

func isTssRegistered(ctx sdk.Context, tss types.Tss, validator exported.Validator) bool {
	return tss.GetValidatorDeregisteredBlockHeight(ctx, validator.GetOperator()) <= 0
}

func (k Keeper) executeSnapshot(ctx sdk.Context, nextCounter int64, validatorCount int64) error {
	var validators []exported.Validator
	snapshotTotalPower, validatorsTotalPower := sdk.ZeroInt(), sdk.ZeroInt()

	validatorIter := func(_ int64, validator sdkExported.ValidatorI) (stop bool) {
		validatorsTotalPower = validatorsTotalPower.AddRaw(validator.GetConsensusPower())

		if !isActive(ctx, k.slasher, validator) {
			return false
		}

		if !hasProxyRegistered(ctx, k.broadcaster, validator) {
			return false
		}

		if !isTssRegistered(ctx, k.tss, validator) {
			return false
		}

		snapshotTotalPower = snapshotTotalPower.AddRaw(validator.GetConsensusPower())
		validators = append(validators, validator)

		return len(validators) == int(validatorCount)
	}
	k.staking.IterateLastValidators(ctx, validatorIter)

	snapshot := exported.Snapshot{
		Validators:           validators,
		Timestamp:            ctx.BlockTime(),
		Height:               ctx.BlockHeight(),
		TotalPower:           snapshotTotalPower,
		ValidatorsTotalPower: validatorsTotalPower,
		Counter:              nextCounter,
	}

	ctx.KVStore(k.storeKey).Set(counterKey(nextCounter), k.cdc.MustMarshalBinaryLengthPrefixed(snapshot))

	return nil
}

func (k Keeper) setLatestCounter(ctx sdk.Context, counter int64) {
	ctx.KVStore(k.storeKey).Set([]byte(lastCounterKey), k.cdc.MustMarshalBinaryLengthPrefixed(counter))
}

func counterKey(counter int64) []byte {
	return []byte(fmt.Sprintf("%s%d", counterPrefix, counter))
}
