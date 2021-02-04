package types

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/balance/exported"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
)

// Default parameter namespace
const (
	DefaultParamspace = ModuleName

	bitcoinDenom  = "satoshi"
	ethereumDenom = "eth"
)

var (

	// KeyChainsCurrencyInfo represents the key for the chains currency info parameter
	KeyChainsCurrencyInfo = []byte("currencyInfo")
)

// KeyTable retrieves a subspace table for the module
func KeyTable() subspace.KeyTable {
	return subspace.NewKeyTable().RegisterParamSet(&Params{})
}

// ChainCurrencyInfo holds information about which forms of currency a chain supports
type ChainCurrencyInfo struct {
	Chain           exported.Chain
	NativeDenom     string
	SupportsForeign bool
}

// Params represent the genesis parameters for the module
type Params struct {
	ChainsCurrencyInfo []ChainCurrencyInfo
}

// DefaultParams creates the default genesis parameters
func DefaultParams() Params {
	return Params{

		ChainsCurrencyInfo: []ChainCurrencyInfo{
			{Chain: exported.Bitcoin, NativeDenom: bitcoinDenom, SupportsForeign: false},
			{Chain: exported.Ethereum, NativeDenom: ethereumDenom, SupportsForeign: true},
		},
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
		subspace.NewParamSetPair(KeyChainsCurrencyInfo, &p.ChainsCurrencyInfo, validateChains),
	}
}

// Validate checks if the parameters are valid
func (p Params) Validate() error {
	return validateChains(p.ChainsCurrencyInfo)
}

func validateChains(infos interface{}) error {

	is, ok := infos.([]ChainCurrencyInfo)
	if !ok {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, fmt.Sprintf("invalid parameter type for chain currency infos: %T", infos))
	}

	for _, i := range is {
		var err error = nil

		switch i.Chain {
		case exported.Bitcoin:
			err = validateBitcoin(i)
		case exported.Ethereum:
			err = validateEthereum(i)
		default:
			err = sdkerrors.Wrap(types.ErrInvalidGenesis, "unknown chain")
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func validateBitcoin(info ChainCurrencyInfo) error {

	if info.NativeDenom != bitcoinDenom {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "incorrect bitcoin denomination")
	}

	if info.SupportsForeign == true {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "bitcoin does not support foreign currency")
	}

	return nil
}

func validateEthereum(info ChainCurrencyInfo) error {

	if info.NativeDenom != ethereumDenom {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "incorrect ethereum denomination")
	}

	if info.SupportsForeign == false {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "ethereum does support foreign currency")
	}

	return nil
}
