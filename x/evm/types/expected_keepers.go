package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	reward "github.com/axelarnetwork/axelar-core/x/reward/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

//go:generate moq -out ./mock/expected_keepers.go -pkg mock . Voter Nexus Snapshotter BaseKeeper ChainKeeper Rewarder StakingKeeper SlashingKeeper MultisigKeeper

// BaseKeeper is implemented by this module's base keeper
type BaseKeeper interface {
	Logger(ctx sdk.Context) log.Logger

	HasChain(ctx sdk.Context, chain nexus.ChainName) bool
	ForChain(chain nexus.ChainName) ChainKeeper

	InitGenesis(ctx sdk.Context, state GenesisState)
	ExportGenesis(ctx sdk.Context) GenesisState
}

// ChainKeeper is implemented by this module's chain keeper
type ChainKeeper interface {
	Logger(ctx sdk.Context) log.Logger

	GetName() string

	SetParams(ctx sdk.Context, p Params)
	GetParams(ctx sdk.Context) Params

	GetNetwork(ctx sdk.Context) (string, bool)
	GetChainID(ctx sdk.Context) (sdk.Int, bool)
	GetRequiredConfirmationHeight(ctx sdk.Context) (uint64, bool)
	GetRevoteLockingPeriod(ctx sdk.Context) (int64, bool)
	GetBurnerByteCode(ctx sdk.Context) ([]byte, bool)
	GetTokenByteCode(ctx sdk.Context) ([]byte, bool)
	SetGateway(ctx sdk.Context, address Address)
	GetGatewayAddress(ctx sdk.Context) (Address, bool)
	GetDeposit(ctx sdk.Context, txID Hash, burnerAddr Address) (ERC20Deposit, DepositStatus, bool)
	GetBurnerInfo(ctx sdk.Context, address Address) *BurnerInfo
	GenerateSalt(ctx sdk.Context, recipient string) Hash
	GetBurnerAddress(ctx sdk.Context, token ERC20Token, salt Hash, gatewayAddr Address) (Address, error)
	SetBurnerInfo(ctx sdk.Context, burnerInfo BurnerInfo)
	DeleteDeposit(ctx sdk.Context, deposit ERC20Deposit)
	SetDeposit(ctx sdk.Context, deposit ERC20Deposit, state DepositStatus)
	GetConfirmedDepositsPaginated(ctx sdk.Context, pageRequest *query.PageRequest) ([]ERC20Deposit, *query.PageResponse, error)
	GetNetworkByID(ctx sdk.Context, id sdk.Int) (string, bool)
	GetChainIDByNetwork(ctx sdk.Context, network string) (sdk.Int, bool)
	GetVotingThreshold(ctx sdk.Context) (utils.Threshold, bool)
	GetMinVoterCount(ctx sdk.Context) (int64, bool)

	CreateERC20Token(ctx sdk.Context, asset string, details TokenDetails, address Address) (ERC20Token, error)
	GetERC20TokenByAsset(ctx sdk.Context, asset string) ERC20Token
	GetERC20TokenBySymbol(ctx sdk.Context, symbol string) ERC20Token
	GetTokens(ctx sdk.Context) []ERC20Token

	EnqueueCommand(ctx sdk.Context, cmd Command) error
	GetCommand(ctx sdk.Context, id CommandID) (Command, bool)
	GetPendingCommands(ctx sdk.Context) []Command
	CreateNewBatchToSign(ctx sdk.Context) (CommandBatch, error)
	SetLatestSignedCommandBatchID(ctx sdk.Context, id []byte)
	GetLatestCommandBatch(ctx sdk.Context) CommandBatch
	GetBatchByID(ctx sdk.Context, id []byte) CommandBatch
	DeleteUnsignedCommandBatchID(ctx sdk.Context)

	GetConfirmedEventQueue(ctx sdk.Context) utils.KVQueue
	GetEvent(ctx sdk.Context, eventID EventID) (Event, bool)
	SetConfirmedEvent(ctx sdk.Context, event Event) error
	SetEventCompleted(ctx sdk.Context, eventID EventID) error
	SetEventFailed(ctx sdk.Context, eventID EventID) error
}

// ParamsKeeper represents a global paramstore
type ParamsKeeper interface {
	Subspace(s string) params.Subspace
	GetSubspace(s string) (params.Subspace, bool)
}

// Voter exposes voting functionality
type Voter interface {
	InitializePoll(ctx sdk.Context, pollBuilder vote.PollBuilder) (vote.PollID, error)
}

