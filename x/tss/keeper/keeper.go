package keeper

import (
	"fmt"
	"math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

const (
	rotationPrefix              = "rotationCount_"
	keygenStartHeight           = "blockHeight_"
	pkPrefix                    = "pk_"
	snapshotForKeyIDPrefix      = "sfkid_"
	sigPrefix                   = "sig_"
	keyIDForSigPrefix           = "kidfs_"
	participatePrefix           = "part_"
	validatorDeregisteredPrefix = "validator_deregistered_block_height_"
)

type Keeper struct {
	broadcaster types.Broadcaster
	slasher     snapshot.Slasher
	params      params.Subspace
	storeKey    sdk.StoreKey
	cdc         *codec.Codec
}

// NewKeeper constructs a tss keeper
func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, paramSpace params.Subspace, broadcaster types.Broadcaster, slasher snapshot.Slasher) Keeper {
	return Keeper{
		broadcaster: broadcaster,
		slasher:     slasher,
		cdc:         cdc,
		params:      paramSpace.WithKeyTable(types.KeyTable()),
		storeKey:    storeKey,
	}
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetParams sets the tss module's parameters
func (k Keeper) SetParams(ctx sdk.Context, set types.Params) {
	k.params.SetParamSet(ctx, &set)
}

// GetParams gets the tss module's parameters
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.params.GetParamSet(ctx, &params)
	return
}

// getLockingPeriod returns the period of blocks that keygen is locked after a new snapshot has been created
func (k Keeper) getLockingPeriod(ctx sdk.Context) int64 {
	var period int64
	k.params.Get(ctx, types.KeyLockingPeriod, &period)
	return period
}

// SetValidatorDeregisteredBlockHeight sets the validator's deregistration block height
func (k Keeper) SetValidatorDeregisteredBlockHeight(ctx sdk.Context, valAddr sdk.ValAddress, blockHeight int64) {
	key := []byte(validatorDeregisteredPrefix + valAddr.String())
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(blockHeight)

	ctx.KVStore(k.storeKey).Set(key, bz)
}

// GetValidatorDeregisteredBlockHeight gets the validator's deregistration block height; 0 if the validator has never deregistered
func (k Keeper) GetValidatorDeregisteredBlockHeight(ctx sdk.Context, valAddr sdk.ValAddress) int64 {
	key := []byte(validatorDeregisteredPrefix + valAddr.String())
	bz := ctx.KVStore(k.storeKey).Get(key)

	if bz == nil {
		return 0
	}

	var blockHeight int64
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &blockHeight)

	return blockHeight
}

// ComputeCorruptionThreshold returns corruption threshold to be used by tss
func (k Keeper) ComputeCorruptionThreshold(ctx sdk.Context, totalvalidators int) int {
	var threshold utils.Threshold
	k.params.Get(ctx, types.KeyCorruptionThreshold, &threshold)
	// threshold = totalValidators * corruption threshold - 1
	return int(math.Ceil(float64(totalvalidators)*float64(threshold.Numerator)/
		float64(threshold.Denominator))) - 1
}
