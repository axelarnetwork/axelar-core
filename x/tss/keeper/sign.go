package keeper

import (
	"fmt"
	"io"
	"math/big"

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

	// populate a []tss.Party with all validator addresses
	// TODO refactor into partyUids := addrToUid(validators) (partyUids []string, myIndex int)
	partyUids := make([]string, 0, len(validators))
	ok := false
	for _, v := range validators {
		partyUids = append(partyUids, v.OperatorAddress.String())
		if v.OperatorAddress.Equals(myAddress) {
			if ok {
				err := fmt.Errorf("cosmos bug: my validator address appears multiple times in the validator list: [%s]", myAddress)
				k.Logger(ctx).Error(err.Error())
				return nil // don't propagate nondeterministic errors
			}
			ok = true
		}
	}
	if !ok {
		err := fmt.Errorf("cosmos bug: my validator address is not in the validator list: [%s]", myAddress)
		k.Logger(ctx).Error(err.Error())
		return nil // don't propagate nondeterministic errors
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
	k.Logger(ctx).Debug("initiate tssd gRPC sign send sign init data")
	if err := k.signStream.Send(signInfo); err != nil {
		wrapErr := sdkerrors.Wrap(err, "failed tssd gRPC sign send sign init data")
		k.Logger(ctx).Error(wrapErr.Error())
		return nil // don't propagate nondeterministic errors
	}
	k.Logger(ctx).Debug("successful tssd gRPC sign send sign init data")

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
			msgOneof, err := k.signStream.Recv() // blocking
			if err == io.EOF {                   // output stream closed by server
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

			k.Logger(ctx).Debug(fmt.Sprintf("handler goroutine: outgoing sign msg: key [%s] from me [%s] broadcast? [%t] to [%s]", info.KeyID, myAddress, msg.IsBroadcast, msg.ToPartyUid))
			tssMsg := types.NewMsgSignTraffic(info.KeyID, msg)
			if err := k.broadcaster.Broadcast(ctx, []broadcast.ValidatorMsg{tssMsg}); err != nil {
				newErr := sdkerrors.Wrap(err, "handler goroutine: failure to broadcast outgoing keygen msg")
				k.Logger(ctx).Error(newErr.Error())
				return
			}
			k.Logger(ctx).Debug(fmt.Sprintf("handler goroutine: successful keygen msg broadcast"))
		}
	}()

	k.Logger(ctx).Debug(fmt.Sprintf("successful StartSign: key [%s] signature [%s]", info.KeyID, info.NewSigID))
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
