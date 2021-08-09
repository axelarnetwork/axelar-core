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
// to ask vald processes about sending their acknowledgments
func (k Keeper) ScheduleKeygen(ctx sdk.Context, req types.StartKeygenRequest) {
	height := k.GetParams(ctx).AckWindowInBlocks + ctx.BlockHeight()
	key := fmt.Sprintf("%s%d_%s_%s", ackPrefix, height, exported.AckKeygen.String(), req.NewKeyID)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(req)

	ctx.KVStore(k.storeKey).Set([]byte(key), bz)
	k.emitAckEvent(ctx, types.AttributeValueKeygen, req.NewKeyID, "")

	k.Logger(ctx).Info(fmt.Sprintf("keygen scheduled for block %d (currently at %d))", height, ctx.BlockHeight()))
}

// StartKeygen starts a keygen protocol with the specified parameters
func (k Keeper) StartKeygen(ctx sdk.Context, voter types.Voter, keyID string, snapshot snapshot.Snapshot) error {
	threshold, set := k.computeAndSetCorruptionThreshold(ctx, snapshot.TotalShareCount, keyID)
	if !set {
		return fmt.Errorf("key ID %s already has a corruption threshold defined", keyID)
	}
	if threshold < 1 || snapshot.TotalShareCount.Int64() <= threshold {
		return fmt.Errorf("invalid threshold: %d, total power: %d", threshold, snapshot.TotalShareCount.Int64())
	}

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

	pollKey := vote.NewPollKey(types.ModuleName, keyID)
	if err := voter.InitializePoll(ctx, pollKey, snapshot.Counter, vote.ExpiryAt(0)); err != nil {
		return err
	}
	return nil
}

