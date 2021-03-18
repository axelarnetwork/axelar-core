package types

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
)

// DefaultParamspace - default parameter namespace
const (
	DefaultParamspace = ModuleName
)

// parameter keys
var (
	KeyConfirmationHeight  = []byte("confirmationHeight")
	KeyNetwork             = []byte("network")
	KeyRevoteLockingPeriod = []byte("RevoteLockingPeriod")
	KeySigCheckInterval    = []byte("KeySigCheckInterval")
)

// KeyTable returns a subspace.KeyTable that has registered all parameter types in this module's parameter set
func KeyTable() subspace.KeyTable {
	return subspace.NewKeyTable().RegisterParamSet(&Params{})
}

// Params is the parameter set for this module
type Params struct {
	ConfirmationHeight  uint64
	Network             Network
	RevoteLockingPeriod int64
	SigCheckInterval    int64
}

// DefaultParams returns the module's parameter set initialized with default values
func DefaultParams() Params {
	return Params{
		ConfirmationHeight:  1,
		Network:             Regtest,
		RevoteLockingPeriod: 50,
		SigCheckInterval:    10,
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
		subspace.NewParamSetPair(KeyRevoteLockingPeriod, &p.RevoteLockingPeriod, validateRevoteLockingPeriod),
		subspace.NewParamSetPair(KeySigCheckInterval, &p.SigCheckInterval, validateSigCheckInterval),
	}
}

func validateConfirmationHeight(height interface{}) error {
	h, ok := height.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type for confirmation height: %T", height)
	}
	if h < 1 {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "transaction confirmation height must be greater than 0")
	}
	return nil
}

func validateNetwork(network interface{}) error {
	n, ok := network.(Network)
	if !ok {
		return sdkerrors.Wrapf(types.ErrInvalidGenesis, "invalid parameter type for network: %T", network)
	}
	return n.Validate()
}

func validateRevoteLockingPeriod(period interface{}) error {
	r, ok := period.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type for revote lock period: %T", r)
	}

	if r <= 0 {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "revote lock period must be greater than 0")
	}

	return nil
}

func validateSigCheckInterval(interval interface{}) error {
	i, ok := interval.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type for signature check interval: %T", i)
	}

	if i <= 0 {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "signature check interval must be greater than 0")
	}

	return nil
}

// Validate checks the validity of the values of the parameter set
func (p Params) Validate() error {
	if err := validateConfirmationHeight(p.ConfirmationHeight); err != nil {
		return err
	}

	if err := validateNetwork(p.Network); err != nil {
		return err
	}

	if err := validateRevoteLockingPeriod(p.RevoteLockingPeriod); err != nil {
		return err
	}
	if err := validateSigCheckInterval(p.SigCheckInterval); err != nil {
		return err
	}

	return nil
}
