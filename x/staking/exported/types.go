package exported

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Validator struct {
	Address sdk.ValAddress `json:"validators"`
	Power   int64          `json:"power"`
}

type Snapshot struct {
	Validators []Validator `json:"validators"`
	Timestamp  time.Time   `json:"timestamp"`
	Height     int64       `json:"height"`
	TotalPower sdk.Int     `json:"totalpower"`
}

type Staker interface {
	Validator(ctx sdk.Context, address sdk.ValAddress) (Validator, bool)
	GetLatestSnapshot(ctx sdk.Context) (Snapshot, bool)
	GetLatestRound(ctx sdk.Context) int64
	GetSnapshot(ctx sdk.Context, round int64) (Snapshot, bool)
}
