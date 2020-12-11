package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/snapshot/exported"

	sdkExported "github.com/cosmos/cosmos-sdk/x/staking/exported"

	"github.com/axelarnetwork/axelar-core/x/snapshot/types"

	"github.com/tendermint/tendermint/libs/log"
)

const lastRound = "lastRound"
const roundPrefix = "r-"

// for now, have a small interval between rounds
// const interval = 7 * 24 * time.Hour
const interval = 10 * time.Second

// Make sure the keeper implements the Staker interface
var _ exported.Snapshotter = Keeper{}

type Keeper struct {
	storeKey sdk.StoreKey
	staking  types.StakingKeeper
	cdc      *codec.Codec
}

// NewKeeper creates a new keeper for the staking module
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, staking types.StakingKeeper) Keeper {

	return Keeper{
		storeKey: key,
		cdc:      cdc,
		staking:  staking,
	}
}

func (k Keeper) Validator(ctx sdk.Context, address sdk.ValAddress) (exported.Validator, bool) {

	result := exported.Validator{}

	validator, ok := k.staking.GetValidator(ctx, address)

	if !ok {
		return result, false
	}

	result.Address = address
	result.Power = validator.GetConsensusPower()

	return result, true
}

// TODO: Investigate if you still need this function
func (k Keeper) IterateValidators(ctx sdk.Context, fn func(_ int64, _ exported.Validator) (stop bool)) {

	fnPrepared := func(i int64, v sdkExported.ValidatorI) (stop bool) {

		validator := exported.Validator{

			Address: v.GetOperator(),
			Power:   v.GetConsensusPower(),
		}

		return fn(i, validator)
	}

	k.staking.IterateValidators(ctx, fnPrepared)

}

// TakeSnapshot attempts to create a new snapshot
func (k Keeper) TakeSnapshot(ctx sdk.Context) error {

	r := k.GetLatestRound(ctx)

	if r == -1 {

		k.executeSnapshot(ctx, r+1)
		return nil
	}

	s, ok := k.GetSnapshot(ctx, r)

	if !ok {

		return fmt.Errorf("Unable to take snapshot: no snapshot for latest round %d", r)
	}

	ts := ctx.BlockTime()

	if s.Timestamp.Add(interval).After(ts) {

		return fmt.Errorf("Unable to take snapshot: %s", "Too soon to take a snapshot")

	}

	k.executeSnapshot(ctx, r+1)
	return nil

}

// GetLatestSnapshot retrieves the last created snapshot
func (k Keeper) GetLatestSnapshot(ctx sdk.Context) (exported.Snapshot, bool) {

	r := k.GetLatestRound(ctx)

	if r == -1 {

		return exported.Snapshot{}, false

	}

	return k.GetSnapshot(ctx, r)
}

// GetSnapshot retrieves a snapshot by round, if it exists
func (k Keeper) GetSnapshot(ctx sdk.Context, round int64) (exported.Snapshot, bool) {

	var snapshot exported.Snapshot

	bz := ctx.KVStore(k.storeKey).Get(roundKey(round))

	if bz == nil {

		return snapshot, false
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &snapshot)

	return snapshot, true
}

// Logger returns the logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) executeSnapshot(ctx sdk.Context, nextRound int64) {

	var validators []exported.Validator

	fnAppend := func(_ int64, v sdkExported.ValidatorI) (stop bool) {

		validator := exported.Validator{

			Address: v.GetOperator(),
			Power:   v.GetConsensusPower(),
		}

		validators = append(validators, validator)
		return false
	}

	k.staking.IterateValidators(ctx, fnAppend)

	snapshot := exported.Snapshot{

		Validators: validators,
		Timestamp:  ctx.BlockTime(),
		Height:     ctx.BlockHeight(),
		TotalPower: k.staking.GetLastTotalPower(ctx),
	}

	k.setLastRound(ctx, nextRound)
	ctx.KVStore(k.storeKey).Set(roundKey(nextRound), k.cdc.MustMarshalBinaryLengthPrefixed(snapshot))

}

func (k Keeper) setLastRound(ctx sdk.Context, round int64) {

	ctx.KVStore(k.storeKey).Set([]byte(lastRound), k.cdc.MustMarshalBinaryLengthPrefixed(round))
}

func (k Keeper) GetLatestRound(ctx sdk.Context) int64 {

	bz := ctx.KVStore(k.storeKey).Get([]byte(lastRound))

	if bz == nil {

		return -1
	}

	var i int64
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &i)
	return i
}

func roundKey(round int64) []byte {
	return []byte(fmt.Sprintf("%s%d", roundPrefix, round))
}
