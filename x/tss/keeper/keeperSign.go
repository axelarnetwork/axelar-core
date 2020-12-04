package keeper

import (
	"fmt"

	"github.com/axelarnetwork/tssd/convert"
	tssd "github.com/axelarnetwork/tssd/pb"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	staking "github.com/axelarnetwork/axelar-core/x/staking/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

func (k Keeper) StartSign(ctx sdk.Context, info types.MsgSignStart, validators []staking.Validator) (<-chan exported.Signature, error) {
	if _, ok := k.signStreams[info.NewSigID]; ok {
		return nil, fmt.Errorf("signing protocol for ID %s already in progress", info.NewSigID)
	}

	k.Logger(ctx).Info(fmt.Sprintf("new Sign: sig_id [%s] key_id [%s] message [%s]", info.NewSigID, info.KeyID, string(info.MsgToSign)))

	// BEGIN: validity check

	// sign cannot proceed unless all validators have registered broadcast proxies
	if err := k.checkProxies(ctx, validators); err != nil {
		return nil, err
	}

	/*
		END: validity check -- any error below this point is local to the specific validator,
		so do not return an error but simply close the result channel
	*/

	sigChan := make(chan exported.Signature)

	stream, signInit := k.prepareSign(ctx, info, validators)
	if stream == nil || signInit == nil {
		close(sigChan)
		return sigChan, nil // don't propagate nondeterministic errors
	}
	k.signStreams[info.NewSigID] = stream

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
				info.NewSigID, msg.IsBroadcast, msg.ToPartyUid))
			// sender is set by broadcaster
			tssMsg := &types.MsgSignTraffic{SessionID: info.NewSigID, Payload: msg}
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

func (k Keeper) prepareSign(ctx sdk.Context, info types.MsgSignStart, validators []staking.Validator) (types.Stream, *tssd.MessageIn_SignInit) {
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

	grpcCtx, _ := k.newContext()
	stream, err := k.client.Sign(grpcCtx)
	if err != nil {
		k.Logger(ctx).Error(sdkerrors.Wrap(err, "failed tssd gRPC call Sign").Error())
		return nil, nil
	}
	k.signStreams[info.NewSigID] = stream
	// TODO refactor
	signInit := &tssd.MessageIn_SignInit{
		SignInit: &tssd.SignInit{
			NewSigUid:     info.NewSigID,
			KeyUid:        info.KeyID,
			PartyUids:     partyUids,
			MessageToSign: info.MsgToSign,
		},
	}

	k.Logger(ctx).Debug(fmt.Sprintf("my uid [%s] of %v", myAddress.String(), partyUids))
	return stream, signInit
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
func (k Keeper) GetSig(ctx sdk.Context, sigID string) (exported.Signature, error) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(sigID))
	if bz == nil {
		return exported.Signature{}, fmt.Errorf("signature not found")
	}
	r, s, err := convert.BytesToSig(bz)
	if err != nil {
		return exported.Signature{}, err
	}

	return exported.Signature{R: r, S: s}, nil
}

func (k Keeper) SetSig(ctx sdk.Context, sigID string, signature exported.Signature) error {
	bz, err := convert.SigToBytes(signature.R.Bytes(), signature.S.Bytes())
	if err != nil {
		return err
	}
	ctx.KVStore(k.storeKey).Set([]byte(sigID), bz)
	return nil
}
