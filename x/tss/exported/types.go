package exported

import (
	"crypto/ecdsa"
	"fmt"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
)

// Handler defines a function that handles a signature after it has
// been generated and voted on
type Handler func(ctx sdk.Context, info SignInfo) error

// Key contains the public key value and corresponding ID
type Key struct {
	ID        KeyID
	Value     ecdsa.PublicKey
	Role      KeyRole
	RotatedAt *time.Time
}

// key id length range bounds dictated by tofnd
const (
	KeyIDLengthMin = 4
	KeyIDLengthMax = 256
)

// KeyID ensures a correctly formatted tss key ID
type KeyID string

// Validate returns an error, if the key ID is too short or too long
func (id KeyID) Validate() error {
	if len(id) < KeyIDLengthMin || len(id) > KeyIDLengthMax {
		return fmt.Errorf("key id length %d not in range [%d,%d]", len(id), KeyIDLengthMin, KeyIDLengthMax)
	}

	return nil
}

// KeyIDsToStrings converts a slice of type KeyID to a slice of strings
func KeyIDsToStrings(keyIDs []KeyID) []string {
	if keyIDs == nil {
		return nil
	}
	strs := make([]string, 0, len(keyIDs))
	for _, id := range keyIDs {
		strs = append(strs, string(id))
	}
	return strs
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

	if err := m.KeyType.Validate(); err != nil {
		return err
	}

	if m.MinKeygenThreshold.Validate() != nil || m.MinKeygenThreshold.GT(utils.OneThreshold) || m.MinKeygenThreshold.LT(utils.ZeroThreshold) {
		return fmt.Errorf("MinKeygenThreshold must be <=1 and >0")
	}

	if m.SafetyThreshold.Validate() != nil || m.SafetyThreshold.GT(utils.OneThreshold) || m.SafetyThreshold.LT(utils.ZeroThreshold) {
		return fmt.Errorf("SafetyThreshold must be <=1 and >0")
	}

	if err := m.KeyShareDistributionPolicy.Validate(); err != nil {
		return err
	}

	if m.KeygenVotingThreshold.Validate() != nil || m.KeygenVotingThreshold.GT(utils.OneThreshold) || m.KeygenVotingThreshold.LT(utils.ZeroThreshold) {
		return fmt.Errorf("KeygenVotingThreshold must be <=1 and >0")
	}

	if m.SignVotingThreshold.Validate() != nil || m.SignVotingThreshold.GT(utils.OneThreshold) || m.SignVotingThreshold.LT(utils.ZeroThreshold) {
		return fmt.Errorf("SignVotingThreshold must be <=1 and >0")
	}

	if m.KeygenTimeout <= 0 {
		return fmt.Errorf("KeygenTimeout must be >0")
	}

	if m.SignTimeout <= 0 {
		return fmt.Errorf("SignTimeout must be >0")
	}

	for totalShareCount := m.MinTotalShareCount; totalShareCount <= m.MaxTotalShareCount; totalShareCount++ {
		corruptionThreshold := ComputeAbsCorruptionThreshold(m.SafetyThreshold, sdk.NewInt(totalShareCount))
		if corruptionThreshold < 0 || corruptionThreshold >= totalShareCount {
			return fmt.Errorf("invalid safety threshold [%s], and corruption threshold [%d] when total share count is [%d]",
				m.SafetyThreshold.SimpleString(),
				corruptionThreshold,
				totalShareCount,
			)
		}

		minSigningThreshold := utils.NewThreshold(corruptionThreshold+1, totalShareCount)

		if m.KeygenVotingThreshold.LT(minSigningThreshold) {
			return fmt.Errorf("invalid keygen voting threshold [%s], safety threshold [%s], and corruption threshold [%d] when total share count is [%d]",
				m.KeygenVotingThreshold.SimpleString(),
				m.SafetyThreshold.SimpleString(),
				corruptionThreshold,
				totalShareCount,
			)
		}

		if m.SignVotingThreshold.GT(minSigningThreshold) {
			return fmt.Errorf("invalid sign voting threshold [%s], safety threshold [%s] and corruption threshold [%d] when total share count is [%d]",
				m.SignVotingThreshold.SimpleString(),
				m.SafetyThreshold.SimpleString(),
				corruptionThreshold,
				totalShareCount,
			)
		}
	}

	return nil
}

