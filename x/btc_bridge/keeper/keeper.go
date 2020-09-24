package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"

	"github.com/axelarnetwork/axelar-core/bridge"
	"github.com/axelarnetwork/axelar-core/x/btc_bridge/types"
)

var (
	_ bridge.Keeper = Keeper{}
)

type Keeper struct {
	client bridge.BridgeClient
	conn   *grpc.ClientConn
}

func (k Keeper) Close() error {
	return k.conn.Close()
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) TrackAddress(ctx sdk.Context, address string) error {
	k.Logger(ctx).Debug(fmt.Sprintf("start tracking address %v", address))

	if _, err := k.client.TrackAddress(context.Background(), &bridge.MsgAddress{Address: address}); err != nil {
		return err
	}

	k.Logger(ctx).Debug(fmt.Sprintf("successfully tracked all past transaction for address %v", address))

	return nil
}

func NewBtcKeeper(address string) Keeper {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return Keeper{conn: conn, client: bridge.NewBridgeClient(conn)}
}
