package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	sdkExported "github.com/cosmos/cosmos-sdk/x/staking/exported"
)

//go:generate moq -pkg mock -out ./mock/expected_keepers.go . StakingKeeper Slasher Broadcaster

// Broadcaster provides broadcasting functionality
type Broadcaster interface {
	GetProxy(ctx sdk.Context, principal sdk.ValAddress) sdk.AccAddress
}

// StakingKeeper adopts the methods from "github.com/cosmos/cosmos-sdk/x/staking/exported" that are
// actually used by this module
type StakingKeeper interface {
	GetLastTotalPower(ctx sdk.Context) (power sdk.Int)
	IterateLastValidators(ctx sdk.Context, fn func(index int64, validator sdkExported.ValidatorI) (stop bool))
	Validator(ctx sdk.Context, addr sdk.ValAddress) sdkExported.ValidatorI
}

// ValidatorInfo adopts the methods from "github.com/cosmos/cosmos-sdk/x/slashing" that are
// actually used by this module
type ValidatorInfo struct {
	slashing.ValidatorSigningInfo
}

// Slasher provides functionality to manage slashing info for a validator
type Slasher interface {
	GetValidatorSigningInfo(ctx sdk.Context, address sdk.ConsAddress) (info ValidatorInfo, found bool)
}
