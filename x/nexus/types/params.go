package types

import (
	fmt "fmt"

	params "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/axelarnetwork/axelar-core/utils"
)

var (

	// KeyChainActivationThreshold represents the key for chain activation threshold
	KeyChainActivationThreshold = []byte("chainActivationThreshold")
)

// KeyTable retrieves a subspace table for the module
func KeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams creates the default genesis parameters
func DefaultParams() Params {
	return Params{
		ChainActivationThreshold: utils.NewThreshold(55, 100),
	}
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of tss module's parameters.
func (m *Params) ParamSetPairs() params.ParamSetPairs {
	/*
		because the subspace package makes liberal use of pointers to set and get values from the store,
		this method needs to have a pointer receiver AND NewParamSetPair needs to receive the
		parameter values as pointer arguments, otherwise either the internal type reflection panics or the value will not be
		set on the correct Params data struct
	*/
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyChainActivationThreshold, &m.ChainActivationThreshold, validateChainActivationThreshold),
	}
}

// Validate checks if the parameters are valid
func (m Params) Validate() error {
	if err := validateChainActivationThreshold(m.ChainActivationThreshold); err != nil {
		return err
	}

	return nil
}

func validateChainActivationThreshold(chainActivationThreshold interface{}) error {
	val, ok := chainActivationThreshold.(utils.Threshold)
	if !ok {
		return fmt.Errorf("invalid parameter type for ChainActivationThreshold: %T", chainActivationThreshold)
	}

	if val.LTE(utils.NewThreshold(0, 1)) || val.GT(utils.NewThreshold(1, 1)) {
		return fmt.Errorf("threshold must be >0 and <=1 for ChainActivationThreshold")
	}

	return nil
}
