package exported

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"golang.org/x/exp/maps"

	"github.com/axelarnetwork/axelar-core/utils"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/utils/slices"
)

//go:generate moq -out ./mock/types.go -pkg mock . SDKValidator Snapshotter ValidatorI

// QuadraticWeightFunc returns floor(sqrt(consensusPower)) as the weight
func QuadraticWeightFunc(consensusPower sdk.Uint) sdk.Uint {
	bigInt := consensusPower.BigInt()

	return sdk.NewUintFromBigInt(bigInt.Sqrt(bigInt))
}

// ValidatorI provides necessary functions to the validator information
type ValidatorI interface {
	GetConsensusPower(sdk.Int) int64       // validation power in tendermint
	GetOperator() sdk.ValAddress           // operator address to receive/return validators coins
	GetConsAddr() (sdk.ConsAddress, error) // validation consensus address
	IsJailed() bool                        // whether the validator is jailed
	IsBonded() bool                        // whether the validator is bonded
}

// NewSnapshot is the constructor of Snapshot
func NewSnapshot(timestamp time.Time, height int64, participants []Participant, bondedWeight sdk.Uint) Snapshot {
	return Snapshot{
		Timestamp:    timestamp,
		Height:       height,
		Participants: slices.ToMap(participants, func(p Participant) string { return p.Address.String() }),
		BondedWeight: bondedWeight,
	}
}

// ValidateBasic returns an error if the given snapshot is invalid; nil otherwise
func (m Snapshot) ValidateBasic() error {
	if len(m.Participants) == 0 {
		return fmt.Errorf("snapshot cannot have no participant")
	}

	if m.BondedWeight.IsZero() {
		return fmt.Errorf("snapshot must have bonded weight >0")
	}

	if m.Height <= 0 {
		return fmt.Errorf("snapshot must have height >0")
	}

	if m.Timestamp.IsZero() {
		return fmt.Errorf("snapshot must have timestamp >0")
	}

	for addr, p := range m.Participants {
		if err := p.ValidateBasic(); err != nil {
			return err
		}

		if addr != p.Address.String() {
			return fmt.Errorf("invalid snapshot")
		}
	}

	if m.GetParticipantsWeight().GT(m.BondedWeight) {
		return fmt.Errorf("snapshot cannot have sum of participants weight greater than bonded weight")
	}

	return nil
}

// NewParticipant is the constructor of Participant
func NewParticipant(address sdk.ValAddress, weight sdk.Uint) Participant {
	return Participant{
		Address: address,
		Weight:  weight,
	}
}

// GetAddress returns the address of the participant
func (m Participant) GetAddress() sdk.ValAddress {
	return m.Address
}

// ValidateBasic returns an error if the given participant is invalid; nil otherwise
func (m Participant) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Address); err != nil {
		return err
	}

	return nil
}

// GetParticipantAddresses returns the addresses of all participants in the snapshot
func (m Snapshot) GetParticipantAddresses() []sdk.ValAddress {
	addresses := slices.Map(maps.Values(m.Participants), Participant.GetAddress)
	sort.SliceStable(addresses, func(i, j int) bool { return bytes.Compare(addresses[i], addresses[j]) < 0 })

	return addresses
}

// GetParticipantsWeight returns the sum of all participants' weights
func (m Snapshot) GetParticipantsWeight() sdk.Uint {
	weight := sdk.ZeroUint()
	for _, p := range m.Participants {
		weight = weight.Add(p.Weight)
	}

	return weight
}

// GetParticipantWeight returns the weight of the given participant
func (m Snapshot) GetParticipantWeight(participant sdk.ValAddress) sdk.Uint {
	if participant, ok := m.Participants[participant.String()]; ok {
		return participant.Weight
	}

	return sdk.ZeroUint()
}

// CalculateMinPassingWeight returns the minimum amount of weights to pass the given threshold
func (m Snapshot) CalculateMinPassingWeight(threshold utils.Threshold) sdk.Uint {
	minPassingWeight := m.BondedWeight.
		MulUint64(uint64(threshold.Numerator)).
		QuoUint64(uint64(threshold.Denominator))

	if minPassingWeight.MulUint64(uint64(threshold.Denominator)).GTE(m.BondedWeight.MulUint64(uint64(threshold.Numerator))) {
		return minPassingWeight
	}

	return minPassingWeight.AddUint64(1)
}

// Validate returns an error if the snapshot is not valid; nil otherwise
// Deprecated
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
	CreateSnapshot(ctx sdk.Context, candidates []sdk.ValAddress, filterFunc func(ValidatorI) bool, weightFunc func(consensusPower sdk.Uint) sdk.Uint, threshold utils.Threshold) (Snapshot, error)
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
