package keeper

import (
	"crypto/ecdsa"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

// StartKeygen starts a keygen protocol with the specified parameters
func (k Keeper) StartKeygen(ctx sdk.Context, voter types.Voter, keyInfo types.KeyInfo, snapshot snapshot.Snapshot) error {
	if _, ok := k.GetKey(ctx, keyInfo.KeyID); ok {
		return fmt.Errorf("key ID %s is already used", keyInfo.KeyID)
	}

	if k.HasKeygenStarted(ctx, keyInfo.KeyID) {
		return fmt.Errorf("keyID %s is already in use", keyInfo.KeyID)
	}

	k.setKeygenStart(ctx, keyInfo.KeyID)
	// store snapshot round to be able to look up the correct validator set when signing with this key
	k.setSnapshotCounterForKeyID(ctx, keyInfo.KeyID, snapshot.Counter)

	// set key info that contains key role and key type
	k.setKeyInfo(ctx, keyInfo)

	keyRequirement, ok := k.GetKeyRequirement(ctx, keyInfo.KeyRole, keyInfo.KeyType)
	if !ok {
		return fmt.Errorf("key requirement for key role %s type %s not found", keyInfo.KeyRole.SimpleString(), keyInfo.KeyType.SimpleString())
	}

	switch keyInfo.KeyType {
	case exported.Multisig:
		// init multisig key info
		multisigKeyInfo := types.MultisigInfo{
			ID:        string(keyInfo.KeyID),
			Timeout:   ctx.BlockHeight() + keyRequirement.KeygenTimeout,
			TargetNum: snapshot.TotalShareCount.Int64(),
		}
		k.SetMultisigKeygenInfo(ctx, multisigKeyInfo)
		// enqueue ongoing multisig keygen
		q := k.GetMultisigKeygenQueue(ctx)
		if err := q.Enqueue(&gogoprototypes.StringValue{Value: string(keyInfo.KeyID)}); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid key type %s", keyInfo.KeyType.SimpleString())
	}

	return nil
}

func (k Keeper) getKeys(ctx sdk.Context) (keys []exported.Key) {
	iter := k.getStore(ctx).Iterator(keyPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var key exported.Key
		iter.UnmarshalValue(&key)

		keys = append(keys, key)
	}

	return keys
}

func (k Keeper) setKey(ctx sdk.Context, key exported.Key) {
	k.getStore(ctx).Set(keyPrefix.AppendStr(string(key.ID)), &key)
}

// SetKey stores the given public key under the given key ID
func (k Keeper) SetKey(ctx sdk.Context, key exported.Key) {
	keyInfo, ok := k.getKeyInfo(ctx, key.ID)
	if ok {
		key.Role = keyInfo.KeyRole
		key.Type = keyInfo.KeyType
	}

	if key.Role != exported.ExternalKey {
		snapshotCounter, ok := k.GetSnapshotCounterForKeyID(ctx, key.ID)
		if !ok {
			panic(fmt.Errorf("snapshot counter for key %s not found", key.ID))
		}

		key.SnapshotCounter = snapshotCounter
	}

	k.setKey(ctx, key)
}

// GetKey returns the key for a given ID, if it exists
func (k Keeper) GetKey(ctx sdk.Context, keyID exported.KeyID) (key exported.Key, ok bool) {
	return key, k.getStore(ctx).Get(keyPrefix.AppendStr(string(keyID)), &key)
}

// GetCurrentKeyID returns the current key ID for given chain and role
func (k Keeper) GetCurrentKeyID(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (exported.KeyID, bool) {
	return k.getKeyID(ctx, chain.Name, k.GetRotationCount(ctx, chain, keyRole), keyRole)
}

// GetCurrentKey returns the current key for given chain and role
func (k Keeper) GetCurrentKey(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (exported.Key, bool) {
	return k.GetKeyByRotationCount(ctx, chain, keyRole, k.GetRotationCount(ctx, chain, keyRole))
}

// GetNextKeyID returns the next key ID for given chain and role
func (k Keeper) GetNextKeyID(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (exported.KeyID, bool) {
	return k.getKeyID(ctx, chain.Name, k.GetRotationCount(ctx, chain, keyRole)+1, keyRole)
}

// GetNextKey returns the next key for given chain and role
func (k Keeper) GetNextKey(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (exported.Key, bool) {
	return k.GetKeyByRotationCount(ctx, chain, keyRole, k.GetRotationCount(ctx, chain, keyRole)+1)
}

// GetKeyByRotationCount returns the key for given chain and key role by rotation count
func (k Keeper) GetKeyByRotationCount(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole, rotationCount int64) (exported.Key, bool) {
	keyID, found := k.getKeyID(ctx, chain.Name, rotationCount, keyRole)
	if !found {
		return exported.Key{}, false
	}

	return k.GetKey(ctx, keyID)
}

// getKeyInfo returns the key info of the given keyID
func (k Keeper) getKeyInfo(ctx sdk.Context, keyID exported.KeyID) (keyInfo types.KeyInfo, ok bool) {
	return keyInfo, k.getStore(ctx).Get(keyInfoPrefix.AppendStr(string(keyID)), &keyInfo)
}

func (k Keeper) setKeyInfo(ctx sdk.Context, keyInfo types.KeyInfo) {
	k.getStore(ctx).Set(keyInfoPrefix.AppendStr(string(keyInfo.KeyID)), &keyInfo)
}

// GetKeyRole returns the role of the given key
func (k Keeper) GetKeyRole(ctx sdk.Context, keyID exported.KeyID) exported.KeyRole {
	if key, ok := k.GetKey(ctx, keyID); ok {
		return key.Role
	}

	return exported.Unknown
}

func (k Keeper) setRotatedAt(ctx sdk.Context, keyID exported.KeyID) {
	key, ok := k.GetKey(ctx, keyID)
	if !ok {
		panic(fmt.Errorf("key %s not found", keyID))
	}

	timestamp := ctx.BlockTime()
	key.RotatedAt = &timestamp

	k.setKey(ctx, key)
}

// AssignNextKey stores a new key for a given chain which will become the default once RotateKey is called
func (k Keeper) AssignNextKey(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole, keyID exported.KeyID) error {
	if _, ok := k.GetKey(ctx, keyID); !ok {
		return fmt.Errorf("key %s does not exist (yet)", keyID)
	}

	// The key entry needs to store the keyID instead of the public key, because the keyID is needed whenever
	// the keeper calls the secure private key store (e.g. for signing) and we would lose the keyID information otherwise
	rotationCount := k.GetRotationCount(ctx, chain, keyRole) + 1
	_, ok := k.getKeyID(ctx, chain.Name, rotationCount, keyRole)
	if ok {
		return fmt.Errorf("next %s key for chain %s already assigned", keyRole.SimpleString(), chain.Name)
	}

	k.setKeyID(ctx, chain.Name, rotationCount, keyRole, keyID)
	k.setRotationOfKey(ctx, keyID, rotationCount, chain)

	k.Logger(ctx).Info(fmt.Sprintf("assigning next key for chain %s for role %s (ID: %s)", chain.Name, keyRole.SimpleString(), keyID))

	return nil
}

// RotateKey rotates to the next stored key. Returns an error if no new key has been prepared
func (k Keeper) RotateKey(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) error {
	r := k.GetRotationCount(ctx, chain, keyRole)

	keyID, ok := k.getKeyID(ctx, chain.Name, r+1, keyRole)
	if !ok {
		return fmt.Errorf("next %s key for chain %s not set", keyRole.SimpleString(), chain.Name)
	}

	k.setRotationCount(ctx, chain.Name, keyRole, r+1)
	k.setRotatedAt(ctx, keyID)

	return nil
}

// HasKeygenStarted returns true if a key session for the given key ID exists; false otherwise
func (k Keeper) HasKeygenStarted(ctx sdk.Context, keyID exported.KeyID) bool {
	return k.getStore(ctx).Has(keygenStartPrefix.AppendStr(string(keyID)))
}

// DeleteKeygenStart deletes the start height for the given key
func (k Keeper) DeleteKeygenStart(ctx sdk.Context, keyID exported.KeyID) {
	k.getStore(ctx).Delete(keygenStartPrefix.AppendStr(string(keyID)))
}

func (k Keeper) setKeygenStart(ctx sdk.Context, keyID exported.KeyID) {
	k.getStore(ctx).SetRaw(keygenStartPrefix.AppendStr(string(keyID)), []byte{1})
}

func (k Keeper) getKeyID(ctx sdk.Context, chain nexus.ChainName, rotation int64, keyRole exported.KeyRole) (exported.KeyID, bool) {
	storageKey := rotationPrefix.
		Append(utils.LowerCaseKey(chain.String())).
		Append(utils.KeyFromStr(keyRole.SimpleString())).
		Append(utils.KeyFromStr(strconv.FormatInt(rotation, 10)))

	keyID := k.getStore(ctx).GetRaw(storageKey)
	if keyID == nil {
		return "", false
	}

	return exported.KeyID(keyID), true
}

func (k Keeper) setKeyID(ctx sdk.Context, chain nexus.ChainName, rotation int64, keyRole exported.KeyRole, keyID exported.KeyID) {
	storageKey := rotationPrefix.
		Append(utils.LowerCaseKey(chain.String())).
		Append(utils.KeyFromStr(keyRole.SimpleString())).
		Append(utils.KeyFromStr(strconv.FormatInt(rotation, 10)))

	k.getStore(ctx).SetRaw(storageKey, []byte(keyID))
}

// GetRotationCount returns the current rotation count for the given chain and key role
func (k Keeper) GetRotationCount(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) int64 {
	storageKey := rotationCountPrefix.Append(utils.LowerCaseKey(chain.Name.String())).Append(utils.KeyFromStr(keyRole.SimpleString()))

	var rotation gogoprototypes.Int64Value
	if ok := k.getStore(ctx).Get(storageKey, &rotation); !ok {
		return 0
	}

	return rotation.Value
}

func (k Keeper) setRotationCount(ctx sdk.Context, chain nexus.ChainName, keyRole exported.KeyRole, rotation int64) {
	storageKey := rotationCountPrefix.Append(utils.LowerCaseKey(chain.String())).Append(utils.KeyFromStr(keyRole.SimpleString()))
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

func (k Keeper) setRotationOfKey(ctx sdk.Context, keyID exported.KeyID, rotationCount int64, chain nexus.Chain) {
	key, ok := k.GetKey(ctx, keyID)
	if !ok {
		panic(fmt.Errorf("key %s not found", keyID))
	}

	key.RotationCount = rotationCount
	key.Chain = chain.Name.String()

	k.setKey(ctx, key)
}

// GetRotationCountOfKeyID returns the rotation count of the given key ID
func (k Keeper) GetRotationCountOfKeyID(ctx sdk.Context, keyID exported.KeyID) (int64, bool) {
	key, ok := k.GetKey(ctx, keyID)
	if !ok || key.RotationCount == 0 {
		return 0, false
	}

	return key.RotationCount, true
}

// GetKeyType returns the key type of the given keyID
func (k Keeper) GetKeyType(ctx sdk.Context, keyID exported.KeyID) exported.KeyType {
	if key, ok := k.GetKey(ctx, keyID); ok {
		return key.Type
	}

	return exported.KEY_TYPE_UNSPECIFIED
}

// SubmitPubKeys stores public keys a validator has under the given multisig key ID
func (k Keeper) SubmitPubKeys(ctx sdk.Context, keyID exported.KeyID, validator sdk.ValAddress, pubKeys ...[]byte) bool {
	var info types.MultisigInfo
	ok := k.getStore(ctx).Get(multiSigKeyPrefix.AppendStr(string(keyID)), &info)
	if !ok {
		// the setter is controlled by keeper
		panic(fmt.Sprintf("MultisigKeygenInfo %s not found", keyID))
	}

	for _, pk := range pubKeys {
		if info.HasData(pk) {
			return false
		}

	}
	info.AddData(validator, pubKeys)
	k.SetMultisigKeygenInfo(ctx, info)

	return true
}

func (k Keeper) getCompletedMultisigKeygenInfos(ctx sdk.Context) (infos []types.MultisigInfo) {
	iter := k.getStore(ctx).Iterator(multiSigKeyPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var info types.MultisigInfo
		iter.UnmarshalValue(&info)

		if !info.IsCompleted() {
			continue
		}

		infos = append(infos, info)
	}

	return infos
}

// GetMultisigKeygenInfo returns the MultisigKeygenInfo
func (k Keeper) GetMultisigKeygenInfo(ctx sdk.Context, keyID exported.KeyID) (types.MultisigKeygenInfo, bool) {
	var info types.MultisigInfo
	ok := k.getStore(ctx).Get(multiSigKeyPrefix.AppendStr(string(keyID)), &info)

	return &info, ok
}

// SetMultisigKeygenInfo store the MultisigKeygenInfo
func (k Keeper) SetMultisigKeygenInfo(ctx sdk.Context, info types.MultisigInfo) {
	k.getStore(ctx).Set(multiSigKeyPrefix.AppendStr(info.ID), &info)
}

// DeleteMultisigKeygen deletes the multisig keygen info for the given key ID
func (k Keeper) DeleteMultisigKeygen(ctx sdk.Context, keyID exported.KeyID) {
	k.getStore(ctx).Delete(multiSigKeyPrefix.AppendStr(string(keyID)))
}

// IsMultisigKeygenCompleted returns true if multisig keygen completed for the given key ID
func (k Keeper) IsMultisigKeygenCompleted(ctx sdk.Context, keyID exported.KeyID) bool {
	keygenInfo, ok := k.GetMultisigKeygenInfo(ctx, keyID)
	if !ok {
		return false
	}
	return keygenInfo.IsCompleted()
}

// GetMultisigPubKeysByValidator returns the pub keys a validator has for the given keyID
func (k Keeper) GetMultisigPubKeysByValidator(ctx sdk.Context, keyID exported.KeyID, val sdk.ValAddress) ([]ecdsa.PublicKey, bool) {
	info, ok := k.GetMultisigKeygenInfo(ctx, keyID)
	if !ok {
		return nil, false
	}

	pubKeys := info.GetPubKeysByValidator(val)
	return pubKeys, len(pubKeys) > 0
}

// GetMultisigKeygenQueue returns the multisig keygen timeout queue
func (k Keeper) GetMultisigKeygenQueue(ctx sdk.Context) utils.SequenceKVQueue {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(multisigKeygenQueue))
	return utils.NewSequenceKVQueue(utils.NewNormalizedStore(store, k.cdc), uint64(k.getMaxSignQueueSize(ctx)), k.Logger(ctx))
}
