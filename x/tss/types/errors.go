package types

import errorsmod "cosmossdk.io/errors"

var (
	// cruft: Code 1 is a reserved code for internal errors and should not be used for anything else
	_ = errorsmod.Register(ModuleName, 1, "internal error")

	// ErrTss generic error because I want to use cosmos-sdk logging without the need to register a million error codes
	ErrTss = errorsmod.Register(ModuleName, 2, "tss error")
)
