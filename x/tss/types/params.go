package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	params "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/axelarnetwork/axelar-core/utils"
	bitcoin "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	ethereum "github.com/axelarnetwork/axelar-core/x/ethereum/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// DefaultParamspace - default parameter namespace
const (
	DefaultParamspace = ModuleName
)

// Parameter keys
var (
	KeyLockingPeriod           = []byte("lockingPeriod")
	KeyMinKeygenThreshold      = []byte("minKeygenThreshold")
	KeyCorruptionThreshold     = []byte("corruptionThreshold")
	KeyKeyRequirements         = []byte("keyRequirements")
	KeyMinBondFractionPerShare = []byte("MinBondFractionPerShare")
)

// KeyTable returns a subspace.KeyTable that has registered all parameter types in this module's parameter set
func KeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns the module's parameter set initialized with default values
func DefaultParams() Params {
	return Params{
		LockingPeriod: 0,
		// Set MinKeygenThreshold >= CorruptionThreshold
		MinKeygenThreshold:  utils.Threshold{Numerator: 9, Denominator: 10},
		CorruptionThreshold: utils.Threshold{Numerator: 2, Denominator: 3},
		KeyRequirements: []exported.KeyRequirement{
			{ChainName: bitcoin.Bitcoin.Name, KeyRole: exported.MasterKey, MinValidatorSubsetSize: 5, KeyShareDistributionPolicy: exported.WeightedByStake},
			{ChainName: bitcoin.Bitcoin.Name, KeyRole: exported.SecondaryKey, MinValidatorSubsetSize: 3, KeyShareDistributionPolicy: exported.OnePerValidator},
			{ChainName: ethereum.Ethereum.Name, KeyRole: exported.MasterKey, MinValidatorSubsetSize: 5, KeyShareDistributionPolicy: exported.WeightedByStake},
		},
		MinBondFractionPerShare: utils.Threshold{Numerator: 1, Denominator: 200},
	}
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of tss module's parameters
func (m *Params) ParamSetPairs() params.ParamSetPairs {
	/*
		because the subspace package makes liberal use of pointers to set and get values from the store,
		this method needs to have a pointer receiver AND NewParamSetPair needs to receive the
		parameter values as pointer arguments, otherwise either the internal type reflection panics or the value will not be
		set on the correct Params data struct
	*/
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyLockingPeriod, &m.LockingPeriod, validateLockingPeriod),
		params.NewParamSetPair(KeyMinKeygenThreshold, &m.MinKeygenThreshold, validateThreshold),
		params.NewParamSetPair(KeyCorruptionThreshold, &m.CorruptionThreshold, validateThreshold),
		params.NewParamSetPair(KeyKeyRequirements, &m.KeyRequirements, validateKeyRequirements),
		params.NewParamSetPair(KeyMinBondFractionPerShare, &m.MinBondFractionPerShare, validateMinBondFractionPerShare),
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
func (m Params) Validate() error {
	if err := validateLockingPeriod(m.LockingPeriod); err != nil {
		return err
	}

	if err := validateThreshold(m.MinKeygenThreshold); err != nil {
		return err
	}

	if err := validateThreshold(m.CorruptionThreshold); err != nil {
		return err
	}

	if err := validateTssThresholds(m.MinKeygenThreshold, m.CorruptionThreshold); err != nil {
		return err
	}

	if err := validateKeyRequirements(m.KeyRequirements); err != nil {
		return err
	}

	if err := validateMinBondFractionPerShare(m.MinBondFractionPerShare); err != nil {
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

func validateKeyRequirements(keyRequirements interface{}) error {
	val, ok := keyRequirements.([]exported.KeyRequirement)
	if !ok {
		return fmt.Errorf("invalid parameter type for keyRequirements: %T", keyRequirements)
	}

	for _, keyRequirement := range val {
		if err := keyRequirement.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func validateMinBondFractionPerShare(MinBondFractionPerShare interface{}) error {
	val, ok := MinBondFractionPerShare.(utils.Threshold)
	if !ok {
		return fmt.Errorf("invalid parameter type for MinBondFractionPerShare: %T", MinBondFractionPerShare)
	}

	if val.Numerator <= 0 {
		return fmt.Errorf("threshold numerator must be a positive integer for MinBondFractionPerShare")
	}

	if val.Denominator <= 0 {
		return fmt.Errorf("threshold denominator must be a positive integer for MinBondFractionPerShare")
	}

	if val.Numerator >= val.Denominator {
		return fmt.Errorf("threshold must be <=1 for MinBondFractionPerShare")
	}

	return nil
}
