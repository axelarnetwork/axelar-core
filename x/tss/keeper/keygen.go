package keeper

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/axelarnetwork/tssd/convert"
	tssd "github.com/axelarnetwork/tssd/pb"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	staking "github.com/axelarnetwork/axelar-core/x/staking/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

const (
	// add the keyID prefix to avoid id conflicts
	regularKeyPrefix    = "regular_"
	masterKeyPrefix     = "master_"
	nextMasterKeyPrefix = "next_master_"
)

// StartKeygen starts a keygen protocol with the specified parameters
func (k Keeper) StartKeygen(ctx sdk.Context, keyID string, threshold int, validators []staking.Validator) (<-chan ecdsa.PublicKey, error) {
	return k.startKeygen(ctx, regularKeyPrefix+keyID, threshold, validators)
}

func (k Keeper) startKeygen(ctx sdk.Context, keyID string, threshold int, validators []staking.Validator) (<-chan ecdsa.PublicKey, error) {
	if _, ok := k.keygenStreams[keyID]; ok {
		return nil, fmt.Errorf("keygen protocol for ID %s already in progress", keyID)
	}

	k.Logger(ctx).Info(fmt.Sprintf("new Keygen: key_id [%s] threshold [%d]", keyID, threshold))

	// BEGIN: validity check

	// keygen cannot proceed unless all validators have registered broadcast proxies
	if err := k.checkProxies(ctx, validators); err != nil {
		return nil, err
	}

	/*
		END: validity check -- any error below this point is local to the specific validator,
		so do not return an error but simply close the result channel
	*/

	pubkeyChan := make(chan ecdsa.PublicKey)

	stream, keygenInit := k.prepareKeygen(ctx, keyID, threshold, validators)
	if stream == nil || keygenInit == nil {
		close(pubkeyChan)
		return pubkeyChan, nil // don't propagate nondeterministic errors
	}
	k.keygenStreams[keyID] = stream

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
			if err := k.broadcaster.Broadcast(ctx, []broadcast.MsgWithSenderSetter{tssMsg}); err != nil {
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
	bz := ctx.KVStore(k.storeKey).Get([]byte(regularKeyPrefix + keyID))
	return convert.BytesToPubkey(bz)
}

func (k Keeper) GetMasterKey(ctx sdk.Context, chain string) (ecdsa.PublicKey, error) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(masterKeyPrefix + chain))
	return convert.BytesToPubkey(bz)
}

// StartKeyRefresh starts a keygen protocol to replace the current master k
func (k Keeper) StartKeyRefresh(ctx sdk.Context, chain string, validators []staking.Validator) (<-chan ecdsa.PublicKey, error) {
	p := k.GetParams(ctx)

	// these casts should never panic because we start out with an int
	threshold := int(p.Threshold.IsMetBy(sdk.NewInt(int64(len(validators)))).Int64())

	return k.startKeygen(ctx, nextMasterKeyPrefix+chain, threshold, validators)
}

func (k Keeper) SetKey(ctx sdk.Context, keyID string, pubkeyBytes []byte) {
	ctx.KVStore(k.storeKey).Set([]byte(regularKeyPrefix+keyID), pubkeyBytes)
}

// RotateMasterKey deletes the current master key  for a given chain and makes the next master key the current one
func (k Keeper) RotateMasterKey(ctx sdk.Context, chain string) error {
	bz := ctx.KVStore(k.storeKey).Get([]byte(nextMasterKeyPrefix + chain))
	if bz == nil {
		return fmt.Errorf("there is no next master key stored for %s", chain)
	}
	ctx.KVStore(k.storeKey).Set([]byte(masterKeyPrefix+chain), bz)
	ctx.KVStore(k.storeKey).Delete([]byte(masterKeyPrefix + chain))
	return nil
}

// SetNextMasterKey stores the next master key to switch to once all funds have been transferred from the old master key
func (k Keeper) SetNextMasterKey(ctx sdk.Context, chain string, pubkeyBytes []byte) {
	mk := types.MasterKey{BlockHeight: ctx.BlockHeight(), PK: pubkeyBytes}
	ctx.KVStore(k.storeKey).Set([]byte(nextMasterKeyPrefix+chain), k.cdc.MustMarshalBinaryLengthPrefixed(mk))
}

// IsKeyRefreshLocked checks if the locking period for the given snapshot has not expired and
// if the current master key for the given chain has been created after the snapshot
func (k Keeper) IsKeyRefreshLocked(ctx sdk.Context, chain string, snapshotHeight int64) bool {
	p := k.GetParams(ctx)
	bz := ctx.KVStore(k.storeKey).Get([]byte(chain))
	var oldBlockHeight int64
	if bz == nil {
		oldBlockHeight = 0
	} else {
		var mk types.MasterKey
		k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &mk)
		oldBlockHeight = mk.BlockHeight
	}

	return snapshotHeight+p.LockingPeriod > ctx.BlockHeight() || oldBlockHeight > snapshotHeight
}
