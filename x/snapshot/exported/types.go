package exported

import (
	"bytes"
	"github.com/axelarnetwork/axelar-core/utils"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/gogo/protobuf/proto"
)

//go:generate moq -out ./mock/types.go -pkg mock . SDKValidator Snapshotter Slasher Broadcaster Tss

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

// Broadcaster provides broadcasting functionality
type Broadcaster interface {
	GetProxy(ctx sdk.Context, principal sdk.ValAddress) sdk.AccAddress
}

// Tss provides functionality to tss module
type Tss interface {
	SetKeyRequirement(ctx sdk.Context, keyRequirement tss.KeyRequirement)
	GetValidatorDeregisteredBlockHeight(ctx sdk.Context, valAddr sdk.ValAddress) int64
	GetMinBondFractionPerShare(ctx sdk.Context) utils.Threshold
	GetTssSuspendedUntil(ctx sdk.Context, validator sdk.ValAddress) int64
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
func HasProxyRegistered(ctx sdk.Context, broadcaster Broadcaster, validator SDKValidator) bool {
	return broadcaster.GetProxy(ctx, validator.GetOperator()) != nil
}

// IsValidatorTssRegistered returns true if the validator is registered to participate in tss key generation; otherwise, false
func IsValidatorTssRegistered(ctx sdk.Context, tss Tss, validator SDKValidator) bool {
	return tss.GetValidatorDeregisteredBlockHeight(ctx, validator.GetOperator()) <= 0
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
}

// GetSDKValidator returns the SdkValidator
func (m Validator) GetSDKValidator() SDKValidator {
	if m.SDKValidator == nil {
		return nil
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
