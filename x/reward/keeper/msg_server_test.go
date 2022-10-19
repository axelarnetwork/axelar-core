package keeper_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	rand2 "github.com/axelarnetwork/axelar-core/testutils/rand"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/axelar-core/x/reward/keeper"
	"github.com/axelarnetwork/axelar-core/x/reward/types"
	"github.com/axelarnetwork/axelar-core/x/reward/types/mock"
	tsstypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	votetypes "github.com/axelarnetwork/axelar-core/x/vote/types"
	testutils "github.com/axelarnetwork/utils/test"
	"github.com/axelarnetwork/utils/test/rand"
)

func TestHandleMsgRefundRequest(t *testing.T) {
	var (
		server       types.MsgServiceServer
		refundKeeper *mock.RefunderMock
		bankKeeper   *mock.BankerMock
		ctx          sdk.Context
		router       sdk.Router
		msg          *types.RefundMsgRequest
	)
	setup := func() {
		refundKeeper = &mock.RefunderMock{
			LoggerFunc: func(ctx sdk.Context) log.Logger { return log.TestingLogger() },
			GetPendingRefundFunc: func(sdk.Context, types.RefundMsgRequest) (types.Refund, bool) {
				return types.Refund{Payer: rand2.AccAddr(), Fees: sdk.NewCoins(sdk.Coin{Denom: "uaxl", Amount: sdk.NewInt(1000)})}, true
			},
			DeletePendingRefundFunc: func(sdk.Context, types.RefundMsgRequest) {},
		}
		bankKeeper = &mock.BankerMock{
			SendCoinsFromModuleToAccountFunc: func(sdk.Context, string, sdk.AccAddress, sdk.Coins) error { return nil },
		}
		var tssHandler = func(_ sdk.Context, _ sdk.Msg) (*sdk.Result, error) {
			return &sdk.Result{}, nil
		}

		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		router = baseapp.NewRouter()
		router.AddRoute(sdk.NewRoute("tss", tssHandler))
		msgServiceRtr := baseapp.NewMsgServiceRouter()

		server = keeper.NewMsgServerImpl(refundKeeper, bankKeeper, msgServiceRtr, router)
	}

	repeatCount := 20

	t.Run("should return error when unpack invalid inner message", testutils.Func(func(t *testing.T) {
		setup()

		any := cdctypes.Any{
			TypeUrl: rand.StrBetween(5, 20),
			Value:   rand.Bytes(int(rand.I64Between(100, 1000))),
		}
		msg = &types.RefundMsgRequest{
			Sender:       rand2.AccAddr(),
			InnerMessage: &any,
		}

		_, err := server.RefundMsg(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("should return error when failed to route inner message", testutils.Func(func(t *testing.T) {
		setup()

		msg = types.NewRefundMsgRequest(rand2.AccAddr(), randomMsgLink())

		_, err := server.RefundMsg(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("should return error when failed to executed inner message", testutils.Func(func(t *testing.T) {
		setup()

		var evmHandler = func(_ sdk.Context, _ sdk.Msg) (*sdk.Result, error) {
			return &sdk.Result{}, fmt.Errorf("failed to execute message")
		}
		router.AddRoute(sdk.NewRoute("evm", evmHandler))
		voteReq := &votetypes.VoteRequest{
			Sender: rand2.AccAddr(),
			PollID: vote.PollID(rand.I64Between(5, 100)),
			Vote:   nil,
		}
		msg = types.NewRefundMsgRequest(rand2.AccAddr(), voteReq)

		_, err := server.RefundMsg(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("should not refund transaction fee when no pending refund", testutils.Func(func(t *testing.T) {
		setup()
		refundKeeper.GetPendingRefundFunc = func(sdk.Context, types.RefundMsgRequest) (types.Refund, bool) { return types.Refund{}, false }

		msg = types.NewRefundMsgRequest(rand2.AccAddr(), &tsstypes.HeartBeatRequest{})

		_, err := server.RefundMsg(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
	}).Repeat(repeatCount))

	t.Run("should refund transaction fee when executed inner message successfully", testutils.Func(func(t *testing.T) {
		setup()

		msg = types.NewRefundMsgRequest(rand2.AccAddr(), &tsstypes.HeartBeatRequest{})

		_, err := server.RefundMsg(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, refundKeeper.GetPendingRefundCalls(), 1)
		assert.Len(t, bankKeeper.SendCoinsFromModuleToAccountCalls(), 1)
	}).Repeat(repeatCount))
}

func randomMsgLink() *axelarnet.LinkRequest {
	return axelarnet.NewLinkRequest(
		rand2.AccAddr(),
		rand.StrBetween(5, 100),
		rand.StrBetween(5, 100),
		rand.StrBetween(5, 100))

}
