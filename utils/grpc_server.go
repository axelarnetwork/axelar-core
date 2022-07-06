package utils

import (
	"context"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/grpc"
	"github.com/tendermint/tendermint/libs/log"
	grpc2 "google.golang.org/grpc"
)

type ErrorWrapper struct {
	grpc.Server
	Err    *errors.Error
	Logger func(ctx sdk.Context) log.Logger
}

func (r ErrorWrapper) RegisterService(sd *grpc2.ServiceDesc, server interface{}) {
	for _, method := range sd.Methods {
		method.Handler = func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc2.UnaryServerInterceptor) (interface{}, error) {
			res, err := method.Handler(srv, ctx, dec, interceptor)
			if err != nil && !IsABCIError(err) {
				err = types.ErrEVM.Wrap(err.Error())
				r.Logger(sdk.UnwrapSDKContext(ctx)).Debug(err.Error())
			}
			return res, err
		}
	}

	r.Server.RegisterService(sd, server)
}
