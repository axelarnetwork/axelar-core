package types

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tendermint/tendermint/libs/log"

	params "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

//go:generate moq -out ./mock/expected_keepers.go -pkg mock . TSS Voter Signer Nexus Snapshotter BaseKeeper ChainKeeper

// BaseKeeper is implemented by this module's base keeper
type BaseKeeper interface {
	Logger(ctx sdk.Context) log.Logger

	HasChain(ctx sdk.Context, chain string) bool
	ForChain(chain string) ChainKeeper

	SetPendingChain(ctx sdk.Context, chain nexus.Chain, p Params)
	GetPendingChain(ctx sdk.Context, chain string) (PendingChain, bool)
	DeletePendingChain(ctx sdk.Context, chain string)
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
	GetChainID(ctx sdk.Context) (*big.Int, bool)
	GetRequiredConfirmationHeight(ctx sdk.Context) (uint64, bool)
	GetRevoteLockingPeriod(ctx sdk.Context) (int64, bool)
	GetGatewayByteCode(ctx sdk.Context) ([]byte, bool)
	GetBurnerByteCode(ctx sdk.Context) ([]byte, bool)
	GetTokenByteCode(ctx sdk.Context) ([]byte, bool)
	GetTransactionFeeRate(ctx sdk.Context) (sdk.Dec, bool)
	SetPendingGateway(ctx sdk.Context, address common.Address)
	ConfirmPendingGateway(ctx sdk.Context) error
	DeletePendingGateway(ctx sdk.Context) error
	GetPendingGatewayAddress(ctx sdk.Context) (common.Address, bool)
	GetGatewayAddress(ctx sdk.Context) (common.Address, bool)
	GetDeposit(ctx sdk.Context, txID common.Hash, burnerAddr common.Address) (ERC20Deposit, DepositStatus, bool)
	GetBurnerInfo(ctx sdk.Context, address Address) *BurnerInfo
	SetPendingDeposit(ctx sdk.Context, key vote.PollKey, deposit *ERC20Deposit)
	GetBurnerAddressAndSalt(ctx sdk.Context, token ERC20Token, recipient string, gatewayAddr common.Address) (Address, Hash, error)
	SetBurnerInfo(ctx sdk.Context, burnerInfo BurnerInfo)
	GetPendingDeposit(ctx sdk.Context, key vote.PollKey) (ERC20Deposit, bool)
	DeletePendingDeposit(ctx sdk.Context, key vote.PollKey)
	DeleteDeposit(ctx sdk.Context, deposit ERC20Deposit)
	SetDeposit(ctx sdk.Context, deposit ERC20Deposit, state DepositStatus)
	GetConfirmedDeposits(ctx sdk.Context) []ERC20Deposit
	GetPendingTransferKey(ctx sdk.Context, key vote.PollKey) (TransferKey, bool)
	SetPendingTransferKey(ctx sdk.Context, key vote.PollKey, transferOwnership *TransferKey)
	GetArchivedTransferKey(ctx sdk.Context, key vote.PollKey) (TransferKey, bool)
	ArchiveTransferKey(ctx sdk.Context, key vote.PollKey)
	DeletePendingTransferKey(ctx sdk.Context, key vote.PollKey)
	GetNetworkByID(ctx sdk.Context, id *big.Int) (string, bool)
	GetChainIDByNetwork(ctx sdk.Context, network string) *big.Int
	GetVotingThreshold(ctx sdk.Context) (utils.Threshold, bool)
	GetMinVoterCount(ctx sdk.Context) (int64, bool)

	CreateERC20Token(ctx sdk.Context, asset string, details TokenDetails, address Address) (ERC20Token, error)
	GetERC20TokenByAsset(ctx sdk.Context, asset string) ERC20Token
	GetERC20TokenBySymbol(ctx sdk.Context, symbol string) ERC20Token
	GetTokens(ctx sdk.Context) []ERC20Token

	EnqueueCommand(ctx sdk.Context, cmd Command) error
	GetCommand(ctx sdk.Context, id CommandID) (Command, bool)
	GetPendingCommands(ctx sdk.Context) []Command
	CreateNewBatchToSign(ctx sdk.Context, signer Signer) (CommandBatch, error)
	SetLatestSignedCommandBatchID(ctx sdk.Context, id []byte)
	GetLatestCommandBatch(ctx sdk.Context) CommandBatch
	GetBatchByID(ctx sdk.Context, id []byte) CommandBatch
	DeleteUnsignedCommandBatchID(ctx sdk.Context)
}

