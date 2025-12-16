package types

import (
	"context"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/reward/exported"
)

//go:generate moq -pkg mock -out ./mock/expected_keepers.go . Rewarder Refunder Nexus Distributor Staker Slasher Banker AccountKeeper MultiSig Snapshotter

// Rewarder provides reward functionality
type Rewarder interface {
	Logger(ctx sdk.Context) log.Logger

	GetParams(ctx sdk.Context) (params Params)
	GetPool(ctx sdk.Context, name string) exported.RewardPool
}

// Refunder provides refunding functionality
type Refunder interface {
	Logger(ctx sdk.Context) log.Logger
	GetPendingRefund(ctx sdk.Context, req RefundMsgRequest) (Refund, bool)
	DeletePendingRefund(ctx sdk.Context, req RefundMsgRequest)
	SetParams(ctx sdk.Context, params Params)
}

// Nexus provides nexus functionality
type Nexus interface {
	GetChains(ctx sdk.Context) []nexus.Chain
	GetChainMaintainers(ctx sdk.Context, chain nexus.Chain) []sdk.ValAddress
	IsChainActivated(ctx sdk.Context, chain nexus.Chain) bool
}

// Distributor provides distribution functionality
type Distributor interface {
	AllocateTokensToValidator(ctx context.Context, val stakingtypes.ValidatorI, tokens sdk.DecCoins) error
}

// Staker provides stake functionality
type Staker interface {
	Validator(ctx context.Context, addr sdk.ValAddress) (stakingtypes.ValidatorI, error)
	PowerReduction(ctx context.Context) math.Int
	IterateBondedValidatorsByPower(ctx context.Context, fn func(index int64, validator stakingtypes.ValidatorI) (stop bool)) error
	StakingTokenSupply(ctx context.Context) (math.Int, error)
	BondedRatio(ctx context.Context) (math.LegacyDec, error)
}

// Slasher provides necessary functions to the validator information
type Slasher interface {
	IsTombstoned(ctx context.Context, consAddr sdk.ConsAddress) bool // whether a validator is tombstoned
}

// Banker provides bank functionality
type Banker interface {
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx context.Context, name string, amt sdk.Coins) error
}

// MultiSig provides mutlisig functionality
type MultiSig interface {
	HasOptedOut(ctx sdk.Context, participant sdk.AccAddress) bool
}

// Snapshotter provides snapshot functionality
type Snapshotter interface {
	GetProxy(ctx sdk.Context, operator sdk.ValAddress) (sdk.AccAddress, bool)
}

// AccountKeeper defines the contract required for account APIs.
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	SetModuleAccount(context.Context, sdk.ModuleAccountI)
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
}
