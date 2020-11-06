package keeper

import (
	"fmt"
	"io"
	"math/big"

	"github.com/tendermint/tendermint/libs/log"

	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	tssd "github.com/axelarnetwork/tssd/pb"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// StartSign TODO refactor code copied from StartKeygen
func (k *Keeper) StartSign(ctx sdk.Context, info types.MsgSignStart) error {
	k.Logger(ctx).Info(fmt.Sprintf("initiate StartSign: signature [%s] key [%s] ", info.NewSigID, info.KeyID))

	// TODO do validity check here, everything else in a separate func win no return value to enforce that we return only nil after the validity check has passed
	// BEGIN: validity check

	// TODO for now assume all validators participate
	validators := k.stakingKeeper.GetAllValidators(ctx)
	if k.broadcaster.GetProxyCount(ctx) != uint32(len(validators)) {
		// keygen cannot proceed unless all validators have registered broadcast proxies
		err := fmt.Errorf("not enough proxies registered: proxies: %d; validators: %d", k.broadcaster.GetProxyCount(ctx), len(validators))
		k.Logger(ctx).Error(err.Error())
		return err
	}

	// END: validity check -- always return nil after this line!

	// TODO call GetLocalPrincipal only once at launch? need to wait until someone pushes a RegisterProxy message on chain...
	myAddress := k.broadcaster.GetLocalPrincipal(ctx)
	if myAddress.Empty() {
		k.Logger(ctx).Info("my validator address is empty; I must not be a validator; ignore StartSign")
		return nil
	}

	// build partyUids by converting validators into a []string
	partyUids := make([]string, 0, len(validators))
	for _, v := range validators {
		partyUids = append(partyUids, v.OperatorAddress.String())
	}

	k.Logger(ctx).Debug("initiate tssd gRPC call Sign")
	var err error
	k.signStream, err = k.client.Sign(k.context)
	if err != nil {
		wrapErr := sdkerrors.Wrap(err, "failed tssd gRPC call Sign")
		k.Logger(ctx).Error(wrapErr.Error())
		return nil // don't propagate nondeterministic errors
	}
	k.Logger(ctx).Debug("successful tssd gRPC call Sign")
	// TODO refactor
	signInfo := &tssd.SignMsgIn{
		Data: &tssd.SignMsgIn_Init{
			Init: &tssd.SignInit{
				NewSigUid:     info.NewSigID,
				KeyUid:        info.KeyID,
				PartyUids:     partyUids,
				MessageToSign: info.MsgToSign,
			},
		},
	}

	k.Logger(ctx).Debug("initiate tssd gRPC sign init goroutine")
	go func(log log.Logger) {
		log.Debug("sign init goroutine: begin")
		defer log.Debug("sign init goroutine: end")
		if err := k.signStream.Send(signInfo); err != nil {
			wrapErr := sdkerrors.Wrap(err, "failed tssd gRPC sign send keygen init data")
			log.Error(wrapErr.Error())
		} else {
			log.Debug("successful tssd gRPC sign init goroutine")
		}
	}(k.Logger(ctx))

	// server handler https://grpc.io/docs/languages/go/basics/#bidirectional-streaming-rpc-1
	// TODO refactor
	k.Logger(ctx).Debug("initiate gRPC handler goroutine")
	go func(log log.Logger) {
		log.Debug("handler goroutine: begin")
		defer func() {
			log.Debug("handler goroutine: end")
		}()
		for {
			log.Debug("handler goroutine: blocking call to gRPC stream Recv...")
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

			msg := msgOneof.GetMsg()
			if msg == nil {
				newErr := sdkerrors.Wrap(types.ErrTss, "handler goroutine: server stream should send only msg type")
				log.Error(newErr.Error())
				return
			}

			log.Debug(fmt.Sprintf("handler goroutine: outgoing sign msg: key [%s] from me [%s] broadcast? [%t] to [%s]", info.KeyID, myAddress, msg.IsBroadcast, msg.ToPartyUid))
			tssMsg := types.NewMsgSignTraffic(info.KeyID, msg)
			if err := k.broadcaster.Broadcast(ctx, []broadcast.ValidatorMsg{tssMsg}); err != nil {
				newErr := sdkerrors.Wrap(err, "handler goroutine: failure to broadcast outgoing keygen msg")
				log.Error(newErr.Error())
				return
			}
			log.Debug(fmt.Sprintf("handler goroutine: successful keygen msg broadcast"))
		}
	}(k.Logger(ctx))

	k.Logger(ctx).Debug(fmt.Sprintf("successful StartSign: key [%s] signature [%s]", info.KeyID, info.NewSigID))
	return nil
}

// SignMsg TODO refactor code copied from keygen
func (k Keeper) SignMsg(ctx sdk.Context, msg types.MsgSignTraffic) error {
	k.Logger(ctx).Debug(fmt.Sprintf("initiate SignMsg: sig_id [%s] from broadcaster [%s] broadcast? [%t] to [%s]", msg.SessionID, msg.Sender, msg.Payload.IsBroadcast, msg.Payload.ToPartyUid))

	// TODO many of these checks apply to both keygen and sign; refactor them into a Msg() method

	// BEGIN: validity check

	// TODO check that msg.SessionID exists; allow concurrent sessions

	senderAddress := k.broadcaster.GetPrincipal(ctx, msg.Sender)
	if senderAddress.Empty() {
		err := fmt.Errorf("invalid message: sender [%s] is not a validator; only validators can send messages of type %T", msg.Sender, msg)
		k.Logger(ctx).Error(err.Error())
		return err
	}

	// END: validity check -- always return nil after this line!

	myAddress := k.broadcaster.GetLocalPrincipal(ctx)
	if myAddress.Empty() {
		k.Logger(ctx).Info(fmt.Sprintf("ignore message: i'm not a validator; only validators care about messages of type %T", msg))
		return nil
	}
	toAddress, err := sdk.ValAddressFromBech32(msg.Payload.ToPartyUid)
	if err != nil {
		newErr := sdkerrors.Wrap(err, fmt.Sprintf("failed to parse [%s] into a validator address", msg.Payload.ToPartyUid))
		k.Logger(ctx).Error(newErr.Error())
		return nil
	}
	k.Logger(ctx).Debug(fmt.Sprintf("myAddress [%s], senderAddress [%s], parsed toAddress [%s]", myAddress, senderAddress, toAddress))
	// TODO this ignore code is buggy but I don't know why
	// if !msg.Payload.IsBroadcast && !myAddress.Equals(toAddress) {
	// 	k.Logger(ctx).Info(fmt.Sprintf("ignore message: msg to [%s] not directed to me [%s]", toAddress, myAddress))
	// 	return nil
	// }
	// if msg.Payload.IsBroadcast && myAddress.Equals(senderAddress) {
	// 	k.Logger(ctx).Info(fmt.Sprintf("ignore message: broadcast message from [%s] came from me [%s]", senderAddress, myAddress))
	// 	return nil
	// }

	// convert the received types.MsgSignTraffic into a tssd.SignMsgIn
	msgIn := &tssd.SignMsgIn{
		Data: &tssd.SignMsgIn_Msg{
			Msg: &tssd.SignTrafficIn{
				Payload:      msg.Payload.Payload,
				IsBroadcast:  msg.Payload.IsBroadcast,
				FromPartyUid: senderAddress.String(),
			},
		},
	}

	k.Logger(ctx).Debug(fmt.Sprintf("initiate forward incoming msg to gRPC server"))
	if k.signStream == nil {
		k.Logger(ctx).Error("nil signStream")
		return nil // don't propagate nondeterministic errors
	}
	if err := k.signStream.Send(msgIn); err != nil {
		newErr := sdkerrors.Wrap(err, "failure to send incoming msg to gRPC server")
		k.Logger(ctx).Error(newErr.Error())
		return nil // don't propagate nondeterministic errors
	}
	k.Logger(ctx).Debug(fmt.Sprintf("successful SignMsg: sig_id [%s] ", msg.SessionID))
	return nil
}

// GetSig returns the signature associated with sigID
// or nil, nil if no such signature exists
// TODO we need a suiable signature struct
// Tendermint uses btcd under the hood:
// https://github.com/tendermint/tendermint/blob/1a8e42d41e9a2a21cb47806a083253ad54c22456/crypto/secp256k1/secp256k1_nocgo.go#L62
// https://github.com/btcsuite/btcd/blob/535f25593d47297f2c7f27fac7725c3b9b05727d/btcec/signature.go#L25-L29
// but we don't want to import btcd everywhere
func (k *Keeper) GetSig(ctx sdk.Context, sigID string) (r *big.Int, s *big.Int) {
	return nil, nil
}
