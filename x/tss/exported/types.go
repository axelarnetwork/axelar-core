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
}

// KeyRole is an enum for the role of the key
type KeyRole int

const (
	MasterKey KeyRole = iota
	SecondaryKey
)

// KeyRoleFromStr creates a KeyRole from string
func KeyRoleFromStr(str string) (KeyRole, error) {
	switch strings.ToLower(str) {
	case MasterKey.String():
		return MasterKey, nil
	case SecondaryKey.String():
		return SecondaryKey, nil
	default:
		return -1, fmt.Errorf("invalid key role %s", str)
	}
}

// Validate validates the KeyRole
func (r KeyRole) Validate() error {
	switch r {
	case MasterKey, SecondaryKey:
		return nil
	default:
		return fmt.Errorf("invalid key role %d", r)
	}
}

// String converts the KeyRole to a string
func (r KeyRole) String() string {
	return [...]string{"master", "secondary"}[r]
}

// KeyRequirement defines requirements for keys
type KeyRequirement struct {
	ChainName              string
	KeyRole                KeyRole
	MinValidatorSubsetSize int64
}

// Validate validates the KeyRequirement
func (r KeyRequirement) Validate() error {
	if r.ChainName == "" {
		return fmt.Errorf("invalid ChainName %s", r.ChainName)
	}

	if err := r.KeyRole.Validate(); err != nil {
		return err
	}

	if r.MinValidatorSubsetSize <= 0 {
		return fmt.Errorf("MinValidatorSubsetSize has to be greater than 0 when the key is required")
	}

	return nil
}
