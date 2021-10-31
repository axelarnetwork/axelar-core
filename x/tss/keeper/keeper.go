package keeper

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"

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

var (
	rotationPrefix             = utils.KeyFromStr("rotation")
	rotationCountPrefix        = utils.KeyFromStr("rotation_count")
	keygenStartPrefix          = utils.KeyFromStr("block_height")
	pkPrefix                   = utils.KeyFromStr("pk")
	groupRecoverPrefix         = utils.KeyFromStr("group_recovery_info")
	privateRecoverPrefix       = utils.KeyFromStr("private_recovery_info")
	snapshotForKeyIDPrefix     = utils.KeyFromStr("sfkid")
	sigPrefix                  = utils.KeyFromStr("sig")
	infoForSigPrefix           = utils.KeyFromStr("info_for_sig")
	participatePrefix          = utils.KeyFromStr("part")
	keyRequirementPrefix       = utils.KeyFromStr("key_requirement")
	keyRolePrefix              = utils.KeyFromStr("key_role")
	keyTssSuspendedUntil       = utils.KeyFromStr("key_tss_suspended_until")
	keyRotatedAtPrefix         = utils.KeyFromStr("key_rotated_at")
	availablePrefix            = utils.KeyFromStr("available")
	presentKeysPrefix          = utils.KeyFromStr("present_keys")
	sigStatusPrefix            = utils.KeyFromStr("sig_status")
	rotationCountOfKeyIDPrefix = utils.KeyFromStr("rotation_count_of_key_id")
	externalKeyIDsPrefix       = utils.KeyFromStr("external_key_ids")
	multiSigPrefix             = utils.KeyFromStr("multi_sig")
	keyTypePrefix              = utils.KeyFromStr("key_type")
	keyInfoPrefix              = utils.KeyFromStr("key_info")
)

// Keeper allows access to the broadcast state
type Keeper struct {
	slasher  snapshot.Slasher
	rewarder types.Rewarder
	params   params.Subspace
	storeKey sdk.StoreKey
	cdc      codec.BinaryCodec
	router   types.Router
}

// AssertMatchesRequirements checks if the properties of the given key match the requirements for the given role
func (k Keeper) AssertMatchesRequirements(ctx sdk.Context, snapshotter snapshot.Snapshotter, chain nexus.Chain, keyID exported.KeyID, keyRole exported.KeyRole) error {
	counter, ok := k.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return fmt.Errorf("could not find snapshot counter for given key ID %s", keyID)
	}

	snap, ok := snapshotter.GetSnapshot(ctx, counter)
	if !ok {
		return fmt.Errorf("could not find snapshot for given key ID %s", keyID)
	}

	if keyRole != k.getKeyRole(ctx, keyID) {
		return fmt.Errorf("key %s is not a %s key", keyID, keyRole.SimpleString())
	}

	if chain.KeyType != k.GetKeyType(ctx, keyID) {
		return fmt.Errorf("chain %s does not accept key type %s", chain.Name, k.GetKeyType(ctx, keyID))
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

	ellegibleShareCount := sdk.ZeroInt()
	for _, validator := range snap.Validators {
		illegibility, err := snapshotter.GetValidatorIllegibility(ctx, validator.GetSDKValidator())
		if err != nil {
			return err
		}

		if illegibility = illegibility.FilterIllegibilityForNewKey(); !illegibility.Is(snapshot.None) {
			k.Logger(ctx).Debug(fmt.Sprintf("validator %s in snapshot %d is not eligible for handling key %s due to [%s]",
				validator.GetSDKValidator().GetOperator().String(),
				counter,
				keyID,
				illegibility.String(),
			))

			continue
		}

		ellegibleShareCount = ellegibleShareCount.AddRaw(validator.ShareCount)
	}

	// ensure that ellegibleShareCount >= snap.CorruptionThreshold + 1
	if ellegibleShareCount.LTE(sdk.NewInt(snap.CorruptionThreshold)) {
		return fmt.Errorf("key %s cannot be assigned due to ellegible share count %d being less than or equal to corruption threshold %d",
			keyID,
			ellegibleShareCount.Int64(),
			snap.CorruptionThreshold,
		)
	}

	return nil
}

