package grpc

import (
	"context"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/grpc"
	"github.com/tendermint/tendermint/libs/log"
	grpc2 "google.golang.org/grpc"

	errors2 "github.com/axelarnetwork/axelar-core/utils/errors"
)

// ServerWithSDKErrors wraps around a grpc server to return registered errors
type ServerWithSDKErrors struct {
	grpc.Server
	Err    *errors.Error
	Logger func(ctx types.Context) log.Logger
}

// RegisterService ensures that every server method that gets registered returns a registered error
func (r ServerWithSDKErrors) RegisterService(sd *grpc2.ServiceDesc, server interface{}) {
	for _, method := range sd.Methods {
		method.Handler = func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc2.UnaryServerInterceptor) (interface{}, error) {
			res, err := method.Handler(srv, ctx, dec, interceptor)
			if err != nil && !errors2.Is[*errors.Error](err) {
				err = r.Err.Wrap(err.Error())
				r.Logger(types.UnwrapSDKContext(ctx)).Debug(err.Error())
			}
			return res, err
		}
	}

	r.Server.RegisterService(sd, server)
}
