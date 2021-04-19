package exported

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
)

// Signature - an ECDSA signature
type Signature struct {
	R *big.Int
	S *big.Int
}

// Key contains the public key value and corresponding ID
type Key struct {
	ID    string
	Value ecdsa.PublicKey
	Role  KeyRole
}

// GetKeyRoles returns an array of all types of key role
func GetKeyRoles() []KeyRole {
	return []KeyRole{MasterKey, SecondaryKey}
}

// KeyRoleFromStr creates a KeyRole from string
func KeyRoleFromStr(str string) (KeyRole, error) {
	switch strings.ToLower(str) {
	case "master":
		return MasterKey, nil
	case "secondary":
		return SecondaryKey, nil
	default:
		return -1, fmt.Errorf("invalid key role %s", str)
	}
}

// Validate validates the KeyRole
func (x KeyRole) Validate() error {
	switch x {
	case MasterKey, SecondaryKey:
		return nil
	default:
		return fmt.Errorf("invalid key role %d", x)
	}
}

// KeyShareDistributionPolicyFromStr creates a KeyShareDistributionPolicy from string
func KeyShareDistributionPolicyFromStr(str string) (KeyShareDistributionPolicy, error) {
	switch strings.ToLower(str) {
	case "weighted-by-stake":
		return WeightedByStake, nil
	case "one-per-validator":
		return OnePerValidator, nil
	default:
		return -1, fmt.Errorf("invalid key share distribution policy %s", str)
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
	if m.ChainName == "" {
		return fmt.Errorf("invalid ChainName %s", m.ChainName)
	}

	if err := m.KeyRole.Validate(); err != nil {
		return err
	}

	if m.MinValidatorSubsetSize <= 0 {
		return fmt.Errorf("MinValidatorSubsetSize has to be greater than 0 when the key is required")
	}

	if err := m.KeyShareDistributionPolicy.Validate(); err != nil {
		return err
	}

	return nil
}
