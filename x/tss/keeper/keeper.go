package keeper

import (
	"context"
	"io"
	"time"

	"github.com/axelarnetwork/tssd/pb"
	"google.golang.org/grpc"
)

// var (
// 	_ axTypes.BridgeKeeper = Keeper{}
// )

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

func (k Keeper) KeygenStart(info *pb.KeygenInfo) error {
	_, err := k.client.KeygenInit(k.context, info)
	if err != nil {
		return err
	}
	k.keygenStream, err = k.client.Keygen(k.context)
	if err != nil {
		return err
	}

	// TODO save my info from info.Parties[info.MyPartyIndex] ?

	// server handler https://grpc.io/docs/languages/go/basics/#bidirectional-streaming-rpc-1
	go func() {
		for {
			msg, err := k.keygenStream.Recv() // blocking
			if err == io.EOF {                // output stream closed by server
				k.keygenStream.CloseSend() // TODO is this the right place to call CloseSend?
				return
			}
			if err != nil {
				// errCh <- err
				// t.Errorf("you should never see this: %s", err)
				// sdkerrors.Wrap(types.ErrConnFailed, fmt.Sprintf("unexpected error when waiting for bitcoin node warmup: %s", err.Error()))
				return
			}

			// TODO deliver msg
			_ = msg
		}
	}()

	return nil
}

func (k Keeper) KeygenMsg(msg *pb.MessageIn) error {
	if err := k.keygenStream.Send(msg); err != nil {
		// log message
		return err
	}
	return nil
}