// NewKeeper constructs a tss keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, paramSpace params.Subspace, slasher snapshot.Slasher, rewarder types.Rewarder) Keeper {
	return Keeper{
		slasher:  slasher,
		rewarder: rewarder,
		cdc:      cdc,
		params:   paramSpace.WithKeyTable(types.KeyTable()),
		storeKey: storeKey,
	}
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetRouter sets the tss router. It will panic if called more than once
func (k *Keeper) SetRouter(router types.Router) {
	if k.router != nil {
		panic("router already set")
	}

	k.router = router

	// In order to avoid invalid or non-deterministic behavior, we seal the router immediately
	// to prevent additionals handlers from being registered after the keeper is initialized.
	k.router.Seal()
}

// GetRouter returns the tss router. If no router was set, it returns a (sealed) router with no handlers
func (k Keeper) GetRouter() types.Router {
	if k.router == nil {
		k.SetRouter(types.NewRouter())
	}

	return k.router
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

// GetExternalMultisigThreshold returns the external multisig threshold
func (k Keeper) GetExternalMultisigThreshold(ctx sdk.Context) utils.Threshold {
	var result utils.Threshold
	k.params.Get(ctx, types.KeyExternalMultisigThreshold, &result)

	return result
}

// GetHeartbeatPeriodInBlocks returns the heartbeat event period
func (k Keeper) GetHeartbeatPeriodInBlocks(ctx sdk.Context) int64 {
	var result int64
	k.params.Get(ctx, types.KeyHeartbeatPeriodInBlocks, &result)

	return result
}

// SetGroupRecoveryInfo sets the group recovery info for a given party
func (k Keeper) SetGroupRecoveryInfo(ctx sdk.Context, keyID exported.KeyID, recoveryInfo []byte) {
	k.getStore(ctx).SetRaw(groupRecoverPrefix.AppendStr(string(keyID)), recoveryInfo)
}

// GetGroupRecoveryInfo returns a party's group recovery info of a specific key ID
func (k Keeper) GetGroupRecoveryInfo(ctx sdk.Context, keyID exported.KeyID) []byte {
	return k.getStore(ctx).GetRaw(groupRecoverPrefix.AppendStr(string(keyID)))
}

// SetPrivateRecoveryInfo sets the private recovery info for a given party
func (k Keeper) SetPrivateRecoveryInfo(ctx sdk.Context, sender sdk.ValAddress, keyID exported.KeyID, recoveryInfo []byte) {
	k.getStore(ctx).SetRaw(privateRecoverPrefix.AppendStr(string(keyID)).AppendStr(sender.String()), recoveryInfo)
}

// GetPrivateRecoveryInfo returns a party's private recovery info of a specific key ID
func (k Keeper) GetPrivateRecoveryInfo(ctx sdk.Context, sender sdk.ValAddress, keyID exported.KeyID) []byte {
	return k.getStore(ctx).GetRaw(privateRecoverPrefix.AppendStr(string(keyID)).AppendStr(sender.String()))
}

// HasPrivateRecoveryInfos returns true if the private recovery infos for a given party exists
func (k Keeper) HasPrivateRecoveryInfos(ctx sdk.Context, sender sdk.ValAddress, keyID exported.KeyID) bool {
	return k.getStore(ctx).Has(privateRecoverPrefix.AppendStr(string(keyID)).AppendStr(sender.String()))
}

// DeleteAllRecoveryInfos removes all recovery infos (private and group) associated to the given key ID
func (k Keeper) DeleteAllRecoveryInfos(ctx sdk.Context, keyID exported.KeyID) {
	store := k.getStore(ctx)
	prefix := privateRecoverPrefix.AppendStr(string(keyID))
	iter := store.Iterator(prefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.GetKey())
	}

	k.getStore(ctx).Delete(prefix)
}

func (k Keeper) setKeyRequirement(ctx sdk.Context, keyRequirement exported.KeyRequirement) {
	key := keyRequirementPrefix.AppendStr(keyRequirement.KeyRole.SimpleString()).AppendStr(keyRequirement.KeyType.SimpleString())
	k.getStore(ctx).Set(key, &keyRequirement)
}

