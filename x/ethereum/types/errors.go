package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	// Code 1 is a reserved code for internal errors and should not be used for anything else
	_                    = sdkerrors.Register(ModuleName, 1, "internal error")
	ErrEthBridge         = sdkerrors.Register(ModuleName, 2, "eth bridge error")
	ErrConnFailed        = sdkerrors.Register(ModuleName, 3, "connection to the ethereum node failed")
	ErrTimeOut           = sdkerrors.Register(ModuleName, 4, "the application timed out")
	ErrAddressNotTracked = sdkerrors.Register(ModuleName, 5, "address is not tracked")
	ErrInvalidConfig     = sdkerrors.Register(ModuleName, 6, "configuration of the ethereum bridge is invalid")
)
