package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

//go:generate moq -out ./mock/expected_keepers.go -pkg mock . Nexus Snapshotter

// Nexus provides functionality to manage cross-chain transfers
type Nexus interface {
	Logger(ctx sdk.Context) log.Logger

	SetParams(ctx sdk.Context, p Params)
	GetParams(ctx sdk.Context) Params

	IsChainActivated(ctx sdk.Context, chain exported.Chain) bool
	ActivateChain(ctx sdk.Context, chain exported.Chain)
	GetChains(ctx sdk.Context) []exported.Chain
	GetChain(ctx sdk.Context, chain string) (exported.Chain, bool)
	IsChainMaintainer(ctx sdk.Context, chain exported.Chain, maintainer sdk.ValAddress) bool
	AddChainMaintainer(ctx sdk.Context, chain exported.Chain, validator sdk.ValAddress) error
	RemoveChainMaintainer(ctx sdk.Context, chain exported.Chain, validator sdk.ValAddress) error
	GetChainMaintainers(ctx sdk.Context, chain exported.Chain) []sdk.ValAddress
}

// Snapshotter provides functionality to the snapshot module
type Snapshotter interface {
	GetOperator(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress
}

// StakingKeeper provides functionality to the staking module
type StakingKeeper interface {
	Validator(ctx sdk.Context, addr sdk.ValAddress) stakingtypes.ValidatorI
	PowerReduction(sdk.Context) sdk.Int
	GetLastTotalPower(sdk.Context) sdk.Int
}
