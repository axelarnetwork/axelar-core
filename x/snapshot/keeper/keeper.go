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

const lastRoundKey = "lastRound"
const roundPrefix = "round_"

// Make sure the keeper implements the Snapshotter interface
var _ exported.Snapshotter = Keeper{}

// Keeper represents the snapshot keeper
type Keeper struct {
	storeKey sdk.StoreKey
	staking  types.StakingKeeper
	slasher  types.Slasher
	cdc      *codec.Codec
	params   subspace.Subspace
}

// NewKeeper creates a new keeper for the staking module
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, paramSpace params.Subspace, staking types.StakingKeeper, slasher types.Slasher) Keeper {
	return Keeper{
		storeKey: key,
		cdc:      cdc,
		staking:  staking,
		params:   paramSpace.WithKeyTable(types.KeyTable()),
		slasher:  slasher,
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

// ComputeActiveValidators returns the subset of all validators that bonded and should be declared active
// and their aggregate staking power
func ComputeActiveValidators(ctx sdk.Context, validators []exported.Validator, slasher types.Slasher) (*[]exported.Validator, *sdk.Int, error) {

	var activeValidators []exported.Validator
	activeStake := sdk.NewInt(int64(0))

	for _, validator := range validators {

		addr := sdk.ConsAddress(validator.GetConsAddr().Bytes())
		signingInfo, found := slasher.GetValidatorSigningInfo(ctx, addr)
		if found == false {
			return nil, nil, fmt.Errorf("snapshot: couldn't retrieve signing info for a validator")
		}

		// check if for any reason the validator should be declared as inactive
		// e.g., the validator missed to vote on blocks
		// TODO: check what interval we're checking missedBlocksCounter for.
		if !(signingInfo.Tombstoned == true || signingInfo.MissedBlocksCounter > 0 ||
			signingInfo.JailedUntil.After(time.Unix(0, 0))) {
			activeValidators = append(activeValidators, validator)
			valstake := sdk.NewInt(validator.GetConsensusPower())
			activeStake = activeStake.Add(valstake)
		}
	}

	return activeValidators, activeStake, nil
}

// TakeSnapshot attempts to create a new snapshot
func (k Keeper) TakeSnapshot(ctx sdk.Context) error {
	s, ok := k.GetLatestSnapshot(ctx)

	if !ok {
		k.executeSnapshot(ctx, 0)
		k.setLatestRound(ctx, 0)
		return nil
	}

	lockingPeriod := k.getLockingPeriod(ctx)
	if s.Timestamp.Add(lockingPeriod).After(ctx.BlockTime()) {
		return fmt.Errorf("not enough time has passed since last snapshot, need to wait %s longer",
			s.Timestamp.Add(lockingPeriod).Sub(ctx.BlockTime()).String())
	}

	k.executeSnapshot(ctx, s.Round+1)
	k.setLatestRound(ctx, s.Round+1)
	return nil
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
func (k Keeper) GetSnapshot(ctx sdk.Context, round int64) (exported.Snapshot, bool) {
	bz := ctx.KVStore(k.storeKey).Get(roundKey(round))
	if bz == nil {

		return exported.Snapshot{}, false
	}

	var snapshot exported.Snapshot
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &snapshot)
	return snapshot, true
}

// GetLatestCounter returns the latest snapshot counter
func (k Keeper) GetLatestCounter(ctx sdk.Context) int64 {
	bz := ctx.KVStore(k.storeKey).Get([]byte(lastRoundKey))

	if bz == nil {

		return -1
	}

	var i int64
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &i)
	return i
}

func (k Keeper) executeSnapshot(ctx sdk.Context, nextRound int64) {
	var validators []exported.Validator
	fnAppend := func(_ int64, v sdkExported.ValidatorI) (stop bool) {
		validators = append(validators, v)
		return false
	}

	k.staking.IterateLastValidators(ctx, fnAppend)

	activeValidators, activeStake, err := ComputeActiveValidators(ctx, validators, k.slasher)
	if err != nil {
		return
	}

	snapshot := exported.Snapshot{
		Validators: *activeValidators,
		Timestamp:  ctx.BlockTime(),
		Height:     ctx.BlockHeight(),
		TotalPower: *activeStake, // k.staking.GetLastTotalPower(ctx),
		Round:      nextRound,
	}

	ctx.KVStore(k.storeKey).Set(roundKey(nextRound), k.cdc.MustMarshalBinaryLengthPrefixed(snapshot))
}

func (k Keeper) setLatestRound(ctx sdk.Context, round int64) {
	ctx.KVStore(k.storeKey).Set([]byte(lastRoundKey), k.cdc.MustMarshalBinaryLengthPrefixed(round))
}

func roundKey(round int64) []byte {
	return []byte(fmt.Sprintf("%s%d", roundPrefix, round))
}
