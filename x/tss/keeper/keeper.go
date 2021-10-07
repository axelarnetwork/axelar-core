package keeper

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
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

const (
	rotationPrefix             = "rotation_"
	rotationCountPrefix        = "rotation_count_"
	keygenStartHeight          = "block_height_"
	pkPrefix                   = "pk_"
	groupRecoverPrefix         = "group_recovery_info_"
	privateRecoverPrefix       = "private_recovery_info_"
	thresholdPrefix            = "threshold_"
	snapshotForKeyIDPrefix     = "sfkid_"
	sigPrefix                  = "sig_"
	infoForSigPrefix           = "info_for_sig_"
	participatePrefix          = "part_"
	keyRequirementPrefix       = "key_requirement_"
	keyRolePrefix              = "key_role_"
	keyTssSuspendedUntil       = "key_tss_suspended_until_"
	keyRotatedAtPrefix         = "key_rotated_at_"
	availablePrefix            = "available_"
	linkedSeqNumPrefix         = "linked_seq_number_"
	scheduledKeygenPrefix      = "scheduled_keygen_"
	scheduledSignPrefix        = "scheduled_sign_"
	sigStatusPrefix            = "sig_status_"
	rotationCountOfKeyIDPrefix = "rotation_count_of_key_id_"
	externalKeyIDsPrefix       = "external_key_ids_"
)

// Keeper allows access to the broadcast state
type Keeper struct {
	slasher  snapshot.Slasher
	params   params.Subspace
	storeKey sdk.StoreKey
	cdc      *codec.LegacyAmino
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

// GetExternalMultisigThreshold returns the external multisig threshold
func (k Keeper) GetExternalMultisigThreshold(ctx sdk.Context) utils.Threshold {
	var result utils.Threshold
	k.params.Get(ctx, types.KeyExternalMultisigThreshold, &result)

	return result
}

// SetGroupRecoveryInfo sets the group recovery info for a given party
func (k Keeper) SetGroupRecoveryInfo(ctx sdk.Context, keyID exported.KeyID, recoveryInfo []byte) {
	key := fmt.Sprintf("%s%s", groupRecoverPrefix, keyID)
	ctx.KVStore(k.storeKey).Set([]byte(key), recoveryInfo)
}

// GetGroupRecoveryInfo returns a party's group recovery info of a specific key ID
func (k Keeper) GetGroupRecoveryInfo(ctx sdk.Context, keyID exported.KeyID) []byte {
	key := fmt.Sprintf("%s%s", groupRecoverPrefix, keyID)
	return ctx.KVStore(k.storeKey).Get([]byte(key))
}

// SetPrivateRecoveryInfo sets the private recovery info for a given party
func (k Keeper) SetPrivateRecoveryInfo(ctx sdk.Context, sender sdk.ValAddress, keyID exported.KeyID, recoveryInfo []byte) {
	key := fmt.Sprintf("%s%s_%s", privateRecoverPrefix, keyID, sender.String())

	// marshal private recover info before storing
	bz := k.cdc.MustMarshalLengthPrefixed(recoveryInfo)

	ctx.KVStore(k.storeKey).Set([]byte(key), bz)
}

// GetPrivateRecoveryInfo returns a party's private recovery info of a specific key ID
func (k Keeper) GetPrivateRecoveryInfo(ctx sdk.Context, sender sdk.ValAddress, keyID exported.KeyID) []byte {
	key := fmt.Sprintf("%s%s_%s", privateRecoverPrefix, keyID, sender.String())
	bz := ctx.KVStore(k.storeKey).Get([]byte(key))

	// private recovery infos has been marshaled in keeper
	var privateRecoveryInfos []byte
	k.cdc.MustUnmarshalLengthPrefixed(bz, &privateRecoveryInfos)

	return privateRecoveryInfos
}

// HasPrivateRecoveryInfos returns true if the private recovery infos for a given party exists
func (k Keeper) HasPrivateRecoveryInfos(ctx sdk.Context, sender sdk.ValAddress, keyID exported.KeyID) bool {
	key := fmt.Sprintf("%s%s_%s", privateRecoverPrefix, keyID, sender.String())
	return ctx.KVStore(k.storeKey).Has([]byte(key))
}

// DeleteAllRecoveryInfos removes all recovery infos (private and group) associated to the given key ID
func (k Keeper) DeleteAllRecoveryInfos(ctx sdk.Context, keyID exported.KeyID) {
	prefix := fmt.Sprintf("%s%s_", privateRecoverPrefix, keyID)
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, []byte(prefix))
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}

	key := fmt.Sprintf("%s%s", groupRecoverPrefix, keyID)
	ctx.KVStore(k.storeKey).Delete([]byte(key))
}

