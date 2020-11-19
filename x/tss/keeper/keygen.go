package keeper

import (
	"crypto/ecdsa"
	"fmt"
	"io"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/tssd/convert"
	tssd "github.com/axelarnetwork/tssd/pb"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

func (k *Keeper) StartKeygen(ctx sdk.Context, info types.MsgKeygenStart) error {
	k.Logger(ctx).Info(fmt.Sprintf("new Keygen: key_id [%s] threshold [%d]", info.NewKeyID, info.Threshold))

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
		k.Logger(ctx).Info("ignore Keygen: my validator address is empty so I must not be a validator")
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
		err := fmt.Errorf("broadcaster module bug: my validator address is not in the validator list: [%s]", myAddress)
		k.Logger(ctx).Error(err.Error())
		return nil // don't propagate nondeterministic errors
	}

	// k.Logger(ctx).Debug("initiate tssd gRPC call Keygen")
	var err error
	k.keygenStream, err = k.client.Keygen(k.context)
	if err != nil {
		wrapErr := sdkerrors.Wrap(err, "failed tssd gRPC call Keygen")
		k.Logger(ctx).Error(wrapErr.Error())
		return nil // don't propagate nondeterministic errors
	}
	// k.Logger(ctx).Debug("successful tssd gRPC call Keygen")
	// TODO refactor
	keygenInfo := &tssd.MessageIn{
		Data: &tssd.MessageIn_KeygenInit{
			KeygenInit: &tssd.KeygenInit{
				NewKeyUid:    info.NewKeyID,
				Threshold:    int32(info.Threshold),
				PartyUids:    partyUids,
				MyPartyIndex: int32(myIndex),
			},
		},
	}

	k.Logger(ctx).Debug(fmt.Sprintf("my uid [%s] index %d of %v", myAddress.String(), myIndex, partyUids))

	// k.Logger(ctx).Debug("initiate tssd gRPC keygen init goroutine")
	go func(log log.Logger) {
		// log.Debug("keygen init goroutine: begin")
		// defer log.Debug("keygen init goroutine: end")
		if err := k.keygenStream.Send(keygenInfo); err != nil {
			wrapErr := sdkerrors.Wrap(err, "failed tssd gRPC keygen send keygen init data")
			log.Error(wrapErr.Error())
		} else {
			// log.Debug("successful tssd gRPC keygen init goroutine")
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
			msgOneof, err := k.keygenStream.Recv() // blocking
			if err == io.EOF {                     // output stream closed by server
				log.Debug("handler goroutine: gRPC stream closed by server")
				return
			}
			if err != nil {
				newErr := sdkerrors.Wrap(err, "handler goroutine: failure to receive msg from gRPC server stream")
				log.Error(newErr.Error())
				return
			}

			if msgResult := msgOneof.GetKeygenResult(); msgResult != nil {
				if err := k.keygenStream.CloseSend(); err != nil {
					newErr := sdkerrors.Wrap(err, "handler goroutine: failure to CloseSend stream")
					log.Error(newErr.Error())
					return
				}
				pubkey, err := convert.BytesToPubkey(msgResult)
				if err != nil {
					newErr := sdkerrors.Wrap(err, "handler goroutine: failure to deserialize pubkey")
					log.Error(newErr.Error())
					return
				}
				// TODO do something with the pubkey
				log.Info(fmt.Sprintf("handler goroutine: received pubkey from server! [%v]", pubkey))
				return
			}

			msg := msgOneof.GetTraffic()
			if msg == nil {
				newErr := sdkerrors.Wrap(types.ErrTss, "handler goroutine: server stream should send only msg type")
				log.Error(newErr.Error())
				return
			}

			log.Debug(fmt.Sprintf("handler goroutine: outgoing keygen msg: key [%s] from me [%s] to [%s] broadcast [%t]", info.NewKeyID, myAddress, msg.ToPartyUid, msg.IsBroadcast))
			tssMsg := types.NewMsgKeygenTraffic(info.NewKeyID, msg)
			if err := k.broadcaster.Broadcast(ctx, []broadcast.ValidatorMsg{tssMsg}); err != nil {
				newErr := sdkerrors.Wrap(err, "handler goroutine: failure to broadcast outgoing keygen msg")
				log.Error(newErr.Error())
				return
			}
			// log.Debug(fmt.Sprintf("handler goroutine: successful keygen msg broadcast"))
		}
	}(k.Logger(ctx))

	// k.Logger(ctx).Debug(fmt.Sprintf("successful StartKeygen: key [%s] ", info.NewKeyID))
	return nil
}

