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
	staking "github.com/cosmos/cosmos-sdk/x/staking/exported"
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
func (k Keeper) TakeSnapshot(ctx sdk.Context, subsetSize int64) (sdk.Int, sdk.Int, error) {
	s, ok := k.GetLatestSnapshot(ctx)

	if !ok {
		k.setLatestCounter(ctx, 0)
		return k.executeSnapshot(ctx, 0, subsetSize)
	}

	lockingPeriod := k.getLockingPeriod(ctx)
	if s.Timestamp.Add(lockingPeriod).After(ctx.BlockTime()) {
		return sdk.ZeroInt(), sdk.ZeroInt(), fmt.Errorf("not enough time has passed since last snapshot, need to wait %s longer",
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

func (k Keeper) executeSnapshot(ctx sdk.Context, counter int64, subsetSize int64) (sdk.Int, sdk.Int, error) {
	var validators []staking.ValidatorI
	snapshotConsensusPower, totalConsensusPower := sdk.ZeroInt(), sdk.ZeroInt()

	validatorIter := func(_ int64, validator staking.ValidatorI) (stop bool) {
		totalConsensusPower = totalConsensusPower.AddRaw(validator.GetConsensusPower())

		// this explicit type cast is necessary, because snapshot needs to call UnpackInterfaces() on the validator
		// and it is not exposed in the ValidatorI interface
		v, ok := validator.(exported.Validator)
		if !ok {
			k.Logger(ctx).Error(fmt.Sprintf("unexpected validator type: expected %T, got %T", stakingtypes.Validator{}, validator))
			return false
		}

		if !exported.IsValidatorActive(ctx, k.slasher, v) {
			return false
		}

		if !exported.HasProxyRegistered(ctx, k.broadcaster, v) {
			return false
		}

		if !exported.IsValidatorTssRegistered(ctx, k.tss, v) {
			return false
		}

		snapshotConsensusPower = snapshotConsensusPower.AddRaw(validator.GetConsensusPower())
		validators = append(validators, v)

		// if subsetSize equals 0, we will iterate through all validators and potentially put them all into the snapshot
		return len(validators) == int(subsetSize)
	}
	// IterateBondedValidatorsByPower(https://github.com/cosmos/cosmos-sdk/blob/7fc7b3f6ff82eb5ede52881778114f6b38bd7dfa/x/staking/keeper/alias_functions.go#L33) iterates validators by power in descending order
	k.staking.IterateBondedValidatorsByPower(ctx, validatorIter)

	minMinBondFractionPerShare := k.tss.GetMinBondFractionPerShare(ctx)

	var participants []exported.Validator
	for _, validator := range validators {
		if !minMinBondFractionPerShare.IsMet(sdk.NewInt(validator.GetConsensusPower()), totalConsensusPower) {
			snapshotConsensusPower = snapshotConsensusPower.SubRaw(validator.GetConsensusPower())
			continue
		}

		participants = append(participants, exported.NewValidator(validator, sdk.ZeroInt()))
	}

	if len(participants) == 0 {
		return sdk.ZeroInt(), sdk.ZeroInt(), fmt.Errorf("no validator is eligible for keygen")
	}

	if subsetSize > 0 && len(participants) != int(subsetSize) {
		return sdk.ZeroInt(), sdk.ZeroInt(), fmt.Errorf("only %d validators are eligible for keygen which is less than desired subset size %d", len(validators), subsetSize)
	}

	// Since IterateBondedValidatorsByPower iterates validators by power in descending order, the last participant is
	// the one with least amount of bond among all participants
	bondPerShare := participants[len(participants)-1].GetConsensusPower()
	totalPower := sdk.ZeroInt()
	for i := range participants {
		participants[i].Power = sdk.NewInt(participants[i].GetConsensusPower()).QuoRaw(bondPerShare)
		totalPower = totalPower.Add(participants[i].Power)
	}

	snapshot := exported.Snapshot{
		Validators: participants,
		Timestamp:  ctx.BlockTime(),
		Height:     ctx.BlockHeight(),
		TotalPower: totalPower,
		Counter:    counter,
	}

	ctx.KVStore(k.storeKey).Set(counterKey(counter), k.cdc.MustMarshalBinaryLengthPrefixed(snapshot))

	return snapshotConsensusPower, totalConsensusPower, nil
}

func (k Keeper) setLatestCounter(ctx sdk.Context, counter int64) {
	ctx.KVStore(k.storeKey).Set([]byte(lastCounterKey), k.cdc.MustMarshalBinaryLengthPrefixed(counter))
}

func counterKey(counter int64) []byte {
	return []byte(fmt.Sprintf("%s%d", counterPrefix, counter))
}