func (k Keeper) setKeyRequirement(ctx sdk.Context, keyRequirement exported.KeyRequirement) {
	key := fmt.Sprintf("%s%s", keyRequirementPrefix, keyRequirement.KeyRole.SimpleString())
	bz := k.cdc.MustMarshalLengthPrefixed(keyRequirement)

	ctx.KVStore(k.storeKey).Set([]byte(key), bz)
}

// GetKeyRequirement gets the key requirement for a given chain of a given role
func (k Keeper) GetKeyRequirement(ctx sdk.Context, keyRole exported.KeyRole) (exported.KeyRequirement, bool) {
	key := fmt.Sprintf("%s%s", keyRequirementPrefix, keyRole.SimpleString())

	bz := ctx.KVStore(k.storeKey).Get([]byte(key))
	if bz == nil {
		return exported.KeyRequirement{}, false
	}

	var keyRequirement exported.KeyRequirement
	k.cdc.MustUnmarshalLengthPrefixed(bz, &keyRequirement)

	return keyRequirement, true
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

// SetAvailableOperator sets the height at which a validator sent his ack for some key/sign ID. Returns an error if
// the validator already submitted a ack for the given ID and type.
func (k Keeper) SetAvailableOperator(ctx sdk.Context, ID string, ackType exported.AckType, validator sdk.ValAddress) error {
	if ID == "" {
		return fmt.Errorf("ID cannot be empty")
	}

	key := fmt.Sprintf("%s%s_%s_%s", availablePrefix, ID, ackType.String(), validator.String())
	bz := ctx.KVStore(k.storeKey).Get([]byte(key))
	if bz != nil {
		return fmt.Errorf("validator already submitted its ack for the specified ID and type")
	}

	bz = make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, uint64(ctx.BlockHeight()))
	ctx.KVStore(k.storeKey).Set([]byte(key), bz)

	return nil
}

// IsOperatorAvailable returns true if the validator already submitted an acknowledgments for the given ID
func (k Keeper) IsOperatorAvailable(ctx sdk.Context, ID string, ackType exported.AckType, validator sdk.ValAddress) bool {
	key := fmt.Sprintf("%s%s_%s_%s", availablePrefix, ID, ackType.String(), validator.String())
	return ctx.KVStore(k.storeKey).Has([]byte(key))
}

// LinkAvailableOperatorsToSnapshot links the available operators of some keygen/sign to a snapshot counter
func (k Keeper) LinkAvailableOperatorsToSnapshot(ctx sdk.Context, sessionID string, ackType exported.AckType, snapshotSeqNo int64) {
	operators := k.GetAvailableOperators(ctx, sessionID, ackType, ctx.BlockHeight())
	if len(operators) > 0 {
		k.setAvailableOperatorsForCounter(ctx, snapshotSeqNo, operators)
	}
}

