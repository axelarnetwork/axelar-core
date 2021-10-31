package types

import (
	"crypto/ecdsa"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	evmTypes "github.com/ethereum/go-ethereum/core/types"
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

	GetParams(ctx sdk.Context) []Params
	SetParams(ctx sdk.Context, params ...Params)

	ForChain(chain string) ChainKeeper
	SetPendingChain(ctx sdk.Context, chain nexus.Chain)
	GetPendingChain(ctx sdk.Context, chain string) (nexus.Chain, bool)
	DeletePendingChain(ctx sdk.Context, chain string)
}

// ChainKeeper is implemented by this module's chain keeper
type ChainKeeper interface {
	Logger(ctx sdk.Context) log.Logger

	GetName() string
	GetNetwork(ctx sdk.Context) (string, bool)
	GetRequiredConfirmationHeight(ctx sdk.Context) (uint64, bool)
	GetRevoteLockingPeriod(ctx sdk.Context) (int64, bool)
	GetGatewayByteCodes(ctx sdk.Context) ([]byte, bool)
	GetBurnerByteCodes(ctx sdk.Context) ([]byte, bool)
	GetTokenByteCodes(ctx sdk.Context) ([]byte, bool)
	GetGatewayAddress(ctx sdk.Context) (common.Address, bool)
	GetDeposit(ctx sdk.Context, txID common.Hash, burnerAddr common.Address) (ERC20Deposit, DepositState, bool)
	GetBurnerInfo(ctx sdk.Context, address common.Address) *BurnerInfo
	SetPendingDeposit(ctx sdk.Context, key vote.PollKey, deposit *ERC20Deposit)
	GetBurnerAddressAndSalt(ctx sdk.Context, tokenAddr Address, recipient string, gatewayAddr common.Address) (common.Address, common.Hash, error)
	SetBurnerInfo(ctx sdk.Context, burnerAddr common.Address, burnerInfo *BurnerInfo)
	GetPendingDeposit(ctx sdk.Context, key vote.PollKey) (ERC20Deposit, bool)
	DeletePendingDeposit(ctx sdk.Context, key vote.PollKey)
	DeleteDeposit(ctx sdk.Context, deposit ERC20Deposit)
	SetDeposit(ctx sdk.Context, deposit ERC20Deposit, state DepositState)
	GetConfirmedDeposits(ctx sdk.Context) []ERC20Deposit
	SetGatewayAddress(ctx sdk.Context, addr common.Address)
	GetPendingTransferKey(ctx sdk.Context, key vote.PollKey) (TransferKey, bool)
	SetPendingTransferKey(ctx sdk.Context, key vote.PollKey, transferOwnership *TransferKey)
	GetArchivedTransferKey(ctx sdk.Context, key vote.PollKey) (TransferKey, bool)
	ArchiveTransferKey(ctx sdk.Context, key vote.PollKey)
	DeletePendingTransferKey(ctx sdk.Context, key vote.PollKey)
	GetNetworkByID(ctx sdk.Context, id *big.Int) (string, bool)
	GetChainIDByNetwork(ctx sdk.Context, network string) *big.Int
	GetVotingThreshold(ctx sdk.Context) (utils.Threshold, bool)
	GetMinVoterCount(ctx sdk.Context) (int64, bool)

	GetHashToSign(ctx sdk.Context, rawTx *evmTypes.Transaction) common.Hash
	SetUnsignedTx(ctx sdk.Context, txID string, tx *evmTypes.Transaction, pk ecdsa.PublicKey) error
	AssembleTx(ctx sdk.Context, txID string, sig tss.Signature) (*evmTypes.Transaction, error)

	CreateERC20Token(ctx sdk.Context, asset string, details TokenDetails) (ERC20Token, error)
	GetERC20Token(ctx sdk.Context, asset string) ERC20Token

	EnqueueCommand(ctx sdk.Context, cmd Command) error
	CreateNewBatchToSign(ctx sdk.Context) ([]byte, error)
	GetLatestCommandBatch(ctx sdk.Context) CommandBatch
	GetBatchByID(ctx sdk.Context, id []byte) CommandBatch
}

// ParamsKeeper represents a global paramstore
type ParamsKeeper interface {
	Subspace(s string) params.Subspace
	GetSubspace(s string) (params.Subspace, bool)
}

// TSS exposes key functionality
type TSS interface {
	GetKeyRequirement(ctx sdk.Context, keyRole tss.KeyRole, keyType tss.KeyType) (tss.KeyRequirement, bool)
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
	LinkAddresses(ctx sdk.Context, sender nexus.CrossChainAddress, recipient nexus.CrossChainAddress)
	GetRecipient(ctx sdk.Context, sender nexus.CrossChainAddress) (nexus.CrossChainAddress, bool)
	EnqueueForTransfer(ctx sdk.Context, sender nexus.CrossChainAddress, amount sdk.Coin) error
	GetTransfersForChain(ctx sdk.Context, chain nexus.Chain, state nexus.TransferState) []nexus.CrossChainTransfer
	ArchivePendingTransfer(ctx sdk.Context, transfer nexus.CrossChainTransfer)
	SetChain(ctx sdk.Context, chain nexus.Chain)
	GetChains(ctx sdk.Context) []nexus.Chain
	GetChain(ctx sdk.Context, chain string) (nexus.Chain, bool)
	IsAssetRegistered(ctx sdk.Context, chainName, denom string) bool
	RegisterAsset(ctx sdk.Context, chainName, denom string)
	GetChainMaintainers(ctx sdk.Context, chain nexus.Chain) []sdk.ValAddress
	IsChainActivated(ctx sdk.Context, chain nexus.Chain) bool
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
	GetCurrentKeyID(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool)
	GetCurrentKey(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool)
	GetNextKey(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool)
	GetKeyForSigID(ctx sdk.Context, sigID string) (tss.Key, bool)
	GetSnapshotCounterForKeyID(ctx sdk.Context, keyID tss.KeyID) (int64, bool)
	AssignNextKey(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole, keyID tss.KeyID) error
	RotateKey(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) error
	AssertMatchesRequirements(ctx sdk.Context, snapshotter Snapshotter, chain nexus.Chain, keyID tss.KeyID, keyRole tss.KeyRole) error
	GetExternalMultisigThreshold(ctx sdk.Context) utils.Threshold
	GetExternalKeyIDs(ctx sdk.Context, chain nexus.Chain) ([]tss.KeyID, bool)
}

// Snapshotter provides access to the snapshot functionality
type Snapshotter = snapshot.Snapshotter
