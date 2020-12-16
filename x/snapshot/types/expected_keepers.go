package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkExported "github.com/cosmos/cosmos-sdk/x/staking/exported"
	typesStaking "github.com/cosmos/cosmos-sdk/x/staking/types"
)

//go:generate moq -pkg mock -out ./mock/expected_keepers.go . StakingKeeper

// StakingKeeper adopts the methods from "github.com/cosmos/cosmos-sdk/x/staking/exported" that are
// actually used by this module
type StakingKeeper interface {
	GetLastTotalPower(ctx sdk.Context) (power sdk.Int)
	IterateValidators(ctx sdk.Context, fn func(index int64, validator sdkExported.ValidatorI) (stop bool))
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator typesStaking.Validator, found bool)
}
