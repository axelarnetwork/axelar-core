package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var KeyAllowedFeeDenoms = []byte("AllowedFeeDenoms")

// KeyTable retrieves a subspace table for the module
func KeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams - the module's default parameters
func DefaultParams() Params {
	return Params{
		AllowedFeeDenoms: []string{"uaxl"},
	}
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value
// pairs of the feepolicy module's parameters.
func (m *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyAllowedFeeDenoms, &m.AllowedFeeDenoms, validateAllowedFeeDenoms),
	}
}

// Validate performs a validation check on the parameters
func (m Params) Validate() error {
	return validateAllowedFeeDenoms(m.AllowedFeeDenoms)
}

func validateAllowedFeeDenoms(i interface{}) error {
	denoms, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if len(denoms) == 0 {
		return fmt.Errorf("allowed fee denoms cannot be empty")
	}

	seen := make(map[string]struct{}, len(denoms))
	for _, denom := range denoms {
		if err := sdk.ValidateDenom(denom); err != nil {
			return fmt.Errorf("invalid fee denom %q: %w", denom, err)
		}
		if _, dup := seen[denom]; dup {
			return fmt.Errorf("duplicate fee denom %q", denom)
		}
		seen[denom] = struct{}{}
	}

	return nil
}
