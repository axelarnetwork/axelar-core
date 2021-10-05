package types

import (
	fmt "fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/axelarnetwork/axelar-core/utils"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

var (

	// KeyChains represents the key for the known chains
	KeyChains = []byte("assetInfo")
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
		Chains:                   []exported.Chain{btc.Bitcoin, evm.Ethereum, axelarnet.Axelarnet},
		ChainActivationThreshold: utils.NewThreshold(25, 100),
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
		params.NewParamSetPair(KeyChains, &m.Chains, validateChains),
		params.NewParamSetPair(KeyChainActivationThreshold, &m.ChainActivationThreshold, validateChainActivationThreshold),
	}
}

// Validate checks if the parameters are valid
func (m Params) Validate() error {
	if err := validateChains(m.Chains); err != nil {
		return err
	}

	if err := validateChainActivationThreshold(m.ChainActivationThreshold); err != nil {
		return err
	}

	return nil
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
