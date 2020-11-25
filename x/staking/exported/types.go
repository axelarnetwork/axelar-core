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
	GetLastTotalPower(ctx sdk.Context) (power sdk.Int)
	Validator(ctx sdk.Context, address sdk.ValAddress) (Validator, error)
	// TODO: check if this is actually the correct function we need (e.g. do we need only bonded?)
	IterateValidators(ctx sdk.Context, fn func(index int64, validator Validator) (stop bool))

	GetLatestSnapshot(ctx sdk.Context) (Snapshot, error)
}
