package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"

	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

var (
	// KeyAssets represents the key for the supported assets
	KeyAssets = []byte("assetInfo")
)

// KeyTable retrieves a subspace table for the module
func KeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams creates the default genesis parameters
func DefaultParams() Params {
	return Params{
		Chains: []exported.Chain{btc.Bitcoin},
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
		params.NewParamSetPair(KeyAssets, &m.Chains, validateChains),
	}
}

// Validate checks if the parameters are valid
func (m Params) Validate() error {
	return validateChains(m.Chains)
}

func validateChains(infos interface{}) error {
	chains, ok := infos.([]exported.Chain)
	if !ok {
		return sdkerrors.Wrapf(types.ErrInvalidGenesis, "invalid parameter type for %T: %T", []exported.Chain{}, infos)
	}

	for _, c := range chains {
		if err := c.Validate(); err != nil {
			return sdkerrors.Wrapf(types.ErrInvalidGenesis, "invalid chain: %v", err)
		}
	}

	return nil
}
