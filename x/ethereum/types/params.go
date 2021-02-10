package types

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
)

// Default parameter namespace
const (
	DefaultParamspace = ModuleName
)

var (
	KeyConfirmationHeight = []byte("confirmationHeight")
	KeyNetwork            = []byte("network")
	KeyBurneable          = []byte("burneable")

	burneable = []byte{} //TODO: ask Sammy for the burner contract byte codes
)

func KeyTable() subspace.KeyTable {
	return subspace.NewKeyTable().RegisterParamSet(&Params{})
}

type Params struct {
	ConfirmationHeight uint64
	Network            Network
	Burneable          []byte
}

func DefaultParams() Params {
	return Params{
		ConfirmationHeight: 1,
		Network:            Ganache,
		Burneable:          burneable,
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
		subspace.NewParamSetPair(KeyConfirmationHeight, &p.ConfirmationHeight, validateConfirmationHeight),
		subspace.NewParamSetPair(KeyNetwork, &p.Network, validateNetwork),
	}
}

func validateNetwork(network interface{}) error {
	n, ok := network.(Network)
	if !ok {
		return fmt.Errorf("invalid parameter type for network: %T", network)
	}
	return n.Validate()
}

func validateConfirmationHeight(height interface{}) error {
	h, ok := height.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type for confirmation height: %T", height)
	}
	if h < 0 {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "transaction confirmation height must be greater than 0")
	}
	return nil
}

func (p Params) Validate() error {
	if err := validateConfirmationHeight(p.ConfirmationHeight); err != nil {
		return err
	}
	if err := validateNetwork(p.Network); err != nil {
		return err
	}
	return nil
}
