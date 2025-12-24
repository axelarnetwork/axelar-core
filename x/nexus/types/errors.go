package types

import errorsmod "cosmossdk.io/errors"

// module errors
var (
	// Code 1 is a reserved code for internal errors and should not be used for anything else
	_        = errorsmod.Register(ModuleName, 1, "internal error")
	ErrNexus = errorsmod.Register(ModuleName, 2, "nexus error")
	// Code 3 was ErrRateLimitExceeded - reserved
	_ = errorsmod.Register(ModuleName, 3, "transfer rate limit exceeded")
)
