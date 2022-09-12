package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/libs/log"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/reward/exported"
)

//go:generate moq -pkg mock -out ./mock/expected_keepers.go . Rewarder Refunder Nexus Minter Distributor Staker Slasher Banker

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
}

// Nexus provides nexus functionality
type Nexus interface {
	GetChains(ctx sdk.Context) []nexus.Chain
	GetChainMaintainers(ctx sdk.Context, chain nexus.Chain) []sdk.ValAddress
	IsChainActivated(ctx sdk.Context, chain nexus.Chain) bool
}

// Minter provides mint functionality
type Minter interface {
	GetParams(ctx sdk.Context) minttypes.Params
	StakingTokenSupply(ctx sdk.Context) sdk.Int
	GetMinter(ctx sdk.Context) minttypes.Minter
}

// Distributor provides distribution functionality
type Distributor interface {
	AllocateTokensToValidator(ctx sdk.Context, val stakingtypes.ValidatorI, tokens sdk.DecCoins)
}

// Staker provides stake functionality
type Staker interface {
	Validator(ctx sdk.Context, addr sdk.ValAddress) stakingtypes.ValidatorI
	PowerReduction(ctx sdk.Context) sdk.Int
	IterateBondedValidatorsByPower(ctx sdk.Context, fn func(index int64, validator stakingtypes.ValidatorI) (stop bool))
}

// Slasher provides necessary functions to the validator information
type Slasher interface {
	IsTombstoned(ctx sdk.Context, consAddr sdk.ConsAddress) bool // whether a validator is tombstoned
}

// Banker provides bank functionality
type Banker interface {
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, name string, amt sdk.Coins) error
}

// MultiSig provides mutlisig functionality
type MultiSig interface {
}

// Snapshotter provides snapshot functionality
type Snapshotter interface {
	GetProxy(ctx sdk.Context, operator sdk.ValAddress) (sdk.AccAddress, bool)
}
