package keeper

import (
	"crypto/ecdsa"
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	voting "github.com/axelarnetwork/axelar-core/x/vote/exported"

	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

// StartKeygen starts a keygen protocol with the specified parameters
func (k Keeper) StartKeygen(ctx sdk.Context, keyID string, threshold int, snapshot snapshot.Snapshot) error {
	if ctx.KVStore(k.storeKey).Has([]byte(keygenStartHeight + keyID)) {
		return fmt.Errorf("keyID %s is already in use", keyID)
	}

	// keygen cannot proceed unless all validators have registered broadcast proxies
	var participants []string
	for _, v := range snapshot.Validators {
		proxy := k.broadcaster.GetProxy(ctx, v.GetOperator())
		if proxy == nil {
			return fmt.Errorf("validator %s has not registered a proxy", v.GetOperator().String())
		}
		participants = append(participants, v.GetOperator().String())
		k.setParticipatesInKeygen(ctx, keyID, v.GetOperator())
	}

	// store block height for this keygen to be able to verify later if the produced key is allowed as a master key
	k.setKeygenStart(ctx, keyID)
	// store snapshot round to be able to look up the correct validator set when signing with this key
	k.setSnapshotCounterForKeyID(ctx, keyID, snapshot.Counter)

	poll := voting.PollMeta{Module: types.ModuleName, Type: types.EventTypeKeygen, ID: keyID}
	if err := k.voter.InitPoll(ctx, poll); err != nil {
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

	if !k.participatesInKeygen(ctx, msg.SessionID, senderAddress) {
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
func (k Keeper) GetKey(ctx sdk.Context, keyID string) (ecdsa.PublicKey, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(pkPrefix + keyID))
	if bz == nil {
		return ecdsa.PublicKey{}, false
	}
	btcecPK, err := btcec.ParsePubKey(bz, btcec.S256())
	// the setter is controlled by the keeper alone, so an error here should be a catastrophic failure
	if err != nil {
		panic(err)
	}
	pk := btcecPK.ToECDSA()
	return *pk, true
}

// SetKey stores the given public key under the given key ID
func (k Keeper) SetKey(ctx sdk.Context, keyID string, key ecdsa.PublicKey) {
	btcecPK := btcec.PublicKey(key)
	ctx.KVStore(k.storeKey).Set([]byte(pkPrefix+keyID), btcecPK.SerializeCompressed())
}

// GetCurrentMasterKey returns the latest master key that was set for the given chain
func (k Keeper) GetCurrentMasterKey(ctx sdk.Context, chain exported.Chain) (ecdsa.PublicKey, bool) {
	return k.GetPreviousMasterKey(ctx, chain, 0)
}

// GetCurrentMasterKeyID returns the ID of the latest master key that was set for the given chain
func (k Keeper) GetCurrentMasterKeyID(ctx sdk.Context, chain exported.Chain) (string, bool) {
	return k.getPreviousMasterKeyId(ctx, chain, 0)
}

// GetNextMasterKey returns the master key for the given chain that will be activated during the next rotation
func (k Keeper) GetNextMasterKey(ctx sdk.Context, chain exported.Chain) (ecdsa.PublicKey, bool) {
	return k.GetPreviousMasterKey(ctx, chain, -1)
}

// GetNextMasterKeyID returns the ID of the master key for the given chain that will be activated during the next rotation
func (k Keeper) GetNextMasterKeyID(ctx sdk.Context, chain exported.Chain) (string, bool) {
	return k.getPreviousMasterKeyId(ctx, chain, -1)
}

/*
GetPreviousMasterKey returns the master key for the given chain x rotations ago, where x is given by offsetFromTop

Example:
	k.GetPreviousMasterKey(ctx, "bitcoin", 3)
returns the master key for Bitcoin three rotations ago.
*/
func (k Keeper) GetPreviousMasterKey(ctx sdk.Context, chain exported.Chain, offsetFromTop int64) (ecdsa.PublicKey, bool) {
	// The master key entry stores the keyID of a previously successfully stored key, so we need to do a second lookup after we retrieve the ID.
	// This indirection is necessary, because we need the keyID for other purposes, eg signing

	keyID, ok := k.getPreviousMasterKeyId(ctx, chain, offsetFromTop)
	if !ok {
		return ecdsa.PublicKey{}, false
	}
	return k.GetKey(ctx, keyID)
}

// AssignNextMasterKey stores a new master key for a given chain which will become the default once RotateMasterKey is called
func (k Keeper) AssignNextMasterKey(ctx sdk.Context, chain exported.Chain, snapshotHeight int64, keyID string) error {
	if _, ok := k.GetKey(ctx, keyID); !ok {
		return fmt.Errorf("key %s does not exist (yet)", keyID)
	}
	keyGenHeight, ok := k.getKeygenStart(ctx, keyID)
	if !ok {
		return fmt.Errorf("there is no key with ID %s", keyID)
	}
	masterKeyHeight := k.getLatestMasterKeyHeight(ctx, chain)

	// key has been generated during locking period or there already is a master key for the current snapshot
	if snapshotHeight+k.getLockingPeriod(ctx) > keyGenHeight || masterKeyHeight > snapshotHeight {
		return fmt.Errorf("key refresh locked")
	}

	// The master key entry needs to store the keyID instead of the public key, because the keyID is needed whenever
	// the keeper calls the secure private key store (e.g. for signing) and we would lose the keyID information otherwise
	r := k.getRotationCount(ctx, chain)
	ctx.KVStore(k.storeKey).Set([]byte(masterKeyStoreKey(r+1, chain)), []byte(keyID))

	k.Logger(ctx).Debug(fmt.Sprintf("prepared master key rotation for chain %s", chain.Name))
	return nil
}

// RotateMasterKey rotates to the next stored master key. Returns an error if no new master key has been prepared
func (k Keeper) RotateMasterKey(ctx sdk.Context, chain exported.Chain) error {
	r := k.getRotationCount(ctx, chain)
	if bz := ctx.KVStore(k.storeKey).Get([]byte(masterKeyStoreKey(r+1, chain))); bz == nil {
		return fmt.Errorf("next master key for chain %s not set", chain.Name)
	}

	k.setRotationCount(ctx, chain, r+1)

	k.Logger(ctx).Debug(fmt.Sprintf("rotated master key for chain %s", chain.Name))
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

func masterKeyStoreKey(rotation int64, chain exported.Chain) string {
	return rotationPrefix + strconv.FormatInt(rotation, 10) + chain.Name
}

func (k Keeper) getPreviousMasterKeyId(ctx sdk.Context, chain exported.Chain, offsetFromTop int64) (string, bool) {
	r := k.getRotationCount(ctx, chain)
	keyId := ctx.KVStore(k.storeKey).Get([]byte(masterKeyStoreKey(r-offsetFromTop, chain)))
	if keyId == nil {
		return "", false
	}
	return string(keyId), true
}

func (k Keeper) getRotationCount(ctx sdk.Context, chain exported.Chain) int64 {
	bz := ctx.KVStore(k.storeKey).Get([]byte(rotationPrefix + chain.Name))
	if bz == nil {
		return 0
	}
	var rotation int64
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &rotation)
	return rotation
}

func (k Keeper) setRotationCount(ctx sdk.Context, chain exported.Chain, rotation int64) {
	ctx.KVStore(k.storeKey).Set([]byte(rotationPrefix+chain.Name), k.cdc.MustMarshalBinaryLengthPrefixed(rotation))
}

func (k Keeper) getLatestMasterKeyHeight(ctx sdk.Context, chain exported.Chain) int64 {
	r := k.getRotationCount(ctx, chain)
	height, ok := k.getKeygenStart(ctx, masterKeyStoreKey(r, chain))
	if !ok {
		return 0
	}
	return height
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

func (k Keeper) participatesInKeygen(ctx sdk.Context, keyID string, validator sdk.ValAddress) bool {
	return ctx.KVStore(k.storeKey).Has([]byte(participatePrefix + "key_" + keyID + validator.String()))
}

// GetMinKeygenThreshold returns minimum threshold of stake that must be met to execute keygen
func (k Keeper) GetMinKeygenThreshold(ctx sdk.Context) utils.Threshold {
	var threshold utils.Threshold
	k.params.Get(ctx, types.KeyMinKeygenThreshold, &threshold)
	return threshold
}
