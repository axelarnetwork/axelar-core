package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/params/subspace"
)

// DefaultParamspace - default parameter namespace
const (
	DefaultParamspace = ModuleName
)

var (
	// KeyLockingPeriod defines the key for the locking period
	KeyLockingPeriod = []byte("lockingPeriod")
)

// KeyTable returns a subspace.KeyTable that has registered all parameter types in this module's parameter set
func KeyTable() subspace.KeyTable {
	return subspace.NewKeyTable().RegisterParamSet(&Params{})
}

// Params is the parameter set for this module
type Params struct {
	LockingPeriod int64
}

// DefaultParams returns the module's parameter set initialized with default values
func DefaultParams() Params {
	return Params{
		LockingPeriod: 1,
	}
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of tss module's parameters
func (p *Params) ParamSetPairs() subspace.ParamSetPairs {
	/*
		because the subspace package makes liberal use of pointers to set and get values from the store,
		this method needs to have a pointer receiver AND NewParamSetPair needs to receive the
		parameter values as pointer arguments, otherwise either the internal type reflection panics or the value will not be
		set on the correct Params data struct
	*/
	return subspace.ParamSetPairs{
		subspace.NewParamSetPair(KeyLockingPeriod, &p.LockingPeriod, validateLockingPeriod),
	}
}

func validateLockingPeriod(period interface{}) error {
	val, ok := period.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type for locking period: %T", period)
	}
	if val < 0 {
		return fmt.Errorf("locking period must be a positive integer")
	}
	return nil
}

// Validate checks the validity of the values of the parameter set
func (p Params) Validate() error {
	if err := validateLockingPeriod(p.LockingPeriod); err != nil {
		return err
	}
	return nil
}
