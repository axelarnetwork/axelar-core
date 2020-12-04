package keeper

import (
	"crypto/ecdsa"
	"fmt"
	"strconv"

	"github.com/axelarnetwork/tssd/convert"
	tssd "github.com/axelarnetwork/tssd/pb"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	staking "github.com/axelarnetwork/axelar-core/x/staking/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

const (
	rotationPrefix    = "round_"
	blockHeightPrefix = "blockHeight_"
	pkPrefix          = "pk_"
)

// StartKeygen starts a keygen protocol with the specified parameters
func (k Keeper) StartKeygen(ctx sdk.Context, keyID string, threshold int, validators []staking.Validator) (<-chan ecdsa.PublicKey, error) {
	// BEGIN: validity check

	// keygen cannot proceed unless all validators have registered broadcast proxies
	if err := k.checkProxies(ctx, validators); err != nil {
		return nil, err
	}

	if ctx.KVStore(k.storeKey).Has([]byte(blockHeightPrefix + keyID)) {
		return nil, fmt.Errorf("keyID %s is already in use", keyID)
	}

	/*
		END: validity check -- any error below this point is local to the specific validator,
		so do not return an error but simply close the result channel
	*/

	// store block height for this key gen to be able to verify later if the produced key is allowed as a master key
	k.setKeygenStart(ctx, keyID)

	k.Logger(ctx).Info(fmt.Sprintf("new Keygen: key_id [%s] threshold [%d]", keyID, threshold))

	pubkeyChan := make(chan ecdsa.PublicKey)
	if _, ok := k.keygenStreams[keyID]; ok {
		k.Logger(ctx).Info(fmt.Sprintf("keygen protocol for ID %s already in progress", keyID))
		return pubkeyChan, nil
	}

	stream, keygenInit := k.prepareKeygen(ctx, keyID, threshold, validators)
	k.keygenStreams[keyID] = stream
	if stream == nil || keygenInit == nil {
		close(pubkeyChan)
		return pubkeyChan, nil // don't propagate nondeterministic errors
	}

	go func() {
		if err := stream.Send(&tssd.MessageIn{Data: keygenInit}); err != nil {
			k.Logger(ctx).Error(sdkerrors.Wrap(err, "failed tssd gRPC keygen send keygen init data").Error())
		}
	}()

	// server handler https://grpc.io/docs/languages/go/basics/#bidirectional-streaming-rpc-1
	broadcastChan, resChan := k.handleStream(ctx, stream)

	// handle intermediate messages
	go func() {
		for msg := range broadcastChan {
			k.Logger(ctx).Debug(fmt.Sprintf(
				"handler goroutine: outgoing keygen msg: key [%s] from me [%s] to [%s] broadcast [%t]",
				keyID, keygenInit.KeygenInit.PartyUids[keygenInit.KeygenInit.MyPartyIndex], msg.ToPartyUid, msg.IsBroadcast))
			// sender is set by broadcaster
			tssMsg := &types.MsgKeygenTraffic{SessionID: keyID, Payload: msg}
			if err := k.broadcaster.BroadcastSync(ctx, []broadcast.MsgWithSenderSetter{tssMsg}); err != nil {
				k.Logger(ctx).Error(sdkerrors.Wrap(err, "handler goroutine: failure to broadcast outgoing sign msg").Error())
				return
			}
		}
	}()

	// handle result
	go func() {
		defer close(pubkeyChan)
		bz := <-resChan
		pubkey, err := convert.BytesToPubkey(bz)
		if err != nil {
			k.Logger(ctx).Error(sdkerrors.Wrap(err, "handler goroutine: failure to deserialize pubkey").Error())
			return
		}

		k.Logger(ctx).Info(fmt.Sprintf("handler goroutine: received pubkey from server! [%v]", pubkey))
		pubkeyChan <- pubkey
	}()

	return pubkeyChan, nil
}

// KeygenMsg takes a types.MsgKeygenTraffic from the chain and relays it to the keygen protocol
func (k Keeper) KeygenMsg(ctx sdk.Context, msg types.MsgKeygenTraffic) error {
	msgIn, err := k.prepareTrafficIn(ctx, msg.Sender, msg.SessionID, msg.Payload)
	if err != nil {
		return err
	}
	if msgIn == nil {
		return nil
	}

	stream, ok := k.keygenStreams[msg.SessionID]
	if !ok {
		k.Logger(ctx).Error(fmt.Sprintf("no keygen session with id %s", msg.SessionID))
		return nil // don't propagate nondeterministic errors
	}

	if err := stream.Send(msgIn); err != nil {
		k.Logger(ctx).Error(sdkerrors.Wrap(err, "failure to send incoming msg to gRPC server").Error())
		return nil // don't propagate nondeterministic errors
	}
	return nil
}

func (k Keeper) GetKey(ctx sdk.Context, keyID string) (ecdsa.PublicKey, error) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(pkPrefix + keyID))
	return convert.BytesToPubkey(bz)
}

// SetKey stores the given public key under the given key ID
func (k Keeper) SetKey(ctx sdk.Context, keyID string, key ecdsa.PublicKey) {
	// Free up the keyID from the stream map because using the keyID here means that the tss protocol has completed
	delete(k.keygenStreams, keyID)

	bz, err := convert.PubkeyToBytes(key)
	if err != nil {
		panic(err)
	}
	ctx.KVStore(k.storeKey).Set([]byte(pkPrefix+keyID), bz)
}

// GetLatestMasterKey returns the latest master key that was set for the given chain
func (k Keeper) GetLatestMasterKey(ctx sdk.Context, chain string) (ecdsa.PublicKey, error) {
	return k.GetPreviousMasterKey(ctx, chain, 0)
}

