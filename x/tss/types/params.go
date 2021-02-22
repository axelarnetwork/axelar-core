package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
)

// DefaultParamspace - default parameter namespace
const (
	DefaultParamspace = ModuleName
)

// Parameter keys
var (
	KeyLockingPeriod       = []byte("lockingPeriod")
	KeyMinKeygenThreshold  = []byte("minKeygenThreshold")
	KeyCorruptionThreshold = []byte("corruptionThreshold")
)

// KeyTable returns a subspace.KeyTable that has registered all parameter types in this module's parameter set
func KeyTable() subspace.KeyTable {
	return subspace.NewKeyTable().RegisterParamSet(&Params{})
}

// Params is the parameter set for this module
type Params struct {
	// KeyLockingPeriod defines the key for the locking period
	LockingPeriod int64
	// MinKeygenThreshold defines the minimum % of stake that must be online
	// to authorize generation of a new key in the system.
	MinKeygenThreshold utils.Threshold
	// CorruptionThreshold defines the corruption threshold with which
	// we'll run keygen protocol.
	CorruptionThreshold utils.Threshold
}

// DefaultParams returns the module's parameter set initialized with default values
func DefaultParams() Params {
	return Params{
		LockingPeriod: 0,
		// Set MinKeygenThreshold >= CorruptionThreshold
		MinKeygenThreshold:  utils.Threshold{Numerator: 9, Denominator: 10},
		CorruptionThreshold: utils.Threshold{Numerator: 2, Denominator: 3},
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
		subspace.NewParamSetPair(KeyMinKeygenThreshold, &p.MinKeygenThreshold, validateThreshold),
		subspace.NewParamSetPair(KeyCorruptionThreshold, &p.CorruptionThreshold, validateThreshold),
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
	if err := validateThreshold(p.MinKeygenThreshold); err != nil {
		return err
	}
	if err := validateThreshold(p.CorruptionThreshold); err != nil {
		return err
	}
	if err := validateTssThresholds(p.MinKeygenThreshold, p.CorruptionThreshold); err != nil {
		return err
	}
	return nil
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

// validateTssThresholds checks that minKeygenThreshold >= corruptionThreshold
func validateTssThresholds(minKeygenThreshold interface{}, corruptionThreshold interface{}) error {
	val1, ok1 := minKeygenThreshold.(utils.Threshold)
	val2, ok2 := corruptionThreshold.(utils.Threshold)

	if !ok1 || !ok2 {
		return fmt.Errorf("invalid parameter types for tss thresholds")
	}
	if !val2.IsMet(sdk.NewInt(val1.Numerator), sdk.NewInt(val1.Denominator)) {
		return fmt.Errorf("min keygen threshold must >= corruption threshold")
	}
	return nil
}
