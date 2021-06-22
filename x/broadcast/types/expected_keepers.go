package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

//go:generate moq -pkg mock -out ./mock/expected_keepers.go . Staker

// Staker gives access to validators of the network
type Staker interface {
	Validator(ctx sdk.Context, addr sdk.ValAddress) stakingtypes.ValidatorI
}
