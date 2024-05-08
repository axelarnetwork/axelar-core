package grpc

import (
	"context"

	"github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/grpc"
	"github.com/tendermint/tendermint/libs/log"
	grpc2 "google.golang.org/grpc"

	"github.com/axelarnetwork/axelar-core/utils/errors"
)

// ServerWithSDKErrors wraps around a grpc server to return registered errors
type ServerWithSDKErrors struct {
	grpc.Server
	Err    *sdkerrors.Error
	Logger func(ctx types.Context) log.Logger
}

// RegisterService ensures that every server method that gets registered returns a registered error
func (r ServerWithSDKErrors) RegisterService(sd *grpc2.ServiceDesc, server interface{}) {
	for i := range sd.Methods {
		method := sd.Methods[i]

		// use the index to modify the actual method in the range, not just a copy
		sd.Methods[i].Handler = func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc2.UnaryServerInterceptor) (interface{}, error) {
			res, err := method.Handler(srv, ctx, dec, interceptor)
			if err != nil && !errors.Is[*sdkerrors.Error](err) {
				err = r.Err.Wrap(err.Error())
				r.Logger(types.UnwrapSDKContext(ctx)).Debug(err.Error())
			}
			return res, err
		}
	}

	r.Server.RegisterService(sd, server)
}
