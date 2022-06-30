package types

import (
	params "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter keys
var ()

// KeyTable returns a subspace.KeyTable that has registered all parameter types in this module's parameter set
func KeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns the module's parameter set initialized with default values
func DefaultParams() Params {
	return Params{}
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of this module's parameters.
func (m *Params) ParamSetPairs() params.ParamSetPairs {
	/*
		because the subspace package makes liberal use of pointers to set and get values from the store,
		this method needs to have a pointer receiver AND NewParamSetPair needs to receive the
		parameter values as pointer arguments, otherwise either the internal type reflection panics or the value will not be
		set on the correct Params data struct
	*/
	return params.ParamSetPairs{}
}

// Validate checks the validity of the values of the parameter set
func (m Params) Validate() error {
	return nil
}
