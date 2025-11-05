package types

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

//go:generate moq -pkg mock -out ./mock/expected_keepers.go . StakingKeeper BankKeeper Slasher

// StakingKeeper adopts the methods from "github.com/cosmos/cosmos-sdk/x/staking/exported" that are
// actually used by this module
type StakingKeeper interface {
	GetLastTotalPower(ctx context.Context) (math.Int, error)
	IterateBondedValidatorsByPower(ctx context.Context, fn func(index int64, validator stakingtypes.ValidatorI) (stop bool)) error
	Validator(ctx context.Context, addr sdk.ValAddress) (stakingtypes.ValidatorI, error)
	PowerReduction(ctx context.Context) math.Int
	BondDenom(ctx context.Context) (string, error)
}

// BankKeeper adopts the GetBalance function of the bank keeper that is used by this module
type BankKeeper interface {
	SpendableBalance(ctx context.Context, address sdk.AccAddress, denom string) sdk.Coin
}

// Nexus provides functionality to manage cross-chain transfers
type Nexus interface {
	GetChains(ctx sdk.Context) []nexus.Chain
	IsChainMaintainer(ctx sdk.Context, chain exported.Chain, maintainer sdk.ValAddress) bool
	RemoveChainMaintainer(ctx sdk.Context, chain nexus.Chain, maintainer sdk.ValAddress) error
}

// Slasher provides functionality to manage slashing info for a validator
type Slasher interface {
	GetValidatorSigningInfo(ctx context.Context, address sdk.ConsAddress) (types.ValidatorSigningInfo, error)
	SignedBlocksWindow(ctx context.Context) (int64, error)
	GetMissedBlockBitmapValue(ctx context.Context, addr sdk.ConsAddress, index int64) (bool, error)
}
