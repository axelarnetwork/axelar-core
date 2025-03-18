package types

import "cosmossdk.io/errors"

// module errors
var (
	// Code 1 is a reserved code for internal errors and should not be used for anything else
	_            = errors.Register(ModuleName, 1, "internal error")
	ErrAuxiliary = errors.Register(ModuleName, 2, "auxiliary module error")
)
