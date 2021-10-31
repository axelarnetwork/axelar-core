package keeper

import (
	"crypto/ecdsa"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// StartKeygen starts a keygen protocol with the specified parameters
func (k Keeper) StartKeygen(ctx sdk.Context, voter types.Voter, keyInfo types.KeyInfo, snapshot snapshot.Snapshot) error {
	if k.hasKeygenStarted(ctx, keyInfo.KeyID) {
		return fmt.Errorf("keyID %s is already in use", keyInfo.KeyID)
	}

	// set keygen participants
	for _, v := range snapshot.Validators {
		k.setParticipatesInKeygen(ctx, keyInfo.KeyID, v.GetSDKValidator().GetOperator())
	}

	k.setKeygenStart(ctx, keyInfo.KeyID)
	// store snapshot round to be able to look up the correct validator set when signing with this key
	k.setSnapshotCounterForKeyID(ctx, keyInfo.KeyID, snapshot.Counter)

	// set key info that contains key role and key type
	k.SetKeyInfo(ctx, keyInfo)

	keyRequirement, ok := k.GetKeyRequirement(ctx, keyInfo.KeyRole, keyInfo.KeyType)
	if !ok {
		return fmt.Errorf("key requirement for key role %s type %s not found", keyInfo.KeyRole.SimpleString(), keyInfo.KeyType.SimpleString())
	}

	switch keyInfo.KeyType {
	case exported.Threshold:
		pollKey := vote.NewPollKey(types.ModuleName, string(keyInfo.KeyID))
		if err := voter.InitializePollWithSnapshot(
			ctx,
			pollKey,
			snapshot.Counter,
			vote.ExpiryAt(0),
			vote.Threshold(keyRequirement.KeygenVotingThreshold),
		); err != nil {
			return err
		}
	case exported.Multisig:
		// init multisig key info
		multisigKeyInfo := types.MultisigKeyInfo{
			KeyID:        keyInfo.KeyID,
			Timeout:      ctx.BlockHeight() + keyRequirement.KeygenTimeout,
			TargetKeyNum: snapshot.TotalShareCount.Int64(),
		}
		k.SetMultisigKeyInfo(ctx, multisigKeyInfo)
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

// SetKeyInfo stores the role and type of the given key
func (k Keeper) SetKeyInfo(ctx sdk.Context, keyInfo types.KeyInfo) {
	k.getStore(ctx).Set(keyInfoPrefix.AppendStr(string(keyInfo.KeyID)), &keyInfo)
}

func (k Keeper) getKeyRole(ctx sdk.Context, keyID exported.KeyID) exported.KeyRole {
	var keyInfo types.KeyInfo
	if ok := k.getStore(ctx).Get(keyInfoPrefix.AppendStr(string(keyID)), &keyInfo); !ok {
		return exported.Unknown
	}

	return keyInfo.KeyRole
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
	switch chain.KeyType {
	case exported.Threshold:
		if _, ok := k.GetKey(ctx, keyID); !ok {
			return fmt.Errorf("key %s does not exist (yet)", keyID)
		}
	case exported.Multisig:
		if !k.IsMultisigKeygenCompleted(ctx, keyID) {
			return fmt.Errorf("key %s does not exist (yet)", keyID)
		}
	default:
		panic(fmt.Sprintf("unrecognized key type %s", chain.KeyType.SimpleString()))
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

func (k Keeper) hasKeygenStarted(ctx sdk.Context, keyID exported.KeyID) bool {
	return k.getStore(ctx).Has(keygenStartPrefix.AppendStr(string(keyID)))
}

// DeleteKeygenStart deletes the start height for the given key
func (k Keeper) DeleteKeygenStart(ctx sdk.Context, keyID exported.KeyID) {
	k.getStore(ctx).Delete(keygenStartPrefix.AppendStr(string(keyID)))
}

func (k Keeper) setKeygenStart(ctx sdk.Context, keyID exported.KeyID) {
	k.getStore(ctx).SetRaw(keygenStartPrefix.AppendStr(string(keyID)), []byte{1})
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

// GetParticipantsInKeygen gets the keygen participants in the given keyID
func (k Keeper) GetParticipantsInKeygen(ctx sdk.Context, keyID exported.KeyID) []sdk.ValAddress {
	store := k.getStore(ctx)
	key := participatePrefix.AppendStr("key").AppendStr(string(keyID))

	iter := store.Iterator(key)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	var participants []sdk.ValAddress
	for ; iter.Valid(); iter.Next() {
		validator := strings.TrimPrefix(string(iter.Key()), string(key.AsKey())+"_")
		address, err := sdk.ValAddressFromBech32(validator)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("ignore participant %s due to parsing error: %s", validator, err.Error()))
			continue
		}

		participants = append(participants, address)
	}

	return participants
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

// GetKeyType returns the key type of the given keyID
func (k Keeper) GetKeyType(ctx sdk.Context, keyID exported.KeyID) exported.KeyType {
	var keyInfo types.KeyInfo
	if ok := k.getStore(ctx).Get(keyInfoPrefix.AppendStr(string(keyID)), &keyInfo); !ok {
		return exported.KEY_TYPE_UNSPECIFIED
	}

	return keyInfo.KeyType
}

// SubmitPubKeys stores public keys a validator has under the given multisig key ID
func (k Keeper) SubmitPubKeys(ctx sdk.Context, keyID exported.KeyID, validator sdk.ValAddress, pubKeys ...[]byte) bool {
	keyInfo, ok := k.getMultisigKeyInfo(ctx, keyID)
	if !ok {
		// the setter is controlled by keeper
		panic(fmt.Sprintf("MultisigKeyInfo %s not found", keyID))
	}

	for _, pk := range pubKeys {
		if keyInfo.HasKey(pk) {
			return false
		}
		keyInfo.AddKey(pk)
	}
	keyInfo.AddParticipant(validator)
	k.SetMultisigKeyInfo(ctx, keyInfo)

	return true
}

func (k Keeper) getMultisigKeyInfo(ctx sdk.Context, keyID exported.KeyID) (types.MultisigKeyInfo, bool) {
	var info types.MultisigKeyInfo
	ok := k.getStore(ctx).Get(multiSigPrefix.AppendStr(string(keyID)), &info)

	return info, ok
}

// SetMultisigKeyInfo store the MultisigKeyInfo
func (k Keeper) SetMultisigKeyInfo(ctx sdk.Context, info types.MultisigKeyInfo) {
	k.getStore(ctx).Set(multiSigPrefix.AppendStr(string(info.KeyID)), &info)
}

// GetMultisigPubKey returns the pub keys for a given keyID, if it exists
func (k Keeper) GetMultisigPubKey(ctx sdk.Context, keyID exported.KeyID) (exported.MultisigKey, bool) {
	var keyInfo types.MultisigKeyInfo
	k.getStore(ctx).Get(multiSigPrefix.AppendStr(string(keyID)), &keyInfo)

	role := k.getKeyRole(ctx, keyID)
	rotatedAt := k.getRotatedAt(ctx, keyID)

	return exported.MultisigKey{ID: keyID, Values: keyInfo.GetKeys(), Role: role, RotatedAt: rotatedAt}, true
}

// IsMultisigKeygenCompleted returns true if the multisig keygen process completed for the given key ID
func (k Keeper) IsMultisigKeygenCompleted(ctx sdk.Context, keyID exported.KeyID) bool {
	var info types.MultisigKeyInfo
	info, ok := k.getMultisigKeyInfo(ctx, keyID)
	if !ok {
		return false
	}

	return info.IsCompleted()
}

// GetMultisigPubKeyCount returns the number of multisig pub keys for the given key ID
func (k Keeper) GetMultisigPubKeyCount(ctx sdk.Context, keyID exported.KeyID) int64 {
	var info types.MultisigKeyInfo
	info, ok := k.getMultisigKeyInfo(ctx, keyID)
	if !ok {
		return 0
	}

	return info.KeyCount()
}

// HasValidatorSubmittedMultisigPubKey returns true if given validator has multisig pub key for the given key ID; otherwise, false
func (k Keeper) HasValidatorSubmittedMultisigPubKey(ctx sdk.Context, keyID exported.KeyID, validator sdk.ValAddress) bool {
	var info types.MultisigKeyInfo
	info, ok := k.getMultisigKeyInfo(ctx, keyID)
	if !ok {
		return false
	}
	return info.DoesParticipate(validator)
}

// DeleteMultisigKeygen deletes the multisig pub key info for the given key ID
func (k Keeper) DeleteMultisigKeygen(ctx sdk.Context, keyID exported.KeyID) {
	k.getStore(ctx).Delete(multiSigPrefix.AppendStr(string(keyID)))
}

// GetMultisigPubKeyTimeout returns the multisig keygen timeout block height for the given keyID
func (k Keeper) GetMultisigPubKeyTimeout(ctx sdk.Context, keyID exported.KeyID) (int64, bool) {
	var info types.MultisigKeyInfo
	ok := k.getStore(ctx).Get(multiSigPrefix.AppendStr(string(keyID)), &info)

	return info.Timeout, ok
}

// GetMultisigKeygenQueue returns the multisig keygen timeout queue
func (k Keeper) GetMultisigKeygenQueue(ctx sdk.Context) utils.SequenceKVQueue {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("multisig"))
	return utils.NewSequenceKVQueue(utils.NewNormalizedStore(store, k.cdc), uint64(k.getMaxSignQueueSize(ctx)), k.Logger(ctx))
}