func (k Keeper) KeygenMsg(ctx sdk.Context, msg types.MsgKeygenTraffic) error {

	// TODO many of these checks apply to both keygen and sign; refactor them into a Msg() method

	// BEGIN: validity check

	// TODO check that msg.SessionID exists; allow concurrent sessions

	senderAddress := k.broadcaster.GetPrincipal(ctx, msg.Sender)
	if senderAddress.Empty() {
		err := fmt.Errorf("invalid message: sender [%s] is not a validator; only validators can send messages of type %T", msg.Sender, msg)
		k.Logger(ctx).Error(err.Error())
		return err
	}
	k.Logger(ctx).Debug(fmt.Sprintf("Keygen message: key [%s] from [%s] to [%s] broadcast? [%t]", msg.SessionID, senderAddress.String(), msg.Payload.ToPartyUid, msg.Payload.IsBroadcast))

	// END: validity check -- always return nil after this line!

	myAddress := k.broadcaster.GetLocalPrincipal(ctx)
	if myAddress.Empty() {
		k.Logger(ctx).Info(fmt.Sprintf("ignore Keygen message: my validator address is empty so I must not be a validator"))
		return nil
	}
	toAddress, err := sdk.ValAddressFromBech32(msg.Payload.ToPartyUid)
	if err != nil {
		newErr := sdkerrors.Wrap(err, fmt.Sprintf("failed to parse [%s] into a validator address", msg.Payload.ToPartyUid))
		k.Logger(ctx).Error(newErr.Error())
		return nil
	}
	if toAddress.String() != msg.Payload.ToPartyUid {
		k.Logger(ctx).Error("Keygen message: address parse discrepancy: given [%s] got [%s]", msg.Payload.ToPartyUid, toAddress.String())
	}
	if !msg.Payload.IsBroadcast && !myAddress.Equals(toAddress) {
		k.Logger(ctx).Info(fmt.Sprintf("I should ignore: msg to [%s] not directed to me [%s]", toAddress, myAddress))
		return nil
	}
	if msg.Payload.IsBroadcast && myAddress.Equals(senderAddress) {
		k.Logger(ctx).Info(fmt.Sprintf("I should ignore: broadcast msg from [%s] came from me [%s]", senderAddress, myAddress))
		return nil
	}
	k.Logger(ctx).Info(fmt.Sprintf("I should NOT ignore: msg from [%s] to [%s] broadcast [%t] me [%s]", senderAddress, toAddress, msg.Payload.IsBroadcast, myAddress))

	// convert the received types.MsgKeygenTraffic into a tssd.KeygenMsgIn
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
	if k.keygenStream == nil {
		k.Logger(ctx).Error("nil keygenStream")
		return nil // don't propagate nondeterministic errors
	}

	k.Logger(ctx).Debug(fmt.Sprintf("Keygen message: forward incoming to tssd: key [%s] from [%s] to [%s] broadcast [%t] me [%s]", msg.SessionID, senderAddress.String(), toAddress.String(), msg.Payload.IsBroadcast, myAddress.String()))

	if err := k.keygenStream.Send(msgIn); err != nil {
		newErr := sdkerrors.Wrap(err, "failure to send incoming msg to gRPC server")
		k.Logger(ctx).Error(newErr.Error())
		return nil // don't propagate nondeterministic errors
	}
	// k.Logger(ctx).Debug(fmt.Sprintf("successful KeygenMsg: key [%s] ", msg.SessionID))
	return nil
}

func (k Keeper) GetKey(ctx sdk.Context, keyUid string) (ecdsa.PublicKey, error) {
	pubkeyBytes, err := k.client.GetKey(
		k.context,
		&tssd.Uid{
			Uid: keyUid,
		},
	)
	if err != nil {
		return ecdsa.PublicKey{}, sdkerrors.Wrapf(err, "failure gRPC get key [%s]", keyUid)
	}
	return convert.BytesToPubkey(pubkeyBytes.Payload)
}
