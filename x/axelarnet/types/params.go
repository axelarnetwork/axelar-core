package types

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter keys
var (
	KeyAssets             = []byte("assetInfo")
	KeyRouteTimeoutWindow = []byte("routeTimeoutWindow")
)

// KeyTable retrieves a subspace table for the module
func KeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams creates the default genesis parameters
func DefaultParams() Params {
	return Params{
		SupportedChains:    []string{"Bitcoin", "Ethereum"},
		RouteTimeoutWindow: 100,
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
		params.NewParamSetPair(KeyAssets, &m.SupportedChains, validateSupportedChains),
		params.NewParamSetPair(KeyRouteTimeoutWindow, &m.RouteTimeoutWindow, validateUint64("RouteTimeoutWindow")),
	}
}

// Validate checks if the parameters are valid
func (m Params) Validate() error {
	return validateSupportedChains(m.SupportedChains)
}

func validateSupportedChains(infos interface{}) error {
	supportedChains, ok := infos.([]string)
	if !ok {
		return sdkerrors.Wrapf(types.ErrInvalidGenesis, "invalid parameter type for %T: %T", []string{}, infos)
	}

	for _, chain := range supportedChains {
		if chain == "" {
			return sdkerrors.Wrap(types.ErrInvalidGenesis, "chain name cannot be an empty string")
		}
	}

	return nil
}

func validateUint64(field string) func(value interface{}) error {
	return func(value interface{}) error {
		_, ok := value.(uint64)
		if !ok {
			return fmt.Errorf("invalid parameter type for %s: %T", field, value)
		}

		return nil
	}
}
