package exported

import (
	"bytes"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

//go:generate moq -out ./mock/types.go -pkg mock . Snapshotter Validator

type Validator interface {
	GetOperator() sdk.ValAddress
	GetConsensusPower() int64
}

type Snapshot struct {
	Validators []Validator `json:"validators"`
	Timestamp  time.Time   `json:"timestamp"`
	Height     int64       `json:"height"`
	TotalPower sdk.Int     `json:"totalpower"`
}

func (s Snapshot) GetValidator(address sdk.ValAddress) (Validator, bool) {
	for _, validator := range s.Validators {
		if bytes.Equal(validator.GetOperator(), address) {
			return validator, true
		}
	}

	return nil, false
}

type Snapshotter interface {
	// GetValidator returns the validator with the given address. Returns false if no validator with that address exists
	GetValidator(ctx sdk.Context, address sdk.ValAddress) (Validator, bool)
	GetLatestSnapshot(ctx sdk.Context) (Snapshot, bool)
	GetLatestRound(ctx sdk.Context) int64
	GetSnapshot(ctx sdk.Context, round int64) (Snapshot, bool)
}
