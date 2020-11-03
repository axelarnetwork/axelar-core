package keeper

import (
	"context"
	"fmt"
	"io"
	"math/big"
	"time"

	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	tssd "github.com/axelarnetwork/tssd/pb"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"
)

type Keeper struct {
	broadcaster   broadcast.Broadcaster
	stakingKeeper types.StakingKeeper // needed only for `GetAllValidators`
	client        tssd.GG18Client
	keygenStream  tssd.GG18_KeygenClient // TODO persist in KV store instead?

	// TODO cruft for grpc; can we get rid of this?
	connection        *grpc.ClientConn
	context           context.Context
	contextCancelFunc context.CancelFunc
}

func NewKeeper(conf types.TssdConfig, logger log.Logger, broadcaster broadcast.Broadcaster, staking types.StakingKeeper) (Keeper, error) {
	logger = prepareLogger(logger)

	// TODO don't start gRPC unless I'm a validator?
	// start a gRPC client
	tssdServerAddress := conf.Host + ":" + conf.Port
	logger.Info(fmt.Sprintf("initiate connection to tssd gRPC server: %s", tssdServerAddress))
	conn, err := grpc.Dial(tssdServerAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return Keeper{}, err
	}
	logger.Debug("successful connection to tssd gRPC server")
	client := tssd.NewGG18Client(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour) // TODO config file

	return Keeper{
		broadcaster:       broadcaster,
		stakingKeeper:     staking,
		client:            client,
		connection:        conn,
		context:           ctx,
		contextCancelFunc: cancel,
	}, nil
}

func prepareLogger(logger log.Logger) log.Logger {
	return logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return prepareLogger(ctx.Logger())
}

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
	parties := make([]*tssd.Party, 0, len(validators))
	ok, myIndex := false, 0
	for i, v := range validators {
		party := &tssd.Party{
			Uid: v.OperatorAddress,
		}
		parties = append(parties, party)
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

	keygenInfo := &tssd.KeygenInfo{
		NewKeyId:     info.NewKeyID,
		Threshold:    int32(info.Threshold),
		Parties:      parties,
		MyPartyIndex: int32(myIndex),
	}

	k.Logger(ctx).Debug("initiate gRPC call KeygenInit")
	_, err := k.client.KeygenInit(k.context, keygenInfo)
	if err != nil {
		wrapErr := sdkerrors.Wrap(err, "KeygenInit failure")
		k.Logger(ctx).Error(wrapErr.Error())
		return nil // don't propagate nondeterministic errors
	}
	k.Logger(ctx).Debug("successful tssd gRPC call KeygenInit")
	k.Logger(ctx).Debug("initiate gRPC call Keygen")
	k.keygenStream, err = k.client.Keygen(k.context) // TODO support concurrent sessions
	if err != nil {
		wrapErr := sdkerrors.Wrap(err, "Keygen failure")
		k.Logger(ctx).Error(wrapErr.Error())
		return nil // don't propagate nondeterministic errors
	}
	k.Logger(ctx).Debug("successful tssd gRPC call Keygen")

	// server handler https://grpc.io/docs/languages/go/basics/#bidirectional-streaming-rpc-1
	k.Logger(ctx).Debug("initiate gRPC handler goroutine")
	go func() {
		k.Logger(ctx).Debug("handler goroutine: begin")
		defer func() {
			k.Logger(ctx).Debug("handler goroutine: end")
		}()
		for {
			k.Logger(ctx).Debug("handler goroutine: blocking call to gRPC stream Recv...")
			msg, err := k.keygenStream.Recv() // blocking
			if err == io.EOF {                // output stream closed by server
				k.Logger(ctx).Debug("handler goroutine: gRPC stream closed by server")
				return
			}
			if err != nil {
				newErr := sdkerrors.Wrap(err, "handler goroutine: failure to receive msg from gRPC server stream")
				k.Logger(ctx).Error(newErr.Error())
				return
			}
			k.Logger(ctx).Debug(fmt.Sprintf("handler goroutine: outgoing keygen msg: key [%s] from [me] broadcast? [%t] to [%s]", keygenInfo.NewKeyId, msg.IsBroadcast, sdk.ValAddress(msg.ToPartyUid)))
			tssMsg := types.NewMsgTSS(keygenInfo.NewKeyId, msg)
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

func (k Keeper) KeygenMsg(ctx sdk.Context, msg *types.MsgTSS) error {
	k.Logger(ctx).Debug(fmt.Sprintf("initiate KeygenMsg: key [%s] from [%s] broadcast? [%t] to [%s]", msg.SessionID, msg.Sender, msg.Payload.IsBroadcast, sdk.ValAddress(msg.Payload.ToPartyUid)))

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
		k.Logger(ctx).Info("ignore message: i'm not a validator; only validators care about messages of type %T", msg)
		return nil
	}
	if !msg.Payload.IsBroadcast && !myAddress.Equals(sdk.ValAddress(msg.Payload.ToPartyUid)) {
		k.Logger(ctx).Info(fmt.Sprintf("ignore message: msg to [%s] not directed to me [%s]", sdk.ValAddress(msg.Payload.ToPartyUid), myAddress))
		return nil
	}
	if msg.Payload.IsBroadcast && myAddress.Equals(senderAddress) {
		k.Logger(ctx).Info(fmt.Sprintf("ignore message: broadcast message from [%s] came from me [%s]", senderAddress, myAddress))
		return nil
	}

	// convert the received MsgTSS into a tss.MessageIn
	msgIn := &tssd.MessageIn{
		SessionId:    msg.SessionID,
		Payload:      msg.Payload.Payload,
		IsBroadcast:  msg.Payload.IsBroadcast,
		FromPartyUid: senderAddress, // TODO convert cosmos address to tss party uid
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

func (k *Keeper) StartSign(ctx sdk.Context, info types.MsgSignStart) error {
	k.Logger(ctx).Info(fmt.Sprintf("TODO not implemented: StartSign: signature [%s] key [%s] ", info.NewSigID, info.KeyID))
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

func (k Keeper) Close(logger log.Logger) error {
	logger = prepareLogger(logger)
	logger.Debug(fmt.Sprintf("initiate Close"))
	k.contextCancelFunc()
	if err := k.connection.Close(); err != nil {
		wrapErr := sdkerrors.Wrap(err, "failure to close connection to server")
		logger.Error(wrapErr.Error())
		return wrapErr
	}
	logger.Debug(fmt.Sprintf("successful Close"))
	return nil
}
