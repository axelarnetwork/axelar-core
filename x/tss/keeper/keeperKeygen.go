package keeper

import (
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

// ScheduleKeygen sets a keygen to start at block currentHeight + AckWindow and emits events
// to ask vald processes about sending their acknowledgments It returns the height at which it was scheduled
func (k Keeper) ScheduleKeygen(ctx sdk.Context, req types.StartKeygenRequest) (int64, error) {
	height := k.GetParams(ctx).AckWindowInBlocks + ctx.BlockHeight()
	key := fmt.Sprintf("%s%d_%s_%s", scheduledKeygenPrefix, height, exported.AckType_Keygen.String(), req.KeyID)
	if ctx.KVStore(k.storeKey).Has([]byte(key)) {
		return -1, fmt.Errorf("keygen for key ID '%s' already set", req.KeyID)
	}
	bz := k.cdc.MustMarshalLengthPrefixed(req)

	ctx.KVStore(k.storeKey).Set([]byte(key), bz)
	k.emitAckEvent(ctx, types.AttributeValueKeygen, req.KeyID, "", height)

	k.Logger(ctx).Info(fmt.Sprintf("keygen for key ID '%s' scheduled for block %d (currently at %d)", req.KeyID, height, ctx.BlockHeight()))
	return height, nil
}

// GetAllKeygenRequestsAtCurrentHeight returns all keygen requests scheduled for the current height
func (k Keeper) GetAllKeygenRequestsAtCurrentHeight(ctx sdk.Context) []types.StartKeygenRequest {
	prefix := fmt.Sprintf("%s%d_%s_", scheduledKeygenPrefix, ctx.BlockHeight(), exported.AckType_Keygen.String())
	store := ctx.KVStore(k.storeKey)
	var requests []types.StartKeygenRequest

	iter := sdk.KVStorePrefixIterator(store, []byte(prefix))
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {

		var request types.StartKeygenRequest
		k.cdc.MustUnmarshalLengthPrefixed(iter.Value(), &request)
		requests = append(requests, request)
	}

	return requests
}

// DeleteScheduledKeygen removes a keygen request for the current height
func (k Keeper) DeleteScheduledKeygen(ctx sdk.Context, keyID exported.KeyID) {
	key := fmt.Sprintf("%s%d_%s_%s", scheduledKeygenPrefix, ctx.BlockHeight(), exported.AckType_Keygen, keyID)
	ctx.KVStore(k.storeKey).Delete([]byte(key))
}

// StartKeygen starts a keygen protocol with the specified parameters
func (k Keeper) StartKeygen(ctx sdk.Context, voter types.Voter, keyID exported.KeyID, keyRole exported.KeyRole, snapshot snapshot.Snapshot) error {
	if _, found := k.getKeygenStart(ctx, keyID); found {
		return fmt.Errorf("keyID %s is already in use", keyID)
	}

	// set keygen participants
	for _, v := range snapshot.Validators {
		k.setParticipatesInKeygen(ctx, keyID, v.GetSDKValidator().GetOperator())
	}

	// store block height for this keygen to be able to confirm later if the produced key is allowed as a master key
	k.setKeygenStart(ctx, keyID)
	// store snapshot round to be able to look up the correct validator set when signing with this key
	k.setSnapshotCounterForKeyID(ctx, keyID, snapshot.Counter)
	// set key role
	k.SetKeyRole(ctx, keyID, keyRole)

	keyRequirement, ok := k.GetKeyRequirement(ctx, keyRole)
	if !ok {
		return fmt.Errorf("key requirement for key role %s not found", keyRole.SimpleString())
	}

	pollKey := vote.NewPollKey(types.ModuleName, string(keyID))
	if err := voter.InitializePoll(
		ctx,
		pollKey,
		snapshot.Counter,
		vote.ExpiryAt(0),
		vote.Threshold(keyRequirement.KeygenVotingThreshold),
	); err != nil {
		return err
	}

	return nil
}

// GetKey returns the key for a given ID, if it exists
func (k Keeper) GetKey(ctx sdk.Context, keyID exported.KeyID) (exported.Key, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(pkPrefix + keyID))
	if bz == nil {
		return exported.Key{}, false
	}

	btcecPK, err := btcec.ParsePubKey(bz, btcec.S256())
	// the setter is controlled by the keeper alone, so an error here should be a catastrophic failure
	if err != nil {
		panic(err)
	}

	pk := btcecPK.ToECDSA()
	role := k.getKeyRole(ctx, keyID)
	rotatedAt := k.getRotatedAt(ctx, keyID)

	return exported.Key{ID: keyID, Value: *pk, Role: role, RotatedAt: rotatedAt}, true
}

// SetKey stores the given public key under the given key ID
func (k Keeper) SetKey(ctx sdk.Context, keyID exported.KeyID, key ecdsa.PublicKey) {
	btcecPK := btcec.PublicKey(key)
	ctx.KVStore(k.storeKey).Set([]byte(pkPrefix+keyID), btcecPK.SerializeCompressed())
}

