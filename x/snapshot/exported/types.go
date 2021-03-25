package exported

import (
	"bytes"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

//go:generate moq -out ./mock/types.go -pkg mock . Validator Snapshotter

// Validator is an interface for a Cosmos validator account
type Validator interface {
	GetOperator() sdk.ValAddress
	GetConsAddr() sdk.ConsAddress
	GetConsensusPower() int64
	IsJailed() bool
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