// GetKeyRequirement gets the key requirement for a given chain of a given role
func (k Keeper) GetKeyRequirement(ctx sdk.Context, keyRole exported.KeyRole, keyType exported.KeyType) (exported.KeyRequirement, bool) {
	var keyRequirement exported.KeyRequirement
	ok := k.getStore(ctx).Get(keyRequirementPrefix.AppendStr(keyRole.SimpleString()).AppendStr(keyType.SimpleString()), &keyRequirement)

	return keyRequirement, ok
}

// GetMaxMissedBlocksPerWindow returns the maximum percent of blocks a validator is allowed
// to miss per signing window
func (k Keeper) GetMaxMissedBlocksPerWindow(ctx sdk.Context) utils.Threshold {
	var threshold utils.Threshold
	k.params.Get(ctx, types.KeyMaxMissedBlocksPerWindow, &threshold)

	return threshold
}

// GetKeyUnbondingLockingKeyRotationCount returns the number of key iterations share
// holds must stay before they can unbond
func (k Keeper) GetKeyUnbondingLockingKeyRotationCount(ctx sdk.Context) int64 {
	var count int64
	k.params.Get(ctx, types.KeyUnbondingLockingKeyRotationCount, &count)

	return count
}

// getMaxSignQueueSize returns the maximum size of sign queue
func (k Keeper) getMaxSignQueueSize(ctx sdk.Context) int64 {
	var size int64
	k.params.Get(ctx, types.KeyMaxSignQueueSize, &size)

	return size
}

// GetMaxSimultaneousSignShares returns the max simultaneous number of sign shares
func (k Keeper) GetMaxSimultaneousSignShares(ctx sdk.Context) int64 {
	var shares int64
	k.params.Get(ctx, types.MaxSimultaneousSignShares, &shares)

	return shares
}

func (k Keeper) setTssSuspendedUntil(ctx sdk.Context, validator sdk.ValAddress, suspendedUntilBlockNumber int64) {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, uint64(suspendedUntilBlockNumber))

	k.getStore(ctx).SetRaw(keyTssSuspendedUntil.AppendStr(validator.String()), bz)
}

// GetTssSuspendedUntil returns the block number at which a validator is released from TSS suspension
func (k Keeper) GetTssSuspendedUntil(ctx sdk.Context, validator sdk.ValAddress) int64 {
	bz := k.getStore(ctx).GetRaw(keyTssSuspendedUntil.AppendStr(validator.String()))
	if bz == nil {
		return 0
	}

	return int64(binary.LittleEndian.Uint64(bz))
}

// SetAvailableOperator signals that a validator sent an ack
func (k Keeper) SetAvailableOperator(ctx sdk.Context, validator sdk.ValAddress, presentKeys ...exported.KeyID) {
	store := k.getStore(ctx)

	// update block height of last seen ack
	key := availablePrefix.AppendStr(validator.String())
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, uint64(ctx.BlockHeight()))
	store.SetRaw(key, bz)

	// garbage collection
	iter := store.Iterator(presentKeysPrefix.AppendStr(validator.String()))
	defer utils.CloseLogError(iter, k.Logger(ctx))
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.GetKey())
	}

	// update keys that operator holds
	for _, keyID := range presentKeys {
		key = presentKeysPrefix.AppendStr(validator.String()).AppendStr(string(keyID))
		store.SetRaw(key, []byte{1})
	}
}

func (k Keeper) operatorHasKeys(ctx sdk.Context, validator sdk.ValAddress, keyIDs ...exported.KeyID) bool {
	for _, keyID := range keyIDs {
		key := presentKeysPrefix.AppendStr(validator.String()).AppendStr(string(keyID))
		bz := k.getStore(ctx).GetRaw(key)
		if bz == nil {
			return false
		}
	}
	return true
}

// IsOperatorAvailable returns true if the validator has a non-stale ack and holds the specified keys
func (k Keeper) IsOperatorAvailable(ctx sdk.Context, validator sdk.ValAddress, keyIDs ...exported.KeyID) bool {
	key := availablePrefix.AppendStr(validator.String())

	bz := k.getStore(ctx).GetRaw(key)
	if bz == nil {
		return false
	}
	height := int64(binary.LittleEndian.Uint64(bz))

	return (ctx.BlockHeight()-height) <= k.GetHeartbeatPeriodInBlocks(ctx) && k.operatorHasKeys(ctx, validator, keyIDs...)
}

