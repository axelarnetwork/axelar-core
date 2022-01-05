package utils

import "github.com/cosmos/cosmos-sdk/types/errors"

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
