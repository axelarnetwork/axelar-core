package exported

import (
	"crypto/ecdsa"
	"fmt"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
)

var _ codectypes.UnpackInterfacesMessage = SignInfo{}

// UnpackInterfaces implements UnpackInterfacesMessage
func (m SignInfo) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var data codec.ProtoMarshaler
	return unpacker.UnpackAny(m.ModuleMetadata, &data)
}

// Handler defines a function that handles a signature after it has
// been generated and voted on
type Handler func(ctx sdk.Context, info SignInfo) error

// key id length range bounds dictated by tofnd
const (
	KeyIDLengthMin = 4
	KeyIDLengthMax = 256
)

// Validate validates the given Signature
func (m Signature) Validate() error {
	if err := utils.ValidateString(m.SigID); err != nil {
		return sdkerrors.Wrap(err, "invalid signature ID")
	}

	if m.SigStatus == SigStatus_Unspecified {
		return fmt.Errorf("sig status must be set")
	}

	if sig := m.GetSingleSig(); sig != nil {
		if err := sig.SigKeyPair.Validate(); err != nil {
			return err
		}
	}

	if sig := m.GetMultiSig(); sig != nil {
		for _, sigKeyPair := range sig.SigKeyPairs {
			if err := sigKeyPair.Validate(); err != nil {
				return err
			}
		}
	}

	return nil
}

// Validate returns an error if the key is not valid; nil otherwise
func (m Key) Validate() error {
	if err := m.ID.Validate(); err != nil {
		return err
	}

	if err := m.Role.Validate(); err != nil {
		return err
	}

	if err := m.Type.Validate(); err != nil {
		return err
	}

	if pub := m.GetECDSAKey(); pub != nil {
		if _, err := pub.GetPubKey(); err != nil {
			return fmt.Errorf("invalid pub key")
		}
	}

	if pubkeys := m.GetMultisigKey(); pubkeys != nil {
		if pubkeys.GetThreshold() <= 0 {
			return fmt.Errorf("invalid threshold")
		}

		pubs, err := pubkeys.GetPubKey()
		if err != nil {
			return fmt.Errorf("invalid multisig pub key")
		}

		if int64(len(pubs)) < pubkeys.GetThreshold() {
			return fmt.Errorf("invalid number of multisig pub keys")
		}
	}

	if m.GetECDSAKey() == nil && m.GetMultisigKey() == nil {
		return fmt.Errorf("pubkey cannot be nil")
	}

	if m.RotationCount < 0 {
		return fmt.Errorf("rotation count must be >=0")
	}

	if err := utils.ValidateString(m.Chain); err != nil {
		return sdkerrors.Wrap(err, "invalid chain")
	}

	if m.SnapshotCounter < 0 {
		return fmt.Errorf("snapshot counter must be >=0")
	}

	return nil
}

// KeyID ensures a correctly formatted tss key ID
type KeyID string

// Validate returns an error, if the key ID is too short or too long
func (id KeyID) Validate() error {
	if err := utils.ValidateString(string(id)); err != nil {
		return sdkerrors.Wrap(err, "invalid key id")
	}

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

	if m.MinKeygenThreshold.Validate() != nil {
		return fmt.Errorf("MinKeygenThreshold must be <=1 and >0")
	}

	if m.SafetyThreshold.Validate() != nil {
		return fmt.Errorf("SafetyThreshold must be <=1 and >0")
	}

	if err := m.KeyShareDistributionPolicy.Validate(); err != nil {
		return err
	}

	if m.KeygenVotingThreshold.Validate() != nil {
		return fmt.Errorf("KeygenVotingThreshold must be <=1 and >0")
	}

	if m.SignVotingThreshold.Validate() != nil {
		return fmt.Errorf("SignVotingThreshold must be <=1 and >0")
	}

	if m.KeygenTimeout <= 0 {
		return fmt.Errorf("KeygenTimeout must be >0")
	}

	if m.SignTimeout <= 0 {
		return fmt.Errorf("SignTimeout must be >0")
	}

	if m.MinTotalShareCount <= 0 || m.MinTotalShareCount > m.MaxTotalShareCount {
		return fmt.Errorf("must satisfy 0 < MinTotalShareCount <= MaxTotalShareCount")
	}

	for totalShareCount := m.MinTotalShareCount; totalShareCount <= m.MaxTotalShareCount; totalShareCount++ {
		corruptionThreshold := ComputeAbsCorruptionThreshold(m.SafetyThreshold, sdk.NewInt(totalShareCount))
		if corruptionThreshold < 0 || corruptionThreshold >= totalShareCount {
			return fmt.Errorf("invalid safety threshold [%s], and corruption threshold [%d] when total share count is [%d]",
				m.SafetyThreshold.String(),
				corruptionThreshold,
				totalShareCount,
			)
		}

		minSigningThreshold := utils.NewThreshold(corruptionThreshold+1, totalShareCount)

		if m.KeygenVotingThreshold.LT(minSigningThreshold) {
			return fmt.Errorf("invalid keygen voting threshold [%s], safety threshold [%s], and corruption threshold [%d] when total share count is [%d]",
				m.KeygenVotingThreshold.String(),
				m.SafetyThreshold.String(),
				corruptionThreshold,
				totalShareCount,
			)
		}

		if m.SignVotingThreshold.GT(minSigningThreshold) {
			return fmt.Errorf("invalid sign voting threshold [%s], safety threshold [%s] and corruption threshold [%d] when total share count is [%d]",
				m.SignVotingThreshold.String(),
				m.SafetyThreshold.String(),
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

// GetPubKey returns the ECDSA public Key
func (m *Key_ECDSAKey) GetPubKey() (*ecdsa.PublicKey, error) {
	pk, err := btcec.ParsePubKey(m.Value, btcec.S256())
	if err != nil {
		return nil, err
	}

	return pk.ToECDSA(), nil
}

// GetPubKey returns the ECDSA public Key
func (m *Key_MultisigKey) GetPubKey() ([]*ecdsa.PublicKey, error) {
	var pks []*ecdsa.PublicKey
	for _, v := range m.Values {
		pk, err := btcec.ParsePubKey(v, btcec.S256())
		if err != nil {
			return nil, err
		}
		pks = append(pks, pk.ToECDSA())
	}

	return pks, nil
}

// GetECDSAPubKey returns public key for ECDSAKey
func (m *Key) GetECDSAPubKey() (ecdsa.PublicKey, error) {
	key := m.GetECDSAKey()
	if key == nil {
		return ecdsa.PublicKey{}, fmt.Errorf("unexpected key type %T", m.PublicKey)
	}

	pk, err := btcec.ParsePubKey(key.Value, btcec.S256())
	if err != nil {
		return ecdsa.PublicKey{}, err
	}

	return *pk.ToECDSA(), nil
}

// GetMultisigPubKey returns public keys for MultisigKey
func (m *Key) GetMultisigPubKey() ([]ecdsa.PublicKey, error) {
	key := m.GetMultisigKey()
	if key == nil {
		return nil, fmt.Errorf("unexpected key type %T", m.PublicKey)
	}

	var pks []ecdsa.PublicKey
	for _, v := range key.Values {
		pk, err := btcec.ParsePubKey(v, btcec.S256())
		if err != nil {
			return nil, err
		}
		pks = append(pks, *pk.ToECDSA())
	}

	return pks, nil
}
