package types

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	// KeyMinProxyBalance is the key for the minimum proxy balance
	KeyMinProxyBalance = []byte("minproxybalance")
)

// KeyTable retrieves a subspace table for the module
func KeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams - the module's default parameters
func DefaultParams() Params {
	return Params{
		MinProxyBalance: 5000000,
	}
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of snapshot module's parameters.
func (m *Params) ParamSetPairs() params.ParamSetPairs {
	/*
		because the subspace package makes liberal use of pointers to set and get values from the store,
		this method needs to have a pointer receiver AND NewParamSetPair needs to receive the
		parameter values as pointer arguments, otherwise either the internal type reflection panics or the value will not be
		set on the correct Params data struct
	*/
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyMinProxyBalance, &m.MinProxyBalance, validateProxyBalance),
	}
}

func validateProxyBalance(balance interface{}) error {
	value, ok := balance.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type for minimum proxy balance: %T", value)
	}
	if value < 0 {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "minimum proxy balance must be greater than 0")
	}
	return nil
}

// Validate performs a validation check on the parameters
func (m Params) Validate() error {
	if err := validateProxyBalance(m.MinProxyBalance); err != nil {
		return err
	}

	return nil
}
