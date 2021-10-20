package keeper

import (
	"crypto/ecdsa"
	"encoding/binary"
	"fmt"
	"strconv"
	"time"

	gogoprototypes "github.com/gogo/protobuf/types"

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
	key := scheduledKeygenPrefix.AppendStr(strconv.FormatInt(ctx.BlockHeight(), 10)).
		AppendStr(exported.AckType_Keygen.String()).AppendStr(string(req.KeyID))
	if k.getStore(ctx).Has(key) {
		return -1, fmt.Errorf("keygen for key ID '%s' already set", req.KeyID)
	}

	k.getStore(ctx).Set(key, &req)
	k.Logger(ctx).Info(fmt.Sprintf("keygen for key ID '%s' scheduled for block %d (currently at %d)", req.KeyID, ctx.BlockHeight(), ctx.BlockHeight()))
	return ctx.BlockHeight(), nil
}

// GetAllKeygenRequestsAtCurrentHeight returns all keygen requests scheduled for the current height
func (k Keeper) GetAllKeygenRequestsAtCurrentHeight(ctx sdk.Context) []types.StartKeygenRequest {
	prefix := scheduledKeygenPrefix.AppendStr(strconv.FormatInt(ctx.BlockHeight(), 10)).AppendStr(exported.AckType_Keygen.String())
	var requests []types.StartKeygenRequest

	iter := k.getStore(ctx).Iterator(prefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var request types.StartKeygenRequest
		iter.UnmarshalValue(&request)
		requests = append(requests, request)
	}

	return requests
}

