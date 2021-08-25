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
	rotationPrefix              = "rotationCount_"
	keygenStartHeight           = "blockHeight_"
	pkPrefix                    = "pk_"
	recoveryPrefix              = "recovery_"
	thresholdPrefix             = "threshold_"
	snapshotForKeyIDPrefix      = "sfkid_"
	sigPrefix                   = "sig_"
	keyIDForSigPrefix           = "kidfs_"
	participatePrefix           = "part_"
	participateShareCountPrefix = "part_share_count_"
	keyRequirementPrefix        = "key_requirement_"
	keyRolePrefix               = "key_role_"
	keyTssSuspendedUntil        = "key_tss_suspended_until_"
	keyRotatedAtPrefix          = "key_rotated_at_"
	availablePrefix             = "available_"
	linkedSeqNumPrefix          = "linked_seq_number_"
	scheduledKeygenPrefix       = "scheduled_keygen_"
	scheduledSignPrefix         = "scheduled_sign_"
	sigStatusPrefix             = "sig_status_"
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

	for _, validator := range snap.Validators {
		illegibility, err := snapshotter.GetValidatorIllegibility(ctx, validator.GetSDKValidator())
		if err != nil {
			return err
		}

		if illegibility = illegibility.FilterIllegibilityForNewKey(); illegibility != snapshot.None {
			return fmt.Errorf("validator %s in snapshot %d is not eligible for handling key %s due to [%s]",
				validator.GetSDKValidator().GetOperator().String(),
				counter,
				keyID,
				illegibility.String(),
			)
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
		k.setKeyRequirement(ctx, keyRequirement)
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

func (k Keeper) setKeyRequirement(ctx sdk.Context, keyRequirement exported.KeyRequirement) {
	key := fmt.Sprintf("%s%s", keyRequirementPrefix, keyRequirement.KeyRole.SimpleString())
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(keyRequirement)

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
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &keyRequirement)

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
func (k Keeper) LinkAvailableOperatorsToSnapshot(ctx sdk.Context, ID string, ackType exported.AckType, snapshotSeqNo int64) {
	operators := k.GetAvailableOperators(ctx, ID, ackType, ctx.BlockHeight())
	if len(operators) > 0 {
		k.setAvailableOperatorsForCounter(ctx, snapshotSeqNo, operators)
	}
}

// GetAvailableOperators gets all operators that sent a acknowledgment for so given keygen/sign ID until some given height
func (k Keeper) GetAvailableOperators(ctx sdk.Context, ID string, ackType exported.AckType, heightLimit int64) []sdk.ValAddress {
	if ID == "" {
		return nil
	}

	prefix := fmt.Sprintf("%s%s_%s_", availablePrefix, ID, ackType.String())
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
func (k Keeper) DeleteAvailableOperators(ctx sdk.Context, ID string, ackType exported.AckType) {
	prefix := fmt.Sprintf("%s%s_%s_", availablePrefix, ID, ackType.String())
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

func (k Keeper) emitAckEvent(ctx sdk.Context, action, keyID, sigID string, height int64) {
	event := sdk.NewEvent(types.EventTypeAck,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, action),
		sdk.NewAttribute(types.AttributeKeyKeyID, keyID),
		sdk.NewAttribute(types.AttributeKeyHeight, strconv.FormatInt(height, 10)),
	)
	if action == types.AttributeValueSign {
		event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeySigID, sigID))
	}

	ctx.EventManager().EmitEvent(event)
}
