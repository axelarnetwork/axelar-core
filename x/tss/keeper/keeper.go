package keeper

import (
	"encoding/binary"
	"fmt"

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
	rotationPrefix         = "rotationCount_"
	keygenStartHeight      = "blockHeight_"
	pkPrefix               = "pk_"
	recoveryPrefix         = "recovery_"
	snapshotForKeyIDPrefix = "sfkid_"
	sigPrefix              = "sig_"
	keyIDForSigPrefix      = "kidfs_"
	participatePrefix      = "part_"
	keyRequirementPrefix   = "key_requirement_"
	keyRolePrefix          = "key_role_"
	keyTssSuspendedUntil   = "key_tss_suspended_until_"
	keyRotatedAtPrefix     = "key_rotated_at_"
)

// Keeper allows access to the broadcast state
type Keeper struct {
	slasher  snapshot.Slasher
	params   params.Subspace
	storeKey sdk.StoreKey
	cdc      *codec.LegacyAmino
}

// AssertMatchesRequirements checks if the properties of the given key match the requirements for the given role
func (k Keeper) AssertMatchesRequirements(ctx sdk.Context, snapshotter snapshot.Snapshotter, chain nexus.Chain, keyID string, keyRole exported.KeyRole) error {
	counter, ok := k.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return fmt.Errorf("could not find snapshot counter for given key ID %s", keyID)
	}

	snap, ok := snapshotter.GetSnapshot(ctx, counter)
	if !ok {
		return fmt.Errorf("could not find snapshot for given key ID %s", keyID)
	}

	currentKeyID, ok := k.GetCurrentKeyID(ctx, chain, keyRole)
	if ok {
		currentCounter, ok := k.GetSnapshotCounterForKeyID(ctx, currentKeyID)
		if !ok {
			return fmt.Errorf("no snapshot associated with the current %s key on chain %s", keyRole.SimpleString(), chain.Name)
		}
		if currentCounter >= counter {
			return fmt.Errorf("choose a key that is newer than the current one for role %s on chain %s", keyRole.SimpleString(), chain.Name)
		}
	}

	keyRequirement, found := k.GetKeyRequirement(ctx, chain, keyRole)
	if !found {
		return fmt.Errorf("%s key is not required for chain %s", keyRole.SimpleString(), chain.Name)
	}

	if len(snap.Validators) < int(keyRequirement.MinValidatorSubsetSize) {
		return fmt.Errorf(
			"expected %s's %s key to be generated with at least %d validators, actual %d",
			chain.Name,
			keyRole.SimpleString(),
			keyRequirement.MinValidatorSubsetSize,
			len(snap.Validators),
		)
	}

	if snap.KeyShareDistributionPolicy != keyRequirement.KeyShareDistributionPolicy {
		return fmt.Errorf(
			"expected %s's %s key to have tss shares distributed with policy %s, actual %s",
			chain.Name,
			keyRole.SimpleString(),
			keyRequirement.KeyShareDistributionPolicy.SimpleString(),
			snap.KeyShareDistributionPolicy.SimpleString(),
		)
	}

	for _, validator := range snap.Validators {
		if !snapshot.IsValidatorEligibleForNewKey(ctx, k.slasher, snapshotter, k, validator.GetSDKValidator()) {
			return fmt.Errorf("validator %s in snapshot %d is not eligible for handling key %s", validator.GetSDKValidator().GetOperator().String(), counter, keyID)
		}
	}

	return nil
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
		k.SetKeyRequirement(ctx, keyRequirement)
	}
}

// GetParams gets the tss module's parameters
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.params.GetParamSet(ctx, &params)
	return
}

// SetRecoveryInfos sets the recovery infos for a given party
func (k Keeper) SetRecoveryInfos(ctx sdk.Context, sender sdk.ValAddress, keyID string, infos [][]byte) {
	key := fmt.Sprintf("%s%s_%s", recoveryPrefix, keyID, sender.String())

	data := types.QueryRecoveryResponse{
		ShareRecoveryInfos: infos,
	}

	bz := k.cdc.MustMarshalBinaryLengthPrefixed(data)

	ctx.KVStore(k.storeKey).Set([]byte(key), bz)
}

