package exported

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/gogo/protobuf/proto"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

//go:generate moq -out ./mock/types.go -pkg mock . SDKValidator Snapshotter Slasher Tss ValidatorI

// ValidatorI provides necessary functions to the validator information
type ValidatorI interface {
}

// NewSnapshot is the constructor of Snapshot
func NewSnapshot(
	validators []Validator,
	timestamp time.Time,
	height int64,
	totalShareCount sdk.Int,
	counter int64,
	keyShareDistributionPolicy tss.KeyShareDistributionPolicy,
	corruptionThreshold int64,
) Snapshot {
	return Snapshot{
		Validators:                 validators,
		Timestamp:                  timestamp,
		Height:                     height,
		TotalShareCount:            totalShareCount,
		Counter:                    counter,
		KeyShareDistributionPolicy: keyShareDistributionPolicy,
		CorruptionThreshold:        corruptionThreshold,
		Participants:               nil,
		BondedWeight:               sdk.ZeroUint(),
	}
}

// ValidateBasic returns an error if the given participant is invalid; nil otherwise
func (m Participant) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Address); err != nil {
		return err
	}

	return nil
}

// ValidateBasic returns an error if the given snapshot is invalid; nil otherwise
func (m Snapshot) ValidateBasic() error {
	if len(m.Participants) == 0 {
		return fmt.Errorf("snapshot cannot have no participant")
	}

	if m.BondedWeight.IsZero() {
		return fmt.Errorf("snapshot cannot have zero total weight")
	}

	if m.Height <= 0 {
		return fmt.Errorf("snapshot must height >0")
	}

	if m.Timestamp.IsZero() {
		return fmt.Errorf("snapshot must timestamp >0")
	}

	sumWeight := sdk.ZeroUint()
	for addr, p := range m.Participants {
		if _, err := sdk.ValAddressFromBech32(addr); err != nil {
			return err
		}

		if err := p.ValidateBasic(); err != nil {
			return err
		}

		sumWeight = sumWeight.Add(p.Weight)
	}

	if sumWeight.GT(m.BondedWeight) {
		return fmt.Errorf("snapshot cannot have sum of participant weight greater than total weight")
	}

	return nil
}

// Validate returns an error if the snapshot is not valid; nil otherwise
func (m Snapshot) Validate() error {
	if len(m.Validators) == 0 {
		return fmt.Errorf("missing validators")
	}

	expectedTotalShareCount := sdk.ZeroInt()
	for _, validator := range m.Validators {
		if err := validator.Validate(); err != nil {
			return err
		}

		expectedTotalShareCount = expectedTotalShareCount.AddRaw(validator.ShareCount)
	}

	if m.Height < 0 {
		return fmt.Errorf("height must be >=0")
	}

	if !m.TotalShareCount.Equal(expectedTotalShareCount) {
		return fmt.Errorf("invalid total share count")
	}

	if m.Counter < 0 {
		return fmt.Errorf("counter must be >=0")
	}

	if m.KeyShareDistributionPolicy == tss.Unspecified {
		return fmt.Errorf("unspecified key distribution policy")
	}

	if m.CorruptionThreshold < 0 || m.CorruptionThreshold >= m.TotalShareCount.Int64() {
		return fmt.Errorf("invalid corruption threshold: %d, total share count: %d", m.CorruptionThreshold, m.TotalShareCount.Int64())
	}

	return nil
}

// Validate returns an error if the validator is not valid; nil otherwise
func (m Validator) Validate() error {
	if m.SDKValidator == nil {
		return fmt.Errorf("missing SDK validator")
	}

	if m.ShareCount <= 0 {
		return fmt.Errorf("share count must be >0")
	}

	return nil
}