/*
GetPreviousMasterKey returns the master key for the given chain x rotations ago, where x is given by beforeCurrent

Example:
	k.GetPreviousMasterKey(ctx, "bitcoin", 3)
returns the master key for Bitcoin three rotations ago.
*/
func (k Keeper) GetPreviousMasterKey(ctx sdk.Context, chain string, beforeCurrent int64) (ecdsa.PublicKey, error) {
	// The master key entry stores the keyID of a previously successfully stored key, so we need to do a second lookup after we retrieve the ID.
	// This indirection is necessary, because we need the keyID for other purposes, eg signing

	r := k.getRotationCount(ctx, chain)
	keyId := ctx.KVStore(k.storeKey).Get([]byte(masterKeyID(r-beforeCurrent, chain)))
	if keyId == nil {
		return ecdsa.PublicKey{}, fmt.Errorf("there is no master key for chain %s %d rotations ago", chain, beforeCurrent)
	}
	return k.GetKey(ctx, string(keyId))
}

// AssignNextMasterKey stores a new master key for a given chain which will become the default once RotateMasterKey is called
func (k Keeper) AssignNextMasterKey(ctx sdk.Context, chain string, snapshotHeight int64, keyID string) error {
	keyGenHeight, ok := k.getKeygenStart(ctx, keyID)
	if !ok {
		return fmt.Errorf("there is no key with ID %s", keyID)
	}
	masterKeyHeight := k.getLatestMasterKeyHeight(ctx, chain)

	p := k.GetParams(ctx)
	// key has been generated during locking period or there already is a master key for the current snapshot
	if snapshotHeight+p.LockingPeriod > keyGenHeight || masterKeyHeight > snapshotHeight {
		return fmt.Errorf("key refresh locked")
	}

	// The master key entry needs to store the keyID instead of the public key, because the keyID is needed whenever
	// the keeper calls the secure private key store (e.g. for signing) and we would lose the keyID information otherwise
	r := k.getRotationCount(ctx, chain)
	ctx.KVStore(k.storeKey).Set([]byte(masterKeyID(r+1, chain)), []byte(keyID))

	k.Logger(ctx).Debug(fmt.Sprintf("prepared master key rotation for chain %s", chain))
	return nil
}

// RotateMasterKey rotates to the next stored master key. Returns an error if no new master key has been prepared
func (k Keeper) RotateMasterKey(ctx sdk.Context, chain string) error {
	r := k.getRotationCount(ctx, chain)
	k.setRotationCount(ctx, chain, r+1)

	k.Logger(ctx).Debug(fmt.Sprintf("rotated master key for chain %s", chain))
	return nil
}

func (k Keeper) setKeygenStart(ctx sdk.Context, keyID string) {
	ctx.KVStore(k.storeKey).Set([]byte(blockHeightPrefix+keyID), k.cdc.MustMarshalBinaryLengthPrefixed(ctx.BlockHeight()))
}

func (k Keeper) getKeygenStart(ctx sdk.Context, keyID string) (int64, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(blockHeightPrefix + keyID))
	if bz == nil {
		return 0, false
	}
	var blockHeight int64
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &blockHeight)
	return blockHeight, true
}

func (k Keeper) prepareKeygen(ctx sdk.Context, keyID string, threshold int, validators []staking.Validator) (types.Stream, *tssd.MessageIn_KeygenInit) {
	// TODO call GetLocalPrincipal only once at launch? need to wait until someone pushes a RegisterProxy message on chain...
	myAddress := k.broadcaster.GetLocalPrincipal(ctx)
	if myAddress.Empty() {
		k.Logger(ctx).Info("ignore Keygen: my validator address is empty so I must not be a validator")
		return nil, nil
	}

	partyUids, myIndex, err := addrToUids(validators, myAddress)
	if err != nil {
		k.Logger(ctx).Error(err.Error())
		return nil, nil
	}

	grpcCtx, _ := k.newContext()
	stream, err := k.client.Keygen(grpcCtx)
	if err != nil {
		k.Logger(ctx).Error(sdkerrors.Wrap(err, "failed tssd gRPC call Keygen").Error())
		return nil, nil
	}
	k.keygenStreams[keyID] = stream
	// TODO refactor
	keygenInit := &tssd.MessageIn_KeygenInit{
		KeygenInit: &tssd.KeygenInit{
			NewKeyUid:    keyID,
			Threshold:    int32(threshold),
			PartyUids:    partyUids,
			MyPartyIndex: myIndex,
		},
	}

	k.Logger(ctx).Debug(fmt.Sprintf("my uid [%s] index %d of %v", myAddress.String(), myIndex, partyUids))
	return stream, keygenInit
}

func masterKeyID(rotation int64, chain string) string {
	return rotationPrefix + strconv.FormatInt(rotation, 10) + chain
}

func (k Keeper) getRotationCount(ctx sdk.Context, chain string) int64 {
	bz := ctx.KVStore(k.storeKey).Get([]byte(rotationPrefix + chain))
	if bz == nil {
		return 0
	}
	var rotation int64
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &rotation)
	return rotation
}

func (k Keeper) setRotationCount(ctx sdk.Context, chain string, rotation int64) {
	ctx.KVStore(k.storeKey).Set([]byte(rotationPrefix+chain), k.cdc.MustMarshalBinaryLengthPrefixed(rotation))
}

func (k Keeper) getLatestMasterKeyHeight(ctx sdk.Context, chain string) int64 {
	r := k.getRotationCount(ctx, chain)
	height, ok := k.getKeygenStart(ctx, masterKeyID(r, chain))
	if !ok {
		return 0
	}
	return height
}