// ParamsKeeper represents a global paramstore
type ParamsKeeper interface {
	Subspace(s string) params.Subspace
	GetSubspace(s string) (params.Subspace, bool)
}

// TSS exposes key functionality
type TSS interface {
	GetKeyRequirement(ctx sdk.Context, keyRole tss.KeyRole, keyType tss.KeyType) (tss.KeyRequirement, bool)
	GetRotationCount(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) int64
}

// Voter exposes voting functionality
type Voter interface {
	InitializePoll(ctx sdk.Context, key vote.PollKey, voters []sdk.ValAddress, pollProperties ...vote.PollProperty) error
	// Deprecated: InitializePollWithSnapshot will be removed soon
	InitializePollWithSnapshot(ctx sdk.Context, key vote.PollKey, snapshotSeqNo int64, pollProperties ...vote.PollProperty) error
	GetPoll(ctx sdk.Context, pollKey vote.PollKey) vote.Poll
}

// Nexus provides functionality to manage cross-chain transfers
type Nexus interface {
	LinkAddresses(ctx sdk.Context, sender nexus.CrossChainAddress, recipient nexus.CrossChainAddress) error
	GetRecipient(ctx sdk.Context, sender nexus.CrossChainAddress) (nexus.CrossChainAddress, bool)
	EnqueueForTransfer(ctx sdk.Context, sender nexus.CrossChainAddress, amount sdk.Coin, feeRate sdk.Dec) (nexus.TransferID, error)
	GetTransfersForChain(ctx sdk.Context, chain nexus.Chain, state nexus.TransferState) []nexus.CrossChainTransfer
	ArchivePendingTransfer(ctx sdk.Context, transfer nexus.CrossChainTransfer)
	SetChain(ctx sdk.Context, chain nexus.Chain)
	GetChains(ctx sdk.Context) []nexus.Chain
	GetChain(ctx sdk.Context, chain string) (nexus.Chain, bool)
	IsAssetRegistered(ctx sdk.Context, chain nexus.Chain, denom string) bool
	RegisterAsset(ctx sdk.Context, chain nexus.Chain, asset nexus.Asset) error
	GetChainMaintainers(ctx sdk.Context, chain nexus.Chain) []sdk.ValAddress
	IsChainActivated(ctx sdk.Context, chain nexus.Chain) bool
	GetChainByNativeAsset(ctx sdk.Context, asset string) (chain nexus.Chain, ok bool)
}

// InitPoller is a minimal interface to start a poll. This must be a type alias instead of a type definition,
// because the concrete implementation of Signer (specifically StartSign) is defined in a different package using another (identical)
// InitPoller interface. Go cannot match the types otherwise
type InitPoller = interface {
	// Deprecated: InitializePollWithSnapshot will be removed soon
	InitializePollWithSnapshot(ctx sdk.Context, key vote.PollKey, snapshotSeqNo int64, pollProperties ...vote.PollProperty) error
}

// Signer provides keygen and signing functionality
type Signer interface {
	StartSign(ctx sdk.Context, info tss.SignInfo, snapshotter Snapshotter, voter InitPoller) error
	GetSig(ctx sdk.Context, sigID string) (tss.Signature, tss.SigStatus)
	GetKey(ctx sdk.Context, keyID tss.KeyID) (tss.Key, bool)
	GetKeyRole(ctx sdk.Context, keyID tss.KeyID) tss.KeyRole
	GetCurrentKeyID(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool)
	GetCurrentKey(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool)
	GetNextKeyID(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool)
	GetSnapshotCounterForKeyID(ctx sdk.Context, keyID tss.KeyID) (int64, bool)
	AssignNextKey(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole, keyID tss.KeyID) error
	RotateKey(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) error
	AssertMatchesRequirements(ctx sdk.Context, snapshotter Snapshotter, chain nexus.Chain, keyID tss.KeyID, keyRole tss.KeyRole) error
	GetExternalMultisigThreshold(ctx sdk.Context) utils.Threshold
	GetExternalKeyIDs(ctx sdk.Context, chain nexus.Chain) ([]tss.KeyID, bool)
	GetKeyRequirement(ctx sdk.Context, keyRole tss.KeyRole, keyType tss.KeyType) (tss.KeyRequirement, bool)
	GetKeyType(ctx sdk.Context, keyID tss.KeyID) tss.KeyType
}

// Snapshotter provides access to the snapshot functionality
type Snapshotter = snapshot.Snapshotter
