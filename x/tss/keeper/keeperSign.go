package keeper

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/axelarnetwork/tssd/convert"
	tssd "github.com/axelarnetwork/tssd/pb"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

// StartMasterKeySign starts a tss signing protocol using the specified key for the given chain.
func (k Keeper) StartSign(ctx sdk.Context, info types.MsgSignStart, validators []snapshot.Validator) (<-chan exported.Signature, error) {
	if _, ok := k.signStreams[info.SigID]; ok {
		return nil, fmt.Errorf("signing protocol for ID %s already in progress", info.SigID)
	}

	k.Logger(ctx).Info(fmt.Sprintf("new Sign: sig_id [%s] key_id [%s] message [%s]", info.SigID, info.KeyID, string(info.MsgToSign)))

	// BEGIN: validity check

	// sign cannot proceed unless all validators have registered broadcast proxies
	if err := k.checkProxies(ctx, validators); err != nil {
		return nil, err
	}

	/*
		END: validity check -- any error below this point is local to the specific validator,
		so do not return an error but simply close the result channel
	*/

	if _, ok := k.getKeyIDForSig(ctx, info.SigID); ok {
		return nil, fmt.Errorf("sigID %s has been used before", info.SigID)
	}
	k.setKeyIDForSig(ctx, info.SigID, info.KeyID)

	sigChan := make(chan exported.Signature)

	stream, signInit := k.prepareSign(ctx, info, validators)
	if stream == nil || signInit == nil {
		close(sigChan)
		return sigChan, nil // don't propagate nondeterministic errors
	}
	k.signStreams[info.SigID] = stream

	go func() {
		if err := stream.Send(&tssd.MessageIn{Data: signInit}); err != nil {
			k.Logger(ctx).Error(sdkerrors.Wrap(err, "failed tssd gRPC sign send sign init data").Error())
		}
	}()

	broadcastChan, resChan := k.handleStream(ctx, stream)

	// handle intermediate messages
	go func() {
		for msg := range broadcastChan {
			k.Logger(ctx).Debug(fmt.Sprintf(
				"handler goroutine: outgoing msg: session id [%s] broadcast? [%t] to [%s]",
				info.SigID, msg.IsBroadcast, msg.ToPartyUid))
			// sender is set by broadcaster
			tssMsg := &types.MsgSignTraffic{SessionID: info.SigID, Payload: msg}
			if err := k.broadcaster.Broadcast(ctx, []broadcast.MsgWithSenderSetter{tssMsg}); err != nil {
				k.Logger(ctx).Error(sdkerrors.Wrap(err, "handler goroutine: failure to broadcast outgoing sign msg").Error())
				return
			}
		}
	}()

	// handle result
	go func() {
		defer close(sigChan)
		bz := <-resChan
		r, s, err := convert.BytesToSig(bz)
		if err != nil {
			k.Logger(ctx).Error(sdkerrors.Wrap(err, "handler goroutine: failure to deserialize sig").Error())
			return
		}

		sigChan <- exported.Signature{R: r, S: s}
		k.Logger(ctx).Info(fmt.Sprintf("handler goroutine: received sig from server! [%s], [%s]", r, s))
	}()
	return sigChan, nil
}

// SignMsg takes a types.MsgSignTraffic from the chain and relays it to the keygen protocol
func (k Keeper) SignMsg(ctx sdk.Context, msg types.MsgSignTraffic) error {
	msgIn, err := k.prepareTrafficIn(ctx, msg.Sender, msg.SessionID, msg.Payload)
	if err != nil {
		return err
	}
	if msgIn == nil {
		return nil // don't propagate nondeterministic errors
	}

	stream, ok := k.signStreams[msg.SessionID]
	if !ok {
		k.Logger(ctx).Error(fmt.Sprintf("no signature session with id %s", msg.SessionID))
		return nil // don't propagate nondeterministic errors
	}

	if err := stream.Send(msgIn); err != nil {
		k.Logger(ctx).Error(sdkerrors.Wrap(err, "failure to send incoming msg to gRPC server").Error())
		return nil // don't propagate nondeterministic errors
	}
	return nil
}

// GetSig returns the signature associated with sigID
// or nil, nil if no such signature exists
func (k Keeper) GetSig(ctx sdk.Context, sigID string) (exported.Signature, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(sigPrefix + sigID))
	if bz == nil {
		return exported.Signature{}, false
	}
	r, s, err := convert.BytesToSig(bz)
	if err != nil {
		// the setter is controlled by the keeper alone, so an error here should be a catastrophic failure
		panic(err)
	}

	return exported.Signature{R: r, S: s}, true
}

func (k Keeper) SetSig(ctx sdk.Context, sigID string, signature exported.Signature) error {
	// Delete the reference to the signing stream with sigID because entering this function means the tss protocol has completed
	delete(k.signStreams, sigID)

	bz, err := convert.SigToBytes(signature.R.Bytes(), signature.S.Bytes())
	if err != nil {
		return err
	}
	ctx.KVStore(k.storeKey).Set([]byte(sigPrefix+sigID), bz)
	return nil
}

func (k Keeper) prepareSign(ctx sdk.Context, info types.MsgSignStart, validators []snapshot.Validator) (types.Stream, *tssd.MessageIn_SignInit) {
	// TODO call GetLocalPrincipal only once at launch? need to wait until someone pushes a RegisterProxy message on chain...
	myAddress := k.broadcaster.GetLocalPrincipal(ctx)
	if myAddress.Empty() {
		k.Logger(ctx).Info("ignore Sign: my validator address is empty so I must not be a validator")
		return nil, nil
	}

	partyUids, _, err := addrToUids(validators, myAddress)
	if err != nil {
		k.Logger(ctx).Error(err.Error())
		return nil, nil
	}

	grpcCtx, _ := k.newGrpcContext()
	stream, err := k.client.Sign(grpcCtx)
	if err != nil {
		k.Logger(ctx).Error(sdkerrors.Wrap(err, "failed tssd gRPC call Sign").Error())
		return nil, nil
	}
	k.signStreams[info.SigID] = stream
	// TODO refactor
	signInit := &tssd.MessageIn_SignInit{
		SignInit: &tssd.SignInit{
			NewSigUid:     info.SigID,
			KeyUid:        info.KeyID,
			PartyUids:     partyUids,
			MessageToSign: info.MsgToSign,
		},
	}

	k.Logger(ctx).Debug(fmt.Sprintf("my uid [%s] of %v", myAddress.String(), partyUids))
	return stream, signInit
}

func (k Keeper) GetKeyForSigID(ctx sdk.Context, sigID string) (ecdsa.PublicKey, bool) {
	keyID, ok := k.getKeyIDForSig(ctx, sigID)
	if !ok {
		return ecdsa.PublicKey{}, false
	}
	return k.GetKey(ctx, keyID)
}

func (k Keeper) setKeyIDForSig(ctx sdk.Context, sigID string, keyID string) {
	ctx.KVStore(k.storeKey).Set([]byte(keyIDForSigPrefix+sigID), []byte(keyID))
}

func (k Keeper) getKeyIDForSig(ctx sdk.Context, sigID string) (string, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(keyIDForSigPrefix + sigID))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}