// DeleteScheduledKeygen removes a keygen request for the current height
func (k Keeper) DeleteScheduledKeygen(ctx sdk.Context, keyID exported.KeyID) {
	key := scheduledKeygenPrefix.AppendStr(strconv.FormatInt(ctx.BlockHeight(), 10)).
		AppendStr(exported.AckType_Keygen.String()).AppendStr(string(keyID))
	k.getStore(ctx).Delete(key)
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
	if err := voter.InitializePollWithSnapshot(
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
	bz := k.getStore(ctx).GetRaw(pkPrefix.AppendStr(string(keyID)))
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
	k.getStore(ctx).SetRaw(pkPrefix.AppendStr(string(keyID)), btcecPK.SerializeCompressed())
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
	bz := make([]byte, 4)
	binary.LittleEndian.PutUint32(bz, uint32(keyRole))

	k.getStore(ctx).SetRaw(keyRolePrefix.AppendStr(string(keyID)), bz)
}

func (k Keeper) getKeyRole(ctx sdk.Context, keyID exported.KeyID) exported.KeyRole {
	bz := k.getStore(ctx).GetRaw(keyRolePrefix.AppendStr(string(keyID)))
	if bz == nil {
		return exported.Unknown
	}

	return exported.KeyRole(binary.LittleEndian.Uint32(bz))
}

func (k Keeper) setRotatedAt(ctx sdk.Context, keyID exported.KeyID) {
	storageKey := keyRotatedAtPrefix.AppendStr(string(keyID))
	k.getStore(ctx).Set(storageKey, &gogoprototypes.Int64Value{Value: ctx.BlockTime().Unix()})
}

func (k Keeper) getRotatedAt(ctx sdk.Context, keyID exported.KeyID) *time.Time {
	storageKey := keyRotatedAtPrefix.AppendStr(string(keyID))

	var seconds gogoprototypes.Int64Value
	if ok := k.getStore(ctx).Get(storageKey, &seconds); !ok {
		return nil
	}

	timestamp := time.Unix(seconds.Value, 0)
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
	return k.getStore(ctx).GetRaw(keygenStartHeight.AppendStr(string(keyID))) != nil
}

// DeleteKeygenStart deletes the start height for the given key
func (k Keeper) DeleteKeygenStart(ctx sdk.Context, keyID exported.KeyID) {
	k.getStore(ctx).Delete(keygenStartHeight.AppendStr(string(keyID)))
}

func (k Keeper) setKeygenStart(ctx sdk.Context, keyID exported.KeyID) {
	k.getStore(ctx).Set(keygenStartHeight.AppendStr(string(keyID)), &gogoprototypes.Int64Value{Value: ctx.BlockHeight()})
}

func (k Keeper) getKeygenStart(ctx sdk.Context, keyID exported.KeyID) (int64, bool) {
	var blockHeight gogoprototypes.Int64Value
	if ok := k.getStore(ctx).Get(keygenStartHeight.AppendStr(string(keyID)), &blockHeight); !ok {
		return 0, false
	}

	return blockHeight.Value, true
}

func (k Keeper) getKeyID(ctx sdk.Context, chain nexus.Chain, rotation int64, keyRole exported.KeyRole) (exported.KeyID, bool) {
	storageKey := rotationPrefix.Append(utils.LowerCaseKey(chain.Name)).
		Append(utils.KeyFromStr(keyRole.SimpleString())).Append(utils.KeyFromStr(strconv.FormatInt(rotation, 10)))

	keyID := k.getStore(ctx).GetRaw(storageKey)
	if keyID == nil {
		return "", false
	}

	return exported.KeyID(keyID), true
}

func (k Keeper) setKeyID(ctx sdk.Context, chain nexus.Chain, rotation int64, keyRole exported.KeyRole, keyID exported.KeyID) {
	storageKey := rotationPrefix.Append(utils.LowerCaseKey(chain.Name)).
		Append(utils.KeyFromStr(keyRole.SimpleString())).Append(utils.KeyFromStr(strconv.FormatInt(rotation, 10)))

	k.getStore(ctx).SetRaw(storageKey, []byte(keyID))
}

// GetRotationCount returns the current rotation count for the given chain and key role
func (k Keeper) GetRotationCount(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) int64 {
	storageKey := rotationCountPrefix.Append(utils.LowerCaseKey(chain.Name)).Append(utils.KeyFromStr(keyRole.SimpleString()))

	var rotation gogoprototypes.Int64Value
	if ok := k.getStore(ctx).Get(storageKey, &rotation); !ok {
		return 0
	}

	return rotation.Value
}

func (k Keeper) setRotationCount(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole, rotation int64) {
	storageKey := rotationCountPrefix.Append(utils.LowerCaseKey(chain.Name)).Append(utils.KeyFromStr(keyRole.SimpleString()))
	k.getStore(ctx).Set(storageKey, &gogoprototypes.Int64Value{Value: rotation})
}

// DeleteSnapshotCounterForKeyID deletes the snapshot counter for the given key
func (k Keeper) DeleteSnapshotCounterForKeyID(ctx sdk.Context, keyID exported.KeyID) {
	k.getStore(ctx).Delete(snapshotForKeyIDPrefix.AppendStr(string(keyID)))
}

func (k Keeper) setSnapshotCounterForKeyID(ctx sdk.Context, keyID exported.KeyID, counter int64) {
	k.getStore(ctx).Set(snapshotForKeyIDPrefix.AppendStr(string(keyID)), &gogoprototypes.Int64Value{Value: counter})
}

// GetSnapshotCounterForKeyID returns the snapshot round in which the key with the given ID was created, if the key exists
func (k Keeper) GetSnapshotCounterForKeyID(ctx sdk.Context, keyID exported.KeyID) (int64, bool) {
	var counter gogoprototypes.Int64Value
	if ok := k.getStore(ctx).Get(snapshotForKeyIDPrefix.AppendStr(string(keyID)), &counter); !ok {
		return 0, false
	}

	return counter.Value, true
}

// DeleteParticipantsInKeygen deletes the participants in the given key genereation
func (k Keeper) DeleteParticipantsInKeygen(ctx sdk.Context, keyID exported.KeyID) {
	store := k.getStore(ctx)

	iter := store.Iterator(participatePrefix.AppendStr("key").AppendStr(string(keyID)))
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.GetKey())
	}
}

func (k Keeper) setParticipatesInKeygen(ctx sdk.Context, keyID exported.KeyID, validator sdk.ValAddress) {
	k.getStore(ctx).SetRaw(participatePrefix.AppendStr("key").AppendStr(string(keyID)).AppendStr(validator.String()), []byte{})
}

func (k Keeper) setRotationCountOfKeyID(ctx sdk.Context, keyID exported.KeyID, rotationCount int64) {
	k.getStore(ctx).Set(rotationCountOfKeyIDPrefix.AppendStr(string(keyID)), &gogoprototypes.Int64Value{Value: rotationCount})
}

// GetRotationCountOfKeyID returns the rotation count of the given key ID
func (k Keeper) GetRotationCountOfKeyID(ctx sdk.Context, keyID exported.KeyID) (int64, bool) {
	var rotationCount gogoprototypes.Int64Value
	if ok := k.getStore(ctx).Get(rotationCountOfKeyIDPrefix.AppendStr(string(keyID)), &rotationCount); !ok {
		return 0, false
	}

	return rotationCount.Value, true
}

// DoesValidatorParticipateInKeygen returns true if given validator participates in key gen for the given key ID; otherwise, false
func (k Keeper) DoesValidatorParticipateInKeygen(ctx sdk.Context, keyID exported.KeyID, validator sdk.ValAddress) bool {
	return k.getStore(ctx).Has(participatePrefix.AppendStr("key").AppendStr(string(keyID)).AppendStr(validator.String()))
}
