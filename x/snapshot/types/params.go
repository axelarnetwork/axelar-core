package types

import (
	"fmt"
	"time"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
)

// DefaultParamspace - default parameter namespace
const (
	DefaultParamspace = ModuleName
)

var (
	// KeyLockingPeriod is the key for the locking period
	KeyLockingPeriod = []byte("locking")
)

// KeyTable retrieves a subspace table for the module
func KeyTable() subspace.KeyTable {
	return subspace.NewKeyTable().RegisterParamSet(&Params{})
}

// Params represents the genesis parameter set for the module
type Params struct {
	LockingPeriod time.Duration
}

// DefaultParams - the module's default parameters
func DefaultParams() Params {
	return Params{
		LockingPeriod: 1 * time.Nanosecond,
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
	}
}

func validateLockingPeriod(period interface{}) error {
	lock, ok := period.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type for locking period: %T", lock)
	}
	if lock < 0 {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "locking period must be greater than 0")
	}
	return nil
}

// Validate performs a validation check on the parameters
func (p Params) Validate() error {
	if err := validateLockingPeriod(p.LockingPeriod); err != nil {
		return err
	}
	return nil
}
