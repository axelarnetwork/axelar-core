package keeper_test

import (
	"context"
	"fmt"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	bam "github.com/cosmos/cosmos-sdk/baseapp"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	appParams "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	rand2 "github.com/axelarnetwork/axelar-core/testutils/rand"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/reward/keeper"
	"github.com/axelarnetwork/axelar-core/x/reward/types"
	"github.com/axelarnetwork/axelar-core/x/reward/types/mock"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	votetypes "github.com/axelarnetwork/axelar-core/x/vote/types"
	. "github.com/axelarnetwork/utils/test"
	"github.com/axelarnetwork/utils/test/rand"
)

func TestHandleMsgRefundRequest(t *testing.T) {
	var (
		server           types.MsgServiceServer
		msgServiceRouter *bam.MsgServiceRouter

		refundKeeper *mock.RefunderMock
		bankKeeper   *mock.BankerMock
		ctx          sdk.Context
		msg          *types.RefundMsgRequest
	)

	givenMsgServer := Given("an Auxiliary msg server", func() {
		ctx = rand2.Context(fake.NewMultiStore(), t)
		msgServiceRouter = bam.NewMsgServiceRouter()

		refundKeeper = &mock.RefunderMock{
			LoggerFunc:              func(ctx sdk.Context) log.Logger { return log.NewTestLogger(t) },
			DeletePendingRefundFunc: func(sdk.Context, types.RefundMsgRequest) {},
		}
		bankKeeper = &mock.BankerMock{
			SendCoinsFromModuleToAccountFunc: func(context.Context, string, sdk.AccAddress, sdk.Coins) error { return nil },
		}
		server = keeper.NewMsgServerImpl(refundKeeper, bankKeeper, msgServiceRouter, appParams.MakeEncodingConfig().Codec)
	})

	failedHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return &sdk.Result{}, fmt.Errorf("failed to execute message")
	}

	succeededHandler := func(ctx context.Context, req interface{}) (interface{}, error) {

		return &sdk.Result{}, nil
	}

	givenMsgServer.
		Branch(
			When("inner message is invalid", func() {
				any := cdctypes.Any{
					TypeUrl: rand.StrBetween(5, 20),
					Value:   rand.Bytes(int(rand.I64Between(100, 1000))),
				}
				msg = &types.RefundMsgRequest{
					Sender:       rand2.AccAddr().String(),
					InnerMessage: &any,
				}
			}).Then("should return error", func(t *testing.T) {
				_, err := server.RefundMsg(sdk.WrapSDKContext(ctx), msg)
				assert.ErrorContains(t, err, "invalid inner message")
			}),
			When("msg sender mismatch", func() {
				msg = types.NewRefundMsgRequest(rand2.AccAddr(), randomMsgLink(rand2.AccAddr()))

			}).Then("should return error", func(t *testing.T) {
				_, err := server.RefundMsg(sdk.WrapSDKContext(ctx), msg)
				assert.ErrorContains(t, err, "signers mismatch")
			}),
			When("inner message is not routable", func() {
				sender := rand2.AccAddr()
				msg = types.NewRefundMsgRequest(sender, randomMsgLink(sender))

			}).Then("should return error", func(t *testing.T) {
				_, err := server.RefundMsg(sdk.WrapSDKContext(ctx), msg)
				assert.ErrorContains(t, err, "can't route message")
			}),
			When("failed to execute inner message", func() {
				sender := rand2.AccAddr()
				msg = types.NewRefundMsgRequest(sender, votetypes.NewVoteRequest(sender, vote.PollID(rand.PosI64()), evmTypes.NewVoteEvents(nexus.ChainName(rand.Str(3)))))

				registerTestService(msgServiceRouter, failedHandler)

			}).Then("should return error", func(t *testing.T) {
				_, err := server.RefundMsg(sdk.WrapSDKContext(ctx), msg)
				assert.ErrorContains(t, err, "failed to execute message")
			}),
			When("no pending refund", func() {
				refundKeeper.GetPendingRefundFunc = func(sdk.Context, types.RefundMsgRequest) (types.Refund, bool) { return types.Refund{}, false }
				sender := rand2.AccAddr()
				msg = types.NewRefundMsgRequest(sender, votetypes.NewVoteRequest(sender, vote.PollID(rand.PosI64()), evmTypes.NewVoteEvents(nexus.ChainName(rand.Str(3)))))

				registerTestService(msgServiceRouter, succeededHandler)

			}).Then("should not refund transaction fee", func(t *testing.T) {
				_, err := server.RefundMsg(sdk.WrapSDKContext(ctx), msg)
				assert.NoError(t, err)
				assert.Len(t, refundKeeper.GetPendingRefundCalls(), 1)
				assert.Len(t, bankKeeper.SendCoinsFromModuleToAccountCalls(), 0)
			}),
			When("executed inner message successfully and there is a pending refund", func() {
				refundKeeper.GetPendingRefundFunc = func(sdk.Context, types.RefundMsgRequest) (types.Refund, bool) { return types.Refund{}, false }
				sender := rand2.AccAddr()
				msg = types.NewRefundMsgRequest(sender, votetypes.NewVoteRequest(sender, vote.PollID(rand.PosI64()), evmTypes.NewVoteEvents(nexus.ChainName(rand.Str(3)))))

				refundKeeper.GetPendingRefundFunc = func(sdk.Context, types.RefundMsgRequest) (types.Refund, bool) {
					return types.Refund{Payer: rand2.AccAddr(), Fees: sdk.NewCoins(sdk.Coin{Denom: "uaxl", Amount: math.NewInt(1000)})}, true
				}
				registerTestService(msgServiceRouter, succeededHandler)

			}).Then("should refund transaction fee", func(t *testing.T) {
				_, err := server.RefundMsg(sdk.WrapSDKContext(ctx), msg)
				assert.NoError(t, err)
				assert.Len(t, refundKeeper.GetPendingRefundCalls(), 1)
				assert.Len(t, bankKeeper.SendCoinsFromModuleToAccountCalls(), 1)
			}),
		).
		Run(t)

}

