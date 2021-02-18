package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"

	"github.com/axelarnetwork/axelar-core/x/balance/exported"
	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	eth "github.com/axelarnetwork/axelar-core/x/ethereum/exported"
)

// DefaultParamspace - default parameter namespace
const (
	DefaultParamspace = ModuleName
)

var (

	// KeyChains represents the key for the known chains
	KeyChains = []byte("assetInfo")
)

// KeyTable retrieves a subspace table for the module
func KeyTable() subspace.KeyTable {
	return subspace.NewKeyTable().RegisterParamSet(&Params{})
}

// Params represent the genesis parameters for the module
type Params struct {
	Chains []exported.Chain
}

// DefaultParams creates the default genesis parameters
func DefaultParams() Params {
	return Params{
		Chains: []exported.Chain{btc.Bitcoin, eth.Ethereum},
	}
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of tss module's parameters.
func (p *Params) ParamSetPairs() subspace.ParamSetPairs {
	/*
		because the subspace package makes liberal use of pointers to set and get values from the store,
		this method needs to have a pointer receiver AND NewParamSetPair needs to receive the
		parameter values as pointer arguments, otherwise either the internal type reflection panics or the value will not be
		set on the correct Params data struct
	*/
	return subspace.ParamSetPairs{
		subspace.NewParamSetPair(KeyChains, &p.Chains, validateChains),
	}
}

// Validate checks if the parameters are valid
func (p Params) Validate() error {
	return validateChains(p.Chains)
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
