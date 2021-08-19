package exported

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/axelarnetwork/axelar-core/utils"
)

// Signature - an ECDSA signature
type Signature struct {
	R *big.Int
	S *big.Int
}

// Key contains the public key value and corresponding ID
type Key struct {
	ID        string
	Value     ecdsa.PublicKey
	Role      KeyRole
	RotatedAt *time.Time
}

// GetKeyRoles returns an array of all types of key role
func GetKeyRoles() []KeyRole {
	return []KeyRole{MasterKey, SecondaryKey, ExternalKey}
}

// KeyRoleFromSimpleStr creates a KeyRole from string
func KeyRoleFromSimpleStr(str string) (KeyRole, error) {
	switch strings.ToLower(str) {
	case MasterKey.SimpleString():
		return MasterKey, nil
	case SecondaryKey.SimpleString():
		return SecondaryKey, nil
	case ExternalKey.SimpleString():
		return ExternalKey, nil
	default:
		return -1, fmt.Errorf("invalid key role %s", str)
	}
}

// SimpleString returns a human-readable string
func (x KeyRole) SimpleString() string {
	switch x {
	case MasterKey:
		return "master"
	case SecondaryKey:
		return "secondary"
	case ExternalKey:
		return "external"
	default:
		return "unknown"
	}
}

// Validate validates the KeyRole
func (x KeyRole) Validate() error {
	switch x {
	case MasterKey, SecondaryKey, ExternalKey:
		return nil
	default:
		return fmt.Errorf("invalid key role %d", x)
	}
}

// KeyShareDistributionPolicyFromSimpleStr creates a KeyShareDistributionPolicy from string
func KeyShareDistributionPolicyFromSimpleStr(str string) (KeyShareDistributionPolicy, error) {
	switch strings.ToLower(str) {
	case WeightedByStake.SimpleString():
		return WeightedByStake, nil
	case OnePerValidator.SimpleString():
		return OnePerValidator, nil
	default:
		return -1, fmt.Errorf("invalid key share distribution policy %s", str)
	}
}

// SimpleString returns a human-readable string
func (r KeyShareDistributionPolicy) SimpleString() string {
	switch r {
	case WeightedByStake:
		return "weighted-by-stake"
	case OnePerValidator:
		return "one-per-validator"
	default:
		return "unknown"
	}
}

// Validate validates the KeyShareDistributionPolicy
func (r KeyShareDistributionPolicy) Validate() error {
	switch r {
	case WeightedByStake, OnePerValidator:
		return nil
	default:
		return fmt.Errorf("invalid key share distribution policy %d", r)
	}
}

// Validate validates the KeyRequirement
func (m KeyRequirement) Validate() error {
	if err := m.KeyRole.Validate(); err != nil {
		return err
	}

	if m.MinKeygenThreshold.Validate() != nil || m.MinKeygenThreshold.GT(utils.Threshold{Numerator: 1, Denominator: 1}) || m.MinKeygenThreshold.LT(utils.Threshold{Numerator: 0, Denominator: 1}) {
		return fmt.Errorf("MinKeygenThreshold must be <=1 and >0")
	}

	if m.SafetyThreshold.Validate() != nil || m.SafetyThreshold.GT(utils.Threshold{Numerator: 1, Denominator: 1}) || m.SafetyThreshold.LT(utils.Threshold{Numerator: 0, Denominator: 1}) {
		return fmt.Errorf("SafetyThreshold must be <=1 and >0")
	}

	if err := m.KeyShareDistributionPolicy.Validate(); err != nil {
		return err
	}

	if m.KeygenVotingThreshold.Validate() != nil || m.KeygenVotingThreshold.GT(utils.Threshold{Numerator: 1, Denominator: 1}) || m.KeygenVotingThreshold.LT(utils.Threshold{Numerator: 0, Denominator: 1}) {
		return fmt.Errorf("KeygenVotingThreshold must be <=1 and >0")
	}

	if m.SignVotingThreshold.Validate() != nil || m.SignVotingThreshold.GT(utils.Threshold{Numerator: 1, Denominator: 1}) || m.SignVotingThreshold.LT(utils.Threshold{Numerator: 0, Denominator: 1}) {
		return fmt.Errorf("SignVotingThreshold must be <=1 and >0")
	}

	if m.KeygenTimeout <= 0 {
		return fmt.Errorf("KeygenTimeout must be >0")
	}

	if m.SignTimeout <= 0 {
		return fmt.Errorf("SignTimeout must be >0")
	}

	return nil
}
