package keeper

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

var (
	// Permanent
	keyRecoveryInfoPrefix  = utils.KeyFromStr("recovery_info")
	keyPrefix              = utils.KeyFromStr("key")
	rotationPrefix         = utils.KeyFromStr("rotation_to_key_id")
	rotationCountPrefix    = utils.KeyFromStr("rotation_count")
	snapshotForKeyIDPrefix = utils.KeyFromStr("snapshot_counter_for_key_id")
	multiSigKeyPrefix      = utils.KeyFromStr("multi_sig_keygen")
	externalKeysPrefix     = utils.KeyFromStr("external_key_ids")
	sigPrefix              = utils.KeyFromStr("sig")
	validatorStatusPrefix  = utils.KeyFromStr("validator_status")
	// temporary
	keyInfoPrefix      = utils.KeyFromStr("info_for_key")
	keygenStartPrefix  = utils.KeyFromStr("block_height")
	availablePrefix    = utils.KeyFromStr("available")
	presentKeysPrefix  = utils.KeyFromStr("present_keys")
	infoForSigPrefix   = utils.KeyFromStr("info_for_sig")
	participatePrefix  = utils.KeyFromStr("part")
	multisigSignPrefix = utils.KeyFromStr("multisig_sign")

	multisigKeygenQueue = "multisig_keygen"
	multisigSignQueue   = "multisig_sign"
)

// Keeper allows access to the broadcast state
type Keeper struct {
	slasher  types.Slasher
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

	if keyRole != k.GetKeyRole(ctx, keyID) {
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
func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, paramSpace params.Subspace, slasher types.Slasher, rewarder types.Rewarder) Keeper {
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

// GetKeyRequirement gets the key requirement for a given chain of a given role
func (k Keeper) GetKeyRequirement(ctx sdk.Context, keyRole exported.KeyRole, keyType exported.KeyType) (exported.KeyRequirement, bool) {
	for _, keyRequirement := range k.GetParams(ctx).KeyRequirements {
		if keyRequirement.KeyRole == keyRole && keyRequirement.KeyType == keyType {
			return keyRequirement, true
		}
	}

	return exported.KeyRequirement{}, false
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

// returns the signed blocks window to be considered when calculating the missed blocks percentage
func (k Keeper) getTssSignedBlocksWindow(ctx sdk.Context) int64 {
	var window int64
	k.params.Get(ctx, types.KeyTssSignedBlocksWindow, &window)

	return window
}

// HasMissedTooManyBlocks returns true if the given validator address missed too many blocks within
// the block window specified by this module. The block window used by this function is either the
// cosmos slashing module window or this own module's window, depending on which one is shorter.
func (k Keeper) HasMissedTooManyBlocks(ctx sdk.Context, address sdk.ConsAddress) (bool, error) {
	signInfo, ok := k.slasher.GetValidatorSigningInfo(ctx, address)
	if !ok {
		return false, fmt.Errorf("signing info not found for validator %s", address.String())
	}

	missedBlocks := k.getMissedBlocksPercent(ctx, address, signInfo)
	maxMissedPerWindow := k.GetMaxMissedBlocksPerWindow(ctx)

	return missedBlocks.GT(maxMissedPerWindow), nil
}

// returns the percentage of blocks signed w.r.t. this module's signed blocks window parameter
func (k Keeper) getMissedBlocksPercent(ctx sdk.Context, address sdk.ConsAddress, signInfo slashingtypes.ValidatorSigningInfo) utils.Threshold {
	counter := int64(0)
	tssWindow := k.getTssSignedBlocksWindow(ctx)
	slasherWindow := k.slasher.SignedBlocksWindow(ctx)
	indexOffset := int64(0)

	// TODO: In order to avoid having to pick the shorter of these two windows, we should implement
	// our own bit arrawy for missed blocks instead of re-purposing the one from cosmos
	window := slasherWindow
	if slasherWindow > tssWindow {
		window = tssWindow
	}
	if signInfo.IndexOffset > window {
		indexOffset = signInfo.IndexOffset - window
	}

	for ; indexOffset < signInfo.IndexOffset; indexOffset++ {
		if missed := k.slasher.GetValidatorMissedBlockBitArray(ctx, address, indexOffset%slasherWindow); missed {
			counter++
		}
	}

	return utils.NewThreshold(counter, tssWindow)
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
			k.Logger(ctx).Debug(fmt.Sprintf("excluding validator %s due to absent keys", validator))
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
		keyID, ok := k.getKeyID(ctx, chain.Name, rotationCount, keyRole)
		if !ok {
			return nil, fmt.Errorf("%s's %s key of rotation count %d not found", chain.Name, keyRole.SimpleString(), rotationCount)
		}

		activeKeyIDs = append(activeKeyIDs, keyID)
	}

	return activeKeyIDs, nil
}

func (k Keeper) getAllExternalKeys(ctx sdk.Context) (results []types.ExternalKeys) {
	iter := k.getStore(ctx).Iterator(externalKeysPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var externalKeys types.ExternalKeys
		iter.UnmarshalValue(&externalKeys)

		results = append(results, externalKeys)
	}

	return results
}

// SetExternalKeyIDs stores the given list of external key IDs
func (k Keeper) SetExternalKeyIDs(ctx sdk.Context, chain nexus.Chain, keyIDs []exported.KeyID) {
	k.setExternalKeys(ctx, types.ExternalKeys{Chain: chain.Name, KeyIDs: keyIDs})
}

func (k Keeper) setExternalKeys(ctx sdk.Context, externalKeys types.ExternalKeys) {
	k.getStore(ctx).Set(externalKeysPrefix.Append(utils.LowerCaseKey(externalKeys.Chain.String())), &externalKeys)
}

// GetExternalKeyIDs retrieves the current list of external key IDs
func (k Keeper) GetExternalKeyIDs(ctx sdk.Context, chain nexus.Chain) ([]exported.KeyID, bool) {
	return k.getExternalKeyIDs(ctx, chain.Name)
}

func (k Keeper) getExternalKeyIDs(ctx sdk.Context, chain nexus.ChainName) ([]exported.KeyID, bool) {
	var externalKeys types.ExternalKeys
	if !k.getStore(ctx).Get(externalKeysPrefix.Append(utils.LowerCaseKey(chain.String())), &externalKeys) {
		return []exported.KeyID{}, false
	}

	return externalKeys.KeyIDs, true
}

func (k Keeper) getStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}
