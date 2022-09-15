package errors

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/grpc"
	"github.com/tendermint/tendermint/libs/log"
	"golang.org/x/xerrors"
	grpc2 "google.golang.org/grpc"
)

type causer interface {
	Cause() error
}

// Is checks if the error is a (wrapped) error of the given type
func Is[T error](err error) bool {
	switch e := err.(type) {
	case nil:
		return false
	case T:
		return true
	case causer:
		return Is[T](e.Cause())
	case xerrors.Wrapper:
		return Is[T](e.Unwrap())
	default:
		return false
	}
}

// ServerWithSDKErrors wraps around a grpc server to return registered errors
type ServerWithSDKErrors struct {
	grpc.Server
	Err    *sdkerrors.Error
	Logger func(ctx sdk.Context) log.Logger
}

// RegisterService ensures that every server method that gets registered returns a registered error
func (r ServerWithSDKErrors) RegisterService(sd *grpc2.ServiceDesc, server interface{}) {
	for _, method := range sd.Methods {
		method.Handler = func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc2.UnaryServerInterceptor) (interface{}, error) {
			res, err := method.Handler(srv, ctx, dec, interceptor)
			if err != nil && !Is[*sdkerrors.Error](err) {
				err = r.Err.Wrap(err.Error())
				r.Logger(sdk.UnwrapSDKContext(ctx)).Debug(err.Error())
			}
			return res, err
		}
	}

	r.Server.RegisterService(sd, server)
}

// ErrWithKeyVals is an error with additional capabilities for easier logging
type ErrWithKeyVals interface {
	error
	KeyVals() []interface{}
	With(keyvals ...interface{}) ErrWithKeyVals
}

var _ ErrWithKeyVals = errorWithKeyVals{}
var _ causer = errorWithKeyVals{}
var _ xerrors.Wrapper = errorWithKeyVals{}

type errorWithKeyVals struct {
	error
	keyVals []interface{}
}

func (e errorWithKeyVals) Unwrap() error {
	return e.error
}

func (e errorWithKeyVals) Cause() error {
	return e.error
}

func (e errorWithKeyVals) KeyVals() []interface{} {
	return e.keyVals
}

func (e errorWithKeyVals) With(keyvals ...interface{}) ErrWithKeyVals {
	kvs := make([]interface{}, len(e.keyVals)+len(keyvals), len(e.keyVals))
	copy(kvs, e.keyVals)
	return errorWithKeyVals{
		error:   e.error,
		keyVals: append(kvs, keyvals...),
	}
}

// With returns an error with attached key values for logging
func With(err error, keyvals ...interface{}) ErrWithKeyVals {
	return errorWithKeyVals{
		error:   err,
		keyVals: append(KeyVals(err), keyvals...),
	}
}

// KeyVals returns key values if error is a (wrapped) ErrWithKeyVals, nil otherwise
func KeyVals(err error) []interface{} {
	switch e := err.(type) {
	case errorWithKeyVals:
		return e.KeyVals()
	case causer:
		return KeyVals(e.Cause())
	case xerrors.Wrapper:
		return KeyVals(e.Unwrap())
	default:
		return nil
	}
}
