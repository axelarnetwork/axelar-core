package keeper

import (
	"context"
	"time"

	"github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/tssd/pb"
	"google.golang.org/grpc"
)

// var (
// 	_ axTypes.BridgeKeeper = Keeper{}
// )

type Keeper struct {
	client pb.GG18Client // TODO `pb` is not a good package name

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

func (k Keeper) KeygenStart(newKeyID string, parties []types.TSSParty, myPartyIndex int, threshold int) error {
	return nil
}

func (k Keeper) KeygenMsgIn(keyID string, payload []byte, isBroadcast bool, from types.TSSPartyID) {

}
