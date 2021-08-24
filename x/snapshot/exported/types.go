package exported

import (
	"bytes"
	"encoding/json"
	"strings"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/gogo/protobuf/proto"
)

//go:generate moq -out ./mock/types.go -pkg mock . SDKValidator Snapshotter Slasher Tss

// SDKValidator is an interface for a Cosmos validator account
type SDKValidator interface {
	proto.Message
	codectypes.UnpackInterfacesMessage
	GetOperator() sdk.ValAddress
	GetConsAddr() (sdk.ConsAddress, error)
	GetConsensusPower() int64
	IsJailed() bool
}

// NewValidator is the constructor for Validator
func NewValidator(validator SDKValidator, shareCount int64) Validator {
	// Pack the validator into an Any
	validatorAny, err := codectypes.NewAnyWithValue(validator)
	if err != nil {
		panic(err)
	}
	return Validator{SDKValidator: validatorAny, ShareCount: shareCount}
}

func (v ValidatorIllegibility) String() string {
	switch v {
	case Tombstoned:
		return "tombstoned"
	case Jailed:
		return "jailed"
	case MissedTooManyBlocks:
		return "missed-too-many-blocks"
	case NoProxyRegistered:
		return "no-proxy-registered"
	case TssSuspended:
		return "tss-suspended"
	default:
		return "unspecified"
	}
}

// IllegibilitiesToString returns a comma-separated string of the list of illegibilities
func IllegibilitiesToString(illegibilities []ValidatorIllegibility) string {
	var illegibilityStrs []string

	for _, illegibility := range illegibilities {
		illegibilityStrs = append(illegibilityStrs, illegibility.String())
	}

	return strings.Join(illegibilityStrs, ",")
}

// GetIllegibilitiesForNewKey returns all the illegibilities for the given validator to control a new key
func (v ValidatorInfo) GetIllegibilitiesForNewKey() []ValidatorIllegibility {
	var illegibilities []ValidatorIllegibility

	if v.Tombstoned {
		illegibilities = append(illegibilities, Tombstoned)
	}

	if v.MissedTooManyBlocks {
		illegibilities = append(illegibilities, MissedTooManyBlocks)
	}

	if v.Jailed {
		illegibilities = append(illegibilities, Jailed)
	}

	if !v.HasProxyRegistered {
		illegibilities = append(illegibilities, NoProxyRegistered)
	}

	if v.TssSuspended {
		illegibilities = append(illegibilities, TssSuspended)
	}

	return illegibilities
}

// GetIllegibilitiesForSigning returns all the illegibilities for the given validator to participate in signing
func (v ValidatorInfo) GetIllegibilitiesForSigning() []ValidatorIllegibility {
	var illegibilities []ValidatorIllegibility

	if v.Tombstoned {
		illegibilities = append(illegibilities, Tombstoned)
	}

	if v.MissedTooManyBlocks {
		illegibilities = append(illegibilities, MissedTooManyBlocks)
	}

	if v.Jailed {
		illegibilities = append(illegibilities, Jailed)
	}

	if v.TssSuspended {
		illegibilities = append(illegibilities, TssSuspended)
	}

	return illegibilities
}

// Slasher provides functionality to manage slashing info for a validator
type Slasher interface {
	GetValidatorSigningInfo(ctx sdk.Context, address sdk.ConsAddress) (info slashingtypes.ValidatorSigningInfo, found bool)
	SignedBlocksWindow(ctx sdk.Context) (res int64)
}

// Tss provides functionality to tss module
type Tss interface {
	GetTssSuspendedUntil(ctx sdk.Context, validator sdk.ValAddress) int64
	GetNextKey(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool)
	OperatorIsAvailableForCounter(ctx sdk.Context, counter int64, validator sdk.ValAddress) bool
	GetMaxMissedBlocksPerWindow(ctx sdk.Context) utils.Threshold
	GetKeyRequirement(ctx sdk.Context, keyRole tss.KeyRole) (tss.KeyRequirement, bool)
}

// GetValidator returns the validator for a given address, if it is part of the snapshot
func (m Snapshot) GetValidator(address sdk.ValAddress) (Validator, bool) {
	for _, validator := range m.Validators {
		if bytes.Equal(validator.GetSDKValidator().GetOperator(), address) {
			return validator, true
		}
	}

	return Validator{}, false
}

// Snapshotter represents the interface for the snapshot module's functionality
type Snapshotter interface {
	GetLatestSnapshot(ctx sdk.Context) (Snapshot, bool)
	GetLatestCounter(ctx sdk.Context) int64
	GetSnapshot(ctx sdk.Context, seqNo int64) (Snapshot, bool)
	TakeSnapshot(ctx sdk.Context, keyRequirement tss.KeyRequirement) (Snapshot, error)
	GetOperator(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress
	GetProxy(ctx sdk.Context, principal sdk.ValAddress) (addr sdk.AccAddress, active bool)
	GetValidatorInfo(ctx sdk.Context, validator SDKValidator) (ValidatorInfo, error)
}

// GetSDKValidator returns the SdkValidator
func (m Validator) GetSDKValidator() SDKValidator {
	if m.SDKValidator == nil {
		panic("SDLValidator cannot be nil")
	}

	return m.SDKValidator.GetCachedValue().(SDKValidator)
}

// UnpackInterfaces implements UnpackInterfacesMessage
func (m Validator) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	if m.SDKValidator != nil {
		var sdkValidator SDKValidator
		return unpacker.UnpackAny(m.SDKValidator, &sdkValidator)
	}
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage
func (m Snapshot) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for i := range m.Validators {
		if err := m.Validators[i].UnpackInterfaces(unpacker); err != nil {
			return err
		}
	}
	return nil
}

// GetSuccinctJSON marshals the snapshot as JSON without including the SDKValidator data
func (m Snapshot) GetSuccinctJSON() ([]byte, error) {
	validators := make([]validator, len(m.Validators))

	for i, val := range m.Validators {
		validators[i].ShareCount = val.ShareCount
		validators[i].Validator = val.GetSDKValidator().GetOperator().String()
	}

	distPolicyStr := strings.ToLower(strings.TrimPrefix(
		m.KeyShareDistributionPolicy.String(), "KEY_SHARE_DISTRIBUTION_POLICY_"))
	timestampStr := m.Timestamp.Format("2 Jan 2006 15:04:05 MST")

	s := struct {
		Validators []validator `json:"validators"`

		Timestamp                  string `json:"timestamp"`
		KeyShareDistributionPolicy string `json:"key_share_distribution_policy"`

		Height          int64 `json:"height"`
		TotalShareCount int64 `json:"total_share_count"`
		Counter         int64 `json:"counter"`
	}{
		Validators: validators,

		Timestamp:                  timestampStr,
		KeyShareDistributionPolicy: distPolicyStr,

		Height:          m.Height,
		TotalShareCount: m.TotalShareCount.Int64(),
		Counter:         m.Counter,
	}

	buff := bytes.NewBuffer([]byte{})
	enc := json.NewEncoder(buff)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	err := enc.Encode(s)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

type validator struct {
	Validator  string `json:"validator"`
	ShareCount int64  `json:"share_count"`
}
