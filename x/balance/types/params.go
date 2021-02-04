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
	ethereumDenom = "wei"
)

var (

	// KeyChainsAssetInfo represents the key for the chains Asset info parameter
	KeyChainsAssetInfo = []byte("assetInfo")
)

// KeyTable retrieves a subspace table for the module
func KeyTable() subspace.KeyTable {
	return subspace.NewKeyTable().RegisterParamSet(&Params{})
}

// ChainAssetInfo holds information about which forms of asset a chain supports
type ChainAssetInfo struct {
	Chain                 exported.Chain
	NativeDenom           string
	SupportsForeignAssets bool
}

// Params represent the genesis parameters for the module
type Params struct {
	ChainsAssetInfo []ChainAssetInfo
}

// DefaultParams creates the default genesis parameters
func DefaultParams() Params {
	return Params{

		ChainsAssetInfo: []ChainAssetInfo{
			{Chain: exported.Bitcoin, NativeDenom: bitcoinDenom, SupportsForeignAssets: false},
			{Chain: exported.Ethereum, NativeDenom: ethereumDenom, SupportsForeignAssets: true},
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
		subspace.NewParamSetPair(KeyChainsAssetInfo, &p.ChainsAssetInfo, validateChains),
	}
}

// Validate checks if the parameters are valid
func (p Params) Validate() error {
	return validateChains(p.ChainsAssetInfo)
}

func validateChains(infos interface{}) error {

	is, ok := infos.([]ChainAssetInfo)
	if !ok {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, fmt.Sprintf("invalid parameter type for chain asset infos: %T", infos))
	}

	for _, i := range is {
		var err error = nil

		switch i.Chain {
		case exported.Bitcoin:
			err = validateBitcoin(i)
		case exported.Ethereum:
			err = validateEthereum(i)
		case exported.NONE:
			err = sdkerrors.Wrap(types.ErrInvalidGenesis, "invalid chain")
		default:
			//non-pre-defined chain
			if i.NativeDenom == "" {
				err = sdkerrors.Wrap(types.ErrInvalidGenesis, "invalid asset denomination")
			}
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func validateBitcoin(info ChainAssetInfo) error {

	if info.NativeDenom != bitcoinDenom {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "incorrect bitcoin denomination")
	}

	if info.SupportsForeignAssets {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "bitcoin does not support foreign assets")
	}

	return nil
}

func validateEthereum(info ChainAssetInfo) error {

	if info.NativeDenom != ethereumDenom {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "incorrect ethereum denomination")
	}

	if !info.SupportsForeignAssets {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "ethereum does support foreign assets")
	}

	return nil
}
