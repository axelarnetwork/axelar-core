package keeper

import (
	"fmt"
	"io"
	"math/big"

	"github.com/cosmos/cosmos-sdk/x/staking/exported"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/tssd/convert"
	tssd "github.com/axelarnetwork/tssd/pb"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

// StartSign TODO refactor code copied from StartKeygen
func (k *Keeper) StartSign(ctx sdk.Context, info types.MsgSignStart) error {
	k.Logger(ctx).Info(fmt.Sprintf("new Sign: sig_id [%s] key_id [%s] message [%s]", info.NewSigID, info.KeyID, string(info.MsgToSign)))

	// TODO do validity check here, everything else in a separate func win no return value to enforce that we return only nil after the validity check has passed
	// BEGIN: validity check

	// TODO for now assume all validators participate
	var validators []exported.ValidatorI
	fnAppend := func(_ int64, v exported.ValidatorI) (stop bool) { validators = append(validators, v); return false }
	k.stakingKeeper.IterateValidators(ctx, fnAppend)
	if k.broadcaster.GetProxyCount(ctx) != uint32(len(validators)) {
		// sign cannot proceed unless all validators have registered broadcast proxies
		err := fmt.Errorf("not enough proxies registered: proxies: %d; validators: %d", k.broadcaster.GetProxyCount(ctx), len(validators))
		k.Logger(ctx).Error(err.Error())
		return err
	}

	// END: validity check -- always return nil after this line!

	// TODO call GetLocalPrincipal only once at launch? need to wait until someone pushes a RegisterProxy message on chain...
	myAddress := k.broadcaster.GetLocalPrincipal(ctx)
	if myAddress.Empty() {
		k.Logger(ctx).Info("ignore Sign: my validator address is empty so I must not be a validator")
		return nil
	}

	// build partyUids by converting validators into a []string
	partyUids := make([]string, 0, len(validators))
	for _, v := range validators {
		partyUids = append(partyUids, v.GetOperator().String())
	}

	// k.Logger(ctx).Debug("initiate tssd gRPC call Sign")
	var err error
	k.signStream, err = k.client.Sign(k.context)
	if err != nil {
		wrapErr := sdkerrors.Wrap(err, "failed tssd gRPC call Sign")
		k.Logger(ctx).Error(wrapErr.Error())
		return nil // don't propagate nondeterministic errors
	}
	// k.Logger(ctx).Debug("successful tssd gRPC call Sign")

	// TODO refactor
	signInfo := &tssd.MessageIn{
		Data: &tssd.MessageIn_SignInit{
			SignInit: &tssd.SignInit{
				NewSigUid:     info.NewSigID,
				KeyUid:        info.KeyID,
				PartyUids:     partyUids,
				MessageToSign: info.MsgToSign,
			},
		},
	}

	k.Logger(ctx).Debug(fmt.Sprintf("my uid [%s] of %v", myAddress.String(), partyUids))

	// k.Logger(ctx).Debug("initiate tssd gRPC sign init goroutine")
	go func(log log.Logger) {
		// log.Debug("sign init goroutine: begin")
		// defer log.Debug("sign init goroutine: end")
		if err := k.signStream.Send(signInfo); err != nil {
			wrapErr := sdkerrors.Wrap(err, "failed tssd gRPC sign send sign init data")
			log.Error(wrapErr.Error())
		} else {
			// log.Debug("successful tssd gRPC sign init goroutine")
		}
	}(k.Logger(ctx))

	// server handler https://grpc.io/docs/languages/go/basics/#bidirectional-streaming-rpc-1
	// TODO refactor
	// k.Logger(ctx).Debug("initiate gRPC handler goroutine")
	go func(log log.Logger) {
		// log.Debug("handler goroutine: begin")
		defer func() {
			// log.Debug("handler goroutine: end")
		}()
		for {
			// log.Debug("handler goroutine: blocking call to gRPC stream Recv...")
			msgOneof, err := k.signStream.Recv() // blocking
			if err == io.EOF {                   // output stream closed by server
				log.Debug("handler goroutine: gRPC stream closed by server")
				return
			}
			if err != nil {
				newErr := sdkerrors.Wrap(err, "handler goroutine: failure to receive msg from gRPC server stream")
				log.Error(newErr.Error())
				return
			}

			if msgResult := msgOneof.GetSignResult(); msgResult != nil {
				if err := k.signStream.CloseSend(); err != nil {
					newErr := sdkerrors.Wrap(err, "handler goroutine: failure to CloseSend stream")
					log.Error(newErr.Error())
					return
				}
				r, s, err := convert.BytesToSig(msgResult)
				if err != nil {
					newErr := sdkerrors.Wrap(err, "handler goroutine: failure to deserialize sig")
					log.Error(newErr.Error())
					return
				}
				// TODO do something with the sig
				log.Info(fmt.Sprintf("handler goroutine: received sigy from server! [%s], [%s]", r, s))
				return
			}

			msg := msgOneof.GetTraffic()
			if msg == nil {
				newErr := sdkerrors.Wrap(types.ErrTss, "handler goroutine: server stream should send only msg type")
				log.Error(newErr.Error())
				return
			}

			log.Debug(fmt.Sprintf("handler goroutine: outgoing sign msg: sig_id [%s] from me [%s] broadcast? [%t] to [%s]", info.NewSigID, myAddress, msg.IsBroadcast, msg.ToPartyUid))
			tssMsg := types.NewMsgSignTraffic(info.NewSigID, msg)
			if err := k.broadcaster.Broadcast(ctx, []broadcast.MsgWithSenderSetter{tssMsg}); err != nil {
				newErr := sdkerrors.Wrap(err, "handler goroutine: failure to broadcast outgoing sign msg")
				log.Error(newErr.Error())
				return
			}
			// log.Debug(fmt.Sprintf("handler goroutine: successful sign msg broadcast"))
		}
	}(k.Logger(ctx))

	// k.Logger(ctx).Debug(fmt.Sprintf("successful StartSign: key [%s] signature [%s]", info.KeyID, info.NewSigID))
	return nil
}

// SignMsg TODO refactor code copied from keygen
func (k Keeper) SignMsg(ctx sdk.Context, msg types.MsgSignTraffic) error {

	// TODO many of these checks apply to both keygen and sign; refactor them into a Msg() method

	// BEGIN: validity check

	// TODO check that msg.SessionID exists; allow concurrent sessions

	senderAddress := k.broadcaster.GetPrincipal(ctx, msg.Sender)
	if senderAddress.Empty() {
		err := fmt.Errorf("invalid message: sender [%s] is not a validator; only validators can send messages of type %T", msg.Sender, msg)
		k.Logger(ctx).Error(err.Error())
		return err
	}
	k.Logger(ctx).Debug(fmt.Sprintf("Sign message: sig [%s] from [%s] to [%s] broadcast? [%t]", msg.SessionID, senderAddress.String(), msg.Payload.ToPartyUid, msg.Payload.IsBroadcast))

	// END: validity check -- always return nil after this line!

	myAddress := k.broadcaster.GetLocalPrincipal(ctx)
	if myAddress.Empty() {
		k.Logger(ctx).Info(fmt.Sprintf("ignore Sign message: my validator address is empty so I must not be a validator"))
		return nil
	}
	toAddress, err := sdk.ValAddressFromBech32(msg.Payload.ToPartyUid)
	if err != nil {
		newErr := sdkerrors.Wrap(err, fmt.Sprintf("failed to parse [%s] into a validator address", msg.Payload.ToPartyUid))
		k.Logger(ctx).Error(newErr.Error())
		return nil
	}
	if toAddress.String() != msg.Payload.ToPartyUid {
		k.Logger(ctx).Error("Sign message: address parse discrepancy: given [%s] got [%s]", msg.Payload.ToPartyUid, toAddress.String())
	}
	// TODO this ignore code is buggy but I don't know why
	if !msg.Payload.IsBroadcast && !myAddress.Equals(toAddress) {
		k.Logger(ctx).Info(fmt.Sprintf("I should ignore: msg to [%s] not directed to me [%s]", toAddress, myAddress))
		return nil
	} else if msg.Payload.IsBroadcast && myAddress.Equals(senderAddress) {
		k.Logger(ctx).Info(fmt.Sprintf("I should ignore: broadcast msg from [%s] came from me [%s]", senderAddress, myAddress))
		return nil
	} else {
		k.Logger(ctx).Info(fmt.Sprintf("I should NOT ignore: msg from [%s] to [%s] broadcast [%t] me [%s]", senderAddress, toAddress, msg.Payload.IsBroadcast, myAddress))
	}

	// convert the received types.MsgSignTraffic into a tssd.SignMsgIn
	msgIn := &tssd.MessageIn{
		Data: &tssd.MessageIn_Traffic{
			Traffic: &tssd.TrafficIn{
				Payload:      msg.Payload.Payload,
				IsBroadcast:  msg.Payload.IsBroadcast,
				FromPartyUid: senderAddress.String(),
			},
		},
	}

	// k.Logger(ctx).Debug(fmt.Sprintf("initiate forward incoming msg to gRPC server"))
	if k.signStream == nil {
		k.Logger(ctx).Error("nil signStream")
		return nil // don't propagate nondeterministic errors
	}

	k.Logger(ctx).Debug(fmt.Sprintf("Sign message: forward incoming to tssd: sig [%s] from [%s] to [%s] broadcast [%t] me [%s]", msg.SessionID, senderAddress.String(), toAddress.String(), msg.Payload.IsBroadcast, myAddress.String()))

	if err := k.signStream.Send(msgIn); err != nil {
		newErr := sdkerrors.Wrap(err, "failure to send incoming msg to gRPC server")
		k.Logger(ctx).Error(newErr.Error())
		return nil // don't propagate nondeterministic errors
	}
	// k.Logger(ctx).Debug(fmt.Sprintf("successful SignMsg: sig_id [%s] ", msg.SessionID))
	return nil
}

// GetSig returns the signature associated with sigID
// or nil, nil if no such signature exists
func (k Keeper) GetSig(ctx sdk.Context, sigUid string) (r *big.Int, s *big.Int, e error) {
	sigBytes, err := k.client.GetSig(
		k.context,
		&tssd.Uid{
			Uid: sigUid,
		},
	)
	if err != nil {
		return nil, nil, sdkerrors.Wrapf(err, "failure gRPC get sig [%s]", sigUid)
	}
	return convert.BytesToSig(sigBytes.Payload)
}
