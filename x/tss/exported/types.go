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
	case MasterKey.SimpleString():
		return MasterKey, nil
	case SecondaryKey.SimpleString():
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

// SimpleString converts the KeyRole to a string
func (x KeyRole) SimpleString() string {
	switch x {
	case MasterKey:
		return "master"
	case SecondaryKey:
		return "secondary"
	default:
		return "unknown"
	}

type KeyShareDistributionPolicy int

const (
	WeightedByStake KeyShareDistributionPolicy = iota
	OnePerValidator
)

// KeyShareDistributionPolicyFromStr creates a KeyShareDistributionPolicy from string
func KeyShareDistributionPolicyFromStr(str string) (KeyShareDistributionPolicy, error) {
	switch strings.ToLower(str) {
	case WeightedByStake.String():
		return WeightedByStake, nil
	case OnePerValidator.String():
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
		return fmt.Errorf("invalid key role %d", r)
	}
}

// String converts the KeyShareDistributionPolicy to a string
func (r KeyShareDistributionPolicy) String() string {
	return [...]string{"weighted-by-stake", "one-per-validator"}[r]
}

// KeyRequirement defines requirements for keys
type KeyRequirement struct {
	ChainName                  string
	KeyRole                    KeyRole
	MinValidatorSubsetSize     int64
	KeyShareDistributionPolicy KeyShareDistributionPolicy
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

	if err := r.KeyShareDistributionPolicy.Validate(); err != nil {
		return err
	}

	return nil
}
