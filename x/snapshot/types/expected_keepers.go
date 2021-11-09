package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

//go:generate moq -pkg mock -out ./mock/expected_keepers.go . StakingKeeper BankKeeper

// StakingKeeper adopts the methods from "github.com/cosmos/cosmos-sdk/x/staking/exported" that are
// actually used by this module
type StakingKeeper interface {
	GetLastTotalPower(ctx sdk.Context) (power sdk.Int)
	IterateBondedValidatorsByPower(ctx sdk.Context, fn func(index int64, validator stakingtypes.ValidatorI) (stop bool))
	Validator(ctx sdk.Context, addr sdk.ValAddress) stakingtypes.ValidatorI
	PowerReduction(ctx sdk.Context) sdk.Int
	BondDenom(ctx sdk.Context) string
}

// BankKeeper adops the GetBalance function of the bank keeper that is used by this module
type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
}
