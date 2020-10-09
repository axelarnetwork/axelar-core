package keeper

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/tssd/pb"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"
)

type Keeper struct {
	client       pb.GG18Client // TODO `pb` is not a good package name
	keygenStream pb.GG18_KeygenClient

	// TODO cruft for grpc; can we get rid of this?
	connection        *grpc.ClientConn
	context           context.Context
	contextCancelFunc context.CancelFunc
}

func NewKeeper() (Keeper, error) {

	// start a gRPC client
	conn, err := grpc.Dial(":50051", grpc.WithInsecure(), grpc.WithBlock()) // TODO hard coded target
	if err != nil {
		return Keeper{}, err
	}
	// defer conn.Close()
	client := pb.NewGG18Client(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour) // TODO hard coded timeout
	// defer cancel()

	return Keeper{
		client:            client,
		connection:        conn,
		context:           ctx,
		contextCancelFunc: cancel,
	}, nil
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) KeygenStart(ctx sdk.Context, info *pb.KeygenInfo) error {
	k.Logger(ctx).Debug(fmt.Sprintf("start keygen protocol: %s", info.NewKeyId))
	_, err := k.client.KeygenInit(k.context, info)
	if err != nil {
		return err
	}
	k.keygenStream, err = k.client.Keygen(k.context)
	if err != nil {
		return err
	}

	// TODO save my info from info.Parties[info.MyPartyIndex]?

	// server handler https://grpc.io/docs/languages/go/basics/#bidirectional-streaming-rpc-1
	go func() {
		for {
			msg, err := k.keygenStream.Recv() // blocking
			if err == io.EOF {                // output stream closed by server
				k.Logger(ctx).Debug("stream closed by server")
				k.keygenStream.CloseSend() // TODO is this the right place to call CloseSend?
				return
			}
			if err != nil {
				newErr := sdkerrors.Wrap(err, "failure to receive streamed message from server")
				k.Logger(ctx).Error(newErr.Error()) // TODO what to do with this error?
				return
			}

			k.Logger(ctx).Debug(fmt.Sprintf("outgoing message:\nbroadcast? %t\nto: %s", msg.IsBroadcast, msg.ToPartyUid))
			// TODO deliver msg
			_ = msg
		}
	}()

	return nil
}

func (k Keeper) KeygenMsg(ctx sdk.Context, msg *pb.MessageIn) error {
	k.Logger(ctx).Debug("incoming message:\nbroadcast? %t\nfrom: %s", msg.IsBroadcast, msg.FromPartyUid)
	if err := k.keygenStream.Send(msg); err != nil {
		newErr := sdkerrors.Wrap(err, "failure to send streamed message to server")
		k.Logger(ctx).Error(newErr.Error()) // TODO Logger forces me to throw away error metadata
		return newErr
	}
	return nil
}
