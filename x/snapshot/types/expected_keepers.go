package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	exported2 "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

//go:generate moq -pkg mock -out ./mock/expected_keepers.go . StakingKeeper BankKeeper Slasher Tss

// StakingKeeper adopts the methods from "github.com/cosmos/cosmos-sdk/x/staking/exported" that are
// actually used by this module
type StakingKeeper interface {
	GetLastTotalPower(ctx sdk.Context) (power sdk.Int)
	IterateBondedValidatorsByPower(ctx sdk.Context, fn func(index int64, validator stakingtypes.ValidatorI) (stop bool))
	Validator(ctx sdk.Context, addr sdk.ValAddress) stakingtypes.ValidatorI
	PowerReduction(ctx sdk.Context) sdk.Int
	BondDenom(ctx sdk.Context) string
}

// BankKeeper adopts the GetBalance function of the bank keeper that is used by this module
type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// Nexus provides functionality to manage cross-chain transfers
type Nexus interface {
	GetChains(ctx sdk.Context) []nexus.Chain
	IsChainMaintainer(ctx sdk.Context, chain exported.Chain, maintainer sdk.ValAddress) bool
	RemoveChainMaintainer(ctx sdk.Context, chain nexus.Chain, maintainer sdk.ValAddress) error
}

// Slasher provides functionality to manage slashing info for a validator
type Slasher interface {
	GetValidatorSigningInfo(ctx sdk.Context, address sdk.ConsAddress) (info types.ValidatorSigningInfo, found bool)
	SignedBlocksWindow(ctx sdk.Context) (res int64)
	GetValidatorMissedBlockBitArray(ctx sdk.Context, address sdk.ConsAddress, index int64) bool
}

// Tss provides functionality to tss module
type Tss interface {
	GetSuspendedUntil(ctx sdk.Context, validator sdk.ValAddress) int64
	GetNextKey(ctx sdk.Context, chain nexus.Chain, keyRole exported2.KeyRole) (exported2.Key, bool)
	IsOperatorAvailable(ctx sdk.Context, validator sdk.ValAddress, keyIDs ...exported2.KeyID) bool
	GetKeyRequirement(ctx sdk.Context, keyRole exported2.KeyRole, keyType exported2.KeyType) (exported2.KeyRequirement, bool)
	HasMissedTooManyBlocks(ctx sdk.Context, address sdk.ConsAddress) (bool, error)
}