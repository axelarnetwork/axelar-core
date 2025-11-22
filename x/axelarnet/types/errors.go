package types

import "cosmossdk.io/errors"

// module errors
var (
	// Code 1 is a reserved code for internal errors and should not be used for anything else
	_                 = errors.Register(ModuleName, 1, "internal error")
	ErrAxelarnet      = errors.Register(ModuleName, 2, "axelarnet error")
	ErrGeneralMessage = errors.Register(ModuleName, 3, "general message error")
)
