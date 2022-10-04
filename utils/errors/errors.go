package errors

import (
	"golang.org/x/xerrors"
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
	kvs := make([]interface{}, len(e.keyVals), len(e.keyVals)+len(keyvals))
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
