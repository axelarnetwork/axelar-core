package types

import (
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	reward "github.com/axelarnetwork/axelar-core/x/reward/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
)

//go:generate moq -out ./mock/expected_keepers.go -pkg mock . Nexus Snapshotter AxelarnetKeeper RewardKeeper SlashingKeeper WasmKeeper AccountKeeper

// Nexus provides functionality to manage cross-chain transfers
type Nexus interface {
	Logger(ctx sdk.Context) log.Logger

	InitGenesis(ctx sdk.Context, genState *GenesisState)
	ExportGenesis(ctx sdk.Context) *GenesisState

	SetParams(ctx sdk.Context, p Params)
	GetParams(ctx sdk.Context) Params

	IsChainActivated(ctx sdk.Context, chain exported.Chain) bool
	ActivateChain(ctx sdk.Context, chain exported.Chain)
	GetChains(ctx sdk.Context) []exported.Chain
	GetChain(ctx sdk.Context, chain exported.ChainName) (exported.Chain, bool)
	IsChainMaintainer(ctx sdk.Context, chain exported.Chain, maintainer sdk.ValAddress) bool
	AddChainMaintainer(ctx sdk.Context, chain exported.Chain, validator sdk.ValAddress) error
	RemoveChainMaintainer(ctx sdk.Context, chain exported.Chain, validator sdk.ValAddress) error
	GetChainMaintainers(ctx sdk.Context, chain exported.Chain) []sdk.ValAddress
	GetChainMaintainerStates(ctx sdk.Context, chain exported.Chain) []exported.MaintainerState
	LinkAddresses(ctx sdk.Context, sender exported.CrossChainAddress, recipient exported.CrossChainAddress) error
	DeactivateChain(ctx sdk.Context, chain exported.Chain)
	RegisterFee(ctx sdk.Context, chain exported.Chain, feeInfo exported.FeeInfo) error
	GetFeeInfo(ctx sdk.Context, chain exported.Chain, asset string) exported.FeeInfo
	SetRateLimit(ctx sdk.Context, chainName exported.ChainName, limit sdk.Coin, window time.Duration) error
	RateLimitTransfer(ctx sdk.Context, chain exported.ChainName, asset sdk.Coin, direction exported.TransferDirection) error
	GenerateMessageID(ctx sdk.Context) (string, []byte, uint64)
	SetNewMessage(ctx sdk.Context, msg exported.GeneralMessage) error
	GetMessage(ctx sdk.Context, id string) (exported.GeneralMessage, bool)
	SetMessageExecuted(ctx sdk.Context, id string) error
	RouteMessage(ctx sdk.Context, id string, routingCtx ...exported.RoutingContext) error
	DequeueRouteMessage(ctx sdk.Context) (exported.GeneralMessage, bool)
}

// Snapshotter provides functionality to the snapshot module
type Snapshotter interface {
	CreateSnapshot(ctx sdk.Context, candidates []sdk.ValAddress, filterFunc func(snapshot.ValidatorI) bool, weightFunc func(consensusPower sdk.Uint) sdk.Uint, threshold utils.Threshold) (snapshot.Snapshot, error)
	GetOperator(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress
	GetProxy(ctx sdk.Context, operator sdk.ValAddress) (addr sdk.AccAddress, active bool)
}

// StakingKeeper provides functionality to the staking module
type StakingKeeper interface {
	Validator(ctx sdk.Context, addr sdk.ValAddress) stakingtypes.ValidatorI
	PowerReduction(sdk.Context) sdk.Int
	GetLastTotalPower(sdk.Context) sdk.Int
}

// AxelarnetKeeper provides functionality to the axelarnet module
type AxelarnetKeeper interface {
	IsCosmosChain(ctx sdk.Context, chain exported.ChainName) bool
}

// RewardKeeper provides functionality to get reward keeper
type RewardKeeper interface {
	GetPool(ctx sdk.Context, name string) reward.RewardPool
}

// SlashingKeeper provides functionality to manage slashing info for a validator
type SlashingKeeper interface {
	IsTombstoned(ctx sdk.Context, consAddr sdk.ConsAddress) bool
}

// WasmKeeper provides functionality to manage wasm contracts
type WasmKeeper interface {
	wasmtypes.ContractOpsKeeper
}

// AccountKeeper provides functionality to get account keeper
type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
}
