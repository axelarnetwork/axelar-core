package grpc

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/grpc"
	grpc2 "google.golang.org/grpc"

	"github.com/axelarnetwork/axelar-core/utils/errors"
)

// ServerWithSDKErrors wraps around a grpc server to return registered errors
type ServerWithSDKErrors struct {
	grpc.Server
	Err    *errorsmod.Error
	Logger func(ctx types.Context) log.Logger
}

// RegisterService ensures that every server method that gets registered returns a registered error
func (r ServerWithSDKErrors) RegisterService(sd *grpc2.ServiceDesc, server interface{}) {
	for i := range sd.Methods {
		method := sd.Methods[i]

		// use the index to modify the actual method in the range, not just a copy
		sd.Methods[i].Handler = func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc2.UnaryServerInterceptor) (interface{}, error) {
			res, err := method.Handler(srv, ctx, dec, interceptor)
			if err != nil && !errors.Is[*errorsmod.Error](err) {
				err = r.Err.Wrap(err.Error())
				r.Logger(types.UnwrapSDKContext(ctx)).Debug(err.Error())
			}
			return res, err
		}
	}

	r.Server.RegisterService(sd, server)
}
