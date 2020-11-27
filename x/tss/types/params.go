package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/params/subspace"

	"github.com/axelarnetwork/axelar-core/utils"
)

// Default parameter namespace
const (
	DefaultParamspace = ModuleName
)

var (
	KeyLockingPeriod = []byte("lockingPeriod")
	KeyThreshold     = []byte("threshold")
)

func KeyTable() subspace.KeyTable {
	return subspace.NewKeyTable().RegisterParamSet(&Params{})
}

type Params struct {
	LockingPeriod int64
	Threshold     utils.Threshold
}

func DefaultParams() Params {
	return Params{
		LockingPeriod: 1,
		Threshold:     utils.Threshold{Numerator: 2, Denominator: 3},
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
		subspace.NewParamSetPair(KeyLockingPeriod, &p.LockingPeriod, validateLockingPeriod),
		subspace.NewParamSetPair(KeyThreshold, &p.Threshold, validateThreshold),
	}
}

func validateThreshold(threshold interface{}) error {
	val, ok := threshold.(utils.Threshold)
	if !ok {
		return fmt.Errorf("invalid parameter type for threshold: %T", threshold)
	}
	if val.Denominator <= 0 {
		return fmt.Errorf("threshold denominator must be a positive integer")
	}

	if val.Numerator < 0 {
		return fmt.Errorf("threshold numerator must be a non-negative integer")
	}

	if val.Numerator >= val.Denominator {
		return fmt.Errorf("threshold must be <1")
	}
	return nil
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

func (p Params) Validate() error {
	if err := validateLockingPeriod(p.LockingPeriod); err != nil {
		return err
	}
	return nil
}
