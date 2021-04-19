package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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
	slasher     exported.Slasher
	broadcaster exported.Broadcaster
	tss         exported.Tss
	cdc         *codec.LegacyAmino
	params      params.Subspace
}

// NewKeeper creates a new keeper for the staking module
func NewKeeper(cdc *codec.LegacyAmino, key sdk.StoreKey, paramSpace params.Subspace, broadcaster exported.Broadcaster, staking types.StakingKeeper, slasher exported.Slasher, tss exported.Tss) Keeper {
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

// TakeSnapshot attempts to create a new snapshot; if subsetSize equals 0, snapshot will be created with all validators
func (k Keeper) TakeSnapshot(ctx sdk.Context, subsetSize int64) error {
	s, ok := k.GetLatestSnapshot(ctx)

	if !ok {
		k.setLatestCounter(ctx, 0)
		return k.executeSnapshot(ctx, 0, subsetSize)
	}

	lockingPeriod := k.getLockingPeriod(ctx)
	if s.Timestamp.Add(lockingPeriod).After(ctx.BlockTime()) {
		return fmt.Errorf("not enough time has passed since last snapshot, need to wait %s longer",
			s.Timestamp.Add(lockingPeriod).Sub(ctx.BlockTime()).String())
	}

	k.setLatestCounter(ctx, s.Counter+1)
	return k.executeSnapshot(ctx, s.Counter+1, subsetSize)
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

func (k Keeper) executeSnapshot(ctx sdk.Context, nextCounter int64, subsetSize int64) error {
	var validators []exported.Validator
	snapshotTotalPower, validatorsTotalPower := sdk.ZeroInt(), sdk.ZeroInt()

	validatorIter := func(_ int64, validator stakingtypes.ValidatorI) (stop bool) {
		validatorsTotalPower = validatorsTotalPower.AddRaw(validator.GetConsensusPower())

		if !exported.IsValidatorActive(ctx, k.slasher, validator) {
			return false
		}

		if !exported.DoesValidatorHasProxyRegistered(ctx, k.broadcaster, validator) {
			return false
		}

		if !exported.IsValidatorTssRegistered(ctx, k.tss, validator) {
			return false
		}

		snapshotTotalPower = snapshotTotalPower.AddRaw(validator.GetConsensusPower())
		validators = append(validators, validator)

		// if subsetSize equals 0, we will iterate through all validators and potentially put them all into the snapshot
		return len(validators) == int(subsetSize)
	}
	// IterateBondedValidatorsByPower(https://github.com/cosmos/cosmos-sdk/blob/7fc7b3f6ff82eb5ede52881778114f6b38bd7dfa/x/staking/keeper/alias_functions.go#L33) iterates validators by power in descending order
	k.staking.IterateBondedValidatorsByPower(ctx, validatorIter)

	if subsetSize > 0 && len(validators) != int(subsetSize) {
		return fmt.Errorf("only %d validators are eligible for keygen which is less than desired subset size %d", len(validators), subsetSize)
	}

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
