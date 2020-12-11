package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	// Code 1 is a reserved code for internal errors and should not be used for anything else
	_                   = sdkerrors.Register(ModuleName, 1, "internal error")
	ErrBroadcast        = sdkerrors.Register(ModuleName, 2, "broadcast error")
	ErrInvalidChain     = sdkerrors.Register(ModuleName, 3, "invalid chain")
	ErrInvalidValidator = sdkerrors.Register(ModuleName, 4, "invalid validator")
)
