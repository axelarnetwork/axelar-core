package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	// Code 1 is a reserved code for internal errors and should not be used for anything else
	_                = sdkerrors.Register(ModuleName, 1, "internal error")
	ErrInvalidConfig = sdkerrors.Register(ModuleName, 2, "configuration of the bitcoin bridge is invalid")
	ErrConnFailed    = sdkerrors.Register(ModuleName, 3, "connection to the bitcoin node failed")
	ErrTimeOut       = sdkerrors.Register(ModuleName, 4, "the application timed out")
)
