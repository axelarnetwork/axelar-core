package keeper

import (
	"crypto/ecdsa"
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	voting "github.com/axelarnetwork/axelar-core/x/vote/exported"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

// StartKeygen starts a keygen protocol with the specified parameters
func (k Keeper) StartKeygen(ctx sdk.Context, voter types.Voter, keyID string, threshold int, snapshot snapshot.Snapshot) error {
	if _, found := k.getKeygenStart(ctx, keyID); found {
		return fmt.Errorf("keyID %s is already in use", keyID)
	}

	// set keygen participants
	var participants []string
	for _, v := range snapshot.Validators {
		participants = append(participants, v.GetOperator().String())
		k.setParticipatesInKeygen(ctx, keyID, v.GetOperator())
	}

	// store block height for this keygen to be able to verify later if the produced key is allowed as a master key
	k.setKeygenStart(ctx, keyID)

	// store snapshot round to be able to look up the correct validator set when signing with this key
	k.setSnapshotCounterForKeyID(ctx, keyID, snapshot.Counter)

	poll := voting.NewPollMeta(types.ModuleName, types.EventTypeKeygen, keyID)
	if err := voter.InitPoll(ctx, poll); err != nil {
		return err
	}

	k.Logger(ctx).Info(fmt.Sprintf("new Keygen: key_id [%s] threshold [%d]", keyID, threshold))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeKeygen,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
			sdk.NewAttribute(types.AttributeKeyKeyID, keyID),
			sdk.NewAttribute(types.AttributeKeyThreshold, strconv.Itoa(threshold)),
			sdk.NewAttribute(types.AttributeKeyParticipants, string(k.cdc.MustMarshalJSON(participants)))))

	return nil
}

// KeygenMsg takes a types.MsgKeygenTraffic from the chain and relays it to the keygen protocol
func (k Keeper) KeygenMsg(ctx sdk.Context, msg types.MsgKeygenTraffic) error {
	senderAddress := k.broadcaster.GetPrincipal(ctx, msg.Sender)
	if senderAddress.Empty() {
		return fmt.Errorf("invalid message: sender [%s] is not a validator", msg.Sender)
	}

	if !k.getParticipatesInKeygen(ctx, msg.SessionID, senderAddress) {
		return fmt.Errorf("invalid message: sender [%.20s] does not participate in keygen [%s] ", senderAddress, msg.SessionID)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeKeygen,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueMsg),
			sdk.NewAttribute(types.AttributeKeySessionID, msg.SessionID),
			sdk.NewAttribute(sdk.AttributeKeySender, senderAddress.String()),
			sdk.NewAttribute(types.AttributeKeyPayload, string(k.cdc.MustMarshalJSON(msg.Payload)))))

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
	return exported.Key{ID: keyID, Value: *pk}, true
}

// SetKey stores the given public key under the given key ID
func (k Keeper) SetKey(ctx sdk.Context, keyID string, key ecdsa.PublicKey) {
	btcecPK := btcec.PublicKey(key)
	ctx.KVStore(k.storeKey).Set([]byte(pkPrefix+keyID), btcecPK.SerializeCompressed())
}

// GetCurrentKeyID returns the current key ID for given chain and role
func (k Keeper) GetCurrentKeyID(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (string, bool) {
	return k.getKeyID(ctx, chain, k.getRotationCount(ctx, chain), keyRole)
}

// GetCurrentKey returns the current key for given chain and role
func (k Keeper) GetCurrentKey(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (exported.Key, bool) {
	if keyID, found := k.GetCurrentKeyID(ctx, chain, keyRole); found {
		return k.GetKey(ctx, keyID)
	}

	return exported.Key{}, false
}

// GetNextKeyID returns the next key ID for given chain and role
func (k Keeper) GetNextKeyID(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (string, bool) {
	return k.getKeyID(ctx, chain, k.getRotationCount(ctx, chain)+1, keyRole)
}

// GetNextKey returns the next key for given chain and role
func (k Keeper) GetNextKey(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (exported.Key, bool) {
	if keyID, found := k.GetNextKeyID(ctx, chain, keyRole); found {
		return k.GetKey(ctx, keyID)
	}

	return exported.Key{}, false
}

// AssignNextKey stores a new key for a given chain which will become the default once RotateKey is called
func (k Keeper) AssignNextKey(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole, keyID string) error {
	if _, ok := k.GetKey(ctx, keyID); !ok {
		return fmt.Errorf("key %s does not exist (yet)", keyID)
	}

	// The key entry needs to store the keyID instead of the public key, because the keyID is needed whenever
	// the keeper calls the secure private key store (e.g. for signing) and we would lose the keyID information otherwise
	k.setKeyID(ctx, chain, k.getRotationCount(ctx, chain)+1, keyRole, keyID)

	return nil
}

// RotateKey rotates to the next stored key. Returns an error if no new key has been prepared
func (k Keeper) RotateKey(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) error {
	r := k.getRotationCount(ctx, chain)
	if _, found := k.getKeyID(ctx, chain, r+1, keyRole); !found {
		return fmt.Errorf("next %s for chain %s not set", keyRole.String(), chain.Name)
	}

	k.setRotationCount(ctx, chain, r+1)

	return nil
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
	storageKey := fmt.Sprintf("%s%d_%s_%s", rotationPrefix, rotation, chain.Name, keyRole.String())

	keyId := ctx.KVStore(k.storeKey).Get([]byte(storageKey))
	if keyId == nil {
		return "", false
	}

	return string(keyId), true
}

func (k Keeper) setKeyID(ctx sdk.Context, chain nexus.Chain, rotation int64, keyRole exported.KeyRole, keyID string) {
	storageKey := fmt.Sprintf("%s%d_%s_%s", rotationPrefix, rotation, chain.Name, keyRole.String())

	ctx.KVStore(k.storeKey).Set([]byte(storageKey), []byte(keyID))
}

func (k Keeper) getRotationCount(ctx sdk.Context, chain nexus.Chain) int64 {
	bz := ctx.KVStore(k.storeKey).Get([]byte(rotationPrefix + chain.Name))
	if bz == nil {
		return 0
	}
	var rotation int64
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &rotation)
	return rotation
}

func (k Keeper) setRotationCount(ctx sdk.Context, chain nexus.Chain, rotation int64) {
	ctx.KVStore(k.storeKey).Set([]byte(rotationPrefix+chain.Name), k.cdc.MustMarshalBinaryLengthPrefixed(rotation))
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

func (k Keeper) setParticipatesInKeygen(ctx sdk.Context, keyID string, validator sdk.ValAddress) {
	ctx.KVStore(k.storeKey).Set([]byte(participatePrefix+"key_"+keyID+validator.String()), []byte{})
}

func (k Keeper) getParticipatesInKeygen(ctx sdk.Context, keyID string, validator sdk.ValAddress) bool {
	return ctx.KVStore(k.storeKey).Has([]byte(participatePrefix + "key_" + keyID + validator.String()))
}

// GetMinKeygenThreshold returns minimum threshold of stake that must be met to execute keygen
func (k Keeper) GetMinKeygenThreshold(ctx sdk.Context) utils.Threshold {
	var threshold utils.Threshold
	k.params.Get(ctx, types.KeyMinKeygenThreshold, &threshold)
	return threshold
}
