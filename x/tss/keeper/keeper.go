package keeper

import (
	"bytes"
	"context"
	"fmt"
	"io"
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
	keygenStream  tssd.GG18_KeygenClient

	// TODO cruft for grpc; can we get rid of this?
	connection        *grpc.ClientConn
	context           context.Context
	contextCancelFunc context.CancelFunc
}

func NewKeeper(logger log.Logger, broadcaster broadcast.Broadcaster, staking types.StakingKeeper) (Keeper, error) {

	// TODO don't start gRPC unless I'm a validator???

	logger = logger.With("module", fmt.Sprintf("x/%s", types.ModuleName)) // TODO horrible copy-paste

	// start a gRPC client
	const tssdServerAddress = "host.docker.internal:50051"
	logger.Debug("dialing tssd to address: %d", tssdServerAddress)                   // TODO logger
	conn, err := grpc.Dial(tssdServerAddress, grpc.WithInsecure(), grpc.WithBlock()) // TODO hard coded target
	if err != nil {
		return Keeper{}, err
	}
	logger.Debug("done")
	client := tssd.NewGG18Client(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour) // TODO hard coded timeout

	return Keeper{
		broadcaster:       broadcaster,
		stakingKeeper:     staking,
		client:            client,
		connection:        conn,
		context:           ctx,
		contextCancelFunc: cancel,
	}, nil
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) StartKeygen(ctx sdk.Context, info types.MsgKeygenStart) error {
	k.Logger(ctx).Debug(fmt.Sprintf("start keygen protocol:\nkey id: %s\nthreshold: %d", info.NewKeyID, info.Threshold))

	// TODO call GetLocalPrincipal only once at launch? need to wait until someone pushes a RegisterProxy message on chain...
	validators := k.stakingKeeper.GetAllValidators(ctx)

	// keygen cannot proceed unless all validators have registered broadcast proxies
	// TODO this breaks if the validator set changes
	if k.broadcaster.GetProxyCount(ctx) != uint32(len(validators)) {
		err := fmt.Errorf("not enough proxies registered:\nvalidators: %d\nproxies: %d", len(validators), k.broadcaster.GetProxyCount(ctx))
		k.Logger(ctx).Error(err.Error())
		return err
	}
	myAddress := k.broadcaster.GetLocalPrincipal(ctx)
	if myAddress.Empty() {
		k.Logger(ctx).Info("my validator address is empty; I must not be a validator; I'm droppping out of this keygen session")
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
			ok, myIndex = true, i
		}
	}
	if !ok {
		err := fmt.Errorf("my address is not in the validator list")
		k.Logger(ctx).Error(err.Error())
		return err
	}

	keygenInfo := &tssd.KeygenInfo{
		NewKeyId:     info.NewKeyID,
		Threshold:    int32(info.Threshold),
		Parties:      parties,
		MyPartyIndex: int32(myIndex),
	}

	_, err := k.client.KeygenInit(k.context, keygenInfo)
	if err != nil {
		k.Logger(ctx).Error(err.Error())
		return err
	}
	k.keygenStream, err = k.client.Keygen(k.context) // TODO support concurrent sessions
	if err != nil {
		k.Logger(ctx).Error(err.Error())
		return err
	}

	// server handler https://grpc.io/docs/languages/go/basics/#bidirectional-streaming-rpc-1
	go func() {
		defer k.keygenStream.CloseSend()
		for {
			msg, err := k.keygenStream.Recv() // blocking
			if err == io.EOF {                // output stream closed by server
				k.Logger(ctx).Debug("stream closed by server")
				return
			}
			if err != nil {
				newErr := sdkerrors.Wrap(err, "failure to receive streamed message from server")
				k.Logger(ctx).Error(newErr.Error()) // TODO what to do with this error?
				return
			}
			k.Logger(ctx).Debug(fmt.Sprintf("outgoing keygen message:\nnew key id:%s\nis broadcast? %t\nto party: %s", keygenInfo.NewKeyId, msg.IsBroadcast, string(msg.ToPartyUid)))
			tssMsg := types.NewMsgTSS(keygenInfo.NewKeyId, msg)
			k.broadcaster.Broadcast(ctx, []broadcast.ValidatorMsg{tssMsg})
		}
	}()

	return nil
}

func (k Keeper) KeygenMsg(ctx sdk.Context, msg *tssd.MessageIn) error {
	k.Logger(ctx).Debug("incoming message:\nkey id: %s\nis broadcast? %t\nfrom party: %s", msg.SessionId, msg.IsBroadcast, string(msg.FromPartyUid))
	// TODO enforce protocol order of operations (eg. check for nil keygenStream)
	// TODO only participate if I'm a validator
	if err := k.keygenStream.Send(msg); err != nil {
		newErr := sdkerrors.Wrap(err, "failure to send streamed message to server")
		k.Logger(ctx).Error(newErr.Error()) // TODO Logger forces me to throw away error metadata
		return newErr
	}
	return nil
}

func (k Keeper) Close() error {
	k.contextCancelFunc()
	if err := k.connection.Close(); err != nil {
		return sdkerrors.Wrap(err, "failure to close connection to server")
	}
	return nil
}

func (k Keeper) EqualsMyUID(uid []byte) bool {
	// TODO how to get my validator address?
	myAddress := sdk.ValAddress{'t', 's', 's'}
	return bytes.Equal(uid, myAddress)
}
