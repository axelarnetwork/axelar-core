package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	// Code 1 is a reserved code for internal errors and should not be used for anything else
	_                         = sdkerrors.Register(ModuleName, 1, "internal error")
	ErrAddressNotTracked      = sdkerrors.Register(ModuleName, 2, "address is not tracked")
	ErrInvalidExternalAddress = sdkerrors.Register(ModuleName, 3, "invalid external address")
	ErrInvalidExternalTx      = sdkerrors.Register(ModuleName, 4, "invalid external transaction")
	ErrInvalidChain           = sdkerrors.Register(ModuleName, 5, "invalid chain")
	ErrInvalidVotes           = sdkerrors.Register(ModuleName, 6, "invalid votes")
	ErrInvalidVoter           = sdkerrors.Register(ModuleName, 7, "invalid voter")
)
