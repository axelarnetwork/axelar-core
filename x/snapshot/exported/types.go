package exported

import (
	"bytes"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
)

//go:generate moq -out ./mock/types.go -pkg mock . Validator Snapshotter Slasher Broadcaster Tss

// Validator is an interface for a Cosmos validator account
type Validator interface {
	GetOperator() sdk.ValAddress
	GetConsAddr() sdk.ConsAddress
	GetConsensusPower() int64
	IsJailed() bool
}

// ValidatorInfo adopts the methods from "github.com/cosmos/cosmos-sdk/x/slashing" that are
// actually used by this module
type ValidatorInfo struct {
	slashing.ValidatorSigningInfo
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
}

// IsValidatorActive returns true if the validator is active; otherwise, false
func IsValidatorActive(ctx sdk.Context, slasher Slasher, validator Validator) bool {
	signingInfo, found := slasher.GetValidatorSigningInfo(ctx, validator.GetConsAddr())

	return found && !signingInfo.Tombstoned && signingInfo.MissedBlocksCounter <= 0 && !validator.IsJailed()
}

// DoesValidatorHasProxyRegistered returns true if the validator has broadcast proxy registered; otherwise, false
func DoesValidatorHasProxyRegistered(ctx sdk.Context, broadcaster Broadcaster, validator Validator) bool {
	return broadcaster.GetProxy(ctx, validator.GetOperator()) != nil
}

// IsValidatorTssRegistered returns true if the validator is registered to participate in tss key generation; otherwise, false
func IsValidatorTssRegistered(ctx sdk.Context, tss Tss, validator Validator) bool {
	return tss.GetValidatorDeregisteredBlockHeight(ctx, validator.GetOperator()) <= 0
}

// Snapshot is a snapshot of the validator set at a given block height.
type Snapshot struct {
	Validators           []Validator `json:"validators"`
	Timestamp            time.Time   `json:"timestamp"`
	Height               int64       `json:"height"`
	TotalPower           sdk.Int     `json:"totalpower"`
	ValidatorsTotalPower sdk.Int     `json:"validatorstotalpower"`
	Counter              int64       `json:"counter"`
}

// GetValidator returns the validator for a given address, if it is part of the snapshot
func (s Snapshot) GetValidator(address sdk.ValAddress) (Validator, bool) {
	for _, validator := range s.Validators {
		if bytes.Equal(validator.GetOperator(), address) {
			return validator, true
		}
	}

	return nil, false
}

// Snapshotter represents the interface for the snapshot module's functionality
type Snapshotter interface {
	GetLatestSnapshot(ctx sdk.Context) (Snapshot, bool)
	GetLatestCounter(ctx sdk.Context) int64
	GetSnapshot(ctx sdk.Context, counter int64) (Snapshot, bool)
	TakeSnapshot(ctx sdk.Context, subsetSize int64) error
}
