package tss

import (
	"context"
	"encoding/hex"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/axelarnetwork/axelar-core/app"
	types2 "github.com/axelarnetwork/axelar-core/x/reward/types"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestGRPCTimeout(t *testing.T) {
	t.Run("connect to server", func(t *testing.T) {
		listener := bufconn.Listen(1)
		server := grpc.NewServer()
		go func() {
			if err := server.Serve(listener); err != nil {
				panic(err)
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		_, err := grpc.DialContext(
			ctx,
			"",
			grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }),
			grpc.WithInsecure(),
			grpc.WithBlock(),
		)
		assert.NoError(t, err)
	})
	t.Run("timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		_, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithBlock())
		assert.Equal(t, context.DeadlineExceeded, err)
	})
}

func TestHeartBeatResponseMarshaling(t *testing.T) {
	heartbeat1 := types.HeartBeatResponse{
		KeygenIllegibility:  7,
		SigningIllegibility: 49,
	}
	wrappedHeartbeat, _ := sdk.WrapServiceResult(sdk.Context{}, &heartbeat1, nil)

	wrappedRefundResponse1, _ := sdk.WrapServiceResult(sdk.Context{}, &types2.RefundMsgResponse{Data: wrappedHeartbeat.Data}, nil)

	heartbeat2 := types.HeartBeatResponse{
		KeygenIllegibility:  8,
		SigningIllegibility: 50,
	}
	wrappedHeartbeat, _ = sdk.WrapServiceResult(sdk.Context{}, &heartbeat2, nil)

	wrappedRefundReponse2, _ := sdk.WrapServiceResult(sdk.Context{}, &types2.RefundMsgResponse{Data: wrappedHeartbeat.Data}, nil)

	// this is what cosmos-sdk does in the background
	txMsgData := &sdk.TxMsgData{Data: []*sdk.MsgData{
		{MsgType: sdk.MsgTypeURL(&types2.RefundMsgRequest{}), Data: wrappedRefundResponse1.Data},
		{MsgType: sdk.MsgTypeURL(&types2.RefundMsgRequest{}), Data: wrappedRefundReponse2.Data},
	}}
	data, _ := proto.Marshal(txMsgData)

	resCommit := coretypes.ResultBroadcastTxCommit{DeliverTx: abci.ResponseDeliverTx{
		Data: data,
	}}

	txResp := &sdk.TxResponse{Data: strings.ToUpper(hex.EncodeToString(resCommit.DeliverTx.Data))}
	// -------------------------------

	cfg := app.MakeEncodingConfig()
	mgr := NewMgr(nil, nil, client.Context{Codec: cfg.Codec}, 0, "", nil, log.TestingLogger(), cfg.Amino)
	heartbeats, err := mgr.extractHeartBeatResponses(txResp)
	assert.NoError(t, err)
	assert.Equal(t, heartbeat1, heartbeats[0])
	assert.Equal(t, heartbeat2, heartbeats[1])
}
