package grpc_test

import (
	"context"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	abci "github.com/tendermint/tendermint/proto/tendermint/types"
	grpc2 "google.golang.org/grpc"

	"github.com/axelarnetwork/axelar-core/utils/errors"
	"github.com/axelarnetwork/axelar-core/utils/grpc"
	"github.com/axelarnetwork/axelar-core/utils/grpc/mock"
	. "github.com/axelarnetwork/utils/test"
)

func TestServerWithSDKErrors(t *testing.T) {
	var (
		serviceDescription *grpc2.ServiceDesc
		unwrappedHandler   func(interface{}, context.Context, func(interface{}) error, grpc2.UnaryServerInterceptor) (interface{}, error)
		wrappedHandler     func(interface{}, context.Context, func(interface{}) error, grpc2.UnaryServerInterceptor) (interface{}, error)
		ctx                context.Context
	)

	Given("a service description with a handler that returns an unregistered error", func() {
		ctx = sdk.WrapSDKContext(sdk.NewContext(nil, abci.Header{}, false, log.NewNopLogger()))

		unwrappedHandler = func(interface{}, context.Context, func(interface{}) error, grpc2.UnaryServerInterceptor) (interface{}, error) {
			return nil, fmt.Errorf("an unregistered error")
		}

		serviceDescription = &grpc2.ServiceDesc{
			Methods: []grpc2.MethodDesc{{
				Handler: unwrappedHandler,
			}},
		}
	}).
		When("the service is registered with a ServerWithSDKErrors", func() {
			registeredError := sdkerrors.Register("test", 0, "test error")

			// capture the handler from the passed in service description
			serviceWrapper := grpc.ServerWithSDKErrors{
				Server: &mock.ServerMock{RegisterServiceFunc: func(service *grpc2.ServiceDesc, _ interface{}) {
					wrappedHandler = service.Methods[0].Handler
				}},
				Err:    registeredError,
				Logger: func(ctx sdk.Context) log.Logger { return ctx.Logger() },
			}

			serviceWrapper.RegisterService(serviceDescription, nil)
		}).
		Then("the unwrapped handler returns the unregistered error", func(t *testing.T) {
			_, err := unwrappedHandler(nil, ctx, nil, nil)
			assert.False(t, errors.Is[*sdkerrors.Error](err))
		}).
		Then("the wrapped handler returns the registered error", func(t *testing.T) {
			_, err := wrappedHandler(nil, ctx, nil, nil)
			assert.True(t, errors.Is[*sdkerrors.Error](err))
		}).Run(t)
}