// GetKey returns the key for a given ID, if it exists
func (k Keeper) GetKey(ctx sdk.Context, keyID string) (exported.Key, bool) {
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
func (k Keeper) SetKey(ctx sdk.Context, keyID string, key ecdsa.PublicKey) {
	btcecPK := btcec.PublicKey(key)
	ctx.KVStore(k.storeKey).Set([]byte(pkPrefix+keyID), btcecPK.SerializeCompressed())
}

// GetCurrentKeyID returns the current key ID for given chain and role
func (k Keeper) GetCurrentKeyID(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (string, bool) {
	return k.getKeyID(ctx, chain, k.GetRotationCount(ctx, chain, keyRole), keyRole)
}

// GetCurrentKey returns the current key for given chain and role
func (k Keeper) GetCurrentKey(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (exported.Key, bool) {
	return k.GetKeyByRotationCount(ctx, chain, keyRole, k.GetRotationCount(ctx, chain, keyRole))
}

// GetNextKeyID returns the next key ID for given chain and role
func (k Keeper) GetNextKeyID(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (string, bool) {
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

func (k Keeper) setKeyRole(ctx sdk.Context, keyID string, keyRole exported.KeyRole) {
	storageKey := fmt.Sprintf("%s%s", keyRolePrefix, keyID)

	ctx.KVStore(k.storeKey).Set([]byte(storageKey), k.cdc.MustMarshalBinaryLengthPrefixed(keyRole))
}

func (k Keeper) getKeyRole(ctx sdk.Context, keyID string) exported.KeyRole {
	storageKey := fmt.Sprintf("%s%s", keyRolePrefix, keyID)

	bz := ctx.KVStore(k.storeKey).Get([]byte(storageKey))
	if bz == nil {
		return exported.Unknown
	}

	var keyRole exported.KeyRole
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &keyRole)

	return keyRole
}

func (k Keeper) setRotatedAt(ctx sdk.Context, keyID string) {
	storageKey := fmt.Sprintf("%s%s", keyRotatedAtPrefix, keyID)
	ctx.KVStore(k.storeKey).Set([]byte(storageKey), k.cdc.MustMarshalBinaryLengthPrefixed(ctx.BlockTime().Unix()))
}

func (k Keeper) getRotatedAt(ctx sdk.Context, keyID string) *time.Time {
	storageKey := fmt.Sprintf("%s%s", keyRotatedAtPrefix, keyID)

	bz := ctx.KVStore(k.storeKey).Get([]byte(storageKey))
	if bz == nil {
		return nil
	}

	var seconds int64
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &seconds)

	timestamp := time.Unix(seconds, 0)
	return &timestamp
}

// AssignNextKey stores a new key for a given chain which will become the default once RotateKey is called
func (k Keeper) AssignNextKey(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole, keyID string) error {
	if _, ok := k.GetKey(ctx, keyID); !ok {
		return fmt.Errorf("key %s does not exist (yet)", keyID)
	}

	// The key entry needs to store the keyID instead of the public key, because the keyID is needed whenever
	// the keeper calls the secure private key store (e.g. for signing) and we would lose the keyID information otherwise
	k.setKeyID(ctx, chain, k.GetRotationCount(ctx, chain, keyRole)+1, keyRole, keyID)
	k.setKeyRole(ctx, keyID, keyRole)

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

// HasKeygenStart returns true if there is a start height for the given key ID
func (k Keeper) HasKeygenStart(ctx sdk.Context, keyID string) bool {
	return ctx.KVStore(k.storeKey).Get([]byte(keygenStartHeight+keyID)) != nil
}

// DeleteKeygenStart deletes the start height for the given key
func (k Keeper) DeleteKeygenStart(ctx sdk.Context, keyID string) {
	ctx.KVStore(k.storeKey).Delete([]byte(keygenStartHeight + keyID))
}

func (k Keeper) setKeygenStart(ctx sdk.Context, keyID string) {
	ctx.KVStore(k.storeKey).Set([]byte(keygenStartHeight+keyID), k.cdc.MustMarshalBinaryLengthPrefixed(ctx.BlockHeight()))
}

func (k Keeper) getKeygenStart(ctx sdk.Context, keyID string) (int64, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(keygenStartHeight + keyID))
	if bz == nil {
		return 0, false
	}

	var blockHeight int64
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &blockHeight)

	return blockHeight, true
}

func (k Keeper) getKeyID(ctx sdk.Context, chain nexus.Chain, rotation int64, keyRole exported.KeyRole) (string, bool) {
	storageKey := fmt.Sprintf("%s%d_%s_%s", rotationPrefix, rotation, chain.Name, keyRole.SimpleString())

	keyID := ctx.KVStore(k.storeKey).Get([]byte(storageKey))
	if keyID == nil {
		return "", false
	}

	return string(keyID), true
}

func (k Keeper) setKeyID(ctx sdk.Context, chain nexus.Chain, rotation int64, keyRole exported.KeyRole, keyID string) {
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
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &rotation)
	return rotation
}

func (k Keeper) setRotationCount(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole, rotation int64) {
	storageKey := fmt.Sprintf("%s%s_%s", rotationPrefix, chain.Name, keyRole.SimpleString())

	ctx.KVStore(k.storeKey).Set([]byte(storageKey), k.cdc.MustMarshalBinaryLengthPrefixed(rotation))
}

// DeleteSnapshotCounterForKeyID deletes the snapshot counter for the given key
func (k Keeper) DeleteSnapshotCounterForKeyID(ctx sdk.Context, keyID string) {
	ctx.KVStore(k.storeKey).Delete([]byte(snapshotForKeyIDPrefix + keyID))
}

func (k Keeper) setSnapshotCounterForKeyID(ctx sdk.Context, keyID string, counter int64) {
	ctx.KVStore(k.storeKey).Set([]byte(snapshotForKeyIDPrefix+keyID), k.cdc.MustMarshalBinaryBare(counter))
}

// GetSnapshotCounterForKeyID returns the snapshot round in which the key with the given ID was created, if the key exists
func (k Keeper) GetSnapshotCounterForKeyID(ctx sdk.Context, keyID string) (int64, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(snapshotForKeyIDPrefix + keyID))
	if bz == nil {
		return 0, false
	}
	var counter int64
	k.cdc.MustUnmarshalBinaryBare(bz, &counter)
	return counter, true
}

// DeleteParticipantsInKeygen deletes the participants in the given key genereation
func (k Keeper) DeleteParticipantsInKeygen(ctx sdk.Context, keyID string) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, []byte(participatePrefix+"key_"+keyID))
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

func (k Keeper) setParticipatesInKeygen(ctx sdk.Context, keyID string, validator sdk.ValAddress) {
	ctx.KVStore(k.storeKey).Set([]byte(participatePrefix+"key_"+keyID+validator.String()), []byte{})
}

// DoesValidatorParticipateInKeygen returns true if given validator participates in key gen for the given key ID; otherwise, false
func (k Keeper) DoesValidatorParticipateInKeygen(ctx sdk.Context, keyID string, validator sdk.ValAddress) bool {
	return ctx.KVStore(k.storeKey).Has([]byte(participatePrefix + "key_" + keyID + validator.String()))
}

// GetMinKeygenThreshold returns minimum threshold of stake that must be met to execute keygen
func (k Keeper) GetMinKeygenThreshold(ctx sdk.Context) utils.Threshold {
	var threshold utils.Threshold
	k.params.Get(ctx, types.KeyMinKeygenThreshold, &threshold)
	return threshold
}

// GetMinBondFractionPerShare returns the % of stake validators have to bond per key share
func (k Keeper) GetMinBondFractionPerShare(ctx sdk.Context) utils.Threshold {
	var threshold utils.Threshold
	k.params.Get(ctx, types.KeyMinBondFractionPerShare, &threshold)

	return threshold
}