// ComputeAbsCorruptionThreshold returns absolute corruption threshold to be used by tss.
// (threshold + 1) shares are required to sign
func ComputeAbsCorruptionThreshold(safetyThreshold utils.Threshold, totalShareCount sdk.Int) int64 {
	return sdk.NewDec(totalShareCount.Int64()).MulInt64(safetyThreshold.Numerator).QuoInt64(safetyThreshold.Denominator).Ceil().TruncateInt().Int64() - 1
}

// MultisigKey contains the public key value and corresponding ID
type MultisigKey struct {
	ID        KeyID
	Values    []ecdsa.PublicKey
	Role      KeyRole
	RotatedAt *time.Time
}

// KeyTypeFromSimpleStr creates a KeyType from string
func KeyTypeFromSimpleStr(str string) (KeyType, error) {
	switch strings.ToLower(str) {
	case Threshold.SimpleString():
		return Threshold, nil
	case Multisig.SimpleString():
		return Multisig, nil
	case None.SimpleString():
		return None, nil
	default:
		return -1, fmt.Errorf("invalid key type %s", str)
	}
}

// SimpleString returns a human-readable string
func (x KeyType) SimpleString() string {
	switch x {
	case Threshold:
		return "threshold"
	case Multisig:
		return "multisig"
	case None:
		return "none"
	default:
		return "unknown"
	}
}

// Validate validates the KeyType
func (x KeyType) Validate() error {
	switch x {
	case Threshold, Multisig, None:
		return nil
	default:
		return fmt.Errorf("invalid key type %d", x)
	}
}

// Validate validates the SigKeyPair
func (m SigKeyPair) Validate() error {
	_, err := btcec.ParsePubKey(m.PubKey, btcec.S256())
	if err != nil {
		return err
	}

	_, err = btcec.ParseDERSignature(m.Signature, btcec.S256())
	if err != nil {
		return err
	}

	return nil
}

// GetKey returns the public key of the SigKeyPair
func (m SigKeyPair) GetKey() (ecdsa.PublicKey, error) {
	btcecKey, err := btcec.ParsePubKey(m.PubKey, btcec.S256())
	if err != nil {
		return ecdsa.PublicKey{}, err
	}
	pk := btcecKey.ToECDSA()

	return *pk, nil
}

// GetSig returns the signature of the SigKeyPair
func (m SigKeyPair) GetSig() (btcec.Signature, error) {
	sig, err := btcec.ParseDERSignature(m.Signature, btcec.S256())
	if err != nil {
		return btcec.Signature{}, err
	}

	return *sig, nil
}

// GetSignature returns btcec Signature for single sig
func (m *Signature_SingleSig_) GetSignature() (btcec.Signature, error) {
	bz := m.SingleSig.SigKeyPair.Signature
	sig, err := btcec.ParseDERSignature(bz, btcec.S256())
	if err != nil {
		return btcec.Signature{}, err
	}

	return *sig, nil
}

// GetSignature returns list of btcec Signatures for multi sig
func (m *Signature_MultiSig_) GetSignature() ([]btcec.Signature, error) {
	var sigs []btcec.Signature
	pairs := m.MultiSig.SigKeyPairs
	for _, pair := range pairs {
		sig, err := btcec.ParseDERSignature(pair.Signature, btcec.S256())
		if err != nil {
			return sigs, err
		}
		sigs = append(sigs, *sig)

	}

	return sigs, nil
}
