package utils

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/grpc"
	"github.com/tendermint/tendermint/libs/log"
	grpc2 "google.golang.org/grpc"
)

type causer interface {
	Cause() error
}

// IsABCIError checks if the error is a (wrapped) registered error
func IsABCIError(err error) bool {
	switch e := err.(type) {
	case nil:
		return false
	case *errors.Error:
		return true
	case causer:
		return IsABCIError(e.Cause())
	default:
		return false
	}
}

// ErrorWrapper wraps around a grpc server to return registered errors
type ErrorWrapper struct {
	grpc.Server
	Err    *errors.Error
	Logger func(ctx sdk.Context) log.Logger
}

// RegisterService ensures that every server method that gets registered returns a registered error
func (r ErrorWrapper) RegisterService(sd *grpc2.ServiceDesc, server interface{}) {
	for _, method := range sd.Methods {
		method.Handler = func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc2.UnaryServerInterceptor) (interface{}, error) {
			res, err := method.Handler(srv, ctx, dec, interceptor)
			if err != nil && !IsABCIError(err) {
				err = r.Err.Wrap(err.Error())
				r.Logger(sdk.UnwrapSDKContext(ctx)).Debug(err.Error())
			}
			return res, err
		}
	}

	r.Server.RegisterService(sd, server)
}
