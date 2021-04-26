package exported

import (
	"bytes"
	"time"

	"github.com/axelarnetwork/axelar-core/utils"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

//go:generate moq -out ./mock/types.go -pkg mock . SDKValidator Snapshotter Slasher Broadcaster Tss

// SDKValidator is an interface for a Cosmos validator account
type SDKValidator interface {
	GetOperator() sdk.ValAddress
	GetConsAddr() (sdk.ConsAddress, error)
	GetConsensusPower() int64
	IsJailed() bool
	UnpackInterfaces(c codectypes.AnyUnpacker) error
}

type Validator struct {
	SDKValidator
	ShareCount int64
}

func NewValidator(validator SDKValidator, shareCount int64) Validator {
	return Validator{SDKValidator: validator, ShareCount: shareCount}
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
	GetValidatorDeregisteredBlockHeight(ctx sdk.Context, valAddr sdk.ValAddress) int64
	GetMinBondFractionPerShare(ctx sdk.Context) utils.Threshold
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

// Snapshot is a snapshot of the validator set at a given block height.
type Snapshot struct {
	Validators                 []Validator                    `json:"validators"`
	Timestamp                  time.Time                      `json:"timestamp"`
	Height                     int64                          `json:"height"`
	TotalShareCount            sdk.Int                        `json:"totalsharecount"`
	Counter                    int64                          `json:"counter"`
	KeyShareDistributionPolicy tss.KeyShareDistributionPolicy `json:"keysharedistributionpolicy"`
}

// GetValidator returns the validator for a given address, if it is part of the snapshot
func (s Snapshot) GetValidator(address sdk.ValAddress) (Validator, bool) {
	for _, validator := range s.Validators {
		if bytes.Equal(validator.GetOperator(), address) {
			return validator, true
		}
	}

	return Validator{}, false
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (s Snapshot) UnpackInterfaces(c codectypes.AnyUnpacker) error {
	for _, v := range s.Validators {
		if err := v.UnpackInterfaces(c); err != nil {
			return err
		}
	}
	return nil
}

// Snapshotter represents the interface for the snapshot module's functionality
type Snapshotter interface {
	GetLatestSnapshot(ctx sdk.Context) (Snapshot, bool)
	GetLatestCounter(ctx sdk.Context) int64
	GetSnapshot(ctx sdk.Context, counter int64) (Snapshot, bool)
	TakeSnapshot(ctx sdk.Context, subsetSize int64, keyShareDistributionPolicy tss.KeyShareDistributionPolicy) (snapshotConsensusPower sdk.Int, totalConsensusPower sdk.Int, err error)
}
