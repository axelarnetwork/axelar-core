package keeper

import (
	"fmt"
	"time"

	"github.com/axelarnetwork/axelar-core/x/staking/exported"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	sdkExported "github.com/cosmos/cosmos-sdk/x/staking/exported"

	"github.com/tendermint/tendermint/libs/log"
)

const lastRound = "lastRound"
const roundPrefix = "r-"

//for now, have a small interval between rounds
//const interval = 7 * 24 * time.Hour
const interval = 10 * time.Second

type Keeper struct {
	storeKey sdk.StoreKey
	staking  staking.Keeper
	cdc      *codec.Codec
}

// NewKeeper creates a new keeper for the staking module
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, staking staking.Keeper) (Keeper, error) {

	return Keeper{
		storeKey: key,
		cdc:      cdc,
		staking:  staking,
	}, nil

}

// TakeSnapshot attempts to create a new snapshot
func (k Keeper) TakeSnapshot(ctx sdk.Context) error {

	r := k.getLastRound(ctx)

	if r == -1 {

		k.executeSnapshot(ctx, r+1)
		return nil
	}

	s, err := k.GetSnapshot(ctx, r)

	if err != nil {

		return fmt.Errorf("Unable to take snapshot: %s", err)
	}

	ts := ctx.BlockTime()

	if s.Timestamp.Add(interval).After(ts) {

		return fmt.Errorf("Unable to take snapshot: %s", "Too soon to take a snapshot")

	}

	k.executeSnapshot(ctx, r+1)
	return nil

}

// GetLastSnapshot retrieves the last created snapshot
func (k Keeper) GetLastSnapshot(ctx sdk.Context) (exported.Snapshot, error) {

	r := k.getLastRound(ctx)

	if r == -1 {

		return exported.Snapshot{}, fmt.Errorf("No snapshots available")

	}

	return k.GetSnapshot(ctx, r)
}

// GetSnapshot retrieves a snapshot by round, if it exists
func (k Keeper) GetSnapshot(ctx sdk.Context, round int64) (exported.Snapshot, error) {

	var snapshot exported.Snapshot

	bz := ctx.KVStore(k.storeKey).Get(roundKey(round))

	if bz == nil {

		return snapshot, fmt.Errorf("Round not found: %d", round)
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &snapshot)

	return snapshot, nil
}

// Logger returns the logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return prepareLogger(ctx.Logger())
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
	}

	ctx.KVStore(k.storeKey).Set([]byte(roundKey(nextRound)), k.cdc.MustMarshalBinaryLengthPrefixed(snapshot))

}

func (k Keeper) setLastRound(ctx sdk.Context, round int64) {

	ctx.KVStore(k.storeKey).Set([]byte(lastRound), k.cdc.MustMarshalBinaryLengthPrefixed(round))
}

func (k Keeper) getLastRound(ctx sdk.Context) int64 {

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

func prepareLogger(logger log.Logger) log.Logger {
	return logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
