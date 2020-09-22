package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrAddressNotTracked = sdkerrors.Register(ModuleName, 1, "address is not tracked")
)
