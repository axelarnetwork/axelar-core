package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// module errors
var (
	// Code 1 is a reserved code for internal errors and should not be used for anything else
	_                = sdkerrors.Register(ModuleName, 1, "internal error")
	ErrBitcoin       = sdkerrors.Register(ModuleName, 2, "btc bridge error")
	ErrConnFailed    = sdkerrors.Register(ModuleName, 3, "connection to the bitcoin node failed")
	ErrTimeOut       = sdkerrors.Register(ModuleName, 4, "the application timed out")
	ErrInvalidConfig = sdkerrors.Register(ModuleName, 6, "configuration of the bitcoin bridge is invalid")
)
