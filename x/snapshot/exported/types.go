package exported

import (
	"bytes"
	"encoding/json"
	"strings"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	"github.com/axelarnetwork/axelar-core/utils"
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

// ValidatorInfo adopts the methods from "github.com/cosmos/cosmos-sdk/x/slashing" that are
// actually used by this module
type ValidatorInfo struct {
	slashingtypes.ValidatorSigningInfo
}

// Slasher provides functionality to manage slashing info for a validator
type Slasher interface {
	GetValidatorSigningInfo(ctx sdk.Context, address sdk.ConsAddress) (info ValidatorInfo, found bool)
}

// Tss provides functionality to tss module
type Tss interface {
	SetKeyRequirement(ctx sdk.Context, keyRequirement tss.KeyRequirement)
	GetMinBondFractionPerShare(ctx sdk.Context) utils.Threshold
	GetTssSuspendedUntil(ctx sdk.Context, validator sdk.ValAddress) int64
	GetNextKey(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool)
}

// IsValidatorEligibleForNewKey returns true if given validator is eligible for handling a new key; otherwise, false
func IsValidatorEligibleForNewKey(ctx sdk.Context, slasher Slasher, snapshotter Snapshotter, tss Tss, validator SDKValidator) bool {
	return IsValidatorActive(ctx, slasher, validator) &&
		HasProxyRegistered(ctx, snapshotter, validator) &&
		!IsValidatorTssSuspended(ctx, tss, validator)
}

// IsValidatorActive returns true if the validator is active; otherwise, false
func IsValidatorActive(ctx sdk.Context, slasher Slasher, validator SDKValidator) bool {
	consAddr, err := validator.GetConsAddr()
	if err != nil {
		return false
	}

	signingInfo, found := slasher.GetValidatorSigningInfo(ctx, consAddr)

	return found && !signingInfo.Tombstoned && signingInfo.MissedBlocksCounter <= 0 && !validator.IsJailed()
}

// HasProxyRegistered returns true if the validator has broadcast proxy registered; otherwise, false
func HasProxyRegistered(ctx sdk.Context, snapshotter Snapshotter, validator SDKValidator) bool {
	_, active := snapshotter.GetProxy(ctx, validator.GetOperator())
	return active
}

// IsValidatorTssSuspended returns true if the validator is suspended from participating TSS ceremonies for committing faulty behaviour; otherwise, false
func IsValidatorTssSuspended(ctx sdk.Context, tss Tss, validator SDKValidator) bool {
	return tss.GetTssSuspendedUntil(ctx, validator.GetOperator()) > ctx.BlockHeight()
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
	GetSnapshot(ctx sdk.Context, counter int64) (Snapshot, bool)
	TakeSnapshot(ctx sdk.Context, subsetSize int64, keyShareDistributionPolicy tss.KeyShareDistributionPolicy) (snapshotConsensusPower sdk.Int, totalConsensusPower sdk.Int, err error)
	GetOperator(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress
	GetProxy(ctx sdk.Context, principal sdk.ValAddress) (addr sdk.AccAddress, active bool)
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

	s := struct {
		Validators []validator `json:"validators"`

		Timestamp                  string `json:"timestamp"`
		KeyShareDistributionPolicy string `json:"key_share_distribution_policy"`

		Height          int64 `json:"height"`
		TotalShareCount int64 `json:"total_share_count"`
		Counter         int64 `json:"counter"`
	}{
		Validators: validators,

		Timestamp:                  m.Timestamp.String(),
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
