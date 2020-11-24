package exported

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
)

type Staker interface {
	GetLastTotalPower(ctx sdk.Context) (power sdk.Int)
	Validator(ctx sdk.Context, address sdk.ValAddress) exported.ValidatorI
	// TODO: check if this is actually the correct function we need (e.g. do we need only bonded?)
	IterateValidators(ctx sdk.Context, fn func(index int64, validator exported.ValidatorI) (stop bool))
}
