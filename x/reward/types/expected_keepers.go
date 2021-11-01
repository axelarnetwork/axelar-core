package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/reward/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

//go:generate moq -pkg mock -out ./mock/expected_keepers.go . Rewarder Nexus Minter Distributor Staker Banker

// Rewarder provides reward functionality
type Rewarder interface {
	Logger(ctx sdk.Context) log.Logger

	GetParams(ctx sdk.Context) (params Params)
	GetPool(ctx sdk.Context, name string) exported.RewardPool
}

// Nexus provides nexus functionality
type Nexus interface {
	GetChains(ctx sdk.Context) []nexus.Chain
	GetChainMaintainers(ctx sdk.Context, chain nexus.Chain) []sdk.ValAddress
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

// Banker provides bank functionality
type Banker interface {
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, name string, amt sdk.Coins) error
}

// Tss provides tss functionality
type Tss interface {
	IsOperatorAvailable(ctx sdk.Context, validator sdk.ValAddress, keyIDs ...tss.KeyID) bool
}

// Snapshotter provides snapshot functionality
type Snapshotter interface {
	GetValidatorIllegibility(ctx sdk.Context, validator snapshot.SDKValidator) (snapshot.ValidatorIllegibility, error)
}
