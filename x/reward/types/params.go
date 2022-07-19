package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var (
	KeyExternalChainVotingInflationRate = []byte("ExternalChainVotingInflationRate")
	KeyKeyMgmtRelativeInflationRate     = []byte("KeyMgmtRelativeInflationRate")
)

// KeyTable retrieves a subspace table for the module
func KeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams - the module's default parameters
func DefaultParams() Params {
	return Params{
		ExternalChainVotingInflationRate: sdk.ZeroDec(),
		KeyMgmtRelativeInflationRate:     sdk.ZeroDec(),
	}
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of reward module's parameters.
func (m *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	/*
		because the subspace package makes liberal use of pointers to set and get values from the store,
		this method needs to have a pointer receiver AND NewParamSetPair needs to receive the
		parameter values as pointer arguments, otherwise either the internal type reflection panics or the value will not be
		set on the correct Params data struct
	*/
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyExternalChainVotingInflationRate, &m.ExternalChainVotingInflationRate, validateExternalChainVotingInflationRate),
		paramtypes.NewParamSetPair(KeyKeyMgmtRelativeInflationRate, &m.KeyMgmtRelativeInflationRate, validateKeyMgmtRelativeInflationRate),
	}
}

// Validate performs a validation check on the parameters
func (m Params) Validate() error {
	if err := validateExternalChainVotingInflationRate(m.ExternalChainVotingInflationRate); err != nil {
		return err
	}

	if err := validateKeyMgmtRelativeInflationRate(m.KeyMgmtRelativeInflationRate); err != nil {
		return err
	}

	return nil
}

func validateExternalChainVotingInflationRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("external chain voting inflation rate cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("external chain voting inflation rate too large: %s", v)
	}

	return nil
}

func validateKeyMgmtRelativeInflationRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("key management inflation rate cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("key management inflation rate too large: %s", v)
	}

	return nil
}
