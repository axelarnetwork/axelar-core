package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type StakingKeeper interface {
	GetAllValidators(ctx sdk.Context) (validators []staking.Validator)
}