// SDKValidator is an interface for a Cosmos validator account
type SDKValidator interface {
	proto.Message
	codectypes.UnpackInterfacesMessage
	GetOperator() sdk.ValAddress
	GetConsAddr() (sdk.ConsAddress, error)
	GetConsensusPower(sdk.Int) int64
	IsJailed() bool
	IsBonded() bool
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

// Is returns true if the illegibility contains the given one; false otherwise
func (v ValidatorIllegibility) Is(illegibility ValidatorIllegibility) bool {
	if illegibility == None {
		return v == None
	}

	return v&illegibility == illegibility
}

// String returns a comma-separated string representation of illegibility
func (v ValidatorIllegibility) String() string {
	var illegibilityStrs []string

	if v.Is(Tombstoned) {
		illegibilityStrs = append(illegibilityStrs, "tombstoned")
	}
	if v.Is(Jailed) {
		illegibilityStrs = append(illegibilityStrs, "jailed")
	}
	if v.Is(MissedTooManyBlocks) {
		illegibilityStrs = append(illegibilityStrs, "missed-too-many-blocks")
	}
	if v.Is(NoProxyRegistered) {
		illegibilityStrs = append(illegibilityStrs, "no-proxy-registered")
	}
	if v.Is(TssSuspended) {
		illegibilityStrs = append(illegibilityStrs, "tss-suspended")
	}
	if v.Is(ProxyInsuficientFunds) {
		illegibilityStrs = append(illegibilityStrs, "proxy-insuficient-funds")
	}

	if len(illegibilityStrs) == 0 {
		illegibilityStrs = append(illegibilityStrs, "none")
	}
	return strings.Join(illegibilityStrs, ",")
}

// FilterIllegibilityForNewKey filters the illegibility to only leave those ones related to handling of new key
func (v ValidatorIllegibility) FilterIllegibilityForNewKey() ValidatorIllegibility {
	mask := None

	for _, illegibility := range GetValidatorIllegibilities() {
		mask |= illegibility
	}

	return v & mask
}

// FilterIllegibilityForTssSigning filters the illegibility to only leave those ones related to handling of signing
func (v ValidatorIllegibility) FilterIllegibilityForTssSigning() ValidatorIllegibility {
	mask := None

	for _, illegibility := range GetValidatorIllegibilities() {
		mask |= illegibility
	}

	return v & mask
}

// FilterIllegibilityForMultisigSigning filters the illegibility to only leave those ones related to handling of signing
// - filter out MissedTooManyBlocks so that even potentially offline validators can submit signature(s)
// - filter out ProxyInsuficientFunds so that validators with proxy account having low balance can submit signature(s)
func (v ValidatorIllegibility) FilterIllegibilityForMultisigSigning() ValidatorIllegibility {
	return v & ^MissedTooManyBlocks & ^ProxyInsuficientFunds
}

// GetValidatorIllegibilities returns all validator illegibilities
func GetValidatorIllegibilities() []ValidatorIllegibility {
	var values []ValidatorIllegibility
	for i := 0; i < len(ValidatorIllegibility_name)-1; i++ {
		values = append(values, ValidatorIllegibility(1<<i))
	}

	return values
}

// Slasher provides functionality to manage slashing info for a validator
type Slasher interface {
	GetValidatorSigningInfo(ctx sdk.Context, address sdk.ConsAddress) (info slashingtypes.ValidatorSigningInfo, found bool)
	SignedBlocksWindow(ctx sdk.Context) (res int64)
	GetValidatorMissedBlockBitArray(ctx sdk.Context, address sdk.ConsAddress, index int64) bool
}

// Tss provides functionality to tss module
type Tss interface {
	GetSuspendedUntil(ctx sdk.Context, validator sdk.ValAddress) int64
	GetNextKey(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool)
	IsOperatorAvailable(ctx sdk.Context, validator sdk.ValAddress, keyIDs ...tss.KeyID) bool
	GetKeyRequirement(ctx sdk.Context, keyRole tss.KeyRole, keyType tss.KeyType) (tss.KeyRequirement, bool)
	HasMissedTooManyBlocks(ctx sdk.Context, address sdk.ConsAddress) (bool, error)
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
	GetSnapshot(ctx sdk.Context, seqNo int64) (Snapshot, bool)
	TakeSnapshot(ctx sdk.Context, keyRequirement tss.KeyRequirement) (Snapshot, error)
	GetOperator(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress
	GetProxy(ctx sdk.Context, principal sdk.ValAddress) (addr sdk.AccAddress, active bool)
	GetValidatorIllegibility(ctx sdk.Context, validator SDKValidator) (ValidatorIllegibility, error)
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
