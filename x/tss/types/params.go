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
	KeyKeyRequirements                  = []byte("keyRequirements")
	KeySuspendDurationInBlocks          = []byte("SuspendDurationInBlocks")
	KeyHeartbeatPeriodInBlocks          = []byte("HeartbeatPeriodInBlocks")
	KeyMaxMissedBlocksPerWindow         = []byte("MaxMissedBlocksPerWindow")
	KeyUnbondingLockingKeyRotationCount = []byte("UnbondingLockingKeyRotationCount")
	KeyExternalMultisigThreshold        = []byte("externalMultisigThreshold")
	KeyMaxSignQueueSize                 = []byte("MaxSignQueueSize")
	MaxSimultaneousSignShares           = []byte("MaxSimultaneousSignShares")
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
				MinKeygenThreshold:         utils.Threshold{Numerator: 5, Denominator: 6},
				SafetyThreshold:            utils.Threshold{Numerator: 2, Denominator: 3},
				KeyShareDistributionPolicy: exported.WeightedByStake,
				MaxTotalShareCount:         50,
				MinTotalShareCount:         4,
				KeygenVotingThreshold:      utils.Threshold{Numerator: 5, Denominator: 6},
				SignVotingThreshold:        utils.Threshold{Numerator: 2, Denominator: 3},
				KeygenTimeout:              50,
				SignTimeout:                50,
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
				KeygenTimeout:              50,
				SignTimeout:                50,
			},
		},
		SuspendDurationInBlocks:          2000,
		HeartbeatPeriodInBlocks:          50,
		MaxMissedBlocksPerWindow:         utils.Threshold{Numerator: 5, Denominator: 100},
		UnbondingLockingKeyRotationCount: 8,
		ExternalMultisigThreshold:        utils.Threshold{Numerator: 3, Denominator: 6},
		MaxSignQueueSize:                 50,
		MaxSimultaneousSignShares:        26,
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

	if val.Numerator <= 0 {
		return fmt.Errorf("threshold numerator must be a positive integer for MaxMissedBlocksPerWindow")
	}

	if val.Denominator <= 0 {
		return fmt.Errorf("threshold denominator must be a positive integer for MaxMissedBlocksPerWindow")
	}

	if val.Numerator > val.Denominator {
		return fmt.Errorf("threshold must be <=1 for MaxMissedBlocksPerWindow")
	}

	return nil
}

func validateExternalMultisigThreshold(externalMultisigThreshold interface{}) error {
	t, ok := externalMultisigThreshold.(utils.Threshold)
	if !ok {
		return fmt.Errorf("invalid parameter type for external multisig threshold: %T", externalMultisigThreshold)
	}

	if t.Numerator <= 0 {
		return fmt.Errorf("numerator must be greater than 0 for external multisig threshold")
	}

	if t.Denominator <= 0 {
		return fmt.Errorf("denominator must be greater than 0 for external multisig threshold")
	}

	if t.Numerator > t.Denominator {
		return fmt.Errorf("threshold must be <=1 for external multisig threshold")
	}

	return nil
}
