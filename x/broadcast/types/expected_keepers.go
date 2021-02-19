package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkExported "github.com/cosmos/cosmos-sdk/x/staking/exported"
)

// Staker gives access to validators of the network
type Staker interface {
	Validator(ctx sdk.Context, addr sdk.ValAddress) sdkExported.ValidatorI
}