// HasRecoveryInfos returns true if the recovery infos for a given party exists
func (k Keeper) HasRecoveryInfos(ctx sdk.Context, sender sdk.ValAddress, keyID string) bool {
	key := fmt.Sprintf("%s%s_%s", recoveryPrefix, keyID, sender.String())
	bz := ctx.KVStore(k.storeKey).Get([]byte(key))
	if bz == nil {
		return false
	}

	return true
}

// GetAllRecoveryInfos returns the recovery infos for all parties of a specific key ID
func (k Keeper) GetAllRecoveryInfos(ctx sdk.Context, keyID string) [][]byte {
	prefix := fmt.Sprintf("%s%s_", recoveryPrefix, keyID)
	store := ctx.KVStore(k.storeKey)
	var infos [][]byte

	iter := sdk.KVStorePrefixIterator(store, []byte(prefix))
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {

		var data types.QueryRecoveryResponse
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iter.Value(), &data)
		infos = append(infos, data.ShareRecoveryInfos...)
	}

	return infos
}

// DeleteAllRecoveryInfos removes all recovery infos associated to the given key ID
func (k Keeper) DeleteAllRecoveryInfos(ctx sdk.Context, keyID string) {
	prefix := fmt.Sprintf("%s%s_", recoveryPrefix, keyID)
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, []byte(prefix))
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

// SetKeyRequirement sets the key requirement for a given chain of a given role
func (k Keeper) SetKeyRequirement(ctx sdk.Context, keyRequirement exported.KeyRequirement) {
	key := fmt.Sprintf("%s%s_%s", keyRequirementPrefix, keyRequirement.ChainName, keyRequirement.KeyRole.SimpleString())
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(keyRequirement)

	ctx.KVStore(k.storeKey).Set([]byte(key), bz)
}

// GetKeyRequirement gets the key requirement for a given chain of a given role
func (k Keeper) GetKeyRequirement(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (exported.KeyRequirement, bool) {
	key := fmt.Sprintf("%s%s_%s", keyRequirementPrefix, chain.Name, keyRole.SimpleString())
	bz := ctx.KVStore(k.storeKey).Get([]byte(key))

	if bz == nil {
		return exported.KeyRequirement{}, false
	}

	var keyRequirement exported.KeyRequirement
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &keyRequirement)

	return keyRequirement, true
}

// ComputeCorruptionThreshold returns corruption threshold to be used by tss
func (k Keeper) ComputeCorruptionThreshold(ctx sdk.Context, totalShareCount sdk.Int) int64 {
	var threshold utils.Threshold
	k.params.Get(ctx, types.KeyCorruptionThreshold, &threshold)

	// (threshold + 1) shares are required to signed
	return totalShareCount.MulRaw(threshold.Numerator).QuoRaw(threshold.Denominator).Int64() - 1
}

func (k Keeper) setTssSuspendedUntil(ctx sdk.Context, validator sdk.ValAddress, suspendedUntilBlockNumber int64) {
	key := fmt.Sprintf("%s%s", keyTssSuspendedUntil, validator.String())
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, uint64(suspendedUntilBlockNumber))

	ctx.KVStore(k.storeKey).Set([]byte(key), bz)
}

// GetTssSuspendedUntil returns the block number at which a validator is released from TSS suspension
func (k Keeper) GetTssSuspendedUntil(ctx sdk.Context, validator sdk.ValAddress) int64 {
	key := fmt.Sprintf("%s%s", keyTssSuspendedUntil, validator.String())
	bz := ctx.KVStore(k.storeKey).Get([]byte(key))
	if bz == nil {
		return 0
	}

	return int64(binary.LittleEndian.Uint64(bz))
}