// GetCurrentKeyID returns the current key ID for given chain and role
func (k Keeper) GetCurrentKeyID(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (exported.KeyID, bool) {
	return k.getKeyID(ctx, chain, k.GetRotationCount(ctx, chain, keyRole), keyRole)
}

// GetCurrentKey returns the current key for given chain and role
func (k Keeper) GetCurrentKey(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (exported.Key, bool) {
	return k.GetKeyByRotationCount(ctx, chain, keyRole, k.GetRotationCount(ctx, chain, keyRole))
}

// GetNextKeyID returns the next key ID for given chain and role
func (k Keeper) GetNextKeyID(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (exported.KeyID, bool) {
	return k.getKeyID(ctx, chain, k.GetRotationCount(ctx, chain, keyRole)+1, keyRole)
}

// GetNextKey returns the next key for given chain and role
func (k Keeper) GetNextKey(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (exported.Key, bool) {
	return k.GetKeyByRotationCount(ctx, chain, keyRole, k.GetRotationCount(ctx, chain, keyRole)+1)
}

// GetKeyByRotationCount returns the key for given chain and key role by rotation count
func (k Keeper) GetKeyByRotationCount(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole, rotationCount int64) (exported.Key, bool) {
	keyID, found := k.getKeyID(ctx, chain, rotationCount, keyRole)
	if !found {
		return exported.Key{}, false
	}

	return k.GetKey(ctx, keyID)
}

// SetKeyRole stores the role of the given key
func (k Keeper) SetKeyRole(ctx sdk.Context, keyID exported.KeyID, keyRole exported.KeyRole) {
	storageKey := fmt.Sprintf("%s%s", keyRolePrefix, keyID)

	ctx.KVStore(k.storeKey).Set([]byte(storageKey), k.cdc.MustMarshalLengthPrefixed(keyRole))
}

func (k Keeper) getKeyRole(ctx sdk.Context, keyID exported.KeyID) exported.KeyRole {
	storageKey := fmt.Sprintf("%s%s", keyRolePrefix, keyID)

	bz := ctx.KVStore(k.storeKey).Get([]byte(storageKey))
	if bz == nil {
		return exported.Unknown
	}

	var keyRole exported.KeyRole
	k.cdc.MustUnmarshalLengthPrefixed(bz, &keyRole)

	return keyRole
}

func (k Keeper) setRotatedAt(ctx sdk.Context, keyID exported.KeyID) {
	storageKey := fmt.Sprintf("%s%s", keyRotatedAtPrefix, keyID)
	ctx.KVStore(k.storeKey).Set([]byte(storageKey), k.cdc.MustMarshalLengthPrefixed(ctx.BlockTime().Unix()))
}

func (k Keeper) getRotatedAt(ctx sdk.Context, keyID exported.KeyID) *time.Time {
	storageKey := fmt.Sprintf("%s%s", keyRotatedAtPrefix, keyID)

	bz := ctx.KVStore(k.storeKey).Get([]byte(storageKey))
	if bz == nil {
		return nil
	}

	var seconds int64
	k.cdc.MustUnmarshalLengthPrefixed(bz, &seconds)

	timestamp := time.Unix(seconds, 0)
	return &timestamp
}

// AssignNextKey stores a new key for a given chain which will become the default once RotateKey is called
func (k Keeper) AssignNextKey(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole, keyID exported.KeyID) error {
	if _, ok := k.GetKey(ctx, keyID); !ok {
		return fmt.Errorf("key %s does not exist (yet)", keyID)
	}

	// The key entry needs to store the keyID instead of the public key, because the keyID is needed whenever
	// the keeper calls the secure private key store (e.g. for signing) and we would lose the keyID information otherwise
	rotationCount := k.GetRotationCount(ctx, chain, keyRole) + 1
	k.setKeyID(ctx, chain, rotationCount, keyRole, keyID)
	k.setRotationCountOfKeyID(ctx, keyID, rotationCount)

	k.Logger(ctx).Info(fmt.Sprintf("assigning next key for chain %s for role %s (ID: %s)", chain.Name, keyRole.SimpleString(), keyID))

	return nil
}

// RotateKey rotates to the next stored key. Returns an error if no new key has been prepared
func (k Keeper) RotateKey(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) error {
	r := k.GetRotationCount(ctx, chain, keyRole)
	keyID, found := k.getKeyID(ctx, chain, r+1, keyRole)
	if !found {
		return fmt.Errorf("next %s key for chain %s not set", keyRole.SimpleString(), chain.Name)
	}

	k.setRotationCount(ctx, chain, keyRole, r+1)
	k.setRotatedAt(ctx, keyID)

	return nil
}

// HasKeygenStarted returns true if a keygen for the given key ID has been started
func (k Keeper) HasKeygenStarted(ctx sdk.Context, keyID exported.KeyID) bool {
	return ctx.KVStore(k.storeKey).Get([]byte(keygenStartHeight+keyID)) != nil
}

// DeleteKeygenStart deletes the start height for the given key
func (k Keeper) DeleteKeygenStart(ctx sdk.Context, keyID exported.KeyID) {
	ctx.KVStore(k.storeKey).Delete([]byte(keygenStartHeight + keyID))
}

func (k Keeper) setKeygenStart(ctx sdk.Context, keyID exported.KeyID) {
	ctx.KVStore(k.storeKey).Set([]byte(keygenStartHeight+keyID), k.cdc.MustMarshalLengthPrefixed(ctx.BlockHeight()))
}

func (k Keeper) getKeygenStart(ctx sdk.Context, keyID exported.KeyID) (int64, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(keygenStartHeight + keyID))
	if bz == nil {
		return 0, false
	}

	var blockHeight int64
	k.cdc.MustUnmarshalLengthPrefixed(bz, &blockHeight)

	return blockHeight, true
}

func (k Keeper) getKeyID(ctx sdk.Context, chain nexus.Chain, rotation int64, keyRole exported.KeyRole) (exported.KeyID, bool) {
	storageKey := fmt.Sprintf("%s%d_%s_%s", rotationPrefix, rotation, chain.Name, keyRole.SimpleString())

	keyID := ctx.KVStore(k.storeKey).Get([]byte(storageKey))
	if keyID == nil {
		return "", false
	}

	return exported.KeyID(keyID), true
}

func (k Keeper) setKeyID(ctx sdk.Context, chain nexus.Chain, rotation int64, keyRole exported.KeyRole, keyID exported.KeyID) {
	storageKey := fmt.Sprintf("%s%d_%s_%s", rotationPrefix, rotation, chain.Name, keyRole.SimpleString())

	ctx.KVStore(k.storeKey).Set([]byte(storageKey), []byte(keyID))
}

// GetRotationCount returns the current rotation count for the given chain and key role
func (k Keeper) GetRotationCount(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) int64 {
	storageKey := fmt.Sprintf("%s%s_%s", rotationPrefix, chain.Name, keyRole.SimpleString())

	bz := ctx.KVStore(k.storeKey).Get([]byte(storageKey))
	if bz == nil {
		return 0
	}
	var rotation int64
	k.cdc.MustUnmarshalLengthPrefixed(bz, &rotation)
	return rotation
}

func (k Keeper) setRotationCount(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole, rotation int64) {
	storageKey := fmt.Sprintf("%s%s_%s", rotationPrefix, chain.Name, keyRole.SimpleString())

	ctx.KVStore(k.storeKey).Set([]byte(storageKey), k.cdc.MustMarshalLengthPrefixed(rotation))
}

// DeleteSnapshotCounterForKeyID deletes the snapshot counter for the given key
func (k Keeper) DeleteSnapshotCounterForKeyID(ctx sdk.Context, keyID exported.KeyID) {
	ctx.KVStore(k.storeKey).Delete([]byte(snapshotForKeyIDPrefix + string(keyID)))
}

func (k Keeper) setSnapshotCounterForKeyID(ctx sdk.Context, keyID exported.KeyID, counter int64) {
	ctx.KVStore(k.storeKey).Set([]byte(snapshotForKeyIDPrefix+keyID), k.cdc.Amino.MustMarshalBinaryBare(counter))
}

// GetSnapshotCounterForKeyID returns the snapshot round in which the key with the given ID was created, if the key exists
func (k Keeper) GetSnapshotCounterForKeyID(ctx sdk.Context, keyID exported.KeyID) (int64, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(snapshotForKeyIDPrefix + keyID))
	if bz == nil {
		return 0, false
	}
	var counter int64
	k.cdc.Amino.MustUnmarshalBinaryBare(bz, &counter)
	return counter, true
}

// DeleteParticipantsInKeygen deletes the participants in the given key genereation
func (k Keeper) DeleteParticipantsInKeygen(ctx sdk.Context, keyID exported.KeyID) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, []byte(participatePrefix+"key_"+string(keyID)))
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

func (k Keeper) setParticipatesInKeygen(ctx sdk.Context, keyID exported.KeyID, validator sdk.ValAddress) {
	ctx.KVStore(k.storeKey).Set([]byte(participatePrefix+"key_"+string(keyID)+validator.String()), []byte{})
}

func (k Keeper) setRotationCountOfKeyID(ctx sdk.Context, keyID exported.KeyID, rotationCount int64) {
	ctx.KVStore(k.storeKey).Set([]byte(fmt.Sprintf("%s%s", rotationCountOfKeyIDPrefix, keyID)), k.cdc.MustMarshalLengthPrefixed(rotationCount))
}

// GetRotationCountOfKeyID returns the rotation count of the given key ID
func (k Keeper) GetRotationCountOfKeyID(ctx sdk.Context, keyID exported.KeyID) (int64, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(fmt.Sprintf("%s%s", rotationCountOfKeyIDPrefix, keyID)))
	if bz == nil {
		return 0, false
	}

	var rotationCount int64
	k.cdc.MustUnmarshalLengthPrefixed(bz, &rotationCount)

	return rotationCount, true
}

// DoesValidatorParticipateInKeygen returns true if given validator participates in key gen for the given key ID; otherwise, false
func (k Keeper) DoesValidatorParticipateInKeygen(ctx sdk.Context, keyID exported.KeyID, validator sdk.ValAddress) bool {
	return ctx.KVStore(k.storeKey).Has([]byte(participatePrefix + "key_" + string(keyID) + validator.String()))
}
