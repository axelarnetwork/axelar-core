package keeper

import (
	"fmt"
	"math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
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
	keyRequirementPrefix        = "key_requirement_"
	keyRolePrefix               = "key_role_"
)

// Keeper allows access to the broadcast state
type Keeper struct {
	slasher  snapshot.Slasher
	params   params.Subspace
	storeKey sdk.StoreKey
	cdc      *codec.LegacyAmino
}

// NewKeeper constructs a tss keeper
func NewKeeper(cdc *codec.LegacyAmino, storeKey sdk.StoreKey, paramSpace params.Subspace, slasher snapshot.Slasher) Keeper {
	return Keeper{
		slasher:  slasher,
		cdc:      cdc,
		params:   paramSpace.WithKeyTable(types.KeyTable()),
		storeKey: storeKey,
	}
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetParams sets the tss module's parameters
func (k Keeper) SetParams(ctx sdk.Context, p types.Params) {
	k.params.SetParamSet(ctx, &p)

	for _, keyRequirement := range p.KeyRequirements {
		// By copying this data to the KV store, we avoid having to iterate across all element
		// in the parameters table when a caller needs to fetch information from it
		k.setKeyRequirement(ctx, keyRequirement)
	}
}

// GetParams gets the tss module's parameters
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.params.GetParamSet(ctx, &params)
	return
}

func (k Keeper) setKeyRequirement(ctx sdk.Context, keyRequirement exported.KeyRequirement) {
	key := fmt.Sprintf("%s%s_%s", keyRequirementPrefix, keyRequirement.ChainName, keyRequirement.KeyRole.String())
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(keyRequirement)

	ctx.KVStore(k.storeKey).Set([]byte(key), bz)
}

// GetKeyRequirement gets the key requirement for a given chain of a given role
func (k Keeper) GetKeyRequirement(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (exported.KeyRequirement, bool) {
	key := fmt.Sprintf("%s%s_%s", keyRequirementPrefix, chain.Name, keyRole.String())
	bz := ctx.KVStore(k.storeKey).Get([]byte(key))

	if bz == nil {
		return exported.KeyRequirement{}, false
	}

	var keyRequirement exported.KeyRequirement
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &keyRequirement)

	return keyRequirement, true
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
