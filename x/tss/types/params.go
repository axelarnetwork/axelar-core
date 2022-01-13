package types

import (
	"fmt"

	params "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// DefaultParamspace - default parameter namespace
const (
	DefaultParamspace = ModuleName
)

// Parameter keys
var (
	KeyKeyRequirements                  = []byte("KeyRequirements")
	KeySuspendDurationInBlocks          = []byte("SuspendDurationInBlocks")
	KeyHeartbeatPeriodInBlocks          = []byte("HeartbeatPeriodInBlocks")
	KeyMaxMissedBlocksPerWindow         = []byte("MaxMissedBlocksPerWindow")
	KeyUnbondingLockingKeyRotationCount = []byte("UnbondingLockingKeyRotationCount")
	KeyExternalMultisigThreshold        = []byte("ExternalMultisigThreshold")
	KeyMaxSignQueueSize                 = []byte("MaxSignQueueSize")
	MaxSimultaneousSignShares           = []byte("MaxSimultaneousSignShares")
	KeyTssSignedBlocksWindow            = []byte("TssSignedBlocksWindow")
)

// KeyTable returns a subspace.KeyTable that has registered all parameter types in this module's parameter set
func KeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns the module's parameter set initialized with default values
func DefaultParams() Params {
	return Params{
		KeyRequirements: []exported.KeyRequirement{
			{
				KeyRole:                    exported.MasterKey,
				KeyType:                    exported.Threshold,
				MinKeygenThreshold:         utils.Threshold{Numerator: 5, Denominator: 6},
				SafetyThreshold:            utils.Threshold{Numerator: 2, Denominator: 3},
				KeyShareDistributionPolicy: exported.WeightedByStake,
				MaxTotalShareCount:         50,
				MinTotalShareCount:         4,
				KeygenVotingThreshold:      utils.Threshold{Numerator: 5, Denominator: 6},
				SignVotingThreshold:        utils.Threshold{Numerator: 2, Denominator: 3},
				KeygenTimeout:              250,
				SignTimeout:                250,
			},
			{
				KeyRole:                    exported.SecondaryKey,
				KeyType:                    exported.Threshold,
				MinKeygenThreshold:         utils.Threshold{Numerator: 15, Denominator: 20},
				SafetyThreshold:            utils.Threshold{Numerator: 11, Denominator: 20},
				KeyShareDistributionPolicy: exported.OnePerValidator,
				MaxTotalShareCount:         20,
				MinTotalShareCount:         4,
				KeygenVotingThreshold:      utils.Threshold{Numerator: 15, Denominator: 20},
				SignVotingThreshold:        utils.Threshold{Numerator: 11, Denominator: 20},
				KeygenTimeout:              150,
				SignTimeout:                150,
			},
			{
				KeyRole:                    exported.MasterKey,
				KeyType:                    exported.Multisig,
				MinKeygenThreshold:         utils.Threshold{Numerator: 31, Denominator: 40},
				SafetyThreshold:            utils.Threshold{Numerator: 11, Denominator: 20},
				KeyShareDistributionPolicy: exported.WeightedByStake,
				MaxTotalShareCount:         50,
				MinTotalShareCount:         4,
				KeygenVotingThreshold:      utils.Threshold{Numerator: 31, Denominator: 40},
				SignVotingThreshold:        utils.Threshold{Numerator: 11, Denominator: 20},
				KeygenTimeout:              20,
				SignTimeout:                20,
			},
			{
				KeyRole:                    exported.SecondaryKey,
				KeyType:                    exported.Multisig,
				MinKeygenThreshold:         utils.Threshold{Numerator: 15, Denominator: 20},
				SafetyThreshold:            utils.Threshold{Numerator: 11, Denominator: 20},
				KeyShareDistributionPolicy: exported.OnePerValidator,
				MaxTotalShareCount:         20,
				MinTotalShareCount:         4,
				KeygenVotingThreshold:      utils.Threshold{Numerator: 15, Denominator: 20},
				SignVotingThreshold:        utils.Threshold{Numerator: 11, Denominator: 20},
				KeygenTimeout:              20,
				SignTimeout:                20,
			},
		},
		SuspendDurationInBlocks:          8500,
		HeartbeatPeriodInBlocks:          50,
		MaxMissedBlocksPerWindow:         utils.Threshold{Numerator: 5, Denominator: 100},
		UnbondingLockingKeyRotationCount: 4,
		ExternalMultisigThreshold:        utils.Threshold{Numerator: 4, Denominator: 8},
		MaxSignQueueSize:                 50,
		MaxSimultaneousSignShares:        100,
		TssSignedBlocksWindow:            100,
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
		params.NewParamSetPair(KeyKeyRequirements, &m.KeyRequirements, validateKeyRequirements),
		params.NewParamSetPair(KeySuspendDurationInBlocks, &m.SuspendDurationInBlocks, validateSuspendDurationInBlocks),
		params.NewParamSetPair(KeyHeartbeatPeriodInBlocks, &m.HeartbeatPeriodInBlocks, validatePosInt64("HeartbeatPeriodInBlocks")),
		params.NewParamSetPair(KeyMaxMissedBlocksPerWindow, &m.MaxMissedBlocksPerWindow, validateMaxMissedBlocksPerWindow),
		params.NewParamSetPair(KeyUnbondingLockingKeyRotationCount, &m.UnbondingLockingKeyRotationCount, validatePosInt64("UnbondingLockingKeyRotationCount")),
		params.NewParamSetPair(KeyExternalMultisigThreshold, &m.ExternalMultisigThreshold, validateExternalMultisigThreshold),
		params.NewParamSetPair(KeyMaxSignQueueSize, &m.MaxSignQueueSize, validatePosInt64("MaxSignQueueSize")),
		params.NewParamSetPair(MaxSimultaneousSignShares, &m.MaxSimultaneousSignShares, validatePosInt64("MaxSimultaneousSignShares")),
		params.NewParamSetPair(KeyTssSignedBlocksWindow, &m.TssSignedBlocksWindow, validatePosInt64("TssSignedBlocksWindow")),
	}
}

// Validate checks the validity of the values of the parameter set
func (m Params) Validate() error {
	if err := validateKeyRequirements(m.KeyRequirements); err != nil {
		return err
	}

	if err := validateSuspendDurationInBlocks(m.SuspendDurationInBlocks); err != nil {
		return err
	}

	if err := validatePosInt64("HeartbeatPeriodInBlocks")(m.HeartbeatPeriodInBlocks); err != nil {
		return err
	}

	if err := validateMaxMissedBlocksPerWindow(m.MaxMissedBlocksPerWindow); err != nil {
		return err
	}

	if err := validatePosInt64("UnbondingLockingKeyRotationCount")(m.UnbondingLockingKeyRotationCount); err != nil {
		return err
	}

	if err := validateExternalMultisigThreshold(m.ExternalMultisigThreshold); err != nil {
		return err
	}

	if err := validatePosInt64("MaxSignQueueSize")(m.MaxSignQueueSize); err != nil {
		return err
	}

	if err := validatePosInt64("MaxSimultaneousSignShares")(m.MaxSimultaneousSignShares); err != nil {
		return err
	}

	if err := validatePosInt64("TssSignedBlocksWindow")(m.TssSignedBlocksWindow); err != nil {
		return err
	}

	return nil
}

func validateKeyRequirements(keyRequirements interface{}) error {
	val, ok := keyRequirements.([]exported.KeyRequirement)
	if !ok {
		return fmt.Errorf("invalid parameter type for keyRequirements: %T", keyRequirements)
	}

	keyRoleSeen := map[string]bool{}
	for _, keyRequirement := range val {
		key := fmt.Sprintf("%s_%s", keyRequirement.KeyRole.SimpleString(), keyRequirement.KeyType.SimpleString())
		if keyRoleSeen[key] {
			return fmt.Errorf("duplicate key role and key type found in KeyRequirements")
		}

		if err := keyRequirement.Validate(); err != nil {
			return err
		}

		keyRoleSeen[key] = true
	}

	return nil
}

func validateSuspendDurationInBlocks(suspendDurationInBlocks interface{}) error {
	val, ok := suspendDurationInBlocks.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type for SuspendDurationInBlocks: %T", suspendDurationInBlocks)
	}

	if val <= 0 {
		return fmt.Errorf("SuspendDurationInBlocks must be a positive integer")
	}

	return nil
}

func validatePosInt64(field string) func(value interface{}) error {
	return func(value interface{}) error {
		val, ok := value.(int64)
		if !ok {
			return fmt.Errorf("invalid parameter type for %s: %T", field, value)
		}

		if val <= 0 {
			return fmt.Errorf("%s must be a positive integer", field)
		}

		return nil
	}
}

func validateMaxMissedBlocksPerWindow(maxMissedBlocksPerWindow interface{}) error {
	val, ok := maxMissedBlocksPerWindow.(utils.Threshold)
	if !ok {
		return fmt.Errorf("invalid parameter type for MaxMissedBlocksPerWindow: %T", maxMissedBlocksPerWindow)
	}

	if val.Validate() != nil {
		return fmt.Errorf("MaxMissedBlocksPerWindow threshold must be >0 and <=1")
	}

	return nil
}

func validateExternalMultisigThreshold(externalMultisigThreshold interface{}) error {
	t, ok := externalMultisigThreshold.(utils.Threshold)
	if !ok {
		return fmt.Errorf("invalid parameter type for external multisig threshold: %T", externalMultisigThreshold)
	}

	if t.Validate() != nil {
		return fmt.Errorf("ExternalMultisigThreshold must be >0 and <=1")
	}

	return nil
}
