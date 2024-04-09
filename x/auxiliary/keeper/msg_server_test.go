package keeper_test

import (
	"context"
	"fmt"
	"testing"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	appParams "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	rand2 "github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/auxiliary/keeper"
	"github.com/axelarnetwork/axelar-core/x/auxiliary/types"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	votetypes "github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestBatching(t *testing.T) {
	var (
		ctx              sdk.Context
		msgServer        types.MsgServiceServer
		msgServiceRouter *bam.MsgServiceRouter

		batchRequest         *types.BatchRequest
		messagehandlerCalled bool
		sender               sdk.AccAddress
		innerMessages        []sdk.Msg
	)

	givenMsgServer := Given("an Auxiliary msg server", func() {
		ctx = rand2.Context(fake.NewMultiStore())
		msgServiceRouter = bam.NewMsgServiceRouter()

		messagehandlerCalled = false
		sender = rand.AccAddr()
		innerMessages = slices.Expand2(func() sdk.Msg {
			return votetypes.NewVoteRequest(sender, vote.PollID(rand.PosI64()), evmTypes.NewVoteEvents(nexus.ChainName(rand.NormalizedStr(3))))
		}, int(rand2.I64Between(2, 10)))

		msgServer = keeper.NewMsgServer(msgServiceRouter)
	})

	withBatchRequest := func() GivenStatement {
		return Given("a batch request", func() {
			batchRequest = types.NewBatchRequest(sender, innerMessages)
		})

	}

	failedHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		messagehandlerCalled = true
		return &sdk.Result{}, fmt.Errorf("failed to execute message")
	}

	succeededHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		messagehandlerCalled = true
		sdk.UnwrapSDKContext(ctx).EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent("executed"),
		})
		return &sdk.Result{}, nil
	}

	givenMsgServer.
		Branch(
			withBatchRequest().
				When("handler is not registered", func() {}).
				Then("should not revert batch message", func(t *testing.T) {
					_, err := msgServer.Batch(sdk.WrapSDKContext(ctx), batchRequest)
					assert.NoError(t, err)
					assert.False(t, messagehandlerCalled)

					events := ctx.EventManager().Events()
					failedMessageEvent := types.BatchedMessageFailed{}
					assert.Equal(t, len(innerMessages), len(events))
					assert.True(t, slices.All(events, func(event sdk.Event) bool {
						return events[0].Type == failedMessageEvent.XXX_MessageName()
					}))

				}),

			withBatchRequest().
				When("failed to executed can fail messages", func() {
					registerTestService(msgServiceRouter, failedHandler)
				}).
				Then("should not revert batch message", func(t *testing.T) {
					_, err := msgServer.Batch(sdk.WrapSDKContext(ctx), batchRequest)
					assert.NoError(t, err)
					assert.True(t, messagehandlerCalled)

					events := ctx.EventManager().Events()
					failedMessageEvent := types.BatchedMessageFailed{}
					assert.Equal(t, len(innerMessages), len(events))
					assert.True(t, slices.All(events, func(event sdk.Event) bool {
						return events[0].Type == failedMessageEvent.XXX_MessageName()
					}))

				}),

			withBatchRequest().
				When("executed can fail message", func() {
					registerTestService(msgServiceRouter, succeededHandler)
				}).
				Then("should emit message execution events", func(t *testing.T) {
					_, err := msgServer.Batch(sdk.WrapSDKContext(ctx), batchRequest)
					assert.NoError(t, err)
					assert.True(t, messagehandlerCalled)

					events := ctx.EventManager().Events()
					assert.Equal(t, len(batchRequest.Messages), len(events))
					assert.Equal(t, "executed", events[0].Type)

				}),
		).
		Run(t)
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