func randomMsgLink(sender sdk.AccAddress) *axelarnet.LinkRequest {
	return axelarnet.NewLinkRequest(
		sender,
		rand.StrBetween(5, 100),
		rand.StrBetween(5, 100),
		rand.StrBetween(5, 100))

}

func registerTestService(msgServiceRouter *bam.MsgServiceRouter, msgHandler func(ctx context.Context, req interface{}) (interface{}, error)) {
	encCfg := appParams.MakeEncodingConfig()
	encCfg.InterfaceRegistry.RegisterImplementations((*sdk.Msg)(nil), &votetypes.VoteRequest{})
	msgServiceRouter.SetInterfaceRegistry(encCfg.InterfaceRegistry)

	handler := func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
		in := new(votetypes.VoteRequest)
		if err := dec(in); err != nil {
			return nil, err
		}

		info := &grpc.UnaryServerInfo{
			Server:     srv,
			FullMethod: "/axelar.vote.v1beta1.MsgService/Vote",
		}

		return interceptor(ctx, in, info, msgHandler)
	}

	var serviceDesc = grpc.ServiceDesc{
		ServiceName: "axelar.vote.v1beta1.MsgService",
		HandlerType: (*votetypes.MsgServiceServer)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "Vote",
				Handler:    handler,
			},
		},
		Streams:  []grpc.StreamDesc{},
		Metadata: "axelar/vote/v1beta1/service.proto",
	}
	msgServiceRouter.RegisterService(&serviceDesc, TestMsgServer{})
}

type TestMsgServer struct{}

var _ votetypes.MsgServiceClient = TestMsgServer{}

func (m TestMsgServer) Vote(_ context.Context, _ *votetypes.VoteRequest, _ ...grpc.CallOption) (*votetypes.VoteResponse, error) {
	return &votetypes.VoteResponse{}, nil
}
