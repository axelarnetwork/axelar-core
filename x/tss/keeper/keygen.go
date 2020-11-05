package keeper

import (
	"crypto/ecdsa"
	"fmt"
	"io"

	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	tssd "github.com/axelarnetwork/tssd/pb"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k *Keeper) StartKeygen(ctx sdk.Context, info types.MsgKeygenStart) error {
	k.Logger(ctx).Info(fmt.Sprintf("initiate StartKeygen: threshold [%d] key [%s] ", info.Threshold, info.NewKeyID))

	// BEGIN: validity check

	validators := k.stakingKeeper.GetAllValidators(ctx)
	if info.Threshold < 1 || info.Threshold > len(validators) {
		err := fmt.Errorf("invalid threshold: %d, validators: %d", info.Threshold, len(validators))
		k.Logger(ctx).Error(err.Error())
		return err
	}
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
		k.Logger(ctx).Info("my validator address is empty; I must not be a validator; ignore StartKeygen")
		return nil
	}

	// populate a []tss.Party with all validator addresses
	// TODO refactor into partyUids := addrToUid(validators) (partyUids []string, myIndex int)
	partyUids := make([]string, 0, len(validators))
	ok, myIndex := false, 0
	for i, v := range validators {
		partyUids = append(partyUids, v.OperatorAddress.String())
		if v.OperatorAddress.Equals(myAddress) {
			if ok {
				err := fmt.Errorf("cosmos bug: my validator address appears multiple times in the validator list: [%s]", myAddress)
				k.Logger(ctx).Error(err.Error())
				return nil // don't propagate nondeterministic errors
			}
			ok, myIndex = true, i
		}
	}
	if !ok {
		err := fmt.Errorf("cosmos bug: my validator address is not in the validator list: [%s]", myAddress)
		k.Logger(ctx).Error(err.Error())
		return nil // don't propagate nondeterministic errors
	}
	k.Logger(ctx).Debug(fmt.Sprintf("partyUids: %v", partyUids))

	k.Logger(ctx).Debug("initiate tssd gRPC call Keygen")
	var err error
	k.keygenStream, err = k.client.Keygen(k.context)
	if err != nil {
		wrapErr := sdkerrors.Wrap(err, "failed tssd gRPC call Keygen")
		k.Logger(ctx).Error(wrapErr.Error())
		return nil // don't propagate nondeterministic errors
	}
	k.Logger(ctx).Debug("successful tssd gRPC call Keygen")
	// TODO refactor
	keygenInfo := &tssd.KeygenMsgIn{
		Data: &tssd.KeygenMsgIn_Init{
			Init: &tssd.KeygenInit{
				NewKeyUid:    info.NewKeyID,
				Threshold:    int32(info.Threshold),
				PartyUids:    partyUids,
				MyPartyIndex: int32(myIndex),
			},
		},
	}
	k.Logger(ctx).Debug("initiate tssd gRPC keygen send keygen init data")
	if err := k.keygenStream.Send(keygenInfo); err != nil {
		wrapErr := sdkerrors.Wrap(err, "failed tssd gRPC keygen send keygen init data")
		k.Logger(ctx).Error(wrapErr.Error())
		return nil // don't propagate nondeterministic errors
	}
	k.Logger(ctx).Debug("successful tssd gRPC keygen send keygen init data")

	// server handler https://grpc.io/docs/languages/go/basics/#bidirectional-streaming-rpc-1
	// TODO refactor
	k.Logger(ctx).Debug("initiate gRPC handler goroutine")
	go func() {
		k.Logger(ctx).Debug("handler goroutine: begin")
		defer func() {
			k.Logger(ctx).Debug("handler goroutine: end")
		}()
		for {
			k.Logger(ctx).Debug("handler goroutine: blocking call to gRPC stream Recv...")
			msgOneof, err := k.keygenStream.Recv() // blocking
			if err == io.EOF {                     // output stream closed by server
				k.Logger(ctx).Debug("handler goroutine: gRPC stream closed by server")
				return
			}
			if err != nil {
				newErr := sdkerrors.Wrap(err, "handler goroutine: failure to receive msg from gRPC server stream")
				k.Logger(ctx).Error(newErr.Error())
				return
			}

			msg := msgOneof.GetMsg()
			if msg == nil {
				newErr := sdkerrors.Wrap(types.ErrTss, "handler goroutine: server stream should send only msg type")
				k.Logger(ctx).Error(newErr.Error())
				return
			}

			k.Logger(ctx).Debug(fmt.Sprintf("handler goroutine: outgoing keygen msg: key [%s] from me [%s] broadcast? [%t] to [%s]", info.NewKeyID, myAddress, msg.IsBroadcast, msg.ToPartyUid))
			tssMsg := types.NewMsgKeygenTraffic(info.NewKeyID, msg)
			if err := k.broadcaster.Broadcast(ctx, []broadcast.ValidatorMsg{tssMsg}); err != nil {
				newErr := sdkerrors.Wrap(err, "handler goroutine: failure to broadcast outgoing keygen msg")
				k.Logger(ctx).Error(newErr.Error())
				return
			}
			k.Logger(ctx).Debug(fmt.Sprintf("handler goroutine: successful keygen msg broadcast"))
		}
	}()

	k.Logger(ctx).Debug(fmt.Sprintf("successful StartKeygen: key [%s] ", info.NewKeyID))
	return nil
}

func (k Keeper) KeygenMsg(ctx sdk.Context, msg types.MsgKeygenTraffic) error {
	k.Logger(ctx).Debug(fmt.Sprintf("initiate KeygenMsg: key [%s] from broadcaster [%s] broadcast? [%t] to [%s]", msg.SessionID, msg.Sender, msg.Payload.IsBroadcast, msg.Payload.ToPartyUid))

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

	// convert the received types.MsgKeygenTraffic into a tssd.KeygenMsgIn
	msgIn := &tssd.KeygenMsgIn{
		Data: &tssd.KeygenMsgIn_Msg{
			Msg: &tssd.KeygenTrafficIn{
				Payload:      msg.Payload.Payload,
				IsBroadcast:  msg.Payload.IsBroadcast,
				FromPartyUid: senderAddress.String(),
			},
		},
	}

	k.Logger(ctx).Debug(fmt.Sprintf("initiate forward incoming msg to gRPC server"))
	if k.keygenStream == nil {
		k.Logger(ctx).Error("nil keygenStream")
		return nil // don't propagate nondeterministic errors
	}
	if err := k.keygenStream.Send(msgIn); err != nil {
		newErr := sdkerrors.Wrap(err, "failure to send incoming msg to gRPC server")
		k.Logger(ctx).Error(newErr.Error())
		return nil // don't propagate nondeterministic errors
	}
	k.Logger(ctx).Debug(fmt.Sprintf("successful KeygenMsg: key [%s] ", msg.SessionID))
	return nil
}

func (k *Keeper) GetKey(ctx sdk.Context, keyID string) ecdsa.PublicKey {
	return ecdsa.PublicKey{}
}
