package exported

import (
	"bytes"
	"time"

	"github.com/axelarnetwork/axelar-core/x/snapshot/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

//go:generate moq -out ./mock/types.go -pkg mock . Validator Snapshotter

// Validator is an interface for a Cosmos validator account
type Validator = types.Validator

// IsValidatorActive returns true if the validator is active; otherwise, false
func IsValidatorActive(ctx sdk.Context, slasher types.Slasher, validator Validator) bool {
	signingInfo, found := slasher.GetValidatorSigningInfo(ctx, validator.GetConsAddr())

	return found && !signingInfo.Tombstoned && signingInfo.MissedBlocksCounter <= 0 && !validator.IsJailed()
}

// DoesValidatorHasProxyRegistered returns true if the validator has broadcast proxy registered; otherwise, false
func DoesValidatorHasProxyRegistered(ctx sdk.Context, broadcaster types.Broadcaster, validator Validator) bool {
	return broadcaster.GetProxy(ctx, validator.GetOperator()) != nil
}

// IsValidatorTssRegistered returns true if the validator is registered to participate in tss key generation; otherwise, false
func IsValidatorTssRegistered(ctx sdk.Context, tss types.Tss, validator Validator) bool {
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
	TakeSnapshot(ctx sdk.Context, validatorCount int64) error
}
