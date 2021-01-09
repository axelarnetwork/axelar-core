package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkExported "github.com/cosmos/cosmos-sdk/x/staking/exported"
)

//go:generate moq -pkg mock -out ./mock/expected_keepers.go . StakingKeeper

// StakingKeeper adopts the methods from "github.com/cosmos/cosmos-sdk/x/staking/exported" that are
// actually used by this module
type StakingKeeper interface {
	GetLastTotalPower(ctx sdk.Context) (power sdk.Int)
	IterateLastValidators(ctx sdk.Context, fn func(index int64, validator sdkExported.ValidatorI) (stop bool))
	Validator(ctx sdk.Context, addr sdk.ValAddress) sdkExported.ValidatorI
}
