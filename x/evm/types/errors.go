package types

import errorsmod "cosmossdk.io/errors"

// module errors
var (
	// Code 1 is a reserved code for internal errors and should not be used for anything else
	_                         = errorsmod.Register(ModuleName, 1, "internal error")
	ErrEVM                    = errorsmod.Register(ModuleName, 2, "bridge error")
	ErrRotationInProgress     = errorsmod.Register(ModuleName, 3, "key rotation in progress")
	ErrSignCommandsInProgress = errorsmod.Register(ModuleName, 4, "signing for command batch in progress")
)