// Nexus provides functionality to manage cross-chain transfers
type Nexus interface {
	LinkAddresses(ctx sdk.Context, sender nexus.CrossChainAddress, recipient nexus.CrossChainAddress) error
	GetRecipient(ctx sdk.Context, sender nexus.CrossChainAddress) (nexus.CrossChainAddress, bool)
	EnqueueTransfer(ctx sdk.Context, senderChain nexus.Chain, recipient nexus.CrossChainAddress, asset sdk.Coin) (nexus.TransferID, error)
	EnqueueForTransfer(ctx sdk.Context, sender nexus.CrossChainAddress, amount sdk.Coin) (nexus.TransferID, error)
	GetTransfersForChainPaginated(ctx sdk.Context, chain nexus.Chain, state nexus.TransferState, pageRequest *query.PageRequest) ([]nexus.CrossChainTransfer, *query.PageResponse, error)
	ArchivePendingTransfer(ctx sdk.Context, transfer nexus.CrossChainTransfer)
	SetChain(ctx sdk.Context, chain nexus.Chain)
	GetChains(ctx sdk.Context) []nexus.Chain
	GetChain(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool)
	IsAssetRegistered(ctx sdk.Context, chain nexus.Chain, denom string) bool
	RegisterAsset(ctx sdk.Context, chain nexus.Chain, asset nexus.Asset) error
	GetChainMaintainers(ctx sdk.Context, chain nexus.Chain) []sdk.ValAddress
	IsChainActivated(ctx sdk.Context, chain nexus.Chain) bool
	GetChainByNativeAsset(ctx sdk.Context, asset string) (chain nexus.Chain, ok bool)
	ComputeTransferFee(ctx sdk.Context, sourceChain nexus.Chain, destinationChain nexus.Chain, asset sdk.Coin) (sdk.Coin, error)
	AddTransferFee(ctx sdk.Context, coin sdk.Coin)
	GetChainMaintainerState(ctx sdk.Context, chain nexus.Chain, address sdk.ValAddress) (nexus.MaintainerState, bool)
	SetChainMaintainerState(ctx sdk.Context, maintainerState nexus.MaintainerState) error
}

// InitPoller is a minimal interface to start a poll. This must be a type alias instead of a type definition,
// because the concrete implementation of Signer (specifically StartSign) is defined in a different package using another (identical)
// InitPoller interface. Go cannot match the types otherwise
type InitPoller = interface {
	InitializePoll(ctx sdk.Context, pollBuilder vote.PollBuilder) (vote.PollID, error)
}

// Snapshotter provides access to the snapshot functionality
type Snapshotter interface {
	CreateSnapshot(ctx sdk.Context, candidates []sdk.ValAddress, filterFunc func(snapshot.ValidatorI) bool, weightFunc func(consensusPower sdk.Uint) sdk.Uint, threshold utils.Threshold) (snapshot.Snapshot, error)
	GetProxy(ctx sdk.Context, principal sdk.ValAddress) (addr sdk.AccAddress, active bool)
}

// Rewarder provides reward functionality
type Rewarder interface {
	GetPool(ctx sdk.Context, name string) reward.RewardPool
}

// StakingKeeper adopts the methods from "github.com/cosmos/cosmos-sdk/x/staking/exported" that are
// actually used by this module
type StakingKeeper interface {
	PowerReduction(ctx sdk.Context) sdk.Int
}

// SlashingKeeper provides functionality to manage slashing info for a validator
type SlashingKeeper interface {
	IsTombstoned(ctx sdk.Context, consAddr sdk.ConsAddress) bool
}

// MultisigKeeper provides functionality to the multisig module
type MultisigKeeper interface {
	GetCurrentKeyID(ctx sdk.Context, chainName nexus.ChainName) (multisig.KeyID, bool)
	GetNextKeyID(ctx sdk.Context, chainName nexus.ChainName) (multisig.KeyID, bool)
	GetKey(ctx sdk.Context, keyID multisig.KeyID) (multisig.Key, bool)
	AssignKey(ctx sdk.Context, chainName nexus.ChainName, keyID multisig.KeyID) error
	RotateKey(ctx sdk.Context, chainName nexus.ChainName) error
	Sign(ctx sdk.Context, keyID multisig.KeyID, payloadHash multisig.Hash, module string, moduleMetadata ...codec.ProtoMarshaler) error
}
