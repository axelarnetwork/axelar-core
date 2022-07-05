package types

import (
	fmt "fmt"

	"github.com/axelarnetwork/axelar-core/utils"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter keys
var (
	KeyKeygenThreshold  = []byte("KeygenThreshold")
	KeySigningThreshold = []byte("SigningThreshold")
	KeyKeygenTimeout    = []byte("KeygenTimeout")
)

// KeyTable returns a subspace.KeyTable that has registered all parameter types in this module's parameter set
func KeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns the module's parameter set initialized with default values
func DefaultParams() Params {
	return Params{
		KeygenThreshold:  utils.NewThreshold(90, 100),
		SigningThreshold: utils.NewThreshold(67, 100),
		KeygenTimeout:    20,
	}
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of this module's parameters.
func (m *Params) ParamSetPairs() params.ParamSetPairs {
	/*
		because the subspace package makes liberal use of pointers to set and get values from the store,
		this method needs to have a pointer receiver AND NewParamSetPair needs to receive the
		parameter values as pointer arguments, otherwise either the internal type reflection panics or the value will not be
		set on the correct Params data struct
	*/
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyKeygenThreshold, &m.KeygenThreshold, validateThreshold),
		params.NewParamSetPair(KeySigningThreshold, &m.SigningThreshold, validateThreshold),
		params.NewParamSetPair(KeyKeygenTimeout, &m.KeygenTimeout, validateKeygenTimeout),
	}
}

// Validate checks the validity of the values of the parameter set
func (m Params) Validate() error {
	if err := validateThreshold(m.KeygenThreshold); err != nil {
		return err
	}

	if err := validateThreshold(m.SigningThreshold); err != nil {
		return err
	}

	if err := validateKeygenTimeout(m.KeygenTimeout); err != nil {
		return err
	}

	return nil
}

func validateThreshold(i interface{}) error {
	threshold, ok := i.(utils.Threshold)
	if !ok {
		return fmt.Errorf("invalid parameter type for threshold: %T", i)
	}

	if err := threshold.Validate(); err != nil {
		return err
	}

	return nil
}

func validateKeygenTimeout(i interface{}) error {
	keygenTimeout, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type for keygen timeout: %T", i)
	}

	if keygenTimeout <= 0 {
		return fmt.Errorf("keygen timeout must be >0")
	}

	return nil
}
