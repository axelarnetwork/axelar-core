package types

import errorsmod "cosmossdk.io/errors"

// module errors
var (
	// Code 1 is a reserved code for internal errors and should not be used for anything else
	_                   = errorsmod.Register(ModuleName, 1, "internal error")
	ErrSnapshot         = errorsmod.Register(ModuleName, 2, "snapshot error")
	ErrInvalidChain     = errorsmod.Register(ModuleName, 3, "invalid chain")
	ErrInvalidValidator = errorsmod.Register(ModuleName, 4, "invalid validator")
)