// GetAvailableOperators gets all operators that sent an acknowledgment for the given keygen/sign ID until some given height
func (k Keeper) GetAvailableOperators(ctx sdk.Context, sessionID string, ackType exported.AckType, heightLimit int64) []sdk.ValAddress {
	if sessionID == "" {
		return nil
	}

	prefix := fmt.Sprintf("%s%s_%s_", availablePrefix, sessionID, ackType.String())
	store := ctx.KVStore(k.storeKey)
	var addresses []sdk.ValAddress

	iter := sdk.KVStorePrefixIterator(store, []byte(prefix))
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {

		validator := strings.TrimPrefix(string(iter.Key()), prefix)
		address, err := sdk.ValAddressFromBech32(validator)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("excluding validator %s due to parsing error: %s", validator, err.Error()))
			continue
		}

		height := int64(binary.LittleEndian.Uint64(iter.Value()))
		if height > heightLimit {
			k.Logger(ctx).Debug(fmt.Sprintf("excluding validator %s due to late acknowledgement"+
				" [received at height %d and height limit %d]", validator, height, heightLimit))
			continue
		}

		addresses = append(addresses, address)
	}

	return addresses
}

// DeleteAvailableOperators removes the validator that sent an ack for some key/sign ID (if it exists)
func (k Keeper) DeleteAvailableOperators(ctx sdk.Context, sessionID string, ackType exported.AckType) {
	prefix := fmt.Sprintf("%s%s_%s_", availablePrefix, sessionID, ackType.String())
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, []byte(prefix))
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

// links a set of available operators to a snapshot counter
func (k Keeper) setAvailableOperatorsForCounter(ctx sdk.Context, counter int64, validators []sdk.ValAddress) {
	key := fmt.Sprintf("%s%d", linkedSeqNumPrefix, counter)

	values := make([]string, len(validators))
	for i, validator := range validators {
		values[i] = validator.String()
	}
	list, _ := json.Marshal(values)

	ctx.KVStore(k.storeKey).Set([]byte(key), list)
}

// OperatorIsAvailableForCounter returns true if the given validator address is available for the specified snapshot counter
func (k Keeper) OperatorIsAvailableForCounter(ctx sdk.Context, counter int64, validator sdk.ValAddress) bool {
	key := fmt.Sprintf("%s%d", linkedSeqNumPrefix, counter)
	bz := ctx.KVStore(k.storeKey).Get([]byte(key))

	if bz == nil {
		return false
	}

	var list []string
	_ = json.Unmarshal(bz, &list)

	for _, value := range list {
		if value == validator.String() {
			return true
		}
	}

	return false
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

// SetExternalKeyIDs stores the given list of external key IDs
func (k Keeper) SetExternalKeyIDs(ctx sdk.Context, chain nexus.Chain, keyIDs []exported.KeyID) {
	storageKey := []byte(fmt.Sprintf("%s%s", externalKeyIDsPrefix, chain.Name))

	ctx.KVStore(k.storeKey).Set(storageKey, k.cdc.MustMarshalLengthPrefixed(keyIDs))
}

// GetExternalKeyIDs retrieves the current list of external key IDs
func (k Keeper) GetExternalKeyIDs(ctx sdk.Context, chain nexus.Chain) ([]exported.KeyID, bool) {
	storageKey := []byte(fmt.Sprintf("%s%s", externalKeyIDsPrefix, chain.Name))

	bz := ctx.KVStore(k.storeKey).Get(storageKey)
	if bz == nil {
		return []exported.KeyID{}, false
	}

	var keyIDs []exported.KeyID
	k.cdc.MustUnmarshalLengthPrefixed(bz, &keyIDs)

	return keyIDs, true
}

func (k Keeper) emitAckEvent(ctx sdk.Context, action string, keyID exported.KeyID, sigID string, height int64) {
	event := sdk.NewEvent(types.EventTypeAck,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, action),
		sdk.NewAttribute(types.AttributeKeyKeyID, string(keyID)),
		sdk.NewAttribute(types.AttributeKeyHeight, strconv.FormatInt(height, 10)),
	)
	if action == types.AttributeValueSign {
		event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeySigID, sigID))
	}

	ctx.EventManager().EmitEvent(event)
}