// GetAvailableOperators gets all operators that still have a non-stale heartbeat
func (k Keeper) GetAvailableOperators(ctx sdk.Context, keyIDs ...exported.KeyID) []sdk.ValAddress {
	iter := k.getStore(ctx).Iterator(availablePrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	var addresses []sdk.ValAddress
	for ; iter.Valid(); iter.Next() {
		validator := strings.TrimPrefix(string(iter.Key()), string(availablePrefix.AsKey())+"_")
		address, err := sdk.ValAddressFromBech32(validator)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("excluding validator %s due to parsing error: %s", validator, err.Error()))
			continue
		}

		height := int64(binary.LittleEndian.Uint64(iter.Value()))
		if (ctx.BlockHeight() - height) > k.GetHeartbeatPeriodInBlocks(ctx) {
			k.Logger(ctx).Debug(fmt.Sprintf("excluding validator %s due to stale heartbeat "+
				"[current height %d, event height %d]", validator, ctx.BlockHeight(), height))
			continue
		}

		if !k.operatorHasKeys(ctx, address, keyIDs...) {
			k.Logger(ctx).Debug(fmt.Sprintf("excluding validator %s due absent keys", validator))
			continue
		}

		addresses = append(addresses, address)
	}

	return addresses
}

// GetOldActiveKeys gets all the old keys of given key role that are still active for chain
func (k Keeper) GetOldActiveKeys(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) ([]exported.Key, error) {
	var activeKeys []exported.Key

	currRotationCount := k.GetRotationCount(ctx, chain, keyRole)
	unbondingLockingKeyRotationCount := k.GetKeyUnbondingLockingKeyRotationCount(ctx)

	rotationCount := currRotationCount - unbondingLockingKeyRotationCount
	if rotationCount <= 0 {
		rotationCount = 1
	}

	for ; rotationCount < currRotationCount; rotationCount++ {
		key, ok := k.GetKeyByRotationCount(ctx, chain, keyRole, rotationCount)
		if !ok {
			return nil, fmt.Errorf("%s's %s key of rotation count %d not found", chain.Name, keyRole.SimpleString(), rotationCount)
		}

		activeKeys = append(activeKeys, key)
	}

	return activeKeys, nil
}

// GetOldActiveKeyIDs gets all the old key IDs of given key role that are still active for chain
func (k Keeper) GetOldActiveKeyIDs(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) ([]exported.KeyID, error) {
	var activeKeyIDs []exported.KeyID

	currRotationCount := k.GetRotationCount(ctx, chain, keyRole)
	unbondingLockingKeyRotationCount := k.GetKeyUnbondingLockingKeyRotationCount(ctx)

	rotationCount := currRotationCount - unbondingLockingKeyRotationCount
	if rotationCount <= 0 {
		rotationCount = 1
	}

	for ; rotationCount < currRotationCount; rotationCount++ {
		keyID, ok := k.getKeyID(ctx, chain, rotationCount, keyRole)
		if !ok {
			return nil, fmt.Errorf("%s's %s key of rotation count %d not found", chain.Name, keyRole.SimpleString(), rotationCount)
		}

		activeKeyIDs = append(activeKeyIDs, keyID)
	}

	return activeKeyIDs, nil
}

// SetExternalKeyIDs stores the given list of external key IDs
func (k Keeper) SetExternalKeyIDs(ctx sdk.Context, chain nexus.Chain, keyIDs []exported.KeyID) {
	storageKey := externalKeyIDsPrefix.Append(utils.LowerCaseKey(chain.Name))
	list, _ := json.Marshal(keyIDs)
	k.getStore(ctx).SetRaw(storageKey, list)
}

// GetExternalKeyIDs retrieves the current list of external key IDs
func (k Keeper) GetExternalKeyIDs(ctx sdk.Context, chain nexus.Chain) ([]exported.KeyID, bool) {
	storageKey := externalKeyIDsPrefix.Append(utils.LowerCaseKey(chain.Name))

	bz := k.getStore(ctx).GetRaw(storageKey)
	if bz == nil {
		return []exported.KeyID{}, false
	}

	var keyIDs []exported.KeyID
	_ = json.Unmarshal(bz, &keyIDs)

	return keyIDs, true
}

func (k Keeper) getStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}
